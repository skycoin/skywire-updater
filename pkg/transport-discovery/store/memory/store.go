package memory

import (
	"context"
	"sync"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type Store struct {
	sync.RWMutex
	db map[cipher.PubKey]store.Nonce
}

func NewStore() *Store {
	return &Store{
		db: make(map[cipher.PubKey]store.Nonce),
	}
}

func (s *Store) IncrementNonce(_ context.Context, key cipher.PubKey) (store.Nonce, error) {
	s.Lock()
	defer s.Unlock()

	nonce, _ := s.db[key]
	nonce++

	s.db[key] = nonce
	return nonce, nil
}

func (s *Store) GetNonce(_ context.Context, key cipher.PubKey) (store.Nonce, error) {
	s.RLock()
	defer s.RUnlock()
	return s.db[key], nil
}
