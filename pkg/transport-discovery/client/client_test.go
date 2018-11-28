package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

var testPubKey, testSecKey = cipher.GenerateKeyPair()

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

	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch url := r.URL.String(); url {
			case "/":
				defer wg.Done()
				assert.Equal(t, testPubKey.Hex(), r.Header.Get("SW-Public"))
				assert.Equal(t, "1", r.Header.Get("SW-Nonce"))
				assert.NotEmpty(t, r.Header.Get("SW-Sig")) // TODO: check for the right key

			case "/incrementing-nonce/" + testPubKey.Hex():
				fmt.Fprintf(w, `{"edge": "%s", "next_nonce": 1}`, testPubKey.Hex())

			default:
				t.Errorf("Don't know how to handle URL = '%s'", url)
			}
		},
	))
	defer srv.Close()

	c := New(srv.URL).WithPubAndSecKey(testPubKey, testSecKey)

	wg.Add(1)
	_, err := c.Post(context.Background(), "/", bytes.NewBufferString("test payload"))
	require.NoError(t, err)

	wg.Wait()
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
