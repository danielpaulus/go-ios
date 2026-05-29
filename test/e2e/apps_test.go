//go:build e2e

package e2e_test

import "testing"

// TestApps uses --system so it is deterministic: a device may have no user
// apps installed (making a plain "apps --list" empty), but system apps always
// exist. It exercises the same installation_proxy lockdown service.
func TestApps(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "apps", "--system", "--list") })
}

// TestAppsAll lists all apps (system, user, and hidden).
func TestAppsAll(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "apps", "--all") })
}
