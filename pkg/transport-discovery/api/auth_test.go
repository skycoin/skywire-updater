package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/mockstore"
)

func validHeaders() http.Header {
	hdr := http.Header{}
	hdr.Set("SW-Public", "pub_key")
	hdr.Set("SW-Sig", "sig")
	hdr.Set("SW-Nonce", "1")
	return hdr
}

func TestAuthFromHeaders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mockstore.NewMockStore(ctrl)

	api := New(mock, APIOptions{})
	ctx := context.Background()

	t.Run("Valid", func(t *testing.T) {
		mock.EXPECT().GetNonce(ctx, "pub_key").Return(store.Nonce(1), nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header = validHeaders()

		api.ServeHTTP(w, r)

		assert.NotEqual(t, 401, w.Code, w.Body.String())
	})
}

func TestAuthFormat(t *testing.T) {
	headers := []string{"SW-Public", "SW-Sig", "SW-Nonce"}
	for _, header := range headers {
		t.Run(header+"-IsMissing", func(t *testing.T) {
			hdr := validHeaders()
			hdr.Del(header)

			_, err := authFromHeaders(hdr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), header)
		})
	}

	t.Run("NonceFormat", func(t *testing.T) {
		nonces := []string{"not_a_number", "-1", "0x0"}
		hdr := validHeaders()
		for _, n := range nonces {
			hdr.Set("SW-Nonce", n)
			_, err := authFromHeaders(hdr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "SW-Nonce: invalid syntax")
		}
	})
}
