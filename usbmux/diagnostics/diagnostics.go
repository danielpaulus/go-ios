package diagnostics

import "github.com/danielpaulus/go-ios/usbmux"

const serviceName = "com.apple.mobile.diagnostics_relay"

type Connection struct {
	deviceConn usbmux.DeviceConnectionInterface
	plistCodec usbmux.PlistCodec
}

func New(deviceID int, udid string, pairRecord usbmux.PairRecord) (*Connection, error) {
	deviceConn, err := usbmux.ConnectToService(deviceID, udid, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn, plistCodec: usbmux.NewPlistCodec()}, nil
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
