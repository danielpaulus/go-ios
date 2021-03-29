// +build integration

package instruments_test

import (
	"log"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/stretchr/testify/assert"
)

func TestLaunchAndKill(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		log.Fatal(err)
	}
	const weatherAppBundleID = "com.apple.weather"
	pControl, err := instruments.NewProcessControl(device)
	defer pControl.Close()
	if !assert.NoError(t, err) {
		log.Fatal(err)
	}
	pid, err := pControl.LaunchApp(weatherAppBundleID)
	if assert.NoError(t, err) {
		assert.Greater(t, pid, uint64(0))
		err := pControl.KillProcess(pid)
		assert.NoError(t, err)
		return
	}
	log.Fatal(err)
}
