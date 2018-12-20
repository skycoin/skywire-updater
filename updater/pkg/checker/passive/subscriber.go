package passive

import (
	"strings"

	"github.com/watercompany/skywire-services/updater/pkg/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/pkg/updater"
)

// Subscriber represents an object that can subscribe for updates of a given service
type Subscriber interface {
	Subscribe(topic string)
	Start()
	Stop()
}

// New returns a subscriber of the given MessageBroker
func New(config config.SubscriberConfig, updater updater.Updater, log *logger.Logger) Subscriber {
	config.MessageBroker = strings.ToLower(config.MessageBroker)
	switch config.MessageBroker {
	case "nats":
		return newNats(updater, config.Urls[0], config.NotifyURL, log)
	}

	return newNats(updater, config.Urls[0], config.NotifyURL, log)
}
