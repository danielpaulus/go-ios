//go:build !fast
// +build !fast

package ios_test

import (
	"os"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

func TestListDevices(t *testing.T) {
	devices, err := ios.ListDevices()
	if err != nil {
		t.Error(err)
		return
	}
	udid := os.Getenv("udid")
	if udid == "" {
		t.Skip("warn no udid specified")
		return
	}
	for _, device := range devices.DeviceList {
		if device.Properties.SerialNumber == udid {
			return
		}
	}
	t.Errorf("device %s not found in list %+v", udid, devices.DeviceList)
}
