package transport

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
)

// Transport represents communication between two nodes via a single hop.
type Transport interface {

	// Read implements io.Reader
	Read(p []byte) (n int, err error)

	// Write implements io.Writer
	Write(p []byte) (n int, err error)

	// Close implements io.Closer
	Close() error

	// Local returns the local transport edge's public key.
	Local() cipher.PubKey

	// Remote returns the remote transport edge's public key.
	Remote() cipher.PubKey

	// Type returns the string representation of the transport type.
	Type() string

	// SetDeadline functions the same as that from net.Conn
	// With a Transport, we don't have a distinction between write and read timeouts.
	SetDeadline(t time.Time) error
}

// Factory generates Transports of a certain type.
type Factory interface {

	// Accept accepts a remotely-initiated Transport.
	Accept(ctx context.Context) (Transport, error)

	// Dial initiates a Transport with a remote node.
	Dial(ctx context.Context, remote cipher.PubKey) (Transport, error)

	// Close implements io.Closer
	Close() error

	// Local returns the local public key.
	Local() cipher.PubKey

	// Type returns the Transport type.
	Type() string
}

// Status represents the current state of a Transport from the perspective
// from a Transport's single edge. Each Transport will have two perspectives;
// one from each of it's edges.
type Status struct {

	// ID is the Transport ID that identifies the Transport that this status is regarding.
	ID uuid.UUID `json:"t_id"`

	// IsUp represents whether the Transport is up.
	// A Transport that is down will fail to forward Packets.
	IsUp bool `json:"is_up"`

	// Updated is the epoch timestamp of when the status is last updated.
	Updated int64 `json:"updated,omitempty"`
}
