// +build !no_ci

package store

import (
	"context"
	"sync"
	"testing"

	"github.com/go-redis/redis"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestRedis(t *testing.T) {
	url := "redis://localhost:6379"
	opt, err := redis.ParseURL(url)
	require.NoError(t, err)

	client := redis.NewClient(opt)
	require.NoError(t, client.FlushDB().Err())

	s, err := New("redis", url)
	require.NoError(t, err)
	suite.Run(t, &TransportSuite{TransportStore: s})
	suite.Run(t, &NonceSuite{NonceStore: s})
}

func TestRedisConcurrency(t *testing.T) {
	wg := sync.WaitGroup{}
	n := 100
	s, err := New("redis", "redis://localhost:6379")
	require.NoError(t, err)
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
