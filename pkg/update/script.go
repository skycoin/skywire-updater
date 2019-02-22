package update

import (
	"bufio"
	"context"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/skycoin/skycoin/src/util/logging"
)

// executes the provided script and logs stdout.
func executeScript(ctx context.Context, cmd *exec.Cmd, log *logging.Logger) (bool, error) {
	script := filepath.Base(cmd.Args[1])
	args   := cmd.Args

	log.Infof("(SCRIPT:%s) START %v", script, args)
	defer log.Infof("(SCRIPT:%s) END %v", script, args)

	// Prepare logging of stdout and stderr.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return false, err
	}

	// Set process group ID so the cmd and all its children become a new process
	// group. This allows Stop to SIGTERM the cmd's process group without
	// killing this process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Start command.
	if err := cmd.Start(); err != nil {
		return false, err
	}

	// Check ctx.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			// Signal the process group (-pid), not just the process, so that
			// the process and all its children are signaled. Else, child procs
			// can keep running and keep the stdout/stderr fd open and cause
			// cmd.Wait to hang.
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
				log.WithError(err).Errorf("(SCRIPT:%s) [ERROR] syscall.Kill returned error", script)
			}
		}
	}()

	// Log stdout and stderr.
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		for s := bufio.NewScanner(stdout); s.Scan(); {
			log.Infof("(SCRIPT:%s) [STDOUT] %s", script, s.Text())
		}
		wg.Done()
	}()
	go func() {
		for s := bufio.NewScanner(stderr); s.Scan(); {
			log.Infof("(SCRIPT:%s) [STDERR] %s", script, s.Text())
		}
		wg.Done()
	}()
	wg.Wait()

	if err = cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok && exitCode(exitErr) != 1 {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func exitCode(exitErr *exec.ExitError) int {
	if exitErr != nil {
		if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return waitStatus.ExitStatus()
		}
	}
	return 0
}
