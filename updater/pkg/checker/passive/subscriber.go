package passive

import (
	"strings"

	"github.com/watercompany/skywire-services/updater/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/pkg/updater"
)

type Subscriber interface {
	Subscribe(topic string)
	Start()
	Stop()
}

func New(config config.SubscriberConfig, updater updater.Updater, log *logger.Logger) Subscriber {
	config.MessageBroker = strings.ToLower(config.MessageBroker)
	switch config.MessageBroker {
	case "nats":
		return newNats(updater, config.Urls[0], config.NotifyUrl, log)
	}

	return newNats(updater, config.Urls[0], config.NotifyUrl, log)
}

