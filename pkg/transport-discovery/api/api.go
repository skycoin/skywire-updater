package api

import (
	"net/http"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type API struct {
	mux   *http.ServeMux
	store store.Store
}

func New(s store.Store) *API {
	mux := http.NewServeMux()
	api := &API{mux: mux}

	mux.HandleFunc("/register", api.handleRegister)
	mux.HandleFunc("/deregister", api.handleDeregister)

	return api
}

func (api *API) handleRegister(w http.ResponseWriter, r *http.Request) {
}

func (api *API) handleDeregister(w http.ResponseWriter, r *http.Request) {
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Add some logging
	r.Header.Set("Content-Type", "application/json")

	// TODO: if Method POST IncrementNonce
	api.mux.ServeHTTP(w, r)
}
