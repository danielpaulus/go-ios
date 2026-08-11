package ios

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRsdPortForService(t *testing.T) {
	device := DeviceEntry{
		Rsd: RsdHandshakeResponse{
			Services: map[string]RsdServiceEntry{
				"com.apple.mobile.notification_proxy.shim.remote": {Port: 50123},
			},
		},
	}

	t.Run("service is listed in rsd", func(t *testing.T) {
		port, err := RsdPortForService(device.Rsd, "com.apple.mobile.notification_proxy.shim.remote")
		require.NoError(t, err)
		assert.Equal(t, 50123, port)
	})

	t.Run("service is missing from rsd", func(t *testing.T) {
		_, err := RsdPortForService(device.Rsd, "com.apple.instruments.dtservicehub")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "com.apple.instruments.dtservicehub")
		assert.Contains(t, err.Error(), "developer disk image")
	})
}

// Services provided by the developer disk image are missing from the RSD service list until the
// image is mounted. Connecting to one of them has to fail with that reason instead of dialing
// port 0, which used to surface as an unrelated 'connection refused' on the tunnel address.
func TestConnectToServiceFailsFastWhenServiceIsMissingFromRsd(t *testing.T) {
	device := DeviceEntry{
		Address: "fd00::1",
		Rsd:     RsdHandshakeResponse{Services: map[string]RsdServiceEntry{}},
	}

	t.Run("ConnectToServiceTunnelIface", func(t *testing.T) {
		_, err := ConnectToServiceTunnelIface(device, "com.apple.instruments.dtservicehub")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not available in RSD")
	})

	t.Run("ConnectToXpcServiceTunnelIface", func(t *testing.T) {
		_, err := ConnectToXpcServiceTunnelIface(device, "com.apple.coredevice.deviceinfo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not available in RSD")
	})

	t.Run("ConnectToShimService", func(t *testing.T) {
		_, err := ConnectToShimService(device, "com.apple.mobile.notification_proxy")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not available in RSD")
	})
}

// A dial to a dead-but-still-routed tunnel address (device rebooted or hung
// while the host-side TUN route stayed up) must fail after go-ios' own dial
// timeout, not the kernel's ~135s SYN timeout, and the error must be
// classifiable as a timeout so staleness handling can key off it (issue #764).
// 192.0.2.1 (TEST-NET-1, RFC 5737) is reserved and never answers, mimicking the
// blackholed-SYN behavior of a dead tunnel.
func TestDialTunnelTCPWithTimeoutFailsFast(t *testing.T) {
	const timeout = 250 * time.Millisecond
	start := time.Now()
	_, err := DialTunnelTCPWithTimeout("192.0.2.1:54321", timeout)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 5*time.Second, "dial must fail well under the kernel SYN timeout")
	if !errors.Is(err, ErrDialTimeout) {
		// Some environments answer TEST-NET-1 with a fast ICMP unreachable or a
		// sandbox denial instead of blackholing the SYN; then this run cannot
		// exercise the timeout classification, only the fast failure above.
		t.Skipf("environment did not blackhole 192.0.2.1, got: %v", err)
	}
}
