//go:build !fast
// +build !fast

package instruments_test

import (
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/stretchr/testify/assert"
)

func TestLaunchAndKill(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		t.Fatal(err)
	}
	const weatherAppBundleID = "com.apple.weather"
	pControl, err := instruments.NewProcessControl(device)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	defer pControl.Close()

	pid, err := pControl.LaunchApp(weatherAppBundleID, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, pid, uint64(0))

	service, err := instruments.NewDeviceInfoService(device)
	if !assert.NoError(t, err) {
		return
	}
	defer service.Close()

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
}

func TestLaunchWithArgsAndKill(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		t.Fatal(err)
	}
	const weatherAppBundleID = "com.apple.weather"
	pControl, err := instruments.NewProcessControl(device)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	defer pControl.Close()

	var args = []interface{}{"-AppleLanguages", "(de-DE)"}
	var env = map[string]interface{}{"SomeRandomValue": "YES"}

	pid, err := pControl.LaunchAppWithArgs(weatherAppBundleID, args, env, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, pid, uint64(0))

	service, err := instruments.NewDeviceInfoService(device)
	if !assert.NoError(t, err) {
		return
	}
	defer service.Close()

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
}
