package ios

import (
	"bytes"

	plist "howett.net/plist"
)

// MuxResponse is a generic response message sent by usbmuxd
// it contains a Number response code
type MuxResponse struct {
	MessageType string
	Number      uint32
}

// MuxResponsefromBytes parses a MuxResponse struct from bytes
func MuxResponsefromBytes(plistBytes []byte) MuxResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var usbMuxResponse MuxResponse
	_ = decoder.Decode(&usbMuxResponse)
	return usbMuxResponse
}

// IsSuccessFull returns UsbMuxResponse.Number==0
func (u MuxResponse) IsSuccessFull() bool {
	return u.Number == 0
}
