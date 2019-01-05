package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/watercompany/skywire/pkg/transport"
)

type TransportSuite struct {
	suite.Suite
	TransportStore
}

func (s *TransportSuite) SetupTest() {
}

func (s *TransportSuite) TestRegister() {
	t := s.T()
	ctx := context.Background()

	pk1, _ := cipher.GenerateKeyPair()
	pk2, _ := cipher.GenerateKeyPair()

	sEntry := &transport.SignedEntry{
		Entry: &transport.Entry{
			ID:     uuid.New(),
			Edges:  [2]string{pk1.Hex(), pk2.Hex()},
			Type:   "messaging",
			Public: true,
		},
		Signatures: [2]string{"foo", "bar"},
	}

	t.Run(".RegisterTransport", func(t *testing.T) {
		require.NoError(t, s.RegisterTransport(ctx, sEntry))
		assert.True(t, sEntry.Registered > 0)

		assert.Equal(t, ErrAlreadyRegistered, s.RegisterTransport(ctx, sEntry))
	})

	t.Run(".GetTransportByID", func(t *testing.T) {
		found, err := s.GetTransportByID(ctx, sEntry.Entry.ID)
		require.NoError(t, err)
		assert.Equal(t, sEntry.Entry, found.Entry)
		assert.True(t, found.IsUp)
	})

	t.Run(".GetTransportsByEdge", func(t *testing.T) {
		entries, err := s.GetTransportsByEdge(ctx, pk1)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, sEntry.Entry, entries[0].Entry)
		assert.True(t, entries[0].IsUp)

		entries, err = s.GetTransportsByEdge(ctx, pk2)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, sEntry.Entry, entries[0].Entry)
		assert.True(t, entries[0].IsUp)

		pk, _ := cipher.GenerateKeyPair()
		_, err = s.GetTransportsByEdge(ctx, pk)
		require.Error(t, err)
	})

	t.Run(".UpdateStatus", func(t *testing.T) {
		_, err := s.UpdateStatus(ctx, sEntry.Entry.ID, false)
		require.Error(t, err)
		assert.Equal(t, "invalid auth", err.Error())

		entry, err := s.UpdateStatus(context.WithValue(ctx, ContextAuthKey, pk1), sEntry.Entry.ID, false)
		require.NoError(t, err)
		assert.Equal(t, sEntry.Entry, entry.Entry)
		assert.False(t, entry.IsUp)
	})

	t.Run(".DeregisterTransport", func(t *testing.T) {
		entry, err := s.DeregisterTransport(ctx, sEntry.Entry.ID)
		require.NoError(t, err)
		assert.Equal(t, sEntry.Entry, entry)

		_, err = s.GetTransportByID(ctx, sEntry.Entry.ID)
		require.Error(t, err)
		assert.Equal(t, "Transport not found", err.Error())
	})
}
