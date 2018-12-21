package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-node/pkg/transport"
)

var (
	// ErrNotEnoughACKs means that we're still waiting for a node to confirm the transport registration
	ErrNotEnoughACKs = errors.New("Not enough ACKs")

	// ErrAlreadyRegistered indicates that transport ID is already in use.
	ErrAlreadyRegistered = errors.New("ID already registered")

	// ErrTransportNotFound indicates that requested transport is not registered.
	ErrTransportNotFound = errors.New("Transport not found")

	// ContextAuthKey stores authenticated PubKey in Context .
	ContextAuthKey = struct{}{}
)

// Nonce is used to sign requests in order to avoid replay attack
type Nonce uint64

func (n Nonce) String() string { return fmt.Sprintf("%d", n) }

// Store stores Transport metadata and generated nonce values.
type Store interface {
	TransportStore
	NonceStore
}

// EntryWithStatus represents accumulated value stored in a Store.
type EntryWithStatus struct {
	Entry      *transport.Entry `json:"entry"`
	IsUp       bool             `json:"is_up"`
	Registered int64            `json:"registered"`
	Statuses   [2]bool          `json:"-"`
}

// TransportStore stores Transport metadata.
type TransportStore interface {
	RegisterTransport(context.Context, *transport.SignedEntry) error
	DeregisterTransport(context.Context, uuid.UUID) (*transport.Entry, error)
	GetTransportByID(context.Context, uuid.UUID) (*EntryWithStatus, error)
	GetTransportsByEdge(context.Context, cipher.PubKey) ([]*EntryWithStatus, error)
	UpdateStatus(context.Context, uuid.UUID, bool) (*EntryWithStatus, error)
}

// NonceStore stores generated nonce values.
type NonceStore interface {
	IncrementNonce(context.Context, cipher.PubKey) (Nonce, error)
	GetNonce(context.Context, cipher.PubKey) (Nonce, error)
}

// New constructs a new Store of requested type.
func New(sType string, args ...string) (Store, error) {
	switch sType {
	case "memory":
		return newMemoryStore(), nil
	case "redis":
		if len(args) != 1 {
			return nil, errors.New("invalid args")
		}

		return newRedisStore(args[0])
	default:
		return nil, errors.New("unknown store type")
	}
}
