package accessibility_test

import (
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/accessibility"
)

func TestIT(t *testing.T) {
	device := usbmux.ListDevices().DeviceList[0]

	conn, err := accessibility.New(device)
	if err != nil {
		log.Fatal(err)
	}

	conn.SwitchToDevice()
	if err != nil {
		log.Fatal(err)
	}
	conn.EnableSelectionMode()
	conn.GetElement()
	conn.GetElement()
	conn.TurnOff()

	//conn.EnableSelectionMode()

}
