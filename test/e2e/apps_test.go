//go:build e2e

package e2e_test

import "testing"

// TestApps uses --system so it is deterministic: a device may have no user apps
// installed, but system apps always exist. "--list" prints a plain-text listing
// (not JSON), so assert it contains a known system bundle prefix.
func TestApps(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeContains(t, udid, "com.apple", "apps", "--system", "--list")
	})
}

// TestAppsAll lists all apps (system, user, and hidden) as JSON.
func TestAppsAll(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "apps", "--all") })
}
