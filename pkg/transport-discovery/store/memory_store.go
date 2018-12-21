package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-node/pkg/transport"
)

type memStore struct {
	transports []*EntryWithStatus
	nonces     map[cipher.PubKey]Nonce

	err error
	mu  sync.Mutex
}

func newMemoryStore() *memStore {
	return &memStore{
		transports: []*EntryWithStatus{},
		nonces:     make(map[cipher.PubKey]Nonce),
	}
}

func (s *memStore) SetError(err error) {
	s.err = err
}

func (s *memStore) RegisterTransport(_ context.Context, entry *transport.SignedEntry) error {
	if s.err != nil {
		return s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.transports {
		if e.Entry.ID == entry.Entry.ID {
			return ErrAlreadyRegistered
		}
	}

	entryWithStatus := &EntryWithStatus{
		Entry:      entry.Entry,
		IsUp:       true,
		Registered: time.Now().Unix(),
		Statuses:   [2]bool{true, true},
	}
	s.transports = append(s.transports, entryWithStatus)

	entry.Registered = entryWithStatus.Registered
	return nil
}

func (s *memStore) DeregisterTransport(ctx context.Context, id uuid.UUID) (*transport.Entry, error) {
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

	return nil, ErrTransportNotFound
}

func (s *memStore) GetTransportByID(_ context.Context, id uuid.UUID) (*EntryWithStatus, error) {
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

	return nil, ErrTransportNotFound
}

func (s *memStore) GetTransportsByEdge(_ context.Context, pk cipher.PubKey) ([]*EntryWithStatus, error) {
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

	if len(res) == 0 {
		return nil, ErrTransportNotFound
	}

	return res, nil
}

func (s *memStore) UpdateStatus(ctx context.Context, id uuid.UUID, isUp bool) (*EntryWithStatus, error) {
	if s.err != nil {
		return nil, s.err
	}

	pk, ok := ctx.Value("auth-pub-key").(cipher.PubKey)
	if !ok {
		return nil, errors.New("invalid auth")
	}

	entry, err := s.GetTransportByID(ctx, id)
	if err != nil {
		return nil, err
	}

	idx := -1
	if entry.Entry.Edges[0] == pk.Hex() {
		idx = 0
	} else if entry.Entry.Edges[1] == pk.Hex() {
		idx = 1
	}

	if idx == -1 {
		return nil, fmt.Errorf("unauthorized")
	}

	entry.Statuses[idx] = isUp
	entry.IsUp = entry.Statuses[0] && entry.Statuses[1]
	return entry, nil
}

func (s *memStore) GetNonce(ctx context.Context, pk cipher.PubKey) (Nonce, error) {
	if s.err != nil {
		return 0, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nonces[pk], nil
}

func (s *memStore) IncrementNonce(ctx context.Context, pk cipher.PubKey) (Nonce, error) {
	if s.err != nil {
		return 0, s.err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nonces[pk] = s.nonces[pk] + 1
	return s.nonces[pk], nil
}
