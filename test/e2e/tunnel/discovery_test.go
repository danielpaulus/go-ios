//go:build e2e

package tunnel_test

import "testing"

// TestRsdLs lists the RemoteServiceDiscovery services exposed over the tunnel.
func TestRsdLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "rsd", "ls") })
}

// TestTunnelLs lists the tunnels the go-ios agent is serving for the device.
func TestTunnelLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "tunnel", "ls") })
}
