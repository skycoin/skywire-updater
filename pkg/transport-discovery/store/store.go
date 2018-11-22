package store

import (
	"context"
)

// ID represent a Transport ID
type ID uint64

// Nonce is used to sign requests in order to avoid replay attack
type Nonce uint64

// Transport represent a single-hop communication between two Skywire Nodes
type Transport struct {
	// ID is the Transport ID
	ID ID
	// Edges are public keys of each Node
	Edges []string
}

type Store interface {
	TransportStore
	NonceStore
}

type TransportStore interface {
	// RegisterTransport
	RegisterTransport(context.Context, *Transport) error
	DeregisterTransport(context.Context, ID) error
	GetTransportByID(context.Context, ID) (*Transport, error)

	// TODO: sorting meta arg
	GetTransportsByEdge(context.Context, string) ([]*Transport, error)
}

type NonceStore interface {
	IncrementNonce(context.Context, string) (Nonce, error)
	GetNonce(context.Context, string) (Nonce, error)
}
