//go:build windows

package remote

import "os/exec"

// configureRunnerProcAttr is a no-op on Windows: there is no POSIX process group
// to establish. CommandContext still terminates the child on ctx cancel.
func configureRunnerProcAttr(cmd *exec.Cmd) {}

// terminateRunnerProc kills the runner. Windows has no SIGTERM; Process.Kill is
// the orderly stop. It does not call Wait (the supervisor owns the single Wait).
func terminateRunnerProc(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
