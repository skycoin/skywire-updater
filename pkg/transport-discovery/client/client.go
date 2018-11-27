package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type Client struct {
	addr   string
	client http.Client
	key    string
}

// Creates
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

// Post POST a resource
func (c *Client) Post(ctx context.Context, path string, payload interface{}) (*http.Response, error) {
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(payload); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.addr+path, body)
	if err != nil {
		return nil, err
	}

	return c.Do(req.WithContext(ctx))
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.key != "" {
		req.Header.Add("SW-Public", c.key)
	}
	// TODO: get nonce and sign the request

	return c.client.Do(req)
}

func (c *Client) RegisterTransport(ctx context.Context, t *store.Transport) error {
	resp, err := c.Post(ctx, "/register", t)
	if err != nil {
		return err
	}

	if resp.StatusCode == 201 {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("status: %d, error: %s", resp.StatusCode, string(body))
}

func (c *Client) DeregisterTransport(ctx context.Context, id store.ID) error {
	return nil
}
