// +build integration

package accessibility_test

import (
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
	log "github.com/sirupsen/logrus"
)

func TestIT(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		log.Fatal(err)
	}

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
