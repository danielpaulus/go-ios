//go:build !fast
// +build !fast

package instruments_test

import (
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLaunchAndKill(t *testing.T) {
	device := TestDevice
	const weatherAppBundleID = "com.apple.weather"
	pControl, err := instruments.NewProcessControl(device)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	defer pControl.Close()
	pid, err := pControl.LaunchApp(weatherAppBundleID)
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, pid, uint64(0))

	service, err := instruments.NewDeviceInfoService(device)
	defer service.Close()
	if !assert.NoError(t, err) {
		return
	}
	processList, err := service.ProcessList()
	if !assert.NoError(t, err) {
		return
	}
	found := false
	for _, proc := range processList {
		if proc.Pid == pid {
			found = true
		}
	}
	if !found {
		t.Errorf("could not find weather app with pid %d in proclist: %+v", pid, processList)
		return
	}
	err = pControl.KillProcess(pid)
	assert.NoError(t, err)
	return
}
