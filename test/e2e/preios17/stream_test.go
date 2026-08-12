//go:build e2e

package preios17_test

import (
	"testing"
	"time"
)

// TestSyslog streams the device syslog (lockdown syslog_relay, no tunnel) and
// asserts output is produced within the window. It uses the resilient stream
// helper because a transient usbmuxd drop on the flaky pre-iOS17 device
// otherwise leaves the stream empty and red-walls the run.
func TestSyslog(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		streamSmokeResilient(t, udid, 8*time.Second, "syslog")
	})
}

// TestOstrace streams os_trace_relay logs and asserts output within the window.
func TestOstrace(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		streamSmoke(t, udid, 8*time.Second, "ostrace")
	})
}
