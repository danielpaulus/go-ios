package house_arrest_test

import (
	"log"
	"testing"

	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/house_arrest"
)

func TestIT(t *testing.T) {
	device := usbmux.ListDevices().DeviceList[0]
	conn, err := house_arrest.New(device.DeviceID, device.Properties.SerialNumber, "d.blaUITests.xctrunner")
	defer conn.Close()
	log.Fatal(err)
}
