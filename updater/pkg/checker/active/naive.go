package active

import (
	"time"

	"fmt"
	"path/filepath"

	"github.com/go-cmd/cmd"
	"github.com/watercompany/skywire-services/updater/pkg/checker"
	"github.com/watercompany/skywire-services/updater/pkg/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
)

type naive struct {
	// URL should be in the format /:owner/:Repository
	service              string
	localName            string
	url                  string
	exit                 chan int
	notifyURL            string
	updateCheckScript    string
	scriptTimeout        time.Duration
	scriptExtraArguments []string
	scriptInterpreter    string
	log                  *logger.Logger
	config.CustomLock
}

func newNaive(service, localName, url, notifyURL, scriptInterpreter, updateCheckScript string,
	scriptExtraArguments []string, scriptTimeout time.Duration, log *logger.Logger) *naive {
	return &naive{
		url:                  filepath.Join("github.com", url),
		exit:                 make(chan int),
		service:              service,
		localName:            localName,
		notifyURL:            notifyURL,
		updateCheckScript:    updateCheckScript,
		scriptTimeout:        scriptTimeout,
		scriptExtraArguments: scriptExtraArguments,
		scriptInterpreter:    scriptInterpreter,
		log:                  log,
	}
}

func (n *naive) Check() error {
	return n.checkIfNew()
}

func (n *naive) checkIfNew() error {
	n.log.Info("checking update...")

	isUpdate, err := n.checkIfUpdate()
	if err != nil {
		return err
	}
	if isUpdate {
		err = checker.NotifyUpdate(n.notifyURL, n.service, "master", "master", "token")
		if err != nil {
			return err
		}
	} else {
		return ErrNoNewVersion
	}
	return nil
}

func (n *naive) checkIfUpdate() (bool, error) {
	var errCh = make(chan error)

	customCmd, statusChan := createAndLaunch(n.localName, "master",
		n.scriptInterpreter, n.updateCheckScript, n.service, n.url, n.scriptExtraArguments, n.log)
	ticker := time.NewTicker(time.Second * 2)

	go logStdout(ticker, customCmd, n.log)

	go timeoutCmd(n.service, n.scriptTimeout, customCmd, errCh)

	return waitForExit(statusChan, errCh)
}

func createAndLaunch(localName, version, scriptInterpreter, script, service, url string, arguments []string, log *logger.Logger) (*cmd.Cmd, <-chan cmd.Status) {
	command := buildCommand(localName, version, script, service, url, arguments)
	log.Info("running command: ", command)
	customCmd := cmd.NewCmd(scriptInterpreter, command...)
	statusChan := customCmd.Start()
	return customCmd, statusChan
}

func buildCommand(localName, version, script, service, url string, arguments []string) []string {
	command := []string{
		script,
		localName,
		version,
		service,
		url,
	}
	return append(command, arguments...)
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

func timeoutCmd(service string, timeout time.Duration, customCmd *cmd.Cmd, errCh chan error) {
	<-time.After(timeout)
	err := customCmd.Stop()
	if err != nil {
		errCh <- err
		return
	}
	errCh <- fmt.Errorf("update script for service %s timed out", service)
}

func waitForExit(statusChan <-chan cmd.Status, errCh chan error) (bool, error) {
	for {
		select {
		case finalStatus := <-statusChan:
			if finalStatus.Exit != 0 && finalStatus.Exit != 1 {
				return false, fmt.Errorf("exit with non-zero status %d", finalStatus.Exit)
			}

			if finalStatus.Exit == 1 {
				return false, nil
			}
			return true, nil
		case err := <-errCh:
			return false, err
		}
	}
}
