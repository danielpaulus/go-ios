//go:build e2e

package e2e_test

import "testing"

func TestCrashLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "crash", "ls") })
}
