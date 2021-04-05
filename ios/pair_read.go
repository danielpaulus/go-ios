package ios

import (
	"bytes"

	plist "github.com/DHowett/go-plist"
)

//ReadPair contains all the Infos necessary
//to request a PairRecord from usbmuxd.
//use newReadPair(udid string) to create one.
type ReadPair struct {
	BundleID            string
	ClientVersionString string
	MessageType         string
	ProgName            string
	LibUSBMuxVersion    uint32 `plist:"kLibUSBMuxVersion"`
	PairRecordID        string
}

func newReadPair(udid string) *ReadPair {
	data := &ReadPair{
		BundleID:            "go.ios.control",
		ClientVersionString: "go-usbmux-0.0.1",
		MessageType:         "ReadPairRecord",
		ProgName:            "go-usbmux",
		LibUSBMuxVersion:    3,
		PairRecordID:        udid,
	}
	return data
}

//PairRecordData only holds a []byte containing the PairRecord data as
//a serialized Plist.
type PairRecordData struct {
	PairRecordData []byte
}

//PairRecord contains the HostID string,
//the SystemBUID string, the HostCertificate []byte
//and the HostPrivateKey []byte.
//It is needed for enabling SSL Connections over Lockdown
type PairRecord struct {
	HostID          string
	SystemBUID      string
	HostCertificate []byte
	HostPrivateKey  []byte
}

func pairRecordDatafromBytes(plistBytes []byte) PairRecordData {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data PairRecordData
	_ = decoder.Decode(&data)
	return data
}

func pairRecordfromBytes(plistBytes []byte) PairRecord {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data PairRecord
	_ = decoder.Decode(&data)
	return data
}

//ReadPair reads the PairRecord from the usbmux socket for the given udid.
//It returns the deserialized PairRecord.
func (muxConn *MuxConnection) ReadPair(udid string) PairRecord {
	muxConn.deviceConn.send(newReadPair(udid))
	resp := <-muxConn.ResponseChannel
	pairRecordData := pairRecordDatafromBytes(resp)
	return pairRecordfromBytes(pairRecordData.PairRecordData)
}
