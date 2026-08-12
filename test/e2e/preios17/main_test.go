//go:build e2e

// Package preios17_test is the comprehensive real-device suite for devices
// older than iOS 17. These devices serve the *entire* go-ios feature set —
// including the instruments/DTX commands (ps, screenshot, launch/kill) that on
// iOS 17+ require the tunnel — directly over usbmuxd/lockdown with only a
// mounted Developer Disk Image and NO tunnel. It runs as its own GitHub Actions
// step against a dedicated pre-iOS17 device list.
//
// The assertions here are deliberately separate from the tunnel-free suite
// (test/e2e): several command outputs differ on old iOS (e.g. MobileGestalt is
// not deprecated and returns real values, Developer Mode does not exist, crash
// reports are not .ips), so dumping these devices into the iOS 18 suite's
// exact-match tests would break it.
package preios17_test

import (
	"syscall"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// Like the tunnel suite, mount the developer disk image up front: the
// instruments services (ps, screenshot, launch) need it. Unlike the tunnel
// suite, no tunnel is started — pre-iOS17 devices reach these services over the
// classic lockdown/usbmux transport.
func TestMain(m *testing.M) { harness.Main(m, harness.MountDeveloperImage) }

// Thin aliases so the per-command test files read cleanly (mirrors the other
// suites' helpers_test.go).
func runIOS(t *testing.T, args ...string) []byte { return harness.RunIOS(t, args...) }

func runIOSForDevice(t *testing.T, udid string, args ...string) []byte {
	return harness.RunForDevice(t, udid, args...)
}

// The *Resilient helpers retry ONLY on a transient usbmuxd / USB-transport drop
// on the flaky pre-iOS17 device (see the harness comment on
// transientTransportSignatures) — never on assertion or genuine command errors.
// The DTX/instruments and syslog tests use them so one infra hiccup mid-suite
// self-heals instead of red-walling the whole PR fleet.
func runIOSForDeviceResilient(t *testing.T, udid string, args ...string) []byte {
	return harness.RunForDeviceResilient(t, udid, args...)
}

func smokeArrResilient(t *testing.T, udid string, elemKeys []string, args ...string) []any {
	return harness.SmokeArrResilient(t, udid, elemKeys, args...)
}

func streamSmokeResilient(t *testing.T, udid string, window time.Duration, args ...string) []byte {
	return harness.StreamSmokeResilient(t, udid, window, args...)
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

func smokeObj(t *testing.T, udid string, requiredKeys []string, args ...string) map[string]any {
	return harness.SmokeJSONObject(t, udid, requiredKeys, args...)
}

func smokeArr(t *testing.T, udid string, elemKeys []string, args ...string) []any {
	return harness.SmokeJSONArray(t, udid, elemKeys, args...)
}

func streamSmoke(t *testing.T, udid string, window time.Duration, args ...string) []byte {
	return harness.StreamSmoke(t, udid, window, args...)
}

func streamNDJSON(t *testing.T, udid string, timeout time.Duration, args ...string) []map[string]any {
	return harness.StreamNDJSON(t, udid, timeout, args...)
}

func startBackground(t *testing.T, udid string, stopSig syscall.Signal, args ...string) (output func() string, stop func()) {
	return harness.StartBackground(t, udid, stopSig, args...)
}

func expectedDevice(udid string) (map[string]string, bool) {
	return harness.ExpectedDevice(udid)
}

func forEachDevice(t *testing.T, fn func(t *testing.T, udid string)) {
	harness.ForEachDevice(t, fn)
}

// assertSnapshot asserts the recorded identity fields in testdata/devices.json
// match the live response for a known device. Keys absent from a particular
// response are skipped (the fixture is a superset across commands); unknown
// devices are skipped so adding a device does not break CI.
func assertSnapshot(t *testing.T, udid string, m map[string]any) {
	t.Helper()
	exp, ok := expectedDevice(udid)
	if !ok {
		return
	}
	for key, want := range exp {
		if got, present := m[key].(string); present && got != want {
			t.Fatalf("%s = %q, want %q (test/e2e/testdata/devices.json)", key, got, want)
		}
	}
}
