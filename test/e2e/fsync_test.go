//go:build e2e

package e2e_test

import "testing"

// TestFsyncTree lists the AFC media directory tree (lockdown/usbmux, no tunnel).
func TestFsyncTree(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "fsync", "tree", "--path=/") })
}
