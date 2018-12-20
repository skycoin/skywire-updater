package supervisor

import (
	"time"

	"github.com/sirupsen/logrus"

	"fmt"
	"path/filepath"
	"sync"

	"errors"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/watercompany/skywire-services/updater/pkg/checker/active"
	"github.com/watercompany/skywire-services/updater/pkg/checker/passive"
	"github.com/watercompany/skywire-services/updater/pkg/config"
	loggerPkg "github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/pkg/store/services"
	"github.com/watercompany/skywire-services/updater/pkg/updater"
)

// supervisor errors
var (
	logger             = logging.MustGetLogger("updater")
	ErrServiceNotFound = errors.New("service definition not found")
)

var defaultSubscriptorConfig = struct {
	kind     string
	interval time.Duration
}{
	kind:     "naive",
	interval: 20 * time.Second,
}

// Supervisor is responsible for spawning fetchers and updaters as well as to allow to modify or update
// updater defined services
type Supervisor struct {
	activeCheckers  map[string]active.Fetcher
	passiveCheckers map[string]passive.Subscriber
	updaters        map[string]updater.Updater
	config          *config.Configuration
	sync.RWMutex
}

func (s *Supervisor) registerChecker(service string, checker active.Fetcher) {
	s.Lock()
	defer s.Unlock()

	s.activeCheckers[fmt.Sprintf("%s-checker", service)] = checker
}

func (s *Supervisor) unregisterChecker(service string) {
	s.Lock()
	defer s.Unlock()

	delete(s.activeCheckers, fmt.Sprintf("%s-checker", service))
}

// New returns a supervisor with the given configuration
func New(conf *config.Configuration) *Supervisor {
	s := &Supervisor{
		activeCheckers:  map[string]active.Fetcher{},
		passiveCheckers: map[string]passive.Subscriber{},
		updaters:        map[string]updater.Updater{},
		config:          conf,
	}

	services.InitStorer("json")

	s.createUpdaters(conf)
	s.createCheckers(conf)

	return s
}

// Register registers a new service to fetch its updates from url, which current version is the given version
// and that on a new update will send a POST request to notifyURL
func (s *Supervisor) Register(service, url, notifyURL, version string) {
	serviceConfig := config.ServiceConfig{
		Repository:                 url,
		UpdateScriptInterpreter:    "/bin/bash",
		LocalName:                  service,
		OfficialName:               service,
		UpdateScriptExtraArguments: []string{service, url},
		UpdateScript:               "generic-service.sh",
		UpdateScriptTimeout:        "6m",
		CheckScriptExtraArguments:  []string{service, url},
		CheckScript:                filepath.Join(s.config.ScriptsDirectory, "generic-service-check-update.sh"),
		CheckScriptTimeout:         "6m",
		Updater:                    "default",
		CheckTag:                   "master",
	}

	s.updaters["default"].RegisterService(serviceConfig, service, s.config.ScriptsDirectory)
	s.config.SubscribeService(service, serviceConfig)

	checker := active.New(defaultSubscriptorConfig.kind,
		service, service, serviceConfig.Repository, notifyURL, serviceConfig,
		time.Minute*6, loggerPkg.NewLogger(service))
	s.registerChecker(service, checker)

	checker.SetInterval(defaultSubscriptorConfig.interval)
	go checker.Start()

}

// Unregister removes the given service from updater, so it won't look for new updates for it
func (s *Supervisor) Unregister(service string) error {
	s.unregisterChecker(service)
	serviceConfig, ok := s.config.Services[service]
	if !ok {
		return ErrServiceNotFound
	}

	s.updaters[serviceConfig.Updater].UnregisterService(service)
	return nil
}

// Start spawns the checkers and updaters
func (s *Supervisor) Start() {
	for _, checker := range s.activeCheckers {
		go checker.Start()
	}

	for _, checker := range s.passiveCheckers {
		go checker.Start()
	}
}

// Stop stops checkers and updaters
func (s *Supervisor) Stop() {
	for _, checker := range s.activeCheckers {
		checker.Stop()
	}

	for _, checker := range s.passiveCheckers {
		checker.Stop()
	}
}

// Update updates the given service
func (s *Supervisor) Update(service string) error {
	logger.Infof("services: %+v\n", s.config.Services)

	logger.Infof("updaters: %+v\n", s.updaters)

	// get updater
	serviceConfig, ok := s.config.Services[service]
	if !ok {
		return ErrServiceNotFound
	}

	updaterInstance := s.updaters[serviceConfig.Updater]

	// Try update
	err := <-updaterInstance.Update(service, serviceConfig.CheckTag, loggerPkg.NewLogger(service))
	if err != nil {
		return err
	}
	return nil
}

func (s *Supervisor) createUpdaters(conf *config.Configuration) {
	defaultUpdater := updater.New("custom", conf)
	s.updaters["default"] = defaultUpdater

	for name, c := range conf.Updaters {
		u := updater.New(c.Kind, conf)
		s.updaters[name] = u
	}
}

func (s *Supervisor) createCheckers(conf *config.Configuration) {
	for name, c := range conf.Services {
		if c.ActiveUpdateChecker != "" {
			s.registerActiveChecker(conf, c, name)
		} else {
			s.registerPassiveChecker(conf, c, name)
		}
	}
}

func (s *Supervisor) registerPassiveChecker(conf *config.Configuration, c config.ServiceConfig, name string) {
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

func (s *Supervisor) registerActiveChecker(conf *config.Configuration, c config.ServiceConfig, name string) {
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
	checkTimeout, err := time.ParseDuration(activeConfig.Interval)
	if err != nil {
		logrus.Fatalf("cannot parse check script timeout %s of active checker configuration %s. %s", activeConfig.Interval,
			c.ActiveUpdateChecker, err)
	}
	log := loggerPkg.NewLogger(name)
	fc, err := conf.FetcherConfig(c.ActiveUpdateChecker)
	if err != nil {
		panic(err)
	}
	fmt.Printf("service config: %+v\n", c)

	c.CheckScript = filepath.Join(conf.ScriptsDirectory, c.CheckScript)
	checker := active.New(fc.Kind, name, c.LocalName, c.Repository, fc.NotifyURL, c, checkTimeout, log)
	checker.SetInterval(interval)
	s.activeCheckers[name] = checker
}
