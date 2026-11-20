//go:build !windows

package remote

import (
	"os/exec"
	"syscall"
	"time"
)

// configureRunnerProcAttr puts the runner in its own process group so a SIGINT
// delivered to `ios remote`'s group (Ctrl-C) does not race our own orderly
// shutdown, and so terminateRunnerProc can signal the whole group.
func configureRunnerProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// terminateRunnerProc asks the runner to stop cleanly (SIGTERM to its process
// group), then hard-kills the group after a grace period if it is still alive.
// It never calls cmd.Wait — the supervisor owns the single Wait that reaps the
// child; here we only signal. Signaling the group also reaps helper processes
// the runner spawned (e.g. the port forward).
func terminateRunnerProc(cmd *exec.Cmd) {
	pid := cmd.Process.Pid
	// Negative pid targets the process group established via Setpgid.
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	// Escalate asynchronously so Stop stays non-blocking; signal 0 probes
	// liveness without affecting the process.
	go func() {
		time.Sleep(5 * time.Second)
		if syscall.Kill(-pid, syscall.Signal(0)) == nil {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
	}()
}
