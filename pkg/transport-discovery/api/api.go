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

func New(s store.Store, opts APIOptions) *API {
	mux := http.NewServeMux()
	api := &API{mux: mux, store: s, opts: opts}

	mux.Handle("/register", apiHandler(api.handleRegister))
	mux.Handle("/deregister", apiHandler(api.handleDeregister))

	return api
}

// Error is the object returned to the client when there's an error.
type Error struct {
	Error string `json:"error"`
}

func renderError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&Error{
		Error: err.Error(),
	})
}

// apiHandler is an adapter to reduce api handler endpoint boilerplate
type apiHandler func(w http.ResponseWriter, r *http.Request) (interface{}, error)

// ServeHTTP implements http.Handler.
func (fn apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")

	res, err := fn(w, r)
	if err != nil {
		renderError(w, 400, err)
	} else {
		json.NewEncoder(w).Encode(res)
	}
}

// ServeHTTP implements http.Handler.
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Add some logging
	ctx := r.Context()

	if !api.opts.DisableSigVerify {
		auth, err := authFromHeaders(r.Header)
		if err != nil {
			renderError(w, 401, err)
			return
		}

		if err := api.verifyAuth(r.Context(), auth); err != nil {
			renderError(w, 401, err)
			return
		}

		if r.Method == "POST" && !auth.Key.Null() {
			_, err := api.store.IncrementNonce(ctx, auth.Key)
			if err != nil {
				renderError(w, 500, err)
				return
			}
		}
	}

	api.mux.ServeHTTP(w, r)
}
