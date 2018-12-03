package api

import (
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type NonceResponse struct {
	Edge      string
	NextNonce uint64
}

type TransportResponse struct {
	store.Transport
	Registered int64
}

func NewTransportResponse(t store.Transport) TransportResponse {
	return TransportResponse{
		Transport:  t,
		Registered: t.Registered.Unix(),
	}
}

type DeletedTransportsResponse struct {
	Deleted []TransportResponse `json:"deleted"`
}
