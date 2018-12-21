package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

var (
	// ErrEmptyPubKey indicates that provided PubKey is empty.
	ErrEmptyPubKey = errors.New("PublicKey can't by empty")
	// ErrInvalidPubKey indicates that provided PubKey is invalid.
	ErrInvalidPubKey = errors.New("PublicKey is invalid")
	// ErrEmptyTransportID indicates that provided TransportID is empty.
	ErrEmptyTransportID = errors.New("TransportID can't by empty")
	// ErrInvalidTransportID indicates that provided TransportID is invalid.
	ErrInvalidTransportID = errors.New("TransportID is invalid")
)

// Options control particular behavior
type Options struct {
	// DisableSigVerify disables signature verification on the request header
	DisableSigVerify bool
}

// API register all the API endpoints.
// It implements a net/http.Handler.
type API struct {
	mux   *http.ServeMux
	store store.Store
	opts  Options
}

// New constructs a new API instance.
func New(s store.Store, opts Options) *API {
	mux := http.NewServeMux()
	api := &API{mux: mux, store: s, opts: opts}

	mux.Handle("/transports/", api.withSigVer(apiHandler(api.handleTransports)))
	mux.Handle("/statuses", api.withSigVer(apiHandler(api.handleStatuses)))
	mux.Handle("/security/nonces/", apiHandler(api.handleIncrementingNonces))

	return api
}

// Error is the object returned to the client when there's an error.
type Error struct {
	Error string `json:"error"`
}

func renderError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&Error{Error: err.Error()}) // nolint
}

// apiHandler is an adapter to reduce api handler endpoint boilerplate
type apiHandler func(w http.ResponseWriter, r *http.Request) (interface{}, error)

// ServeHTTP implements http.Handler.
func (fn apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")

	res, err := fn(w, r)
	if err == nil {
		json.NewEncoder(w).Encode(res) // nolint
		return
	}

	var status int

	switch err {
	case ErrEmptyPubKey, ErrEmptyTransportID, ErrInvalidTransportID, ErrInvalidPubKey:
		status = http.StatusBadRequest
	case context.DeadlineExceeded:
		status = http.StatusRequestTimeout
	}

	// we still haven't found the error
	if status == 0 {
		switch err.(type) {
		case *json.SyntaxError:
			status = http.StatusBadRequest
		}
	}

	// we fallback to 500
	if status == 0 {
		status = http.StatusInternalServerError
	}

	renderError(w, status, err)
}

func (api *API) withSigVer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if api.opts.DisableSigVerify {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		auth, err := authFromHeaders(r.Header)
		if err != nil {
			renderError(w, http.StatusUnauthorized, err)
			return
		}

		if err := api.VerifyAuth(r, auth); err != nil {
			renderError(w, http.StatusUnauthorized, err)
			return
		}

		if r.Method == "POST" && (auth.Key != cipher.PubKey{}) {
			_, err := api.store.IncrementNonce(ctx, auth.Key)
			if err != nil {
				renderError(w, http.StatusInternalServerError, err)
				return
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, store.ContextAuthKey, auth.Key)))
	}

	return http.HandlerFunc(fn)
}

// ServeHTTP implements http.Handler.
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Add some logging
	api.mux.ServeHTTP(w, r)
}
