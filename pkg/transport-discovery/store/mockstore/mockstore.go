package mockstore

import (
	"context"
	"sync"
	"time"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

func newID() store.ID { return store.ID(time.Now().UnixNano()) }

type Store struct {
	idsMutex sync.Mutex
	idsIndex map[store.ID]*store.Transport

	edgesMutex sync.Mutex
	edgesIndex map[string][]*store.Transport

	noncesMutex sync.Mutex
	noncesIndex map[string]store.Nonce

	err error
}

var _ store.Store = &Store{}

func NewStore() *Store {
	return &Store{
		idsIndex:    make(map[store.ID]*store.Transport),
		edgesIndex:  make(map[string][]*store.Transport),
		noncesIndex: make(map[string]store.Nonce),
	}
}

func (s *Store) SetError(err error) { s.err = err }

func (s *Store) RegisterTransport(_ context.Context, t *store.Transport) error {
	if s.err != nil {
		return s.err
	}

	if t.ID == 0 {
		t.ID = newID()
	}

	s.idsMutex.Lock()
	s.idsIndex[t.ID] = t
	s.idsMutex.Unlock()

	s.edgesMutex.Lock()
	for _, edge := range t.Edges {
		key := edge.Hex()
		s.edgesIndex[key] = append(s.edgesIndex[key], t)
	}
	s.edgesMutex.Unlock()

	return nil
}

func (s *Store) DeregisterTransport(ctx context.Context, ID store.ID) (*store.Transport, error) {
	s.idsMutex.Lock()
	defer s.idsMutex.Unlock()
	t, err := s.getTransportByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	delete(s.idsIndex, ID)

	return t, nil

}

func (s *Store) GetTransportByID(ctx context.Context, ID store.ID) (*store.Transport, error) {
	s.idsMutex.Lock()
	defer s.idsMutex.Unlock()
	return s.getTransportByID(ctx, ID)

}

func (s *Store) getTransportByID(_ context.Context, ID store.ID) (*store.Transport, error) {
	t, ok := s.idsIndex[ID]
	if !ok {
		return nil, store.ErrNotEnoughACKs
	}

	return t, nil
}

func (s *Store) GetTransportsByEdge(_ context.Context, pub cipher.PubKey) ([]*store.Transport, error) {
	return nil, nil
}

func (s *Store) GetNonce(ctx context.Context, pub cipher.PubKey) (store.Nonce, error) {
	if s.err != nil {
		return 0, s.err
	}

	s.noncesMutex.Lock()
	defer s.noncesMutex.Unlock()
	return s.getNonce(ctx, pub)
}

func (s *Store) getNonce(_ context.Context, pub cipher.PubKey) (store.Nonce, error) {
	return s.noncesIndex[pub.Hex()], nil
}

func (s *Store) setNonce(_ context.Context, pub cipher.PubKey, nonce store.Nonce) error {
	s.noncesIndex[pub.Hex()] = nonce
	return nil
}

func (s *Store) IncrementNonce(ctx context.Context, pub cipher.PubKey) (store.Nonce, error) {
	s.noncesMutex.Lock()
	defer s.noncesMutex.Unlock()

	nonce, _ := s.getNonce(ctx, pub)
	nonce++
	s.setNonce(ctx, pub, nonce)
	return nonce, nil
}
