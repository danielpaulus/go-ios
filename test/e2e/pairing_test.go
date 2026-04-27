//go:build e2e

package e2e_test

import "testing"

func TestReadpair(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "readpair") })
}

func TestProfileList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "profile", "list") })
}
