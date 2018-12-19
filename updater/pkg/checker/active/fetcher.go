package active

import (
	"time"

	"github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/config"
)

type Fetcher interface {
	SetInterval(duration time.Duration)
	Start()
	Stop()
}


func New(kind, service, localName, repository, notifyUrl string, c config.ServiceConfig,
	scriptTimeout time.Duration, log *logger.Logger) Fetcher {

		switch kind {
	case "git":
		return NewGit(service, repository, notifyUrl, log)
	case "naive":
		return NewNaive(service, localName, repository, notifyUrl, c.CheckScriptInterpreter, c.CheckScript,
			c.CheckScriptExtraArguments, scriptTimeout,log)
	}
	return NewNaive(service, localName, repository, notifyUrl, c.CheckScriptInterpreter, c.CheckScript,
		c.CheckScriptExtraArguments, scriptTimeout,log)
}
