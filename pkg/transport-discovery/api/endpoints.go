package api

import (
	"encoding/json"
	"net/http"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

func (api *API) handleRegister(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	t := store.Transport{}
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return nil, err
	}

	if err := api.store.RegisterTransport(r.Context(), &t); err != nil {
		return nil, err
	}

	w.WriteHeader(201)
	return t, nil
}

func (api *API) handleDeregister(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return nil, nil
}
