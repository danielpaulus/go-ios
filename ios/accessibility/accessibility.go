package accessibility

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

const serviceName string = "com.apple.accessibility.axAuditDaemon.remoteserver"

// New creates and connects to the given device, a new ControlInterface instance
func New(device ios.DeviceEntry) (ControlInterface, error) {
	conn, err := dtx.NewConnection(device, serviceName)
	if err != nil {
		return ControlInterface{}, err
	}
	control := ControlInterface{conn.GlobalChannel()}
	err = control.init()
	return control, err
}
