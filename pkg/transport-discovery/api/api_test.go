package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-node/pkg/transport"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type errorSetter interface {
	SetError(error)
}

func newTestEntry() *transport.Entry {
	pk1, _ := cipher.GenerateKeyPair()
	return &transport.Entry{
		ID:     uuid.New(),
		Edges:  [2]string{pk1.Hex(), testPubKey.Hex()},
		Type:   "messaging",
		Public: true,
	}
}

func TestBadRequest(t *testing.T) {
	mock, _ := store.New("memory") // nolint

	api := New(mock, Options{DisableSigVerify: true})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/transports/", bytes.NewBufferString("not-a-json"))

	api.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterTransport(t *testing.T) {
	mock, _ := store.New("memory") // nolint
	sEntry := &transport.SignedEntry{Entry: newTestEntry(), Signatures: [2]string{"foo", "bar"}}

	api := New(mock, Options{DisableSigVerify: true})
	w := httptest.NewRecorder()

	body := bytes.NewBuffer(nil)
	require.NoError(t, json.NewEncoder(body).Encode([]*transport.SignedEntry{sEntry}))
	r := httptest.NewRequest("POST", "/transports/", body)
	api.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var resp []*transport.SignedEntry
	require.NoError(t, json.NewDecoder(bytes.NewBuffer(w.Body.Bytes())).Decode(&resp))

	require.Len(t, resp, 1)
	assert.Equal(t, sEntry.Entry, resp[0].Entry)
	assert.True(t, resp[0].Registered > 0)
}

func TestRegisterTimeout(t *testing.T) {
	timeout := 10 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mock, _ := store.New("memory") // nolint
	sEntry := &transport.SignedEntry{Entry: newTestEntry(), Signatures: [2]string{"foo", "bar"}}
	api := New(mock, Options{DisableSigVerify: true})

	// after this ctx's deadline will be exceeded
	time.Sleep(timeout * 2)

	mock.(errorSetter).SetError(ctx.Err())

	w := httptest.NewRecorder()
	body := bytes.NewBuffer(nil)
	require.NoError(t, json.NewEncoder(body).Encode([]*transport.SignedEntry{sEntry}))
	r := httptest.NewRequest("POST", "/transports/", body)

	api.ServeHTTP(w, r.WithContext(ctx))

	require.Equal(t, http.StatusRequestTimeout, w.Code, w.Body.String())
}

func TestGETTransportByID(t *testing.T) {
	mock, _ := store.New("memory") // nolint

	api := New(mock, Options{DisableSigVerify: true})

	ctx := context.Background()

	entry := newTestEntry()
	sEntry := &transport.SignedEntry{Entry: entry, Signatures: [2]string{"foo", "bar"}}
	require.NoError(t, mock.RegisterTransport(ctx, sEntry))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/transports/id:%s", entry.ID), nil)
	api.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp *store.EntryWithStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, entry, resp.Entry)
	assert.True(t, resp.IsUp)

	t.Run("Persistence", func(t *testing.T) {
		found, err := mock.GetTransportByID(ctx, entry.ID)
		require.NoError(t, err)
		assert.Equal(t, found.Entry, entry)
	})
}

func TestUpdateStatus(t *testing.T) {
	mock, _ := store.New("memory") // nolint
	sEntry := &transport.SignedEntry{Entry: newTestEntry(), Signatures: [2]string{"foo", "bar"}}
	require.NoError(t, mock.RegisterTransport(context.Background(), sEntry))

	api := New(mock, Options{})
	w := httptest.NewRecorder()

	body := bytes.NewBuffer(nil)
	require.NoError(t, json.NewEncoder(body).Encode([]*transport.Status{{ID: sEntry.Entry.ID, IsUp: false}}))
	r := httptest.NewRequest("POST", "/statuses", body)
	r.Header = validHeaders(body.Bytes())
	api.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp []*store.EntryWithStatus
	require.NoError(t, json.NewDecoder(bytes.NewBuffer(w.Body.Bytes())).Decode(&resp))

	require.Len(t, resp, 1)
	assert.Equal(t, sEntry.Entry, resp[0].Entry)
	assert.False(t, resp[0].IsUp)
}

func TestGETTransportByEdge(t *testing.T) {
	mock, _ := store.New("memory") // nolint

	api := New(mock, Options{DisableSigVerify: true})

	ctx := context.Background()

	entry := newTestEntry()
	sEntry := &transport.SignedEntry{Entry: entry, Signatures: [2]string{"foo", "bar"}}
	require.NoError(t, mock.RegisterTransport(ctx, sEntry))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/transports/edge:%s", entry.Edges[0]), nil)
	api.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp []*store.EntryWithStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	require.Len(t, resp, 1)
	assert.Equal(t, entry, resp[0].Entry)
	assert.True(t, resp[0].IsUp)

	t.Run("Persistence", func(t *testing.T) {
		found, err := mock.GetTransportByID(ctx, entry.ID)
		require.NoError(t, err)
		assert.Equal(t, found.Entry, entry)
	})
}

func TestGETIncrementingNonces(t *testing.T) {
	mock, _ := store.New("memory") // nolint

	pubKey, _ := cipher.GenerateKeyPair()

	api := New(mock, Options{})

	t.Run("ValidRequest", func(t *testing.T) {
		ctx := context.Background()
		for range [0xff]bool{} {
			_, err := mock.IncrementNonce(context.Background(), pubKey)
			require.NoError(t, err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/security/nonces/"+pubKey.Hex(), nil)
		api.ServeHTTP(w, r.WithContext(ctx))
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		var resp NonceResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, pubKey.Hex(), resp.Edge)
		assert.Equal(t, uint64(0xff), resp.NextNonce)
	})

	t.Run("StoreError", func(t *testing.T) {
		boom := errors.New("boom")
		mock.(errorSetter).SetError(boom)
		defer mock.(errorSetter).SetError(nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/security/nonces/"+pubKey.Hex(), nil)
		api.ServeHTTP(w, r)
		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), boom.Error())
	})

	t.Run("EmptyKey", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/security/nonces/", nil)
		api.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "empty")
	})

	t.Run("InvalidKey", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/security/nonces/foo-bar", nil)
		api.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "PublicKey is invalid")
	})
}
