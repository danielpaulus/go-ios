//go:build !fast

package appservice_test

import (
	"os"
	"slices"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/appservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLaunchAndKillApps(t *testing.T) {
	rsdPath := os.Getenv("GO_IOS_RSD")
	if len(rsdPath) == 0 {
		t.Skipf("GO_IOS_RSD missing")
	}
	address := os.Getenv("GO_IOS_ADDRESS")
	if len(address) == 0 {
		t.Skipf("GO_IOS_ADDRESS missing")
	}

	f, err := os.Open(rsdPath)
	require.NoError(t, err)
	defer f.Close()
	rsd, err := ios.NewRsdPortProvider(f)
	require.NoError(t, err)

	device, err := ios.GetDeviceWithAddress("", address, rsd)
	require.NoError(t, err)
	version, err := ios.GetProductVersion(device)
	require.NoError(t, err)
	if version.Major() < 17 {
		t.Skipf("Device has version %s. Skipping test on devices lower than iOS 17", version.String())
	}

	t.Run("launch and kill app", func(t *testing.T) {
		testLaunchAndKillApp(t, device)
	})
	t.Run("kill invalid pid", func(t *testing.T) {
		testKillInvalidPidReturnsError(t, device)
	})
}

func testLaunchAndKillApp(t *testing.T, device ios.DeviceEntry) {
	as, err := appservice.New(device)
	require.NoError(t, err)
	defer as.Close()

	_, err = as.LaunchApp("com.apple.mobilesafari", nil, nil, nil, true)
	require.NoError(t, err)

	processes, err := as.ListProcesses()
	require.NoError(t, err)

	idx := slices.IndexFunc(processes, func(e appservice.Process) bool {
		return e.ExecutableName() == "MobileSafari"
	})
	assert.NotEqual(t, -1, idx)

	process := processes[idx]

	err = as.KillProcess(process.Pid)
	assert.NoError(t, err)
}

func testKillInvalidPidReturnsError(t *testing.T, device ios.DeviceEntry) {
	as, err := appservice.New(device)
	require.NoError(t, err)
	defer as.Close()

	pid, err := as.LaunchApp("com.apple.mobilesafari", nil, nil, nil, true)
	require.NoError(t, err)

	err = as.KillProcess(pid)
	require.NoError(t, err)

	err = as.KillProcess(pid)
	assert.Error(t, err)
}
