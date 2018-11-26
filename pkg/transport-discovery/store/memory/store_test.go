package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/tests"
)

func TestMemory(t *testing.T) {
	suite.Run(t, &tests.NonceSuite{Store: NewStore()})
}

func TestMemoryConcurrency(t *testing.T) {
	wg := sync.WaitGroup{}
	n := 100
	mem := NewStore()

	ctx := context.Background()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := mem.IncrementNonce(ctx, "concurrent-test")
			require.NoError(t, err)
		}()
	}
	wg.Wait()

	nonce, err := mem.GetNonce(ctx, "concurrent-test")
	require.NoError(t, err)
	assert.Equal(t, store.Nonce(n), nonce)
}
