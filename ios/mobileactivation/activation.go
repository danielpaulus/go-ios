package mobileactivation

import (
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

const serviceName string = "com.apple.mobileactivationd"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	var activationdConn Connection
	activationdConn.deviceConn = deviceConn
	activationdConn.plistCodec = ios.NewPlistCodec()

	return &activationdConn, nil
}
func (activationdConn *Connection) Close() error {
	return activationdConn.deviceConn.Close()
}

func Activate(device ios.DeviceEntry) error {
	conn, err := New(device)
	if err != nil {
		return err
	}
	resp, err := conn.sendAndReceive(map[string]interface{}{"Command": "CreateTunnel1SessionInfoRequest"})
	if err != nil {
		return err
	}
	log.Infof("resp: %v", resp)
	defer conn.Close()
	return nil
}

func (mcInstallConn *Connection) sendAndReceive(request map[string]interface{}) (map[string]interface{}, error) {
	reader := mcInstallConn.deviceConn.Reader()
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return map[string]interface{}{}, err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return map[string]interface{}{}, err
	}
	responseBytes, err := mcInstallConn.plistCodec.Decode(reader)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return ios.ParsePlist(responseBytes)

}
