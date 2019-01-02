package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"sync"

	"errors"

	"gopkg.in/yaml.v2"
)

// Configuration errors
var (
	ErrNoConfigForThatService = errors.New("no configuration defined for given service name")
)

// Configuration represents an updater service configuration
type Configuration struct {
	ScriptsDirectory      string                      `yaml:"scripts_directory"`
	Port                  uint16                      `yaml:"port"`
	Updaters              map[string]UpdaterConfig    `yaml:"updaters"`
	ActiveUpdateCheckers  map[string]FetcherConfig    `yaml:"active_update_checkers"`
	PassiveUpdateCheckers map[string]SubscriberConfig `yaml:"passive_update_checkers"`
	Services              map[string]ServiceConfig    `yaml:"services"`
	sync.RWMutex
}

// UpdaterConfig represents an user-defined updater of one of the available kinds
type UpdaterConfig struct {
	Retries   int    `yaml:"retries"`
	RetryTime string `yaml:"retry_time"`
	Kind      string `yaml:"kind"`
}

// FetcherConfig represents an user-defined fetcher of one of the available kinds
type FetcherConfig struct {
	Kind      string `yaml:"kind"`
	NotifyURL string `yaml:"notify_url"`
}

// SubscriberConfig represents an user-defined subscriber of one of the available kinds
type SubscriberConfig struct {
	MessageBroker string   `yaml:"message-broker"`
	Topic         string   `yaml:"topic"`
	Urls          []string `yaml:"urls"`
	NotifyURL     string   `yaml:"notify_url"`
}

// ServiceConfig represents one of the services to be updated
type ServiceConfig struct {
	OfficialName               string   `yaml:"official_name"`
	LocalName                  string   `yaml:"local_name"`
	UpdateScript               string   `yaml:"update_script"`
	UpdateScriptTimeout        string   `yaml:"update_script_timeout"`
	UpdateScriptInterpreter    string   `yaml:"update_script_interpreter"`
	UpdateScriptExtraArguments []string `yaml:"update_script_extra_arguments"`
	CheckScript                string   `yaml:"check_script"`
	CheckScriptTimeout         string   `yaml:"check_script_timeout"`
	CheckScriptInterpreter     string   `yaml:"check_script_interpreter"`
	CheckScriptExtraArguments  []string `yaml:"check_script_extra_arguments"`
	ActiveUpdateChecker        string   `yaml:"active_update_checker"`
	PassiveUpdateChecker       string   `yaml:"passive_update_checker"`
	CheckTag                   string   `yaml:"check_tag"`
	Updater                    string   `yaml:"updater"`
	Repository                 string   `yaml:"repository"`
}

// New creates and returns a new configuration
func New() *Configuration {
	return &Configuration{
		ActiveUpdateCheckers:  make(map[string]FetcherConfig),
		PassiveUpdateCheckers: make(map[string]SubscriberConfig),
		Services:              make(map[string]ServiceConfig),
	}
}

// NewFromFile loads a configuration from the given yaml config filepath
func NewFromFile(path string) *Configuration {
	confPath := defaultPathIfNil(path)

	b, err := ioutil.ReadFile(confPath)
	if err != nil {
		panic(err)
	}

	conf := &Configuration{
		ActiveUpdateCheckers:  make(map[string]FetcherConfig),
		PassiveUpdateCheckers: make(map[string]SubscriberConfig),
		Services:              make(map[string]ServiceConfig),
	}

	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		panic(err)
	}

	conf.ScriptsDirectory = os.ExpandEnv(conf.ScriptsDirectory)

	return conf
}

// SubscribeService adds a new service config with the given name
func (c *Configuration) SubscribeService(name string, serviceConfig ServiceConfig) {
	c.Lock()
	defer c.Unlock()

	c.Services[name] = serviceConfig
}

// ServiceConfig returns the service config of the given service
func (c *Configuration) ServiceConfig(service string) (*ServiceConfig, error) {
	c.RLock()
	defer c.RUnlock()

	sc, ok := c.Services[service]
	if !ok {
		return nil, ErrNoConfigForThatService
	}

	return &sc, nil
}

// FetcherConfig returns the fetcher config of the given name
func (c *Configuration) FetcherConfig(name string) (*FetcherConfig, error) {
	c.RLock()
	defer c.RUnlock()

	ac, ok := c.ActiveUpdateCheckers[name]
	if !ok {
		return nil, ErrNoConfigForThatService
	}

	return &ac, nil
}

// UpdaterConfig returns the updater config of the given name
func (c *Configuration) UpdaterConfig(name string) (*UpdaterConfig, error) {
	c.RLock()
	defer c.RUnlock()

	uc, ok := c.Updaters[name]
	if !ok {
		return nil, ErrNoConfigForThatService
	}

	return &uc, nil
}

func defaultPathIfNil(path string) string {
	if path == "" {
		gopath := os.Getenv("$GOPATH")
		confPath := filepath.Join(gopath, "pkg", "github.com", "skycoin", "services", "updater",
			"configuration.yml")

		return confPath
	}
	return path
}
