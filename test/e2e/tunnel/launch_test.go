//go:build e2e

package tunnel_test

import (
	"testing"
	"time"
)

// TestLaunchKill launches a system app over the tunnel and confirms it actually
// runs (appears in ps), then kills it and confirms it is gone. Uses Settings
// (com.apple.Preferences), which is always installed.
func TestLaunchKill(t *testing.T) {
	const bundleID, procName = "com.apple.Preferences", "Preferences"
	forEachDevice(t, func(t *testing.T, udid string) {
		running := func() bool {
			for _, p := range smokeArr(t, udid, []string{"Name", "Pid", "IsApplication"}, "ps") {
				m, ok := p.(map[string]any)
				if !ok || m["Name"] != procName {
					continue
				}
				if app, _ := m["IsApplication"].(bool); app {
					return true
				}
			}
			return false
		}
		waitFor := func(want bool, what string) {
			for i := 0; i < 10; i++ {
				if running() == want {
					return
				}
				time.Sleep(time.Second)
			}
			t.Fatalf("%s: %q running=%v, want %v", what, procName, !want, want)
		}

		runIOSForDevice(t, udid, "launch", bundleID)
		waitFor(true, "launch")

		runIOSForDevice(t, udid, "kill", bundleID)
		waitFor(false, "kill")
	})
}
