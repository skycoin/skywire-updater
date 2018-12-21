package transport

import (
	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
)

// Entry is the unsigned representation of a Transport.
type Entry struct {

	// ID is the Transport ID that uniquely identifies the Transport.
	ID uuid.UUID `json:"t_id"`

	// Edges contains the public keys of the Transport's edge nodes (should only have 2 edges and the least-significant edge should come first).
	Edges [2]string `json:"edges"`

	// Type represents the transport type.
	Type string `json:"type"`

	// Public determines whether the transport is to be exposed to other nodes or not.
	// Public transports are to be registered in the Transport Discovery.
	Public bool `json:"public"`
}

// ToBinary returns binary representation of a Signature.
func (e *Entry) ToBinary() []byte {
	bEntry := e.ID[:]
	for _, edge := range e.Edges {
		pk := cipher.MustPubKeyFromHex(edge)
		bEntry = append(bEntry, pk[:]...)
	}
	return append(bEntry, []byte(e.Type)...)
}

// Signature returns signature for Entry calculated from binary
// representation.
func (e *Entry) Signature(secKey cipher.SecKey) string {
	return cipher.SignHash(cipher.SumSHA256(e.ToBinary()), secKey).Hex()
}

// SignedEntry holds an Entry and it's associated signatures.
// The signatures should be ordered as the contained 'Entry.Edges'.
type SignedEntry struct {
	Entry      *Entry    `json:"entry"`
	Signatures [2]string `json:"signatures"`
	Registered int64     `json:"registered,omitempty"`
}
