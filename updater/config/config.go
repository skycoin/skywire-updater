package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"sync"
	"github.com/pkg/errors"
)

var (
	ErrNoConfigForThatService = errors.New("no configuration defined for given service name")
)

type Configuration struct {
	ScriptsDirectory string `yaml:"scripts_directory"`
	Port				  uint16 `yaml:"port"`
	Updaters              map[string]UpdaterConfig    `yaml:"updaters"`
	ActiveUpdateCheckers  map[string]FetcherConfig    `yaml:"active_update_checkers"`
	PassiveUpdateCheckers map[string]SubscriberConfig `yaml:"passive_update_checkers"`
	Services              map[string]ServiceConfig    `yaml:"services"`
	sync.RWMutex
}

type UpdaterConfig struct {
	Retries   int    `yaml:"retries"`
	RetryTime string `yaml:"retry_time"`
	Kind string `yaml:"kind"`
}

type FetcherConfig struct {
	Interval  string `yaml:"interval"`
	Kind      string `yaml:"kind"`
	NotifyUrl string `yaml:"notify_url"`
}

type SubscriberConfig struct {
	MessageBroker string   `yaml:"message-broker"`
	Topic         string   `yaml:"topic"`
	Urls          []string `yaml:"urls"`
	NotifyUrl string `yaml:"notify_url"`
}

type ServiceConfig struct {
	OfficialName         string   `yaml:"official_name"`
	LocalName            string   `yaml:"local_name"`
	UpdateScript         string   `yaml:"update_script"`
	UpdateScriptTimeout        string   `yaml:"update_script_timeout"`
	UpdateScriptInterpreter    string   `yaml:"update_script_interpreter"`
	UpdateScriptExtraArguments []string `yaml:"update_script_extra_arguments"`
	CheckScript         string   `yaml:"check_script"`
	CheckScriptTimeout        string   `yaml:"check_script_timeout"`
	CheckScriptInterpreter    string   `yaml:"check_script_interpreter"`
	CheckScriptExtraArguments []string `yaml:"check_script_extra_arguments"`
	ActiveUpdateChecker  string   `yaml:"active_update_checker"`
	PassiveUpdateChecker string   `yaml:"passive_update_checker"`
	CheckTag             string   `yaml:"check_tag"`
	Updater              string   `yaml:"updater"`
	Repository           string   `yaml:"repository"`
}

func New() *Configuration {
	return &Configuration{
		ActiveUpdateCheckers:  make(map[string]FetcherConfig),
		PassiveUpdateCheckers: make(map[string]SubscriberConfig),
		Services:              make(map[string]ServiceConfig),
	}
}

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

	return conf
}

func (c *Configuration) SubscribeService(name string, serviceConfig ServiceConfig) {
	c.Lock()
	defer c.Unlock()

	c.Services[name] = serviceConfig
}

func (c *Configuration) ServiceConfig(service string) (*ServiceConfig, error) {
	c.RLock()
	defer c.RUnlock()

	sc, ok := c.Services[service]
	if !ok {
		return nil, ErrNoConfigForThatService
	}

	return &sc, nil
}


func (c *Configuration) FetcherConfig(name string) (*FetcherConfig, error) {
	c.RLock()
	defer c.RUnlock()

	ac, ok := c.ActiveUpdateCheckers[name]
	if !ok {
		return nil, ErrNoConfigForThatService
	}

	return &ac, nil
}

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
