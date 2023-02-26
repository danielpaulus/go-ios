package ios

import (
	"bytes"

	plist "howett.net/plist"
)

type stopSessionRequest struct {
	Label     string
	Request   string
	SessionID string
}

func newStopSessionRequest(sessionID string) stopSessionRequest {
	data := stopSessionRequest{
		Label:     "go.ios.control",
		Request:   "StopSession",
		SessionID: sessionID,
	}
	return data
}

type stopSessionResponse struct {
	Request string
}

func stopSessionResponsefromBytes(plistBytes []byte) stopSessionResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data stopSessionResponse
	_ = decoder.Decode(&data)
	return data
}

// StopSession sends a Lockdown StopSessionRequest to the device
func (lockDownConn *LockDownConnection) StopSession() {
	if lockDownConn.sessionID == "" {
		return
	}
	lockDownConn.Send(newStopSessionRequest(lockDownConn.sessionID))
	// this returns a stopSessionResponse which we currently do not care about
	lockDownConn.ReadMessage()
}
