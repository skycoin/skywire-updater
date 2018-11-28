package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/mockstore"
)

var testPubKey, testSec = cipher.GenerateKeyPair()

// validHeaders returns a valid set of headers
func validHeaders(t *testing.T, payload string) http.Header {
	nonce := "1"
	hash := cipher.SumSHA256([]byte(
		fmt.Sprintf("%s%s", payload, nonce),
	))

	sig, err := cipher.SignHash(hash, testSec)
	require.NoError(t, err)

	hdr := http.Header{}
	hdr.Set("SW-Public", testPubKey.Hex())
	hdr.Set("SW-Sig", sig.Hex())
	hdr.Set("SW-Nonce", nonce)

	return hdr
}

func TestAuthFromHeaders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mockstore.NewMockStore(ctrl)

	api := New(mock, APIOptions{})
	ctx := context.Background()

	t.Run("Valid", func(t *testing.T) {
		mock.EXPECT().GetNonce(ctx, testPubKey).Return(store.Nonce(1), nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header = validHeaders(t, "")

		api.ServeHTTP(w, r)

		assert.NotEqual(t, 401, w.Code, w.Body.String())
	})
}

func TestAuthFormat(t *testing.T) {
	headers := []string{"SW-Public", "SW-Sig", "SW-Nonce"}
	for _, header := range headers {
		t.Run(header+"-IsMissing", func(t *testing.T) {
			hdr := validHeaders(t, "")
			hdr.Del(header)

			_, err := authFromHeaders(hdr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), header)
		})
	}

	t.Run("NonceFormat", func(t *testing.T) {
		nonces := []string{"not_a_number", "-1", "0x0"}
		hdr := validHeaders(t, "")
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
	payload := "test payload"

	hash := cipher.SumSHA256([]byte(
		fmt.Sprintf("%s%d", payload, nonce),
	))

	pub, sec := cipher.GenerateKeyPair()
	sig, err := cipher.SignHash(hash, sec)
	require.NoError(t, err)

	auth := &Auth{
		Key:   pub,
		Nonce: nonce,
		Sig:   sig,
	}

	assert.NoError(t, auth.Verify([]byte(payload)))
	assert.Error(t, auth.Verify([]byte("other payload")), ".Validate should return an error for this payload")
}
