//go:build e2e

// Package e2e_test is the tunnel-free real-device suite: commands served over
// the classic lockdown/usbmux services, which work on every device regardless
// of iOS version. Tests that need the iOS 17+ tunnel live in test/e2e/tunnel.
package e2e_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

func TestMain(m *testing.M) { harness.Main(m) }

// Thin aliases so the per-command test files read cleanly.
func runIOS(t *testing.T, args ...string) []byte { return harness.RunIOS(t, args...) }

func runIOSForDevice(t *testing.T, udid string, args ...string) []byte {
	return harness.RunForDevice(t, udid, args...)
}

func smoke(t *testing.T, udid string, args ...string) []byte {
	return harness.Smoke(t, udid, args...)
}

func smokeJSON(t *testing.T, udid string, args ...string) []byte {
	return harness.SmokeJSON(t, udid, args...)
}

func smokeContains(t *testing.T, udid, want string, args ...string) []byte {
	return harness.SmokeContains(t, udid, want, args...)
}

func expectedDevice(udid string) (map[string]string, bool) {
	return harness.ExpectedDevice(udid)
}

func smokeObj(t *testing.T, udid string, requiredKeys []string, args ...string) map[string]any {
	return harness.SmokeJSONObject(t, udid, requiredKeys, args...)
}

func smokeArr(t *testing.T, udid string, elemKeys []string, args ...string) []any {
	return harness.SmokeJSONArray(t, udid, elemKeys, args...)
}

func forEachDevice(t *testing.T, fn func(t *testing.T, udid string)) {
	harness.ForEachDevice(t, fn)
}
