//go:build e2e

package preios17_test

import (
	"encoding/json"
	"testing"
)

// TestDiskspace asserts the reported capacities are internally consistent.
func TestDiskspace(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"BlockSize", "FreeBytes", "Model", "TotalBytes"}, "diskspace")
		free, _ := m["FreeBytes"].(float64)
		total, _ := m["TotalBytes"].(float64)
		if total < 1e9 {
			t.Fatalf("diskspace: implausible TotalBytes %.0f (< 1GB)", total)
		}
		if free <= 0 || free > total {
			t.Fatalf("diskspace: FreeBytes %.0f out of range (0, TotalBytes=%.0f]", free, total)
		}
	})
}

// TestBatterycheck asserts a battery is present and reports a sane charge level.
func TestBatterycheck(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"BatteryCurrentCapacity", "BatteryIsCharging", "ExternalConnected",
			"FullyCharged", "HasBattery",
		}, "batterycheck")
		if has, _ := m["HasBattery"].(bool); !has {
			t.Fatalf("batterycheck: HasBattery is false")
		}
		if lvl, _ := m["BatteryCurrentCapacity"].(float64); lvl < 0 || lvl > 100 {
			t.Fatalf("batterycheck: BatteryCurrentCapacity %.0f not in 0..100", lvl)
		}
	})
}

func TestBatteryregistry(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{
			"CurrentCapacity", "CycleCount", "DesignCapacity", "Temperature", "Voltage",
		}, "batteryregistry")
	})
}

func TestReadpair(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{
			"DeviceCertificate", "HostCertificate", "HostID", "RootCertificate", "SystemBUID",
		}, "readpair")
	})
}

// TestProfileList lists configuration profiles and asserts the output is the
// profile service's JSON array (legitimately empty here) — i.e. its own handler
// reached, not the global device-list command.
func TestProfileList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := smokeJSON(t, udid, "profile", "list")
		var profiles []any
		if err := json.Unmarshal(out, &profiles); err != nil {
			t.Fatalf("profile list: output is not a JSON array: %v\n%s", err, out)
		}
	})
}

// TestCrashLs lists crash/diagnostic reports. Old iOS uses report names such as
// "Analytics-...ips.ca.synced" rather than plain ".ips", so assert only that
// the list is non-empty and length is consistent with the files array.
func TestCrashLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"files", "length"}, "crash", "ls")
		files, _ := m["files"].([]any)
		length, _ := m["length"].(float64)
		if int(length) != len(files) {
			t.Fatalf("crash ls: length=%d but files has %d entries", int(length), len(files))
		}
		if len(files) == 0 {
			t.Fatalf("crash ls: no crash/diagnostic reports found")
		}
	})
}

// TestImageList runs after the harness mounted the developer disk image; assert
// the command succeeds (output shape varies by iOS version).
func TestImageList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { runIOSForDevice(t, udid, "image", "list") })
}
