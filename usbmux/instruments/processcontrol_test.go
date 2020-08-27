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
	options := map[string]interface{}{}
	options["StartSuspendedKey"] = uint64(0)
	pid, err := processControl.StartProcess("com.netflix.Netflix", map[string]interface{}{}, []interface{}{}, options)
	if err != nil {
		log.Fatal(err)
	}
	assert.NotEqual(t, 0, pid)
}
