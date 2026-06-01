//go:build e2e

package e2e_test

import "testing"

// TestLockdownGet dumps all lockdown values; assert the stable identity keys are
// present and, for known devices, match the recorded snapshot.
func TestLockdownGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"BuildVersion", "CPUArchitecture", "DeviceClass", "DeviceName",
			"ProductType", "ProductVersion", "UniqueDeviceID",
		}, "lockdown", "get")
		if exp, ok := expectedDevice(udid); ok {
			for _, key := range []string{"ProductType", "ProductVersion", "BuildVersion"} {
				if got, _ := m[key].(string); got != exp[key] {
					t.Fatalf("lockdown get %s = %q, want %q", key, got, exp[key])
				}
			}
		}
	})
}

// TestMobilegestalt queries a key via the diagnostics relay.
func TestMobilegestalt(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{"Diagnostics", "Status"}, "mobilegestalt", "ProductVersion")
	})
}
