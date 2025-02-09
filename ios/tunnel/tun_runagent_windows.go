//go:build windows

package tunnel

import (
	"syscall"
)

// For Windows OS, create and return *syscall.SysProcAttr by adding flag CREATE_NEW_PROCESS_GROUP for persistency
func createSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
