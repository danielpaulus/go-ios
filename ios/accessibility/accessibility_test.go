package accessibility_test

import (
	"testing"

	log "github.com/sirupsen/logrus"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
)

func TestIT(t *testing.T) {
	device := ios.ListDevices().DeviceList[0]

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
