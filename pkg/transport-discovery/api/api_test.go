package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/mockstore"
)

func TestPOSTRegisterTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trans := store.Transport{
		Edges: []string{"pub_key_1", "pub_key_2"},
	}

	mock := mockstore.NewMockStore(ctrl)
	mock.EXPECT().GetNonce(gomock.Any(), gomock.Any()).Return(store.Nonce(1), nil)
	mock.EXPECT().IncrementNonce(gomock.Any(), "pub_key").Return(store.Nonce(2), nil)
	mock.EXPECT().RegisterTransport(gomock.Any(), gomock.Any()).Do(
		func(_ context.Context, in *store.Transport) error {
			in.ID = 0xff
			return nil
		},
	)

	api := New(mock)
	w := httptest.NewRecorder()

	post := bytes.NewBuffer(nil)
	json.NewEncoder(post).Encode(trans)
	r := httptest.NewRequest("POST", "/register", post)
	r.Header = validHeaders()
	api.ServeHTTP(w, r)

	assert.Equal(t, 201, w.Code, w.Body.String())

	json.NewDecoder(bytes.NewBuffer(w.Body.Bytes())).Decode(&trans)
	assert.NotEmpty(t, trans.ID)
}
