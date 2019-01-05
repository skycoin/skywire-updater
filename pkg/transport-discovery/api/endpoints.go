package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire/pkg/transport"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

func (api *API) handleTransports(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "POST":
		return api.registerTransport(w, r)
	case "GET":
		return api.getTransport(w, r)
	}

	return nil, errors.New("Invalid HTTP Method")
}

func (api *API) registerTransport(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	entries := []*transport.SignedEntry{}
	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if err := api.store.RegisterTransport(r.Context(), entry); err != nil {
			return nil, err
		}
	}

	w.WriteHeader(http.StatusCreated)
	return entries, nil
}

func (api *API) getTransport(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	components := strings.Split(strings.Replace(r.URL.Path, "/transports/", "", -1), ":")

	if len(components) != 2 {
		return nil, ErrEmptyTransportID
	}

	switch components[0] {
	case "id":
		id, err := uuid.Parse(components[1])
		if err != nil {
			return nil, ErrInvalidTransportID
		}

		entry, err := api.store.GetTransportByID(r.Context(), id)
		if err != nil {
			return nil, err
		}

		return entry, nil
	case "edge":
		pk, err := cipher.PubKeyFromHex(components[1])
		if err != nil {
			return nil, ErrInvalidPubKey
		}

		entries, err := api.store.GetTransportsByEdge(r.Context(), pk)
		if err != nil {
			return nil, err
		}

		return entries, nil
	default:
		return nil, ErrEmptyTransportID
	}
}

func (api *API) handleStatuses(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if r.Method != "POST" {
		return nil, errors.New("Invalid HTTP Method")
	}

	statuses := []*transport.Status{}
	if err := json.NewDecoder(r.Body).Decode(&statuses); err != nil {
		return nil, err
	}

	res := []*store.EntryWithStatus{}
	for _, status := range statuses {
		entry, err := api.store.UpdateStatus(r.Context(), status.ID, status.IsUp)
		if err != nil {
			return nil, err
		}

		res = append(res, entry)
	}

	return res, nil
}

func (api *API) handleIncrementingNonces(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if r.Method != "GET" {
		return nil, errors.New("Invalid HTTP Method")
	}

	key := strings.Replace(r.URL.Path, "/security/nonces/", "", -1)
	if key == "" {
		return nil, ErrEmptyPubKey
	}

	pubKey, err := cipher.PubKeyFromHex(key)
	if err != nil {
		return nil, ErrInvalidPubKey
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
