package api

import (
	"encoding/json"
	"net/http"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

func (api *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	t := store.Transport{}
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		json.NewEncoder(w).Encode(Error{Error: err.Error()})
		return
	}

	if err := api.store.RegisterTransport(r.Context(), &t); err != nil {
		json.NewEncoder(w).Encode(Error{Error: err.Error()})
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(t)
}

func (api *API) handleDeregister(w http.ResponseWriter, r *http.Request) {
}
