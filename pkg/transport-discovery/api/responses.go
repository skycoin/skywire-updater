package api

import (
	"time"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type NonceResponse struct {
	Edge      string `json:"edge"`
	NextNonce uint64 `json:"next_nonce"`
}

type TransportResponse struct {
	ID         uint64   `json:"id"`
	Edges      []string `json:"edges"`
	Registered int64    `json:"registered"`
}

func NewTransportResponse(t store.Transport) *TransportResponse {
	var edges []string
	for _, e := range t.Edges {
		edges = append(edges, e.Hex())
	}

	return &TransportResponse{
		ID:         uint64(t.ID),
		Edges:      edges,
		Registered: t.Registered.Unix(),
	}
}

func (t *TransportResponse) Model() *store.Transport {
	var edges []cipher.PubKey

	for _, e := range t.Edges {
		p, err := cipher.PubKeyFromHex(e)
		if err == nil {
			edges = append(edges, p)
		}
	}

	return &store.Transport{
		ID:         store.ID(t.ID),
		Edges:      edges,
		Registered: time.Unix(t.Registered, 0),
	}
}

type DeletedTransportsResponse struct {
	Deleted []*TransportResponse `json:"deleted"`
}
