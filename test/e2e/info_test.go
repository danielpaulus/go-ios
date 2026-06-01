//go:build e2e

package e2e_test

import "testing"

// assertSnapshot, for a known (static) device, asserts every recorded identity
// field in testdata/devices.json matches the live response. Unknown devices are
// skipped so adding a device doesn't break CI.
func assertSnapshot(t *testing.T, udid string, m map[string]any) {
	t.Helper()
	exp, ok := expectedDevice(udid)
	if !ok {
		return
	}
	for key, want := range exp {
		if got, _ := m[key].(string); got != want {
			t.Fatalf("%s = %q, want %q (test/e2e/testdata/devices.json)", key, got, want)
		}
	}
}

func TestInfo(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"ProductType", "ProductVersion", "BuildVersion", "DeviceClass",
			"CPUArchitecture", "HardwareModel", "ProductName", "ModelNumber",
			"SerialNumber", "UniqueDeviceID", "DeviceName",
		}, "info")
		assertSnapshot(t, udid, m)
	})
}

func TestInfoLockdown(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"BuildVersion", "CPUArchitecture", "DeviceClass", "DeviceName",
			"ProductType", "ProductVersion", "UniqueDeviceID",
		}, "info", "lockdown")
		assertSnapshot(t, udid, m)
	})
}

func TestDevicename(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"devicename"}, "devicename")
		if name, _ := m["devicename"].(string); name == "" {
			t.Fatalf("devicename: empty")
		}
	})
}

// TestDate reads the device clock; the values are volatile, so only assert the
// response shape.
func TestDate(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{"TimeIntervalSince1970", "formatedDate"}, "date")
	})
}
