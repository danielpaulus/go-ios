package misagent

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.misagent"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	var c Connection
	c.deviceConn = deviceConn
	c.plistCodec = ios.NewPlistCodec()

	return &c, nil
}

func (c *Connection) CopyAll() error {
	msg := map[string]interface{}{
		"MessageType": "CopyAll",
		"ProfileType": "Provisioning",
	}
	reader := c.deviceConn.Reader()
	requestBytes, err := c.plistCodec.Encode(msg)
	if err != nil {
		return err
	}
	err = c.deviceConn.Send(requestBytes)
	if err != nil {
		return err
	}
	responseBytes, err := c.plistCodec.Decode(reader)
	if err != nil {
		return err
	}

	resp, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return err
	}
	t, ok := resp["Status"]
	if !ok {
		return fmt.Errorf("misagent invalid response %v", resp)
	}
	i, ok := t.(int)
	if !ok {
		return fmt.Errorf("misagent invalid status in response %v", resp)
	}
	if i == 0 {
		return nil
	}
	return fmt.Errorf("misagent returned error code %d in response %v", i, resp)
}
