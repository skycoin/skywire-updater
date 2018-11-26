package api

import (
	"encoding/json"
	"net/http"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

// APIOptions control particular behavior
type APIOptions struct {
	// DisableSigVerify disables signature verification on the request header
	DisableSigVerify bool
}

// API register all the API endpoints.
// It implements a net/http.Handler.
type API struct {
	mux   *http.ServeMux
	store store.Store
	opts  APIOptions
}

type Error struct {
	Error string `json:"error"`
}

func New(s store.Store, opts APIOptions) *API {
	mux := http.NewServeMux()
	api := &API{mux: mux, store: s, opts: opts}

	mux.Handle("/register", apiHandler(api.handleRegister))
	mux.Handle("/deregister", apiHandler(api.handleDeregister))

	return api
}

type apiHandler func(w http.ResponseWriter, r *http.Request) (interface{}, error)

func (fn apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	res, err := fn(w, r)
	if err != nil {
		res = &Error{Error: err.Error()}
		w.WriteHeader(400)
	}

	enc.Encode(res)
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Add some logging
	ctx := r.Context()

	if !api.opts.DisableSigVerify {
		auth, err := authFromHeaders(r.Header)
		if err != nil {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(&Error{Error: err.Error()})
			return
		}

		if err := api.verifyAuth(r.Context(), auth); err != nil {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(&Error{Error: err.Error()})
			return
		}

		if r.Method == "POST" && auth.Key != "" {
			_, err := api.store.IncrementNonce(ctx, auth.Key)
			if err != nil {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(&Error{Error: err.Error()})
				return
			}
		}
	}

	api.mux.ServeHTTP(w, r)
}
