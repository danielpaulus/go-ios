//go:build !fast

package deviceinfo

import (
	"os"
	"strconv"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDisplayInfos(t *testing.T) {
	rsdPort, err := strconv.Atoi(os.Getenv("GO_IOS_RSD_PORT"))
	require.NoError(t, err)
	address := os.Getenv("GO_IOS_ADDRESS")
	if len(address) == 0 {
		t.Skipf("GO_IOS_ADDRESS missing")
	}

	rsdService, err := ios.NewWithAddrPort(address, rsdPort)
	require.NoError(t, err)

	defer rsdService.Close()
	rsdProvider, err := rsdService.Handshake()
	device, err := ios.GetDeviceWithAddress("", address, rsdProvider)

	deviceInfo, err := NewDeviceInfo(device)
	require.NoError(t, err)
	defer deviceInfo.Close()

	displayInfo, err := deviceInfo.GetDisplayInfo()
	assert.NoError(t, err)
	if _, ok := displayInfo["displays"].(interface{}); !ok {
		assert.Fail(t, "could not find 'displays' entry")
	}
}
