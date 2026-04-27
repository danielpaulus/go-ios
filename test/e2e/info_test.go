//go:build e2e

package e2e_test

import "testing"

func TestInfo(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "info") })
}

func TestInfoDisplay(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "info", "display") })
}

func TestInfoLockdown(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "info", "lockdown") })
}

func TestDevicename(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "devicename") })
}

func TestDate(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "date") })
}
