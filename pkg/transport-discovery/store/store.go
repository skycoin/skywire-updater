package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/skycoin/skycoin/src/cipher"
)

var (
	// ErrNotEnoughACKs means that we're still waiting for a node to confirm the transport registration
	ErrNotEnoughACKs = errors.New("Not enough ACKs")
)

// ID represent a Transport ID
type ID uint64

// Nonce is used to sign requests in order to avoid replay attack
type Nonce uint64

func (n Nonce) String() string { return fmt.Sprintf("%d", n) }

// Transport represent a single-hop communication between two Skywire Nodes
type Transport struct {
	// ID is the Transport ID
	ID ID
	// Edges are public keys of each Node
	Edges []cipher.PubKey
}

//go:generate mockgen -package=mockstore -destination=mockstore/mockstore.go github.com/watercompany/skywire-services/pkg/transport-discovery/store Store
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
	GetTransportsByEdge(context.Context, cipher.PubKey) ([]*Transport, error)
}

type NonceStore interface {
	IncrementNonce(context.Context, cipher.PubKey) (Nonce, error)
	GetNonce(context.Context, cipher.PubKey) (Nonce, error)
}
