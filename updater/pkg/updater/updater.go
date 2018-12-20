package updater

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/watercompany/skywire-services/updater/pkg/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
)

// Updater represents an object capable of updating services
type Updater interface {
	Update(service, version string, log *logger.Logger) chan error
	RegisterService(conf config.ServiceConfig, officialName, scriptsDirectory string)
	UnregisterService(officialName string)
}

// New returns an Updater of the given kind
func New(kind string, conf *config.Configuration) Updater {

	normalized := strings.ToLower(kind)
	logrus.Infof("updater: %s", normalized)

	switch normalized {
	case "custom":
		return newCustomUpdater(conf.ScriptsDirectory, conf.Services)
	}

	return newCustomUpdater(conf.ScriptsDirectory, conf.Services)
}
