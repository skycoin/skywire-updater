package api

import (
	"encoding/json"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

type NonceResponse struct {
	Edge      string
	NextNonce uint64
}

type TransportResponse struct {
	store.Transport
	Registered int64
}

func (t *TransportResponse) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &t.Transport); err != nil {
		return err
	}

	t.Registered = t.Transport.Registered.Unix()
	return nil
}
