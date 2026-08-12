//go:build e2e

package e2e_test

import (
	"testing"
	"time"
)

// assertSnapshot, for a known (static) device, asserts the recorded identity
// fields in testdata/devices.json match the live response. The fixture is a
// superset across commands (e.g. ConnectionType comes from `list`, not `info`),
// so keys absent from this particular response are skipped — the keys a command
// must return are enforced separately via smokeObj's requiredKeys. Unknown
// devices are skipped so adding a device doesn't break CI.
func assertSnapshot(t *testing.T, udid string, m map[string]any) {
	t.Helper()
	exp, ok := expectedDevice(udid)
	if !ok {
		return
	}
	for key, want := range exp {
		got, present := m[key].(string)
		if present && got != want {
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

// TestDate reads the device clock. The exact value is volatile, but assert
// formatedDate parses as a real date (RFC850, e.g. "Tuesday, 02-Jun-26
// 09:49:13 CEST") and TimeZone is a loadable IANA name — i.e. the command
// returns a genuine, well-formed date in the device's own timezone.
func TestDate(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"TimeIntervalSince1970", "formatedDate", "TimeZone"}, "date")
		fd, _ := m["formatedDate"].(string)
		if _, err := time.Parse(time.RFC850, fd); err != nil {
			t.Fatalf("date: formatedDate %q is not a parseable date: %v", fd, err)
		}
		tz, _ := m["TimeZone"].(string)
		if _, err := time.LoadLocation(tz); err != nil {
			t.Fatalf("date: TimeZone %q is not a valid IANA timezone: %v", tz, err)
		}
	})
}
