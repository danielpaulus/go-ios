package instruments_test

import (
	"log"
	"testing"

	"github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	"github.com/danielpaulus/go-ios/usbmux/instruments"
	"github.com/stretchr/testify/assert"
)

func TestIt(t *testing.T) {
	device := usbmux.ListDevices().DeviceList[0]
	conn, _ := dtx.NewDtxConnection(device.DeviceID, device.Properties.SerialNumber, "com.apple.instruments.remoteserver")
	processControl := instruments.NewProcessControl(conn)

	pid, err := processControl.StartProcess("bla.test", "bla.test.de", map[string]string{}, []string{}, map[string]interface{}{})
	if err != nil {
		log.Fatal(err)
	}
	assert.NotEqual(t, 0, pid)
}
