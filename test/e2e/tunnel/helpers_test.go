//go:build e2e

// Package tunnel_test is the real-device suite for commands that require the
// iOS 17+ tunnel (RemoteServiceDiscovery / CoreDevice / instruments) and a
// mounted Developer Disk Image. It runs as a separate GitHub Actions step
// against tunnel-capable devices only; the tunnel-free suite (test/e2e) runs
// against every device.
package tunnel_test

import (
	"syscall"
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// The tunnel suite mounts the developer disk image up front: CoreDevice
// services (e.g. info display) need it, and it is unmounted by device reboots.
func TestMain(m *testing.M) { harness.Main(m, harness.MountDeveloperImage) }

func smoke(t *testing.T, udid string, args ...string) []byte {
	return harness.Smoke(t, udid, args...)
}

func smokeJSON(t *testing.T, udid string, args ...string) []byte {
	return harness.SmokeJSON(t, udid, args...)
}

func smokeObj(t *testing.T, udid string, requiredKeys []string, args ...string) map[string]any {
	return harness.SmokeJSONObject(t, udid, requiredKeys, args...)
}

func smokeArr(t *testing.T, udid string, elemKeys []string, args ...string) []any {
	return harness.SmokeJSONArray(t, udid, elemKeys, args...)
}

func runIOSForDevice(t *testing.T, udid string, args ...string) []byte {
	return harness.RunForDevice(t, udid, args...)
}

func startBackground(t *testing.T, udid string, stopSig syscall.Signal, args ...string) (output func() string, stop func()) {
	return harness.StartBackground(t, udid, stopSig, args...)
}

func forEachDevice(t *testing.T, fn func(t *testing.T, udid string)) {
	harness.ForEachDevice(t, fn)
}
