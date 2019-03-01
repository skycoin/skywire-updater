package update

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// DefaultConfig is the configuration that is shared across all services (as default).
type DefaultConfig struct {
	MainBranch  string   `yaml:"main-branch"`
	Interpreter string   `yaml:"interpreter"`
	Envs        []string `yaml:"envs"`
}

// ServiceConfig represents one of the services to be updated
type ServiceConfig struct {
	Repo        string        `yaml:"repo"`
	MainBranch  string        `yaml:"main-branch"`
	MainProcess string        `yaml:"main-process"`
	Checker     CheckerConfig `yaml:"checker"`
	Updater     UpdaterConfig `yaml:"updater"`
}

// CheckerConfig is the configuration for a service's checker.
type CheckerConfig struct {
	Type CheckerType `yaml:"type"`

	// script checker fields:
	Interpreter string   `yaml:"interpreter"`
	Script      string   `yaml:"script"`
	Args        []string `yaml:"args"`
	Envs        []string `yaml:"envs"`
}

// UpdaterConfig is the configuration for a service's updater.
type UpdaterConfig struct {
	Type UpdaterType `json:"type"`

	// script updater fields:
	Interpreter string   `yaml:"interpreter"`
	Script      string   `yaml:"script"`
	Args        []string `yaml:"args"`
	Envs        []string `yaml:"envs"`
}

// Config represents an updater service configuration
type Config struct {
	Default  DefaultConfig             `yaml:"default"`
	Services map[string]*ServiceConfig `yaml:"services"`
}

// NewConfig loads a configuration from the given yaml config filepath
func NewConfig(path string) (*Config, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &Config{Services: make(map[string]*ServiceConfig)}

	if err := yaml.Unmarshal(f, &conf); err != nil {
		return nil, err
	}

	for name, srv := range conf.Services {
		srv.fillDefaults(&conf.Default)
		if err := srv.validate(); err != nil {
			return nil, fmt.Errorf("invalid service %s: %s", name, err.Error())
		}
	}

	return conf, nil
}

func (sc *ServiceConfig) fillDefaults(d *DefaultConfig) {
	if sc.Repo != "" {
		if sc.MainBranch == "" {
			if d.MainBranch != "" {
				sc.MainBranch = d.MainBranch
			} else {
				sc.MainBranch = "master"
			}
		}
		if sc.Checker.Type == "" {
			sc.Checker.Type = ScriptCheckerType
		}
		if sc.Checker.Type == ScriptCheckerType && sc.Checker.Interpreter == "" {
			if d.Interpreter != "" {
				sc.Checker.Interpreter = d.Interpreter
			} else {
				sc.Checker.Interpreter = "/bin/bash"
			}
		}
		if sc.Updater.Type == "" {
			sc.Updater.Type = ScriptUpdaterType
		}
		if sc.Updater.Type == ScriptUpdaterType && sc.Updater.Interpreter == "" {
			if d.Interpreter != "" {
				sc.Updater.Interpreter = d.Interpreter
			} else {
				sc.Updater.Interpreter = "/bin/bash"
			}
		}
	}
}

func (sc *ServiceConfig) validate() error {
	if sc.Checker.Type == ScriptCheckerType && sc.Checker.Script == "" {
		return errors.New("checker.script needs to be defined")
	}
	if sc.Updater.Type == ScriptUpdaterType && sc.Updater.Script == "" {
		return errors.New("checker.script needs to be defined")
	}
	return nil
}
