package accessibility

import (
	"github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
)

const serviceName string = "com.apple.accessibility.axAuditDaemon.remoteserver"

func New(device usbmux.DeviceEntry) (AccessibilityControl, error) {
	conn, err := dtx.NewDtxConnection(device.DeviceID, device.Properties.SerialNumber, serviceName)
	if err != nil {
		return AccessibilityControl{}, err
	}
	return AccessibilityControl{conn.GlobalChannel()}, nil
}
