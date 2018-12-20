package passive

import (
	"sync"

	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"

	"github.com/watercompany/skywire-services/updater/pkg/checker"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/pkg/updater"
)

type nats struct {
	updater    updater.Updater
	url        string
	connection *gonats.Conn
	closer     chan int
	topic      string
	notifyURL  string
	log        *logger.Logger
	sync.Mutex
}

func newNats(u updater.Updater, url string, notifyURL string, log *logger.Logger) *nats {
	connection, err := gonats.Connect(url)
	if err != nil {
		log.Fatal(err)
	}
	return &nats{
		updater:    u,
		url:        url,
		connection: connection,
		closer:     make(chan int),
		notifyURL:  notifyURL,
		log:        log,
	}
}

func (n *nats) Subscribe(topic string) {
	n.Lock()
	defer n.Unlock()
	n.topic = topic
}

func (n *nats) Start() {
	_, err := n.connection.Subscribe(n.topic, n.onUpdate)
	if err != nil {
		panic(err)
	}

	n.log.Infof("subscribed to %s", n.topic)
	<-n.closer
	n.log.Info("stop")
}

func (n *nats) Stop() {
	n.closer <- 1
}

func (n *nats) onUpdate(msg *gonats.Msg) {
	n.log.Info("received update notification")
	err := checker.NotifyUpdate(n.notifyURL, msg.Subject, msg.Subject, msg.Subject, "token")
	if err != nil {
		logrus.Fatal(err)
	}
}
