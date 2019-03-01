package update

import (
	"context"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/skycoin/skycoin/src/util/logging"
)

// ExecuteScript executes the provided script and logs stdout.
func ExecuteScript(ctx context.Context, log *logging.Logger, cmd *exec.Cmd) (bool, error) {
	l := log.WithField("script", filepath.Base(cmd.Args[1]))

	l.Infof("START %v", cmd.Args)
	defer l.Infof("END %v", cmd.Args)

	// Prepare logging.
	cmd.Stdout = l.WithField("source", "stdout").Writer()
	cmd.Stderr = l.WithField("source", "stderr").Writer()

	// Set process group ID so the cmd and all its children become a new process
	// group. This allows Stop to SIGTERM the command's process group without
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
			l.Info("Context closed")
			// Signal the process group (-pid), not just the process, so that
			// the process and all its children are signaled. Else, child procs
			// can keep running and keep the stdout/stderr fd open and cause
			// cmd.Wait to hang.
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
				l.WithError(err).Error("syscall.Kill returned error")
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitCode(exitErr) != 1 {
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
