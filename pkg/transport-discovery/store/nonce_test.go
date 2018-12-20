package store

import (
	"context"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type NonceSuite struct {
	suite.Suite
	NonceStore
}

func (s *NonceSuite) SetupTest() {
	// Setup goes here if required
}

func (s *NonceSuite) TestNonce() {
	t := s.T()
	ctx := context.Background()

	t.Run("GetUnexistingNonce", func(t *testing.T) {
		pub, _ := cipher.GenerateKeyPair()
		nonce, err := s.GetNonce(ctx, pub)
		require.NoError(t, err)
		assert.Equal(t, Nonce(0), nonce)
	})

	t.Run("IncrementNonce", func(t *testing.T) {
		var (
			nonce Nonce
			err   error
		)

		pub, _ := cipher.GenerateKeyPair()

		nonce, err = s.IncrementNonce(ctx, pub)
		require.NoError(t, err)
		assert.Equal(t, Nonce(1), nonce)

		nonce, err = s.IncrementNonce(ctx, pub)
		require.NoError(t, err)
		assert.Equal(t, Nonce(2), nonce)

		nonce, err = s.GetNonce(ctx, pub)
		require.NoError(t, err)
		assert.Equal(t, Nonce(2), nonce)
	})
}
