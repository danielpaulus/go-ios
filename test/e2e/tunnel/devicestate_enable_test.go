//go:build e2e

package tunnel_test

import (
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestDevicestateEnable activates a device condition profile and verifies the
// command reports it active, then stops it. devicestate enable holds the
// condition only while running, so SIGTERM (via startBackground's stop) reverts
// the device — the test is self-restoring.
//
// Verification reads the enable command's own output rather than issuing a
// concurrent "devicestate list": go-ios does not handle two simultaneous
// RSD/tunnel sessions reliably.
func TestDevicestateEnable(t *testing.T) {
	const typeID, profileID = "SlowNetworkCondition", "SlowNetwork3GGood"
	forEachDevice(t, func(t *testing.T, udid string) {
		output, stop := startBackground(t, udid, syscall.SIGTERM, "devicestate", "enable", typeID, profileID)
		defer stop()

		// Wait for the command to report the profile active.
		var out string
		for i := 0; i < 20; i++ {
			out = output()
			if strings.Contains(out, "is active") {
				return
			}
			time.Sleep(time.Second)
		}
		t.Fatalf("devicestate enable %s %s: never reported active. output:\n%s", typeID, profileID, out)
	})
}
