//go:build e2e

package e2e_test

import "testing"

// TestCrashLs lists crash reports. Which reports exist is volatile, so assert
// the response shape (files + length) rather than contents.
func TestCrashLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{"files", "length"}, "crash", "ls")
	})
}
