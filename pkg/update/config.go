package update

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	yaml "gopkg.in/yaml.v2"
)

// PathsConfig configures the paths for the updater.
type PathsConfig struct {
	DBFile      string `yaml:"db-file"`
	ScriptsPath string `yaml:"scripts-path"`
}

// InterfacesConfig configures the http interface for the updater.
type InterfacesConfig struct {
	Addr       string `yaml:"addr"`
	EnableREST bool   `yaml:"enable-rest"`
	EnableRPC  bool   `yaml:"enable-rpc"`
}

// DefaultsConfig is the configuration that is shared across all services (as default).
type DefaultsConfig struct {
	MainBranch  string   `yaml:"main-branch"`
	Interpreter string   `yaml:"interpreter"`
	Envs        []string `yaml:"envs"`
}

// CheckerConfig is the configuration for a service's checker.
type CheckerConfig struct {
	Type CheckerType `yaml:"type"`

	// script checker fields:
	Interpreter string   `yaml:"interpreter,omitempty"`
	Script      string   `yaml:"script,omitempty"`
	Args        []string `yaml:"args,omitempty"`
	Envs        []string `yaml:"envs,omitempty"`
}

// UpdaterConfig is the configuration for a service's updater.
type UpdaterConfig struct {
	Type UpdaterType `yaml:"type"`

	// script updater fields:
	Interpreter string   `yaml:"interpreter,omitempty"`
	Script      string   `yaml:"script,omitempty"`
	Args        []string `yaml:"args,omitempty"`
	Envs        []string `yaml:"envs,omitempty"`
}

// ServiceConfig represents one of the services to be updated
type ServiceConfig struct {
	Repo        string        `yaml:"repo,omitempty"`
	MainBranch  string        `yaml:"main-branch,omitempty"`
	MainProcess string        `yaml:"main-process"`
	Checker     CheckerConfig `yaml:"checker"`
	Updater     UpdaterConfig `yaml:"updater"`
}

func (sc *ServiceConfig) process(scriptsPath string, d *DefaultsConfig) error {
	if sc.Repo != "" {
		if sc.MainBranch == "" {
			sc.MainBranch = d.MainBranch
		}
		if sc.Checker.Type == "" {
			sc.Checker.Type = ScriptCheckerType
		}
		if sc.Checker.Type == ScriptCheckerType {
			if sc.Checker.Interpreter == "" {
				sc.Checker.Interpreter = d.Interpreter
			}
			if sc.Checker.Script == "" {
				return errors.New("checker.script needs to be defined")
			}
			if scriptsPath != "" {
				sc.Checker.Script = filepath.Join(scriptsPath, sc.Checker.Script)
			}
			if err := scriptOK(sc.Checker.Script); err != nil {
				return fmt.Errorf("checker.script cannot be accessed: %s", err.Error())
			}
		}
		if sc.Updater.Type == "" {
			sc.Updater.Type = ScriptUpdaterType
		}
		if sc.Updater.Type == ScriptUpdaterType {
			if sc.Updater.Interpreter == "" {
				sc.Updater.Interpreter = d.Interpreter
			}
			if sc.Updater.Script == "" {
				return errors.New("updater.script needs to be defined")
			}
			if scriptsPath != "" {
				sc.Updater.Script = filepath.Join(scriptsPath, sc.Updater.Script)
			}
			if err := scriptOK(sc.Updater.Script); err != nil {
				return fmt.Errorf("updater.script cannot be accessed: %s", err.Error())
			}
		}
	}
	return nil
}

// Config represents an updater service configuration
type Config struct {
	Paths      PathsConfig               `yaml:"paths"`
	Interfaces InterfacesConfig          `yaml:"interfaces"`
	Defaults   DefaultsConfig            `yaml:"defaults"`
	Services   map[string]*ServiceConfig `yaml:"services"`
}

// NewConfig returns a config with default values.
func NewConfig() *Config {
	rootDir := filepath.Join(userHomeDir(), ".skywire/updater")
	return &Config{
		Paths: PathsConfig{
			DBFile:      filepath.Join(rootDir, "db.json"),
			ScriptsPath: filepath.Join(rootDir, "scripts"),
		},
		Interfaces: InterfacesConfig{
			Addr:       ":7280",
			EnableREST: true,
			EnableRPC:  true,
		},
		Defaults: DefaultsConfig{
			MainBranch:  "master",
			Interpreter: "/bin/bash",
			Envs:        []string{},
		},
		Services: make(map[string]*ServiceConfig),
	}
}

// ParseConfig loads a configuration from the given json config filepath
func ParseConfig(path string) (*Config, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := NewConfig()
	if err := yaml.Unmarshal(f, &conf); err != nil {
		return nil, err
	}
	for name, srv := range conf.Services {
		if err := srv.process(conf.Paths.ScriptsPath, &conf.Defaults); err != nil {
			return nil, fmt.Errorf("invalid service %s: %s", name, err.Error())
		}
	}
	{
		out, err := yaml.Marshal(conf)
		if err != nil {
			log.WithError(err).Fatal()
		}
		log.Printf("Parsed Configuration:\n%s", string(out))
	}

	return conf, nil
}

// SRC: https://github.com/spf13/viper/blob/80ab6657f9ec7e5761f6603320d3d58dfe6970f6/util.go#L144-L153
func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// determines if the script can be accessed.
func scriptOK(name string) error {
	_, err := os.Stat(name)
	return err
}
