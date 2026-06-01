//go:build e2e

package e2e_test

import "testing"

func TestLockdownGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "lockdown", "get") })
}

func TestMobilegestalt(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "mobilegestalt", "ProductVersion") })
}
