package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type NonceSuite struct {
	suite.Suite
	Store store.NonceStore
}

func (s *NonceSuite) SetupTest() {
	// Setup goes here if required
}

func (s *NonceSuite) TestNonce() {
	t := s.T()
	ctx := context.Background()

	t.Run("GetUnexistingNonce", func(t *testing.T) {
		nonce, err := s.Store.GetNonce(ctx, "not-a-registered-key")
		require.NoError(t, err)
		assert.Equal(t, store.Nonce(0), nonce)
	})

	t.Run("IncrementNonce", func(t *testing.T) {
		var (
			nonce store.Nonce
			err   error
		)

		nonce, err = s.Store.IncrementNonce(ctx, "test-key")
		require.NoError(t, err)
		assert.Equal(t, store.Nonce(1), nonce)

		nonce, err = s.Store.IncrementNonce(ctx, "test-key")
		require.NoError(t, err)
		assert.Equal(t, store.Nonce(2), nonce)

		nonce, err = s.Store.GetNonce(ctx, "test-key")
		require.NoError(t, err)
		assert.Equal(t, store.Nonce(2), nonce)
	})
}
