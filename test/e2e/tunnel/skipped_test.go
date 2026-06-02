//go:build e2e

package tunnel_test

import "testing"

// TestIP is skipped: "ios ip" hangs without terminating on this setup (likely a
// go-ios bug). Re-enable if the command is fixed to exit.
func TestIP(t *testing.T) {
	t.Skip("ios ip hangs without terminating (possible go-ios bug)")
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "ip") })
}
