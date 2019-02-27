package update

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const (

	// EnvRepo can be used by a script to determine the repository URL of the
	// service.
	EnvRepo = "SKYUPD_REPO"

	// EnvToVersion can be used by an updater script to determine the version
	// to update the service to.
	EnvToVersion = "SKYUPD_TO_VERSION"
)

func cmdEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// Config represents an updater service configuration
type Config struct {
	Services map[string]*ServiceConfig `yaml:"services"`
}

// ServiceConfig represents one of the services to be updated
type ServiceConfig struct {
	Repo    string        `yaml:"repo"`
	Checker CheckerConfig `yaml:"checker"`
	Updater UpdaterConfig `yaml:"updater"`
}

func (sc *ServiceConfig) fillDefaults() {
	if sc.Repo != "" {
		if sc.Checker.Type == "" {
			sc.Checker.Type = ScriptCheckerType
		}
		if sc.Checker.Type == ScriptCheckerType && sc.Checker.Interpreter == "" {
			sc.Checker.Interpreter = "/bin/bash"
		}
		if sc.Updater.Type == "" {
			sc.Updater.Type = ScriptUpdaterType
		}
		if sc.Updater.Type == ScriptUpdaterType && sc.Updater.Interpreter == "" {
			sc.Updater.Interpreter = "/bin/bash"
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

// Envs ...
func (sc ServiceConfig) Envs() []string {
	return append(os.Environ(), []string{
		cmdEnv(EnvRepo, sc.Repo),
	}...)
}

// CheckerEnvs ...
func (sc ServiceConfig) CheckerEnvs() []string {
	return append(sc.Envs(), sc.Checker.Envs...)
}

// UpdaterEnvs ...
func (sc ServiceConfig) UpdaterEnvs() []string {
	return append(sc.Envs(), sc.Updater.Envs...)
}

// CheckerConfig ...
type CheckerConfig struct {
	Type CheckerType `yaml:"type"`

	// script checker fields:
	Interpreter string   `yaml:"interpreter"`
	Script      string   `yaml:"script"`
	Args        []string `yaml:"args"`
	Envs        []string `yaml:"envs"`
}

// UpdaterConfig ...
type UpdaterConfig struct {
	Type UpdaterType `json:"type"`

	// script updater fields:
	Interpreter string   `yaml:"interpreter"`
	Script      string   `yaml:"script"`
	Args        []string `yaml:"args"`
	Envs        []string `yaml:"envs"`
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
		srv.fillDefaults()
		if err := srv.validate(); err != nil {
			return nil, fmt.Errorf("invalid service %s: %s", name, err.Error())
		}
	}

	return conf, nil
}
