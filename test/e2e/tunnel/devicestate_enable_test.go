//go:build e2e

package tunnel_test

import (
	"testing"
	"time"
)

// TestDevicestateEnable activates a device condition profile and verifies a
// separate "devicestate list" reports it active, then stops it. devicestate
// enable only holds the condition while running, so SIGTERM (via startBackground)
// reverts the device — the test is self-restoring.
func TestDevicestateEnable(t *testing.T) {
	const typeID, profileID = "SlowNetworkCondition", "SlowNetwork3GGood"
	forEachDevice(t, func(t *testing.T, udid string) {
		stop := startBackground(t, udid, "devicestate", "enable", typeID, profileID)
		defer stop()

		// Poll until the condition becomes active (it takes a moment to apply).
		var gotProfile string
		var active bool
		for i := 0; i < 15; i++ {
			gotProfile, active = "", false
			for _, e := range smokeArr(t, udid, []string{"Identifier", "ActiveProfile", "IsActive"}, "devicestate", "list") {
				m, _ := e.(map[string]any)
				if m["Identifier"] == typeID {
					gotProfile, _ = m["ActiveProfile"].(string)
					active, _ = m["IsActive"].(bool)
				}
			}
			if active {
				break
			}
			time.Sleep(time.Second)
		}

		if !active || gotProfile != profileID {
			t.Fatalf("devicestate enable %s %s: condition not active (ActiveProfile=%q, IsActive=%v)",
				typeID, profileID, gotProfile, active)
		}
	})
}
