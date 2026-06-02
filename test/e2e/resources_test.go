//go:build e2e

package e2e_test

import "testing"

// TestDiskspace asserts the reported capacities are internally consistent and
// plausible (free <= total, total is at least a few GB).
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
