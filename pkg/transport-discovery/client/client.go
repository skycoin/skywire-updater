package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire/pkg/transport"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/api"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

// Client performs Transport discovery operations.
type Client interface {
	RegisterTransports(ctx context.Context, entries ...*transport.SignedEntry) error
	GetTransportByID(ctx context.Context, id uuid.UUID) (*store.EntryWithStatus, error)
	GetTransportsByEdge(ctx context.Context, pk cipher.PubKey) ([]*store.EntryWithStatus, error)
	UpdateStatuses(ctx context.Context, statuses ...*transport.Status) ([]*store.EntryWithStatus, error)
}

// APIClient implements Client for discovery API.
type APIClient struct {
	addr   string
	client http.Client
	key    cipher.PubKey
	sec    cipher.SecKey
}

func sanitizedAddr(addr string) string {
	if addr == "" {
		return "http://localhost"
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "http://localhost"
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	u.Path = strings.TrimSuffix(u.Path, "/")
	return u.String()
}

// New creates a new client instance.
func New(addr string) Client {
	return &APIClient{
		addr:   sanitizedAddr(addr),
		client: http.Client{},
	}
}

// NewWithAuth creates a new client setting a public key to the client to be used for auth.
// When keys are set, the client will sign request before submitting.
// The signature information is transmitted in the header using:
// * SW-Public: The specified public key
// * SW-Nonce:  The nonce for that public key
// * SW-Sig:    The signature of the payload + the nonce
func NewWithAuth(addr string, key cipher.PubKey, sec cipher.SecKey) Client {
	return &APIClient{
		addr:   sanitizedAddr(addr),
		client: http.Client{},
		key:    key,
		sec:    sec,
	}
}

// Post POST a resource
func (c *APIClient) Post(ctx context.Context, path string, payload interface{}) (*http.Response, error) {
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

// Get performs a new GET request.
func (c *APIClient) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.addr+path, new(bytes.Buffer))
	if err != nil {
		return nil, err
	}

	return c.Do(req.WithContext(ctx))
}

// Do performs a new Request.
func (c *APIClient) Do(req *http.Request) (*http.Response, error) {
	if (c.key == cipher.PubKey{}) {
		return c.client.Do(req)
	}

	req.Header.Add("SW-Public", c.key.Hex())
	nonce, err := c.getNextNonce(req.Context(), c.key)
	if err != nil {
		return nil, err
	}
	req.Header.Add("SW-Nonce", strconv.FormatUint(uint64(nonce), 10))

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	hash := cipher.SumSHA256([]byte(
		fmt.Sprintf("%s%d", string(body), nonce),
	))

	sig := cipher.MustSignHash(hash, c.sec)
	if err != nil {
		return nil, err
	}
	req.Header.Add("SW-Sig", sig.Hex())

	return c.client.Do(req)
}

func (c *APIClient) getNextNonce(_ context.Context, key cipher.PubKey) (store.Nonce, error) {
	resp, err := c.client.Get(c.addr + "/security/nonces/" + key.Hex())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error getting current nonce: status: %d <- %v", resp.StatusCode, extractError(resp.Body))
	}

	var nr api.NonceResponse
	if err := json.NewDecoder(resp.Body).Decode(&nr); err != nil {
		return 0, err
	}

	return store.Nonce(nr.NextNonce), nil
}

// RegisterTransports registers new Transports.
func (c *APIClient) RegisterTransports(ctx context.Context, entries ...*transport.SignedEntry) error {
	if len(entries) == 0 {
		return nil
	}

	resp, err := c.Post(ctx, "/transports/", entries)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("status: %d, error: %v", resp.StatusCode, extractError(resp.Body))
}

// GetTransportByID returns Transport for corresponding ID.
func (c *APIClient) GetTransportByID(ctx context.Context, id uuid.UUID) (*store.EntryWithStatus, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/transports/id:%s", id.String()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, error: %v", resp.StatusCode, extractError(resp.Body))
	}

	entry := &store.EntryWithStatus{}
	if err := json.NewDecoder(resp.Body).Decode(entry); err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}

	return entry, nil
}

// GetTransportsByEdge returns all Transport registered for the edge.
func (c *APIClient) GetTransportsByEdge(ctx context.Context, pk cipher.PubKey) ([]*store.EntryWithStatus, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/transports/edge:%s", pk.Hex()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, error: %v", resp.StatusCode, extractError(resp.Body))
	}

	entry := []*store.EntryWithStatus{}
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}

	return entry, nil
}

// UpdateStatuses updates statuses of transports in discovery.
func (c *APIClient) UpdateStatuses(ctx context.Context, statuses ...*transport.Status) ([]*store.EntryWithStatus, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	resp, err := c.Post(ctx, "/statuses", statuses)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, error: %v", resp.StatusCode, extractError(resp.Body))
	}

	entries := []*store.EntryWithStatus{}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}

	return entries, nil
}

// extractError returns the decoded error message from Body.
func extractError(r io.Reader) error {
	var apiError api.Error

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &apiError); err != nil {
		return errors.New(string(body))
	}

	return errors.New(apiError.Error)
}
