package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/auth"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/mockstore"
)

var testPubKey, testSec = cipher.GenerateKeyPair()

// validHeaders returns a valid set of headers
func validHeaders(t *testing.T, payload []byte) http.Header {
	nonce := store.Nonce(0)
	sig, err := auth.Sign(payload, nonce, testSec)
	require.NoError(t, err)

	hdr := http.Header{}
	hdr.Set("SW-Public", testPubKey.Hex())
	hdr.Set("SW-Sig", sig.Hex())
	hdr.Set("SW-Nonce", nonce.String())

	return hdr
}

func TestAuthFromHeaders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mockstore.NewMockStore(ctrl)

	api := New(mock, APIOptions{})
	ctx := context.Background()

	t.Run("Valid", func(t *testing.T) {
		mock.EXPECT().IncrementNonce(gomock.Any(), gomock.Any()).AnyTimes()
		mock.EXPECT().GetNonce(ctx, testPubKey).Return(store.Nonce(0), nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/register", nil)
		r.Header = validHeaders(t, nil)

		api.ServeHTTP(w, r)

		assert.NotEqual(t, 401, w.Code, w.Body.String())
	})
}

func TestAuthFormat(t *testing.T) {
	headers := []string{"SW-Public", "SW-Sig", "SW-Nonce"}
	for _, header := range headers {
		t.Run(header+"-IsMissing", func(t *testing.T) {
			hdr := validHeaders(t, nil)
			hdr.Del(header)

			_, err := authFromHeaders(hdr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), header)
		})
	}

	t.Run("NonceFormat", func(t *testing.T) {
		nonces := []string{"not_a_number", "-1", "0x0"}
		hdr := validHeaders(t, nil)
		for _, n := range nonces {
			hdr.Set("SW-Nonce", n)
			_, err := authFromHeaders(hdr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "SW-Nonce: invalid syntax")
		}
	})
}

func TestAuthSignatureVerification(t *testing.T) {
	nonce := store.Nonce(0xdeadbeef)
	payload := []byte("dead beed")

	sig, err := auth.Sign(payload, nonce, testSec)
	require.NoError(t, err)

	auth := &Auth{
		Key:   testPubKey,
		Nonce: nonce,
		Sig:   sig,
	}

	assert.NoError(t, auth.Verify([]byte(payload)))
	assert.Error(t, auth.Verify([]byte("other payload")), ".Validate should return an error for this payload")
}
