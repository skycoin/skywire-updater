package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/skycoin/skycoin/src/cipher"
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

	w.WriteHeader(http.StatusCreated)
	return NewTransportResponse(t), nil
}

func (api *API) handleTransports(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	split := strings.Split(r.URL.String(), "/")
	if len(split) < 3 || split[2] == "" {
		return nil, ErrEmptyTransportID
	}
	id, err := strconv.ParseUint(split[2], 10, 64)
	if err != nil {
		return nil, err
	}

	switch r.Method {
	case "GET":
		t, err := api.store.GetTransportByID(r.Context(), store.ID(id))
		if err != nil {
			return nil, err
		}
		return NewTransportResponse(*t), nil
	case "DELETE":
		t, err := api.store.DeregisterTransport(r.Context(), store.ID(id))
		if err != nil {
			return nil, err
		}

		resp := DeletedTransportsResponse{
			Deleted: []*TransportResponse{
				NewTransportResponse(*t),
			},
		}

		return resp, nil
	}

	return nil, errors.New("Invalid HTTP Method")
}

func (api *API) handleIncrementingNonces(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	split := strings.Split(r.URL.String(), "/")
	if len(split) < 3 || split[2] == "" {
		return nil, ErrEmptyPubKey
	}

	pubKey, err := cipher.PubKeyFromHex(split[2])
	if err != nil {
		return nil, err
	}

	nonce, err := api.store.GetNonce(r.Context(), pubKey)
	if err != nil {
		return nil, err
	}

	return NonceResponse{
		Edge:      pubKey.Hex(),
		NextNonce: uint64(nonce),
	}, nil
}
