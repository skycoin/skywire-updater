package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

// Auth struct maps SW-{Key,Nonce,Sig} headers
type Auth struct {
	Key   cipher.PubKey
	Nonce store.Nonce
	Sig   string
}

func authFromHeaders(hdr http.Header) (*Auth, error) {
	a := &Auth{}
	if pub := hdr.Get("SW-Public"); pub == "" {
		return nil, errors.New("SW-Public missing")
	} else {
		key, err := cipher.PubKeyFromHex(pub)
		if err != nil {
			return nil, fmt.Errorf("Error parsing SW-Public: %s", err.Error())
		}
		a.Key = key
	}

	if a.Sig = hdr.Get("SW-Sig"); a.Sig == "" {
		return nil, errors.New("SW-Sig missing")
	}

	nonceStr := hdr.Get("SW-Nonce")
	if nonceStr == "" {
		return nil, errors.New("SW-Nonce missing")
	}

	nonceUint, err := strconv.ParseUint(nonceStr, 10, 64)
	if err != nil {
		if numErr, ok := err.(*strconv.NumError); ok {
			return nil, fmt.Errorf("Error parsing SW-Nonce: %s", numErr.Err.Error())
		}

		return nil, fmt.Errorf("Error parsing SW-Nonce: %s", err.Error())
	}
	a.Nonce = store.Nonce(nonceUint)
	return a, nil
}

func (api *API) verifyAuth(ctx context.Context, auth *Auth) error {
	cur, err := api.store.GetNonce(ctx, auth.Key)
	if err != nil {
		return err
	}

	if auth.Nonce != store.Nonce(cur) {
		return errors.New("SW-Nonce does not match")
	}

	// TODO: Signature verification

	return nil
}
