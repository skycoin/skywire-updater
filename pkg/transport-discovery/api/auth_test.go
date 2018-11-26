package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	api := New(mock)
	ctx := context.Background()

	t.Run("Valid", func(t *testing.T) {
		mock.EXPECT().GetNonce(ctx, "pub_key").Return(store.Nonce(1), nil)
		auth, err := api.auth(ctx, validHeaders())
		require.NoError(t, err)

		assert.Equal(t, "pub_key", auth.Key)
		assert.Equal(t, store.Nonce(1), auth.Nonce)
		assert.Equal(t, "sig", auth.Sig)
	})

	headers := []string{"SW-Public", "SW-Sig", "SW-Nonce"}
	for _, header := range headers {
		t.Run(header+"-IsMissing", func(t *testing.T) {
			hdr := validHeaders()
			hdr.Del(header)
			_, err := api.auth(ctx, hdr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), header)
		})
	}

	t.Run("NonceFormat", func(t *testing.T) {
		nonces := []string{"not_a_number", "-1", "0x0"}
		hdr := validHeaders()
		for _, n := range nonces {
			hdr.Set("SW-Nonce", n)
			_, err := api.auth(ctx, hdr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "SW-Nonce: invalid syntax")
		}
	})
}
