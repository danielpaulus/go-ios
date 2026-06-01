//go:build e2e

package e2e_test

import "testing"

func TestDiskspace(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{"BlockSize", "FreeBytes", "Model", "TotalBytes"}, "diskspace")
	})
}

func TestBatterycheck(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"BatteryCurrentCapacity", "BatteryIsCharging", "ExternalConnected",
			"FullyCharged", "HasBattery",
		}, "batterycheck")
		if has, _ := m["HasBattery"].(bool); !has {
			t.Fatalf("batterycheck: HasBattery is false")
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
