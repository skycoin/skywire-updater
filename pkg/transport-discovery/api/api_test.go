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

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/mockstore"
)

func newTestTransport() *store.Transport {
	pk1, _ := cipher.GenerateKeyPair()
	pk2, _ := cipher.GenerateKeyPair()
	return &store.Transport{
		ID:         0xff,
		Edges:      []cipher.PubKey{pk1, pk2},
		Registered: time.Now().Add(1 * time.Minute),
	}
}

func TestBadRequest(t *testing.T) {
	mock := mockstore.NewStore()

	api := New(mock, APIOptions{DisableSigVerify: true})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/entries", bytes.NewBufferString("not-a-json"))

	api.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterTransport(t *testing.T) {
	mock := mockstore.NewStore()
	trans := newTestTransport()

	api := New(mock, APIOptions{DisableSigVerify: true})
	w := httptest.NewRecorder()

	post := bytes.NewBuffer(nil)
	json.NewEncoder(post).Encode(trans)
	r := httptest.NewRequest("POST", "/entries", post)
	api.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var resp TransportResponse
	json.NewDecoder(bytes.NewBuffer(w.Body.Bytes())).Decode(&resp)

	m := resp.Model()
	assert.Equal(t, trans.ID, m.ID)
	assert.Equal(t, trans.Edges, m.Edges)
	assert.Equal(t, trans.Registered.Unix(), m.Registered.Unix())
}

func TestRegisterTimeout(t *testing.T) {
	timeout := 10 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	mock := mockstore.NewStore()
	api := New(mock, APIOptions{DisableSigVerify: true})

	// after this ctx's deadline will be exceeded
	time.Sleep(timeout * 2)

	mock.SetError(ctx.Err())

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/entries", bytes.NewBufferString("{}"))

	api.ServeHTTP(w, r.WithContext(ctx))

	require.Equal(t, http.StatusRequestTimeout, w.Code, w.Body.String())
}

func TestGETTransportByID(t *testing.T) {
	mock := mockstore.NewStore()

	api := New(mock, APIOptions{DisableSigVerify: true})

	ctx := context.Background()

	expected := newTestTransport()
	mock.RegisterTransport(ctx, expected)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/ids/%d", expected.ID), nil)
	api.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp TransportResponse
	require.NoError(t,
		json.Unmarshal(w.Body.Bytes(), &resp),
	)

	m := resp.Model()
	assert.Equal(t, expected.ID, m.ID)
	assert.Equal(t, expected.Edges, m.Edges)
	assert.Equal(t, expected.Registered.Unix(), m.Registered.Unix())

	t.Run("Persistence", func(t *testing.T) {
		found, err := mock.GetTransportByID(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, found, expected)
	})
}

func TestDELETETransportByID(t *testing.T) {
	mock := mockstore.NewStore()

	expected := newTestTransport()
	api := New(mock, APIOptions{DisableSigVerify: true})

	mock.RegisterTransport(context.Background(), expected)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", fmt.Sprintf("/ids/%d", expected.ID), nil)
	api.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp DeletedTransportsResponse
	require.NoError(t,
		json.Unmarshal(w.Body.Bytes(), &resp),
		w.Body.String(),
	)

	got := NewTransportResponse(*expected)
	require.Len(t, resp.Deleted, 1)

	assert.Equal(t, resp.Deleted[0].ID, got.ID)
}

func TestGETIncrementingNonces(t *testing.T) {
	mock := mockstore.NewStore()

	pubKey, _ := cipher.GenerateKeyPair()

	api := New(mock, APIOptions{})

	t.Run("ValidRequest", func(t *testing.T) {
		ctx := context.Background()
		for _ = range [0xff]bool{} {
			mock.IncrementNonce(context.Background(), pubKey)
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
		mock.SetError(boom)
		defer mock.SetError(nil)

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
		assert.Contains(t, w.Body.String(), "Invalid")
	})
}
