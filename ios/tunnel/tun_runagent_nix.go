//go:build !windows

package tunnel

import (
	"syscall"
)

// For *nix OS, create and return *syscall.SysProcAttr with Setsid set to true for running go-ios agent in a new session for persistency.
func createSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
