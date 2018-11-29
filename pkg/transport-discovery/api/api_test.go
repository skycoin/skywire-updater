package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/mockstore"
)

func TestBadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mockstore.NewMockStore(ctrl)

	api := New(mock, APIOptions{DisableSigVerify: true})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/register", bytes.NewBufferString("not-a-json"))

	api.ServeHTTP(w, r)

	assert.Equal(t, 400, w.Code)
}

func TestPOSTRegisterTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	pk1, _ := cipher.GenerateKeyPair()
	pk2, _ := cipher.GenerateKeyPair()
	trans := store.Transport{
		ID:         0xff,
		Edges:      []cipher.PubKey{pk1, pk2},
		Registered: now,
	}

	mock := mockstore.NewMockStore(ctrl)
	mock.EXPECT().RegisterTransport(gomock.Any(), gomock.Any()).Return(nil)

	api := New(mock, APIOptions{DisableSigVerify: true})
	w := httptest.NewRecorder()

	post := bytes.NewBuffer(nil)
	json.NewEncoder(post).Encode(trans)
	r := httptest.NewRequest("POST", "/register", post)
	api.ServeHTTP(w, r)

	assert.Equal(t, 201, w.Code, w.Body.String())

	var resp TransportResponse
	json.NewDecoder(bytes.NewBuffer(w.Body.Bytes())).Decode(&resp)
	assert.Equal(t, trans.ID, resp.ID)
	assert.Equal(t, trans.Edges, resp.Edges)
	assert.Equal(t, trans.Registered.Unix(), resp.Registered)
}

func TestGETIncrementingNonces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mockstore.NewMockStore(ctrl)

	pubKey, _ := cipher.GenerateKeyPair()

	api := New(mock, APIOptions{})

	t.Run("ValidRequest", func(t *testing.T) {
		ctx := context.Background()

		mock.EXPECT().GetNonce(ctx, pubKey).Return(store.Nonce(0xff), nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/incrementing-nonces/"+pubKey.Hex(), nil)
		api.ServeHTTP(w, r.WithContext(ctx))
		require.Equal(t, 200, w.Code, w.Body.String())

		var resp NonceResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, pubKey.Hex(), resp.Edge)
		assert.Equal(t, uint64(0xff), resp.NextNonce)
	})

	t.Run("StoreError", func(t *testing.T) {
		boom := errors.New("boom")
		mock.EXPECT().GetNonce(gomock.Any(), gomock.Any()).Return(store.Nonce(0), boom)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/incrementing-nonces/"+pubKey.Hex(), nil)
		api.ServeHTTP(w, r)
		require.Equal(t, 500, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), boom.Error())
	})

	t.Run("EmptyKey", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/incrementing-nonces/", nil)
		api.ServeHTTP(w, r)
		require.Equal(t, 400, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "empty")
	})

	t.Run("InvalidKey", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/incrementing-nonces/foo-bar", nil)
		api.ServeHTTP(w, r)
		require.Equal(t, 400, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "Invalid")
	})
}
