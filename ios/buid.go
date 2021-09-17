package ios

import (
	"bytes"

	plist "howett.net/plist"
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

func newReadBuid() readBuid {
	data := readBuid{
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
func (muxConn *UsbMuxConnection) ReadBuid() (string, error) {
	err:= muxConn.Send(newReadBuid())
	if err != nil {
		return "", err
	}
	resp, err := muxConn.ReadMessage()
	if err != nil {
		return "", err
	}
	buidResponse := readBuidResponsefromBytes(resp.Payload)
	return buidResponse.BUID, nil
}
