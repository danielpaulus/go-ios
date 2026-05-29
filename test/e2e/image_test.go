//go:build e2e

package e2e_test

import "testing"

// TestImageList lists mounted developer disk image signatures. The output is
// legitimately empty when no image is mounted, so this only asserts the
// command succeeds (runIOSForDevice fails on a non-zero exit).
func TestImageList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { runIOSForDevice(t, udid, "image", "list") })
}
