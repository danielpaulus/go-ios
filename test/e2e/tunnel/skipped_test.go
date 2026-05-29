//go:build e2e

package tunnel_test

import "testing"

// TestSetLocation is skipped: setlocation runs persistently to hold the
// simulated location open and never exits cleanly (like "ios ip"), so it can't
// be asserted with a simple exit-code check. Enable once a stream-style harness
// (start, observe, kill) is wired up, then assert resetlocation restores state.
func TestSetLocation(t *testing.T) {
	t.Skip("setlocation runs persistently and does not exit; needs a stream-style harness")
	forEachDevice(t, func(t *testing.T, udid string) {
		runIOSForDevice(t, udid, "setlocation", "--lat=51.5", "--lon=-0.12")
		runIOSForDevice(t, udid, "resetlocation")
	})
}

// TestIP is skipped: "ios ip" hangs without terminating on this setup (likely a
// go-ios bug). Re-enable if the command is fixed to exit.
func TestIP(t *testing.T) {
	t.Skip("ios ip hangs without terminating (possible go-ios bug)")
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "ip") })
}
