//go:build e2e

package tunnel_test

import (
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestSetLocation simulates a device location and verifies it took effect by
// observing the device's own logs: while a location is simulated, locationd logs
// under subsystem com.apple.locationd.Core / category "Simulation". There is no
// "get location" command, so this independent log signal is the verification.
//
// setlocation (iOS 17+ RSD path) holds the simulation until it receives SIGINT,
// at which point it reverts — so the test is self-restoring.
func TestSetLocation(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		_, stopLoc := startBackground(t, udid, syscall.SIGINT, "setlocation", "--lat=51.5", "--lon=-0.12")
		defer stopLoc()
		time.Sleep(2 * time.Second) // let the simulation start

		logs, stopLogs := startBackground(t, udid, syscall.SIGTERM, "ostrace")
		defer stopLogs()

		for i := 0; i < 20; i++ {
			out := logs()
			if strings.Contains(out, "com.apple.locationd.Core") && strings.Contains(out, `"category":"Simulation"`) {
				return
			}
			time.Sleep(time.Second)
		}
		t.Fatalf("setlocation: no locationd Simulation activity observed in device logs")
	})
}
