package supervisor

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/watercompany/skywire-services/updater/config"
	"github.com/watercompany/skywire-services/updater/pkg/checker/active"
	"github.com/watercompany/skywire-services/updater/pkg/checker/passive"
	"github.com/watercompany/skywire-services/updater/pkg/updater"
	loggerPkg "github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/store/services"
	"github.com/pkg/errors"
	"github.com/skycoin/skycoin/src/util/logging"
	"sync"
	"fmt"
)

var (
	logger = logging.MustGetLogger("updater")
	ErrServiceNotFound = errors.New("service definition not found")
)

var defaultSubscriptorConfig = struct {
	kind string
}{
	kind: "naive",
}

type Supervisor struct {
	activeCheckers  map[string]active.Fetcher
	passiveCheckers map[string]passive.Subscriber
	updaters        map[string]updater.Updater
	defaultFetcherConfig active.Fetcher
	defaultService updater.Updater
	config config.Configuration
	sync.RWMutex
}

func New(conf config.Configuration) *Supervisor {
	s := &Supervisor{
		activeCheckers:  map[string]active.Fetcher{},
		passiveCheckers: map[string]passive.Subscriber{},
		updaters:        map[string]updater.Updater{},
		config: conf,
	}

	services.InitStorer("json")

	s.createUpdaters(conf)
	s.createCheckers(conf)

	return s
}

func (s *Supervisor) Register(service, url, notifyUrl, version string) {
	s.Lock()
	defer s.Unlock()
	checker := active.New(defaultSubscriptorConfig.kind,
		service, url, notifyUrl, loggerPkg.NewLogger(service))

	s.activeCheckers[fmt.Sprintf("%s-checker",service)] = checker
	go checker.Start()
}

func (s *Supervisor) Start() {
	for _, checker := range s.activeCheckers {
		go checker.Start()
	}

	for _, checker := range s.passiveCheckers {
		go checker.Start()
	}
}

func (s *Supervisor) Stop() {
	for _, checker := range s.activeCheckers {
		checker.Stop()
	}

	for _, checker := range s.passiveCheckers {
		checker.Stop()
	}
}

func (s *Supervisor) Update(service string) error {
	logger.Infof("%+v\n",s.config.Services)

	// get updater
	serviceConfig, ok := s.config.Services[service]
	if !ok {
		return ErrServiceNotFound
	}

	updater := s.updaters[serviceConfig.Updater]

	// Try update
	err := <- updater.Update(service, serviceConfig.CheckTag, loggerPkg.NewLogger(service))
	if err != nil {
		return err
	}
	return nil
}

func (s *Supervisor) createUpdaters(conf config.Configuration) {
	for name, c := range conf.Updaters {
		u := updater.New(c.Kind, conf)
		s.updaters[name] = u
	}
}

func (s *Supervisor) createCheckers(conf config.Configuration) {
	for name, c := range conf.Services {
		if c.ActiveUpdateChecker != "" {
			s.registerActiveChecker(conf, c, name)
		} else {
			s.registerPassiveChecker(conf, c, name)
		}
	}
}

func (s *Supervisor) registerPassiveChecker(conf config.Configuration, c config.ServiceConfig, name string) {
	passiveConfig, ok := conf.PassiveUpdateCheckers[c.PassiveUpdateChecker]
	if !ok {
		logrus.Fatalf("%s checker not defined for service %s",
			c.ActiveUpdateChecker, name)
	}
	log := loggerPkg.NewLogger(name)
	sub := passive.New(passiveConfig, s.updaters[c.Updater], log)
	s.passiveCheckers[name] = sub
	sub.Subscribe(passiveConfig.Topic)
}

func (s *Supervisor) registerActiveChecker(conf config.Configuration, c config.ServiceConfig, name string) {
	activeConfig, ok := conf.ActiveUpdateCheckers[c.ActiveUpdateChecker]
	if !ok {
		logrus.Fatalf("%s checker not defined for service %s",
			c.ActiveUpdateChecker, name)
	}
	interval, err := time.ParseDuration(activeConfig.Interval)
	if err != nil {
		logrus.Fatalf("cannot parse interval %s of active checker configuration %s. %s", activeConfig.Interval,
			c.ActiveUpdateChecker, err)
	}
	log := loggerPkg.NewLogger(name)
	fc, err := conf.FetcherConfig(c.ActiveUpdateChecker)
	if err != nil {
		panic(err)
	}
	checker := active.New(fc.Kind, name, c.Repository, fc.NotifyUrl, log)
	checker.SetInterval(interval)
	s.activeCheckers[name] = checker
}
