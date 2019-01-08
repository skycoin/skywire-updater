package updater

import (
	"fmt"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"

	"path/filepath"
	"sync"

	"errors"

	"github.com/watercompany/skywire-updater/pkg/config"
	"github.com/watercompany/skywire-updater/pkg/logger"
)

// This package implements a custom updater. This means, a script that would be launched upon
// update notify. Two arguments would always be passed to the script: Name of the service + version.

var defaultScriptTimeout = time.Minute * 10

// Custom related errors
var (
	ErrNoServiceWithThatName = errors.New("no service registered with that name")
)

// Custom is an updater that runs an user-given custom script in order to update services
type Custom struct {
	services map[string]customServiceConfig
	sync.RWMutex
}

type customServiceConfig struct {
	officialName         string
	localName            string
	scriptInterpreter    string
	scriptExtraArguments []string
	updateScript         string
	tag                  string
	scriptTimeout        time.Duration
}

func newCustomUpdater(scriptsDirectory string, services map[string]config.ServiceConfig) *Custom {
	customServices := make(map[string]customServiceConfig)
	custom := &Custom{services: customServices}

	for officialName, c := range services {
		custom.RegisterService(c, officialName, scriptsDirectory)
	}

	return &Custom{
		services: customServices,
	}
}

// RegisterService allows to register a new service to update
func (c *Custom) RegisterService(conf config.ServiceConfig, officialName, scriptsDirectory string) {
	c.Lock()
	defer c.Unlock()

	customService := c.parseServiceConfig(conf, officialName, scriptsDirectory)
	c.services[officialName] = customService
}

// UnregisterService allows to remove a service to update
func (c *Custom) UnregisterService(officialName string) {
	c.Lock()
	defer c.Unlock()

	delete(c.services, officialName)
}

func (c *Custom) service(officialName string) (customServiceConfig, bool) {
	c.RLock()
	defer c.RUnlock()

	serviceConfig, ok := c.services[officialName]
	return serviceConfig, ok
}

func (c *Custom) parseServiceConfig(conf config.ServiceConfig, officialName, scriptsDirectory string) customServiceConfig {
	duration, err := time.ParseDuration(conf.UpdateScriptTimeout)
	if err != nil {
		duration = defaultScriptTimeout
		logrus.Warnf("cannot parse timeout duration %s of service %s configuration."+
			" setting default timeout %s", conf.UpdateScriptTimeout, conf.OfficialName, duration.String())
	}
	return customServiceConfig{
		officialName:         officialName,
		localName:            conf.LocalName,
		scriptExtraArguments: conf.UpdateScriptExtraArguments,
		scriptInterpreter:    conf.UpdateScriptInterpreter,
		scriptTimeout:        duration,
		tag:                  conf.CheckTag,
		updateScript:         filepath.Join(scriptsDirectory, conf.UpdateScript),
	}
}

// Update updates the given service
func (c *Custom) Update(service, version string, log *logger.Logger) chan error {
	errCh := make(chan error)
	localService, ok := c.service(service)
	if !ok {
		errCh <- ErrNoServiceWithThatName
		return errCh
	}

	customCmd, statusChan := createAndLaunch(localService, version, log)
	ticker := time.NewTicker(time.Second * 2)

	go logStdout(ticker, customCmd, log)

	go timeoutCmd(localService, customCmd, errCh)

	go waitForExit(statusChan, errCh, log)

	return errCh
}

func createAndLaunch(service customServiceConfig, version string, log *logger.Logger) (*cmd.Cmd, <-chan cmd.Status) {
	command := buildCommand(service, version)
	log.Info("running command: ", command)
	customCmd := cmd.NewCmd(service.scriptInterpreter, command...)
	statusChan := customCmd.Start()
	return customCmd, statusChan
}

func buildCommand(service customServiceConfig, version string) []string {
	command := []string{
		service.updateScript,
		service.localName,
		version,
	}
	return append(command, service.scriptExtraArguments...)
}

func logStdout(ticker *time.Ticker, customCmd *cmd.Cmd, log *logger.Logger) {
	var previousLastLine int

	for range ticker.C {
		status := customCmd.Status()
		currentLastLine := len(status.Stdout)

		if currentLastLine != previousLastLine {
			for _, line := range status.Stdout[previousLastLine:] {
				log.Infof("script stdout: %s", line)
			}
			previousLastLine = currentLastLine
		}

	}
}

func timeoutCmd(service customServiceConfig, customCmd *cmd.Cmd, errCh chan error) {
	<-time.After(service.scriptTimeout)
	err := customCmd.Stop()
	if err != nil {
		errCh <- err
		return
	}
	errCh <- fmt.Errorf("update script for service %s timed out", service.officialName)
}

func waitForExit(statusChan <-chan cmd.Status, errCh chan error, log *logger.Logger) {
	finalStatus := <-statusChan
	log.Infof("%s exit with: %d", finalStatus.Cmd, finalStatus.Exit)
	if finalStatus.Exit != 0 {
		errCh <- fmt.Errorf("exit with non-zero status %d", finalStatus.Exit)
	}
	errCh <- nil
}
