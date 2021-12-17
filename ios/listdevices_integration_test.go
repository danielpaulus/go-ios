package ios_test

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"strings"
	"testing"
)

func TestListDevices(t *testing.T) {
	devices, err := ios.ListDevices()
	if err != nil {
		t.Error(err)
		return
	}
	outBytes, err := exec.Command("ioreg", "-p", "IOUSB", "-l", "-b").CombinedOutput()
	usbDeviceDetails := string(outBytes)
	assert.Greater(t, len(devices.DeviceList), 0, "at least one device should be connected")
	for _, dev := range devices.DeviceList {
		sn := dev.Properties.SerialNumber
		sn = strings.Replace(sn, "-", "", 1)

		assert.Contains(t, usbDeviceDetails, sn, "check that sn %s is connected")
	}
}
