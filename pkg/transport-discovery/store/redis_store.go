package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-node/pkg/transport"
)

type redisStore struct {
	client *redis.Client
}

func newRedisStore(url string) (*redisStore, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("url: %s", err)
	}

	return &redisStore{redis.NewClient(opt)}, nil
}

func (s *redisStore) RegisterTransport(_ context.Context, sEntry *transport.SignedEntry) error {
	entry := sEntry.Entry
	entryWithStatus := &EntryWithStatus{
		Entry:      entry,
		IsUp:       true,
		Registered: time.Now().Unix(),
		Statuses:   [2]bool{true, true},
	}
	data, err := json.Marshal(entryWithStatus)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	var res *redis.BoolCmd
	_, err = s.client.TxPipelined(func(pipe redis.Pipeliner) error {
		res = pipe.SetNX(fmt.Sprintf("entries:%s", entry.ID), data, 0)
		pipe.SAdd(fmt.Sprintf("transports:%s", entry.Edges[0]), entry.ID.String())
		pipe.SAdd(fmt.Sprintf("transports:%s", entry.Edges[1]), entry.ID.String())
		return nil
	})
	if err != nil {
		return fmt.Errorf("redis: %s", err)
	}

	if !res.Val() {
		return ErrAlreadyRegistered
	}

	sEntry.Registered = entryWithStatus.Registered
	return nil
}

func (s *redisStore) DeregisterTransport(ctx context.Context, id uuid.UUID) (*transport.Entry, error) {
	entry, err := s.GetTransportByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.client.Del(fmt.Sprintf("entries:%s", id)).Err(); err != nil {
		return nil, fmt.Errorf("redis: %s", err)
	}

	return entry.Entry, nil
}

func (s *redisStore) GetTransportByID(_ context.Context, id uuid.UUID) (*EntryWithStatus, error) {
	data, err := s.client.Get(fmt.Sprintf("entries:%s", id)).Result()
	if err != nil {
		return nil, ErrTransportNotFound
	}

	var entry *EntryWithStatus
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}

	return entry, nil
}

func (s *redisStore) GetTransportsByEdge(_ context.Context, pk cipher.PubKey) ([]*EntryWithStatus, error) {
	trIDs, err := s.client.SMembers(fmt.Sprintf("transports:%s", pk.Hex())).Result()
	if err != nil {
		return nil, ErrTransportNotFound
	}

	keys := []string{}
	for _, trID := range trIDs {
		keys = append(keys, fmt.Sprintf("entries:%s", string(trID)))
	}

	data, err := s.client.MGet(keys...).Result()
	if err != nil {
		return nil, ErrTransportNotFound
	}

	entries := []*EntryWithStatus{}
	for _, e := range data {
		var entry *EntryWithStatus
		if err := json.Unmarshal([]byte(e.(string)), &entry); err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *redisStore) UpdateStatus(ctx context.Context, id uuid.UUID, isUp bool) (*EntryWithStatus, error) {
	pk, ok := ctx.Value(ContextAuthKey).(cipher.PubKey)
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

func (s *redisStore) GetNonce(ctx context.Context, pk cipher.PubKey) (Nonce, error) {
	nonce, err := s.client.Get(fmt.Sprintf("nonces:%s", pk.Hex())).Result()
	if err != nil {
		return 0, nil
	}

	n, _ := strconv.Atoi(nonce) // nolint
	return Nonce(n), nil
}

func (s *redisStore) IncrementNonce(ctx context.Context, pk cipher.PubKey) (Nonce, error) {
	nonce, err := s.client.Incr(fmt.Sprintf("nonces:%s", pk.Hex())).Result()
	if err != nil {
		return 0, fmt.Errorf("redis: %s", err)
	}

	return Nonce(nonce), nil
}
