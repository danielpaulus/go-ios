//go:build e2e

package e2e_test

import "testing"

func TestDiskspace(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "diskspace") })
}

func TestBatterycheck(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "batterycheck") })
}

func TestBatteryregistry(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "batteryregistry") })
}
