package diagnostics

import (
	ios "github.com/danielpaulus/go-ios/ios"
)

type Request struct {
	Request string `plist:"Request"`
}
type IORegistryRequest struct {
	Request
	CurrentPlane string `plist:"CurrentPlane,omitempty"`
	EntryName    string `plist:"EntryName,omitempty"`
	EntryClass   string `plist:"EntryClass,omitempty"`
}

func Battery(device ios.DeviceEntry) (interface{}, error) {
	conn, _ := New(device)
	req := &IORegistryRequest{
		Request:      Request{"IORegistry"},
		CurrentPlane: "",
		EntryName:    "",
		EntryClass:   "IOPMPowerSource",
	}
	err := conn.deviceConn.SendAny(req)
	if err != nil {
		return "", err
	}
	respBytes, err := conn.plistCodec.RecvBytes(conn.deviceConn.Reader())
	if err != nil {
		return "", err
	}
	plist, err := ios.ParsePlist(respBytes)
	return plist, err
}
