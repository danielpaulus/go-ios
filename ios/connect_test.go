package ios

import (
	"testing"

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
