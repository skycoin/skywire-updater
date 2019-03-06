package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/skycoin/skywire-updater/pkg/update"
)

const rpcPrefix = "updater"

var log = logging.MustGetLogger("api")

// Gateway provides the API gateway.
type Gateway interface {
	Services() []string
	Check(ctx context.Context, srvName string) (*update.Release, error)
	Update(ctx context.Context, srvName, toVersion string) (bool, error)
}

// Handle makes a http.Handler from a Gateway implementation.
func Handle(g Gateway, enableAPI, enableRPC bool) http.Handler {
	r := chi.NewRouter()
	if enableAPI {
		r.Mount("/api", handleREST(g))
	}
	if enableRPC {
		r.Mount("/rpc", handleRPC(g))
	}
	return r
}
