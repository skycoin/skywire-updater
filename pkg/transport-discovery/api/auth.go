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

	body, err := peekBody(r)
	if err != nil {
		return err
	}

	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	return auth.Verify(payload)
}

// peekBody reads the body from http.Request without removing the data from the original body.
func peekBody(r *http.Request) (io.ReadCloser, error) {
	var err error
	save := r.Body
	save, r.Body, err = drainBody(r.Body)

	return save, err
}

// drainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes.

// It returns an error if the initial slurp of all bytes fails. It does not attempt
// to make the returned ReadClosers have identical error-matching behavior.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}

	if err = b.Close(); err != nil {
		return nil, b, err
	}

	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
