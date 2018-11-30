package active

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/watercompany/skywire-services/updater/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/pkg/checker"
)

type naive struct {
	// Url should be in the format /:owner/:Repository
	service   string
	url       string
	interval  time.Duration
	ticker    *time.Ticker
	lock      sync.Mutex
	exit      chan int
	notifyUrl string
	log       *logger.Logger
	config.CustomLock
}

func NewNaive(service, url string, notifyUrl string, log *logger.Logger) *naive {
	return &naive{
		url:       "github.com" + url,
		exit:      make(chan int),
		service:   service,
		notifyUrl:	notifyUrl,
		log:       log,
	}
}

func (n *naive) SetInterval(t time.Duration) {
	n.interval = t

	n.lock.Lock()
	if n.ticker != nil {
		n.ticker = time.NewTicker(n.interval)
	}
	n.lock.Unlock()
}

func (n *naive) Start() {
	n.ticker = time.NewTicker(n.interval)
	go func() {
		for {
			select {
			case t := <-n.ticker.C:
				n.log.Info("looking for new version at: ", t)
				// Try to fetch new version
				go n.checkIfNew()
			}
		}
	}()
	<-n.exit
}

func (n *naive) Stop() {
	n.ticker.Stop()
	n.exit <- 1
}

func (n *naive) checkIfNew() {
	if n.IsLock() {
		n.log.Warnf("service %s is already being updated... waiting for it to finish", n.service)
	}
	n.Lock()
	defer n.Unlock()

	n.log.Info("updating...")

	err := checker.NotifyUpdate(n.notifyUrl, n.service, "master", "master", "token")

	if err != nil {
		logrus.Error(err)
	}
}

/*func (n *naive) tryUpdate() error {
	for i := 0; i < n.retries; i++ {
		err := <-n.updater.Update(n.service, n.service, n.log)
		if err != nil {
			n.log.Errorf("error on update %s", err)

			if i == (n.retries - 1) {
				return fmt.Errorf("maximum retries attempted, service %s failed to update", n.service)
			} else {
				n.log.Infof("retry again in %s", n.retryTime.String())
			}
		} else {
			break
		}

		time.Sleep(n.retryTime)
	}
	return nil
}*/
