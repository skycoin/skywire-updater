package active

import (
	"sync"
	"time"

	"github.com/watercompany/skywire-services/updater/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/pkg/checker"
	"github.com/go-cmd/cmd"
	"fmt"
	"path/filepath"
)

type naive struct {
	// Url should be in the format /:owner/:Repository
	service   string
	localName string
	url       string
	interval  time.Duration
	ticker    *time.Ticker
	lock      sync.Mutex
	exit      chan int
	notifyUrl string
	updateCheckScript string
	scriptTimeout time.Duration
	scriptExtraArguments []string
	scriptInterpreter string
	log       *logger.Logger
	config.CustomLock
}

func NewNaive(service, localName, url, notifyUrl, scriptInterpreter, updateCheckScript string,
	scriptExtraArguments []string, scriptTimeout time.Duration, log *logger.Logger) *naive {
	return &naive{
		url:       filepath.Join("github.com", url),
		exit:      make(chan int),
		service:   service,
		localName: localName,
		notifyUrl:	notifyUrl,
		updateCheckScript: updateCheckScript,
		scriptTimeout: scriptTimeout,
		scriptExtraArguments: scriptExtraArguments,
		scriptInterpreter: scriptInterpreter,
		log:       log,
	}
}

func (n *naive) SetInterval(t time.Duration) {
	n.interval = t

	n.lock.Lock()
	if n.ticker != nil {
		n.ticker = time.NewTicker(n.interval)
	}
	n.lock.Unlock()
}

func (n *naive) Start() {
	n.ticker = time.NewTicker(n.interval)
	go func() {
		for {
			select {
			case t := <-n.ticker.C:
				n.log.Info("looking for new version at: ", t)
				// Try to fetch new version
				go n.checkIfNew()
			}
		}
	}()
	<-n.exit
}

func (n *naive) Stop() {
	n.ticker.Stop()
	n.exit <- 1
}

func (n *naive) checkIfNew() {
	if n.IsLock() {
		n.log.Warnf("service %s is already being updated... waiting for it to finish", n.service)
	}
	n.Lock()
	defer n.Unlock()

	n.log.Info("updating...")

	isUpdate, err := n.checkIfUpdate()
	if err != nil {
		n.log.Error(err)
	}
	if isUpdate {
		err = checker.NotifyUpdate(n.notifyUrl, n.service, "master", "master", "token")
		if err != nil {
			n.log.Error(err)
		}
	} else {
		n.log.Info("up to date")
	}
}

func (n *naive) checkIfUpdate() (bool, error) {
	var errCh = make(chan error)

	customCmd, statusChan := createAndLaunch(n.localName, "master",
		n.scriptInterpreter, n.updateCheckScript, n.service, n.url,  n.scriptExtraArguments, n.log)
	ticker := time.NewTicker(time.Second * 2)

	go logStdout(ticker, customCmd, n.log)

	go timeoutCmd(n.service, n.scriptTimeout, customCmd, errCh)

	return waitForExit(statusChan, errCh, n.log)
}

func createAndLaunch(localName, version, scriptInterpreter, script, service, url string, arguments []string, log *logger.Logger) (*cmd.Cmd, <-chan cmd.Status) {
	command := buildCommand(localName, version, script, service, url, arguments)
	log.Info("running command: ", command)
	customCmd := cmd.NewCmd(scriptInterpreter, command...)
	statusChan := customCmd.Start()
	return customCmd, statusChan
}

func buildCommand(localName, version,  script, service ,url string, arguments []string) []string {
	fmt.Println("localName: ", localName)
	fmt.Println("service: ", service)
	fmt.Println("url: ", url)
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
	customCmd.Stop()
	errCh <- fmt.Errorf("update script for service %s timed out", service)
}

func waitForExit(statusChan <-chan cmd.Status, errCh chan error, log *logger.Logger) (bool, error) {
	for {
		select {
		case finalStatus := <-statusChan:
			log.Infof("%s exit with: %d", finalStatus.Cmd, finalStatus.Exit)
			if finalStatus.Exit != 0 {
				return false, fmt.Errorf("exit with non-zero status %d", finalStatus.Exit)
			}
			return true, nil
		case err := <- errCh:
			return false, err
		}
	}
}
