package ios

import (
	"bytes"

	plist "github.com/DHowett/go-plist"
)

type readBuid struct {
	BundleID            string
	ClientVersionString string
	MessageType         string
	ProgName            string
	LibUSBMuxVersion    uint32 `plist:"kLibUSBMuxVersion"`
}

type readBuidResponse struct {
	BUID string
}

func newReadBuid() *readBuid {
	data := &readBuid{
		BundleID:            "go.ios.control",
		ClientVersionString: "go-usbmux-0.0.1",
		MessageType:         "ReadBUID",
		ProgName:            "go-usbmux",
		LibUSBMuxVersion:    3,
	}
	return data
}

func readBuidResponsefromBytes(plistBytes []byte) readBuidResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data readBuidResponse
	_ = decoder.Decode(&data)
	return data
}

//ReadBuid requests the BUID of the host
//It returns the deserialized BUID as a string.
func (muxConn *MuxConnection) ReadBuid() string {
	muxConn.deviceConn.send(newReadBuid())
	resp := <-muxConn.ResponseChannel
	buidResponse := readBuidResponsefromBytes(resp)
	return buidResponse.BUID
}
