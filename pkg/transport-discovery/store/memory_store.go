package store

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-node/pkg/transport"
)

type mStore struct {
	transports []*EntryWithStatus
	nonces     map[cipher.PubKey]Nonce

	err error
	mu  sync.Mutex
}

func NewMemoryStore() *mStore {
	return &mStore{
		transports: []*EntryWithStatus{},
		nonces:     make(map[cipher.PubKey]Nonce),
	}
}

func (s *mStore) SetError(err error) {
	s.err = err
}

func (s *mStore) RegisterTransport(_ context.Context, entry *transport.SignedEntry) error {
	if s.err != nil {
		return s.err
	}

	s.mu.Lock()
	for _, e := range s.transports {
		if e.Entry.ID == entry.Entry.ID {
			return errors.New("ID already registered")
		}
	}

	s.transports = append(s.transports, &EntryWithStatus{
		Entry:      entry.Entry,
		IsUp:       true,
		Registered: time.Now().Unix(),
		Statuses:   [2]bool{true, true},
	})
	s.mu.Unlock()

	entry.Registered = time.Now().Unix()
	return nil
}

func (s *mStore) DeregisterTransport(_ context.Context, id uuid.UUID) (*transport.Entry, error) {
	if s.err != nil {
		return nil, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for idx, entry := range s.transports {
		if entry == nil {
			continue
		}

		if entry.Entry.ID == id {
			s.transports[idx] = nil
			return entry.Entry, nil
		}
	}

	return nil, errors.New("Transport not found")
}

func (s *mStore) GetTransportByID(_ context.Context, id uuid.UUID) (*EntryWithStatus, error) {
	if s.err != nil {
		return nil, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range s.transports {
		if entry == nil {
			continue
		}

		if entry.Entry.ID == id {
			return entry, nil
		}
	}

	return nil, errors.New("Transport not found")
}

func (s *mStore) GetTransportsByEdge(_ context.Context, pk cipher.PubKey) ([]*EntryWithStatus, error) {
	if s.err != nil {
		return nil, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	res := []*EntryWithStatus{}
	for _, entry := range s.transports {
		if entry == nil {
			continue
		}

		if entry.Entry.Edges[0] == pk.Hex() || entry.Entry.Edges[1] == pk.Hex() {
			res = append(res, entry)
		}
	}

	return res, nil
}

func (s *mStore) UpdateStatus(ctx context.Context, id uuid.UUID, isUp bool) (*EntryWithStatus, error) {
	if s.err != nil {
		return nil, s.err
	}

	pk, ok := ctx.Value("auth-pub-key").(cipher.PubKey)
	if !ok {
		return nil, errors.New("invalid auth")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.transports {
		if entry == nil {
			continue
		}

		if entry.Entry.ID == id {
			idx := 0
			if entry.Entry.Edges[1] == pk.Hex() {
				idx = 1
			}

			entry.Statuses[idx] = isUp
			entry.IsUp = entry.Statuses[0] && entry.Statuses[1]
			return entry, nil
		}
	}

	return nil, errors.New("Transport not found")
}

func (s *mStore) GetNonce(ctx context.Context, pk cipher.PubKey) (Nonce, error) {
	if s.err != nil {
		return 0, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nonces[pk], nil
}

func (s *mStore) IncrementNonce(ctx context.Context, pk cipher.PubKey) (Nonce, error) {
	if s.err != nil {
		return 0, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nonces[pk] = s.nonces[pk] + 1
	return s.nonces[pk], nil
}
