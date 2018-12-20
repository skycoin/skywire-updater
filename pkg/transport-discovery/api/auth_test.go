package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/auth"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/mockstore"
)

var testPubKey, testSec = cipher.GenerateKeyPair()

// validHeaders returns a valid set of headers
func validHeaders(t *testing.T, payload []byte) http.Header {
	nonce := store.Nonce(0)
	sig := auth.Sign(payload, nonce, testSec)

	hdr := http.Header{}
	hdr.Set("SW-Public", testPubKey.Hex())
	hdr.Set("SW-Sig", sig.Hex())
	hdr.Set("SW-Nonce", nonce.String())

	return hdr
}

func TestAuthFromHeaders(t *testing.T) {
	mock := mockstore.NewStore()

	api := New(mock, APIOptions{})

	t.Run("Valid", func(t *testing.T) {

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/entries", nil)
		r.Header = validHeaders(t, nil)

		api.ServeHTTP(w, r)

		assert.NotEqual(t, http.StatusUnauthorized, w.Code, w.Body.String())
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

	sig := auth.Sign(payload, nonce, testSec)

	auth := &Auth{
		Key:   testPubKey,
		Nonce: nonce,
		Sig:   sig,
	}

	assert.NoError(t, auth.Verify([]byte(payload)))
	assert.Error(t, auth.Verify([]byte("other payload")), "Validate should return an error for this payload")
}
