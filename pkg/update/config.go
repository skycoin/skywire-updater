package update

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/skycoin/skywire-updater/internal/pathutil"
)

// Config represents an updater service configuration
type Config struct {
	Paths      PathsConfig      `yaml:"paths"`
	Interfaces InterfacesConfig `yaml:"interfaces"`
	Services   ServicesConfig   `yaml:"services"`
}

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

// ServicesConfig configures all the services.
type ServicesConfig struct {
	Defaults ServiceDefaultsConfig     `yaml:"defaults"`
	Services map[string]*ServiceConfig `yaml:"services"`
}

// ServiceDefaultsConfig is the configuration that is shared across all services (as default).
type ServiceDefaultsConfig struct {
	MainBranch  string   `yaml:"main-branch"`
	BinDir      string   `yaml:"bin-dir"`
	Interpreter string   `yaml:"interpreter"`
	Envs        []string `yaml:"envs"`
}

// ServiceConfig represents one of the services to be updated.
type ServiceConfig struct {
	Repo        string        `yaml:"repo,omitempty"`
	MainBranch  string        `yaml:"main-branch,omitempty"`
	MainProcess string        `yaml:"main-process"`
	BinDir      string        `yaml:"bin-dir,omitempty"`
	Checker     CheckerConfig `yaml:"checker"`
	Updater     UpdaterConfig `yaml:"updater"`
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

// NewConfig returns a config with default values (from the provided root
// directory and bin directory).
func NewConfig(rootDir, binDir string) *Config {
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
		Services: ServicesConfig{
			Defaults: ServiceDefaultsConfig{
				MainBranch:  "master",
				BinDir:      binDir,
				Interpreter: "/bin/bash",
				Envs:        []string{},
			},
			Services: make(map[string]*ServiceConfig),
		},
	}
}

// NewLocalConfig returns a config with default values, suitable for use by all
// users of a system.
func NewLocalConfig() *Config {
	var (
		rootDir = "/usr/local/skycoin/skywire-updater"
		binDir  = "/usr/local/skycoin/bin"
	)
	return NewConfig(rootDir, binDir)
}

// NewHomeConfig returns a config with default values, suitable for use by the
// local user.
func NewHomeConfig() *Config {
	var (
		homeDir = pathutil.HomeDir()
		rootDir = filepath.Join(homeDir, ".skycoin/skywire-updater")
		binDir  = filepath.Join(homeDir, ".skycoin/bin")
	)
	return NewConfig(rootDir, binDir)
}

// Parse parses the config from a given yaml file path.
func (c *Config) Parse(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(raw, c); err != nil {
		return err
	}
	for name, srv := range c.Services.Services {
		if err := processServiceConfig(srv, c.Paths.ScriptsPath, &c.Services.Defaults); err != nil {
			return fmt.Errorf("invalid service %s: %s", name, err.Error())
		}
	}
	{
		out, err := yaml.Marshal(c)
		if err != nil {
			log.WithError(err).Fatal()
		}
		log.Printf("Parsed Configuration:\n%s", string(out))
	}
	return nil
}

// Checks for errors and fills unspecified fields with default values.
func processServiceConfig(sc *ServiceConfig, scriptsPath string, d *ServiceDefaultsConfig) error {
	if sc.BinDir == "" {
		sc.BinDir = d.BinDir
	}
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
			if _, err := os.Stat(sc.Checker.Script); err != nil {
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
			if _, err := os.Stat(sc.Updater.Script); err != nil {
				return fmt.Errorf("updater.script cannot be accessed: %s", err.Error())
			}
		}
	}
	return nil
}
