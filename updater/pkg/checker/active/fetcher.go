package active

import (
	"time"

	"github.com/watercompany/skywire-services/updater/pkg/logger"
)

type Fetcher interface {
	SetInterval(duration time.Duration)
	Start()
	Stop()
}


func New(kind, service, repository, notifyUrl string, log *logger.Logger) Fetcher {
	switch kind {
	case "git":
		return NewGit(service, repository, notifyUrl, log)
	case "naive":
		return NewNaive(service, repository, notifyUrl, log)
	}
	return NewNaive(service, repository, notifyUrl, log)
}
