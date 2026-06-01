//go:build e2e

package e2e_test

import (
	"encoding/json"
	"testing"
)

// TestInfo checks the lockdown info dump is valid JSON and carries a real
// ProductVersion (e.g. "18.5"), not just non-empty output.
func TestInfo(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		var m map[string]any
		if err := json.Unmarshal(smokeJSON(t, udid, "info"), &m); err != nil {
			t.Fatalf("info: parse: %v", err)
		}
		if v, _ := m["ProductVersion"].(string); v == "" {
			t.Fatalf("info: missing/empty ProductVersion in response")
		}
		// For known (static) devices, assert the identity fields match the
		// recorded snapshot exactly.
		if exp, ok := expectedDevice(udid); ok {
			for _, key := range []string{"ProductType", "ProductVersion"} {
				if got, _ := m[key].(string); got != exp[key] {
					t.Fatalf("info %s = %q, want %q (test/e2e/testdata/devices.json)", key, got, exp[key])
				}
			}
		}
	})
}

func TestInfoLockdown(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "info", "lockdown") })
}

func TestDevicename(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "devicename") })
}

func TestDate(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "date") })
}
