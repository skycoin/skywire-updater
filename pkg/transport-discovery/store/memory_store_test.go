package store

import (
	"context"
	"sync"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestMemory(t *testing.T) {
	s, _ := New("memory")
	suite.Run(t, &TransportSuite{TransportStore: s})
	suite.Run(t, &NonceSuite{NonceStore: s})
}

func TestMemoryConcurrency(t *testing.T) {
	wg := sync.WaitGroup{}
	n := 100
	s, _ := New("memory")
	pub, _ := cipher.GenerateKeyPair()

	ctx := context.Background()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.IncrementNonce(ctx, pub)
			require.NoError(t, err)
		}()
	}
	wg.Wait()

	nonce, err := s.GetNonce(ctx, pub)
	require.NoError(t, err)
	assert.Equal(t, Nonce(n), nonce)
}
