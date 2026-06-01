//go:build e2e

package tunnel_test

import "testing"

// TestRsdLs lists the RemoteServiceDiscovery services exposed over the tunnel.
// Assert a few stable services are advertised, including developer services
// that only appear with Developer Mode + a mounted DDI.
func TestRsdLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		smokeObj(t, udid, []string{
			"com.apple.instruments.dtservicehub",
			"com.apple.coredevice.deviceinfo",
			"com.apple.springboardservices.shim.remote",
		}, "rsd", "ls")
	})
}

// TestTunnelLs lists the tunnels the go-ios agent serves; assert this device's
// tunnel is present.
func TestTunnelLs(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		a := smokeArr(t, udid, []string{"address", "rsdPort", "udid"}, "tunnel", "ls")
		found := false
		for _, e := range a {
			if m, ok := e.(map[string]any); ok && m["udid"] == udid {
				found = true
			}
		}
		if !found {
			t.Fatalf("tunnel ls: no tunnel for udid %s", udid)
		}
	})
}
