package deviceinfo

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/coredevice"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

type Connection struct {
	conn     *xpc.Connection
	deviceId string
}

func New(deviceEntry ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(deviceEntry, "com.apple.coredevice.deviceinfo")
	if err != nil {
		return nil, err
	}

	return &Connection{conn: xpcConn, deviceId: uuid.New().String()}, nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c Connection) GetDisplayInfo() (map[string]interface{}, error) {
	request := coredevice.BuildRequest(c.deviceId, "com.apple.coredevice.feature.getdisplayinfo", map[string]interface{}{})
	err := c.conn.Send(request, xpc.HeartbeatRequestFlag)
	if err != nil {
		return nil, err
	}

	response, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return nil, err
	}

	if output, ok := response["CoreDevice.output"].(map[string]interface{}); ok {
		return output, nil
	}

	return nil, fmt.Errorf("could not parse response %+v", response)
}
