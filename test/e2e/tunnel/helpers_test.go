//go:build e2e

// Package tunnel_test is the real-device suite for commands that require the
// iOS 17+ tunnel (RemoteServiceDiscovery / CoreDevice / instruments) and a
// mounted Developer Disk Image. It runs as a separate GitHub Actions step
// against tunnel-capable devices only; the tunnel-free suite (test/e2e) runs
// against every device.
package tunnel_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

func TestMain(m *testing.M) { harness.Main(m) }

func smoke(t *testing.T, udid string, args ...string) []byte {
	return harness.Smoke(t, udid, args...)
}

func runIOSForDevice(t *testing.T, udid string, args ...string) []byte {
	return harness.RunForDevice(t, udid, args...)
}

func forEachDevice(t *testing.T, fn func(t *testing.T, udid string)) {
	harness.ForEachDevice(t, fn)
}
