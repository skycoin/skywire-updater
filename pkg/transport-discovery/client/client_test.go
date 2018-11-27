package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
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
	c.Post(context.Background(), "/", nil)
}

func TestRegisterTransport(t *testing.T) {
	wg := sync.WaitGroup{}

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		assert  func(err error)
	}{
		{
			"Status201",
			func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201) },
			func(err error) { require.NoError(t, err) },
		},
		{
			"Status200",
			func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) },
			func(err error) { require.Error(t, err) },
		},
		{
			"Status500",
			func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) },
			func(err error) { require.Error(t, err) },
		},
		{
			"Request",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/register", r.URL.String())
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wg.Add(1)

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					defer wg.Done()
					test.handler(w, r)
				},
			))
			defer srv.Close()

			c := New(srv.URL)
			err := c.RegisterTransport(context.Background(), &store.Transport{})
			if test.assert != nil {
				test.assert(err)
			}

			wg.Wait()
		})
	}
}
