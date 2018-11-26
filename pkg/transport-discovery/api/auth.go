package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

func (api *API) auth(ctx context.Context, hdr http.Header) (*Auth, error) {
	a := &Auth{}
	if a.Key = hdr.Get("SW-Public"); a.Key == "" {
		return nil, errors.New("SW-Public missing")
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
	nonce := store.Nonce(nonceUint)

	cur, err := api.store.GetNonce(ctx, a.Key)
	if err != nil {
		return nil, err
	}

	if nonce != store.Nonce(cur) {
		return nil, errors.New("SW-Nonce does not match")
	}
	a.Nonce = nonce

	return a, nil
}
