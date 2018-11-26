package api

import (
	"encoding/json"
	"net/http"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

// API register all the API endpoints.
// It implements a net/http.Handler.
type API struct {
	mux   *http.ServeMux
	store store.Store
}

// Auth struct maps SW-{Key,Nonce,Sig} headers
type Auth struct {
	Key   string
	Nonce store.Nonce
	Sig   string
}

type Error struct {
	Error string `json:"error"`
}

func New(s store.Store) *API {
	mux := http.NewServeMux()
	api := &API{mux: mux, store: s}

	mux.HandleFunc("/register", api.handleRegister)
	mux.HandleFunc("/deregister", api.handleDeregister)

	return api
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Add some logging
	ctx := r.Context()

	r.Header.Set("Content-Type", "application/json")

	auth, err := api.auth(ctx, r.Header)
	if err != nil {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(&Error{Error: err.Error()})
		return
	}

	// TODO: Verify signature
	_ = auth

	if r.Method == "POST" {
		_, err := api.store.IncrementNonce(ctx, auth.Key)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(&Error{Error: err.Error()})
			return
		}
	}

	api.mux.ServeHTTP(w, r)
}
