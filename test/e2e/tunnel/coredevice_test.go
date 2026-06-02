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

// TestPs lists processes; assert launchd (pid 1) is present — it always is.
func TestPs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		ps := smokeArr(t, udid, []string{"IsApplication", "Name", "Pid", "RealAppName", "StartDate"}, "ps")
		for _, p := range ps {
			if m, ok := p.(map[string]any); ok && m["Pid"] == float64(1) && m["Name"] == "launchd" {
				return
			}
		}
		t.Fatalf("ps: launchd (pid 1) not found")
	})
}

// TestPsApps lists app processes; assert at least one is flagged IsApplication.
func TestPsApps(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		ps := smokeArr(t, udid, []string{"IsApplication", "Name", "Pid", "RealAppName", "StartDate"}, "ps", "--apps")
		for _, p := range ps {
			if m, ok := p.(map[string]any); ok {
				if app, _ := m["IsApplication"].(bool); app {
					return
				}
			}
		}
		t.Fatalf("ps --apps: no entry flagged IsApplication")
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
