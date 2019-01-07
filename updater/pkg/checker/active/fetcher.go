package active

import (
	"errors"
	"time"

	"github.com/watercompany/skywire-services/updater/pkg/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
)

// Fetcher errors
var (
	ErrNoNewVersion = errors.New("no new version")
)

// Fetcher represents an update available checker, which will fetch the information of such update every
// given interval
type Fetcher interface {
	Check() error
}

// New returns a new Fetcher of the given kind type
func New(kind, service, localName, repository, notifyURL string, c config.ServiceConfig,
	scriptTimeout time.Duration, log *logger.Logger) Fetcher {

	switch kind {
	case "git":
		return newGit(service, repository, notifyURL, log)
	case "naive":
		return newNaive(service, localName, repository, notifyURL, c.CheckScriptInterpreter, c.CheckScript,
			c.CheckScriptExtraArguments, scriptTimeout, log)
	}
	return newNaive(service, localName, repository, notifyURL, c.CheckScriptInterpreter, c.CheckScript,
		c.CheckScriptExtraArguments, scriptTimeout, log)
}
