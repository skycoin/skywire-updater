package auth

import (
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

func TestSignatureVerification(t *testing.T) {
	pub, sec := cipher.GenerateKeyPair()
	payload := []byte("payload to sign")
	nonce := store.Nonce(0xff)

	sig, err := Sign(payload, nonce, sec)
	require.NoError(t, err)
	require.NoError(t, Verify(payload, nonce, pub, sig))
	require.Error(t, Verify(payload, nonce+1, pub, sig))
}
