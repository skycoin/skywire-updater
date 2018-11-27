package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		addr     string
		expected string
	}{
		{"", "http://localhost"},
		{"http://localhost:8080", "http://localhost:8080"},
		{"localhost", "http://localhost"},
		{"http://localhost/path/", "http://localhost/path"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, New(test.addr).addr)
	}
}

func TestClientAuth(t *testing.T) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer wg.Done()
			assert.Equal(t, "pub_key", r.Header.Get("SW-Public"))
		},
	))
	defer srv.Close()

	c := New(srv.URL).WithPubKey("pub_key")

	wg.Add(1)
	c.Post(context.Background(), "/", struct{ Field string }{"x"})
}
