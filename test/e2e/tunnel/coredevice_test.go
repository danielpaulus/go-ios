//go:build e2e

package tunnel_test

import "testing"

// These commands reach the device over the iOS 17+ tunnel (RemoteServiceDiscovery
// + CoreDevice/instruments) and require a mounted Developer Disk Image. They
// fail with "missing tunnel address" / "InvalidService: Have you mounted the
// Developer Image?" when the tunnel daemon is not running on the host.

func TestInfoDisplay(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "info", "display") })
}

func TestPs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "ps") })
}

func TestDevicestateList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "devicestate", "list") })
}
