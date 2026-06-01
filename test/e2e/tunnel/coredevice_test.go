//go:build e2e

package tunnel_test

import "testing"

// These commands reach the device over the iOS 17+ tunnel (RemoteServiceDiscovery
// + CoreDevice/instruments) and require a mounted Developer Disk Image.

func TestInfoDisplay(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"backlightState", "current", "displays", "orientation"}, "info", "display")
		if d, ok := m["displays"].([]any); !ok || len(d) == 0 {
			t.Fatalf("info display: expected a non-empty displays array, got %v", m["displays"])
		}
	})
}

func TestPs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeArr(t, udid, []string{"IsApplication", "Name", "Pid", "RealAppName", "StartDate"}, "ps")
	})
}

func TestPsApps(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeArr(t, udid, []string{"IsApplication", "Name", "Pid", "RealAppName", "StartDate"}, "ps", "--apps")
	})
}

func TestDevicestateList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		a := smokeArr(t, udid, []string{"Identifier", "Name", "IsActive", "Profiles"}, "devicestate", "list")
		// The supported condition profiles are stable; SlowNetworkCondition is
		// always present.
		found := false
		for _, e := range a {
			if m, ok := e.(map[string]any); ok && m["Identifier"] == "SlowNetworkCondition" {
				found = true
			}
		}
		if !found {
			t.Fatalf("devicestate list: SlowNetworkCondition not present")
		}
	})
}
