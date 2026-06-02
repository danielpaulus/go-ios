//go:build e2e

package e2e_test

import "testing"

// knownSystemApp is always installed, so the apps tests can assert it is listed.
const knownSystemApp = "com.apple.Preferences"

// TestApps lists system apps in compact text form (--list) and asserts a known
// system app appears.
func TestApps(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeContains(t, udid, knownSystemApp, "apps", "--system", "--list")
	})
}

// TestAppsAll lists all apps as JSON and asserts a known system app is present
// with a sane bundle entry.
func TestAppsAll(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		apps := smokeArr(t, udid, []string{"CFBundleIdentifier", "CFBundleName", "ApplicationType"}, "apps", "--all")
		for _, a := range apps {
			if m, ok := a.(map[string]any); ok && m["CFBundleIdentifier"] == knownSystemApp {
				return
			}
		}
		t.Fatalf("apps --all: %s not found in app list", knownSystemApp)
	})
}
