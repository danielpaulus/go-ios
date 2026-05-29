//go:build e2e

package tunnel_test

import "testing"

// TestLaunchKill launches a system app over the tunnel (dvt/instruments) and
// kills it again. It uses Settings (com.apple.Preferences), which is always
// installed; kill succeeding confirms the app was actually running.
func TestLaunchKill(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		runIOSForDevice(t, udid, "launch", "com.apple.Preferences")
		runIOSForDevice(t, udid, "kill", "com.apple.Preferences")
	})
}
