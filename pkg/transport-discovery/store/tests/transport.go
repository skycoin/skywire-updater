package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type TransportSuite struct {
	suite.Suite
	Store store.TransportStore
}

func (s *TransportSuite) SetupTest() {
}

func (s *TransportSuite) TestRegister() {
	t := s.T()
	ctx := context.Background()

	pk1, _ := cipher.GenerateKeyPair()
	pk2, _ := cipher.GenerateKeyPair()

	var tr1 = &store.Transport{Edges: []cipher.PubKey{pk1, pk2}}
	var tr2 = &store.Transport{Edges: []cipher.PubKey{pk2, pk1}}

	wg := sync.WaitGroup{}

	// This goroutine represent the first node registering a transport
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, s.Store.RegisterTransport(ctx, tr1))
	}()

	// Simulate 1 second delay between both nodes
	time.Sleep(time.Second)

	// This goroutine represent the second node registering a transport
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, s.Store.RegisterTransport(ctx, tr2))
	}()

	wg.Wait()
	assert.Equal(t, tr1.ID, tr2.ID)

	t.Run(".Registered has been set", func(t *testing.T) {
		assert.Equal(t, tr1.Registered, tr2.Registered)
		assert.False(t, tr1.Registered.IsZero(), "Can't be zero")
		assert.True(t, time.Now().After(tr1.Registered))
	})

	t.Run("Transport should be found", func(t *testing.T) {
		found, err := s.Store.GetTransportByID(ctx, tr1.ID)
		require.NoError(t, err)
		assert.Equal(t, tr1.ID, found.ID, "IDs should be equal")
		assert.ElementsMatch(t, tr1.Edges, found.Edges, "Edges should contain the same PKs")
		assert.True(t, len(tr1.Edges) == len(found.Edges))
		assert.Equal(t, tr1.Registered, found.Registered)
	})

	t.Run("Transport should be deleted", func(t *testing.T) {
		trans, err := s.Store.DeregisterTransport(ctx, tr1.ID)
		require.NoError(t, err)
		assert.Equal(t, tr1, trans)

		_, err = s.Store.GetTransportByID(ctx, tr1.ID)
		require.Error(t, err)
		assert.Equal(t, store.ErrNotEnoughACKs, err, "Can't be found after .DeregisterTransport")
	})
}
