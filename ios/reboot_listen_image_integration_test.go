//go:build !fast
// +build !fast

package ios_test

import (
	"os"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO: add image mounting
func TestRebootListenAndImage(t *testing.T) {
	muxConnection, err := ios.NewUsbMuxConnectionSimple()
	if err != nil {
		t.Error("Failed connecting usbmux", err)
		return
	}
	attachedReceiver, err := muxConnection.Listen()
	if err != nil {
		t.Error("Failed issuing Listen command, will retry in 3 seconds", err)
		return
	}
	udid := os.Getenv("udid")
	var device ios.DeviceEntry
	for {
		msg, err := attachedReceiver()
		if err != nil {
			t.Error("Failed listen", err)
			return
		}
		if msg.Properties.SerialNumber == udid {
			assert.Equal(t, true, msg.DeviceAttached())
			device = msg.DeviceEntry()
			break
		}
	}

	log.Infof("rebooting device: %s", device.Properties.SerialNumber)
	err = diagnostics.Reboot(device)
	if err != nil {
		t.Error("Failed rebooting", err)
		return
	}

	for {
		msg, err := attachedReceiver()
		log.Infof("attached: %+v", msg)
		if err != nil {
			t.Error("Failed listen", err)
			return
		}
		if msg.DeviceDetached() && msg.DeviceID == device.DeviceID {
			break
		}
	}
	for {
		msg, err := attachedReceiver()
		if err != nil {
			t.Error("Failed listen", err)
			return
		}
		if msg.DeviceAttached() && msg.Properties.SerialNumber == device.Properties.SerialNumber {
			break
		}
	}
}
