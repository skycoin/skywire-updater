package update

import (
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

func (sc *ServiceConfig) FillDefaults() error {
	if sc.Repo != "" {
		if sc.Checker.Type == "" {
			sc.Checker.Type = ScriptCheckerType
		}
		if sc.Checker.Interpreter == "" {
			sc.Checker.Interpreter = "/bin/bash"
		}
		if sc.Checker.Script == "" {
			sc.Checker.Script = "check_generic"
		}
		if sc.Updater.Type == "" {
			sc.Updater.Type = ScriptUpdaterType
		}
		if sc.Updater.Interpreter == "" {
			sc.Updater.Interpreter = "/bin/bash"
		}
		if sc.Updater.Script == "" {
			sc.Updater.Script = "update_generic"
		}
	}
	return nil
}

func (sc ServiceConfig) Envs() []string {
	return append(os.Environ(), []string{
		cmdEnv(EnvRepo, sc.Repo),
	}...)
}

func (sc ServiceConfig) CheckerEnvs() []string {
	return append(sc.Envs(), sc.Checker.Envs...)
}

func (sc ServiceConfig) UpdaterEnvs() []string {
	return append(sc.Envs(), sc.Updater.Envs...)
}

type CheckerConfig struct {
	Type CheckerType `yaml:"type"`

	// script checker fields:
	Interpreter string   `yaml:"interpreter"`
	Script      string   `yaml:"script"`
	Args        []string `yaml:"args"`
	Envs        []string `yaml:"envs"`
}

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
		if err := srv.FillDefaults(); err != nil {
			return nil, fmt.Errorf("invalid service %s: %s", name, err.Error())
		}
	}

	return conf, nil
}
