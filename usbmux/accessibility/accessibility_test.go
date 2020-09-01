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
	conn.Init()
	if err != nil {
		log.Fatal(err)
	}

}
