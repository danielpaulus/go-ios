package diagnostics

import (
	"fmt"

	ios "github.com/danielpaulus/go-ios/ios"
)

const serviceName = "com.apple.mobile.diagnostics_relay"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn, plistCodec: ios.NewPlistCodec()}, nil
}

func Reboot(device ios.DeviceEntry) error {
	service, err := New(device)
	if err != nil {
		return err
	}
	return service.Reboot()
}

func (diagnosticsConn *Connection) Reboot() error {
	req := rebootRequest{Request: "Restart", WaitForDisconnect: true, DisplayFail: true, DisplayPass: true}
	reader := diagnosticsConn.deviceConn.Reader()
	bytes, err := diagnosticsConn.plistCodec.Encode(req)
	if err != nil {
		return err
	}
	diagnosticsConn.deviceConn.Send(bytes)
	response, err := diagnosticsConn.plistCodec.Decode(reader)
	if err != nil {
		return err
	}
	plist, err := ios.ParsePlist(response)
	if err != nil {
		return err
	}
	if val, ok := plist["Status"]; ok {
		if statusString, yes := val.(string); yes {
			if "Success" == statusString {
				return nil
			}

		}

	}
	return fmt.Errorf("Could not reboot, response: %+v", plist)
}

func (diagnosticsConn *Connection) MobileGestaltQuery(keys []string) (interface{}, error) {
	err := diagnosticsConn.deviceConn.Send(gestaltRequest(keys))
	if err != nil {
		return "", err
	}
	respBytes, err := diagnosticsConn.plistCodec.Decode(diagnosticsConn.deviceConn.Reader())
	if err != nil {
		return "", err
	}
	err = diagnosticsConn.deviceConn.Send(goodBye())
	if err != nil {
		return "", err
	}
	plist, err := ios.ParsePlist(respBytes)
	return plist, err
}

func (diagnosticsConn *Connection) AllValues() (allDiagnosticsResponse, error) {
	allReq := diagnosticsRequest{"All"}
	reader := diagnosticsConn.deviceConn.Reader()
	bytes, err := diagnosticsConn.plistCodec.Encode(allReq)
	if err != nil {
		return allDiagnosticsResponse{}, err
	}
	diagnosticsConn.deviceConn.Send(bytes)
	response, err := diagnosticsConn.plistCodec.Decode(reader)
	if err != nil {
		return allDiagnosticsResponse{}, err
	}
	return diagnosticsfromBytes(response), nil
}

func (diagnosticsConn *Connection) Close() error {
	reader := diagnosticsConn.deviceConn.Reader()
	closeReq := diagnosticsRequest{"Goodbye"}
	bytes, err := diagnosticsConn.plistCodec.Encode(closeReq)
	if err != nil {
		return err
	}
	diagnosticsConn.deviceConn.Send(bytes)
	_, err = diagnosticsConn.plistCodec.Decode(reader)
	diagnosticsConn.deviceConn.Close()
	return err
}
