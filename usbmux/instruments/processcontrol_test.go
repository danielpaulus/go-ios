package instruments_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	"github.com/danielpaulus/go-ios/usbmux/instruments"
)

func TestIt(t *testing.T) {
	device := usbmux.ListDevices().DeviceList[0]
	conn, _ := dtx.NewDtxConnection(device.DeviceID, device.Properties.SerialNumber, "com.apple.instruments.remoteserver")
	instruments.NewProcessControl(conn)
	for {
	}
}
