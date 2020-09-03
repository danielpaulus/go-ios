package accessibility

import (
	ios "github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
)

const serviceName string = "com.apple.accessibility.axAuditDaemon.remoteserver"

//New creates and connects to the given device, a new ControlInterface instance
func New(device ios.DeviceEntry) (ControlInterface, error) {
	conn, err := dtx.NewConnection(device, serviceName)
	if err != nil {
		return ControlInterface{}, err
	}
	control := ControlInterface{conn.GlobalChannel()}
	err = control.init()
	return control, err
}
