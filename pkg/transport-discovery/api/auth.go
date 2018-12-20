package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/auth"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

// Auth struct maps SW-{Key,Nonce,Sig} headers
type Auth struct {
	Key   cipher.PubKey
	Nonce store.Nonce
	Sig   cipher.Sig
}

func authFromHeaders(hdr http.Header) (*Auth, error) {
	a := &Auth{}
	var v string

	if v = hdr.Get("SW-Public"); v == "" {
		return nil, errors.New("SW-Public missing")
	}
	key, err := cipher.PubKeyFromHex(v)
	if err != nil {
		return nil, fmt.Errorf("Error parsing SW-Public: %s", err.Error())
	}
	a.Key = key

	if v = hdr.Get("SW-Sig"); v == "" {
		return nil, errors.New("SW-Sig missing")
	}
	sig, err := cipher.SigFromHex(v)
	if err != nil {
		return nil, fmt.Errorf("Error parsing SW-Sig:'%s': %s", v, err.Error())
	}
	a.Sig = sig

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

func (a *Auth) Verify(in []byte) error {
	return auth.Verify(in, a.Nonce, a.Key, a.Sig)
}

func (api *API) VerifyAuth(r *http.Request, auth *Auth) error {
	cur, err := api.store.GetNonce(r.Context(), auth.Key)
	if err != nil {
		return err
	}

	if auth.Nonce != store.Nonce(cur) {
		return errors.New("SW-Nonce does not match")
	}

	var buf bytes.Buffer
	body := io.TeeReader(r.Body, &buf)

	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	// close the original body cause it will be replaced
	if err := r.Body.Close(); err != nil {
		return err
	}

	r.Body = ioutil.NopCloser(&buf)
	return auth.Verify(payload)
}
