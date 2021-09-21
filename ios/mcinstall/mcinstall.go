package mcinstall

import (
	"bytes"
	"fmt"
	"io"

	ios "github.com/danielpaulus/go-ios/ios"
	plist "howett.net/plist"
)

type plistArray []interface{}

const serviceName string = "com.apple.mobile.MCInstall"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	var mcInstallConn Connection
	mcInstallConn.deviceConn = deviceConn
	mcInstallConn.plistCodec = ios.NewPlistCodec()

	return &mcInstallConn, nil
}

func (mcInstallConn *Connection) readExchangeResponse(reader io.Reader) error {
	responseBytes, err := mcInstallConn.plistCodec.Decode(reader)
	fmt.Printf("Go %u %s", responseBytes, err)
	if err != nil {
		return err
	}

	response := getArrayFromBytes(responseBytes)
	readyMessage, ok := response[0].(string)

	fmt.Printf("Go %s - %s", readyMessage, ok)
	return nil
}

func getArrayFromBytes(plistBytes []byte) plistArray {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data plistArray
	_ = decoder.Decode(&data)
	return data
}

func (mcInstallConn *Connection) HandleList() error {
	reader := mcInstallConn.deviceConn.Reader()
	bytes, err := mcInstallConn.plistCodec.Encode([]interface{}{"RequestType", "GetProfileList"})
	if err != nil {
		return err
	}
	mcInstallConn.deviceConn.Send(bytes)
	mcInstallConn.readExchangeResponse(reader)
	return nil
}

//Close closes the underlying DeviceConnection
func (mcInstallConn *Connection) Close() {
	mcInstallConn.deviceConn.Close()
}
