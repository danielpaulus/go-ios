package usbmux

import (
	"bytes"

	plist "howett.net/plist"
)

type startSessionRequest struct {
	Label           string
	ProtocolVersion string
	Request         string
	HostID          string
	SystemBUID      string
}

func newStartSessionRequest(hostID string, systemBuid string) *startSessionRequest {
	data := &startSessionRequest{
		Label:           "go.ios.control",
		ProtocolVersion: "2",
		Request:         "StartSession",
		HostID:          hostID,
		SystemBUID:      systemBuid,
	}
	return data
}

//StartSessionResponse contains the information sent by the device as a response to a StartSessionRequest.
type StartSessionResponse struct {
	EnableSessionSSL bool
	Request          string
	SessionID        string
}

func startSessionResponsefromBytes(plistBytes []byte) StartSessionResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data StartSessionResponse
	_ = decoder.Decode(&data)
	return data
}

//StartSession will send a StartSession Request to Lockdown, wait for the response and enable
//SSL on the underlying connection if necessary. The devices usually always requests to enable
//SSL.
//It returns a StartSessionResponse
func (lockDownConn *LockDownConnection) StartSession(pairRecord PairRecord) StartSessionResponse {
	deviceConnection := lockDownConn.deviceConnection
	return deviceConnection.SendForSslUpgrade(lockDownConn, pairRecord)
}
