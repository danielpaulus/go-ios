//go:build e2e

package e2e_test

import "testing"

// TestDevmodeGet asserts Developer Mode is reported enabled (the test devices
// have it on; the tunnel suite depends on it).
func TestDevmodeGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"DeveloperModeEnabled"}, "devmode", "get")
		if on, _ := m["DeveloperModeEnabled"].(bool); !on {
			t.Fatalf("devmode get: DeveloperModeEnabled is false")
		}
	})
}

func TestLang(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"Language", "Locale", "SupportedLocales"}, "lang")
		if lang, _ := m["Language"].(string); lang == "" {
			t.Fatalf("lang: empty Language")
		}
	})
}

// TestDiagnosticsList asserts the diagnostics relay reports overall success.
func TestDiagnosticsList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"Diagnostics", "Status"}, "diagnostics", "list")
		if status, _ := m["Status"].(string); status != "Success" {
			t.Fatalf("diagnostics list: Status = %q, want Success", status)
		}
	})
}
