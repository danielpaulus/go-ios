package accessibility

import (
	"time"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

const serviceName string = "com.apple.accessibility.axAuditDaemon.remoteserver"

// NewWithoutEventChangeListeners creates and connects to the given device, a new ControlInterface instance
// without setting accessibility event change listeners to avoid keeping constant connection.
func NewWithoutEventChangeListeners(device ios.DeviceEntry, timeout time.Duration) (*ControlInterface, error) {
	conn, err := dtx.NewUsbmuxdConnection(device, serviceName)
	if err != nil {
		return nil, err
	}
	control := &ControlInterface{
		cm:               conn,
		channel:          conn.GlobalChannel(),
		broadcastTimeout: timeout,
	}
	return control, nil
}

// New creates and connects to the given device, a new ControlInterface instance
func New(device ios.DeviceEntry, timeout time.Duration) (*ControlInterface, error) {
	conn, err := dtx.NewUsbmuxdConnection(device, serviceName)
	if err != nil {
		return nil, err
	}
	control := &ControlInterface{
		cm:               conn,
		channel:          conn.GlobalChannel(),
		broadcastTimeout: timeout,
	}

	err = control.init()
	return control, err
}
