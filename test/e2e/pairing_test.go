//go:build e2e

package e2e_test

import "testing"

func TestReadpair(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{
			"DeviceCertificate", "HostCertificate", "HostID", "RootCertificate", "SystemBUID",
		}, "readpair")
	})
}

// TestProfileList lists configuration profiles. The list is legitimately empty
// on these devices, so only assert it is valid JSON (an array).
func TestProfileList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "profile", "list") })
}
