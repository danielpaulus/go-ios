package instruments

import (
	"bufio"

	"github.com/danielpaulus/go-ios/usbmux"
)

const serviceName string = "com.apple.instruments.remoteserver"

type Connection struct {
	deviceConn     usbmux.DeviceConnectionInterface
	bufferedReader *bufio.Reader
}

func New(deviceID int, udid string) error {
	_, err := usbmux.ConnectToService(deviceID, udid, serviceName)
	if err != nil {
		return err
	}
	return nil
}

func (connection *Connection) Close() {
	connection.deviceConn.Close()
}
