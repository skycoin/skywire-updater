package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type Client struct {
	addr   string
	client http.Client
	key    string
}

func New(addr string) *Client {
	// Sanitize addr
	if addr == "" {
		addr = "http://localhost"
	}

	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	if strings.HasSuffix(addr, "/") {
		addr = addr[:len(addr)-1]
	}

	return &Client{
		addr:   addr,
		client: http.Client{},
	}
}

// WithPubKey set a public key to the client.
// When PubKey is set, the client will sign request before submitting.
// The signature information is transmitted in the header using:
// * SW-Public: The specified public key
// * SW-Nonce:  The nonce for that public key
// * SW-Sig:    The signature of the payload + the nonce
func (c *Client) WithPubKey(key string) *Client {
	c.key = key
	return c
}

func (c *Client) Post(ctx context.Context, path string, payload interface{}) (*http.Response, error) {
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(payload); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.addr+path, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.key != "" {
		req.Header.Add("SW-Public", c.key)
	}

	return c.client.Do(req)
}

func (c *Client) RegisterTransport(ctx context.Context, t *store.Transport) error {
	return nil
}

func (c *Client) DeregisterTransport(ctx context.Context, id store.ID) error {
	return nil
}
