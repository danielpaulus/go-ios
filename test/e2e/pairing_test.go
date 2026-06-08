//go:build e2e

package e2e_test

import (
	"encoding/json"
	"testing"
)

func TestReadpair(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{
			"DeviceCertificate", "HostCertificate", "HostID", "RootCertificate", "SystemBUID",
		}, "readpair")
	})
}

// TestProfileList lists configuration profiles and asserts the output is the
// profile service's JSON array (legitimately empty on these unmanaged devices) —
// i.e. its own handler reached, not the global device-list command.
func TestProfileList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := smokeJSON(t, udid, "profile", "list")
		var profiles []any
		if err := json.Unmarshal(out, &profiles); err != nil {
			t.Fatalf("profile list: output is not a JSON array: %v\n%s", err, out)
		}
	})
}
