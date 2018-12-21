// Package auth provides a set of convenience functions that wrap cipher signature functions.
package auth

import (
	"fmt"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

// Hash returns the SHA256 of the concatenation of payload and nonce.
func Hash(payload []byte, nonce store.Nonce) cipher.SHA256 {
	return cipher.SumSHA256([]byte(
		fmt.Sprintf("%s%d", string(payload), nonce),
	))
}

// Sign signs the Hash of payload and nonce
func Sign(payload []byte, nonce store.Nonce, sec cipher.SecKey) cipher.Sig {
	return cipher.SignHash(Hash(payload, nonce), sec)
}

// Verify verifies the signature of the hash of payload and nonce
func Verify(payload []byte, nonce store.Nonce, pub cipher.PubKey, sig cipher.Sig) error {
	return cipher.VerifySignature(pub, sig, Hash(payload, nonce))
}
