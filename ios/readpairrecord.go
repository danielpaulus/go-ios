package ios

import (
	"bytes"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
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

func newReadPair(udid string) ReadPair {
	data := ReadPair{
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
	HostID            string
	SystemBUID        string
	HostCertificate   []byte
	HostPrivateKey    []byte
	DeviceCertificate []byte
	EscrowBag         []byte
	WiFiMACAddress    string
	RootCertificate   []byte
	RootPrivateKey    []byte
}

func pairRecordDatafromBytes(plistBytes []byte) PairRecordData {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data PairRecordData
	err := decoder.Decode(&data)
	if err != nil {
		log.Fatalf("Failed decoding pair record plist %x", plistBytes)
	}
	if data.PairRecordData == nil {
		resp := MuxResponsefromBytes(plistBytes)
		log.Fatalf("ReadPair failed with errorcode '%d', is the device paired?", resp.Number)
	}
	return data
}

//PairRecordfromBytes parsed a plist into a PairRecord
func PairRecordfromBytes(plistBytes []byte) PairRecord {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data PairRecord
	err := decoder.Decode(&data)
	if err != nil {
		log.Fatalf("Failed decoding pair record plist %x", plistBytes)
	}
	return data
}

//ReadPair reads the PairRecord from the usbmux socket for the given udid.
//It returns the deserialized PairRecord.
func (muxConn *UsbMuxConnection) ReadPair(udid string) PairRecord {
	muxConn.Send(newReadPair(udid))
	resp, err := muxConn.ReadMessage()
	if err != nil {
		log.Fatal("Error reading PairRecord", err)
	}
	pairRecordData := pairRecordDatafromBytes(resp.Payload)
	return PairRecordfromBytes(pairRecordData.PairRecordData)
}

//ReadPairRecord creates a new USBMuxConnection just to read the pair record and closes it right after than.
func ReadPairRecord(udid string) (PairRecord, error) {
	muxConnection, err := NewUsbMuxConnectionSimple()
	defer muxConnection.ReleaseDeviceConnection()
	return muxConnection.ReadPair(udid), err
}
