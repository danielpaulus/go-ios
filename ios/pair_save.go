package ios

type SavePair struct {
	BundleID            string
	ClientVersionString string
	MessageType         string
	ProgName            string
	LibUSBMuxVersion    uint32 `plist:"kLibUSBMuxVersion"`
	PairRecordID        string
	PairRecordData      []byte
}

type savePairRecordData struct {
	DeviceCertificate []byte
	HostPrivateKey    []byte
	HostCertificate   []byte
	RootPrivateKey    []byte
	RootCertificate   []byte
	EscrowBag         []byte
	WiFiMACAddress    string
	HostID            string
	SystemBUID        string
}

func newSavePair(udid string, savePairRecordData []byte) SavePair {
	data := SavePair{
		BundleID:            "go.ios.control",
		ClientVersionString: "go-ios-1.0.0",
		MessageType:         "SavePairRecord",
		ProgName:            "go-ios",
		LibUSBMuxVersion:    3,
		PairRecordID:        udid,
		PairRecordData:      savePairRecordData,
	}
	return data

}

func newSavePairRecordData(DeviceCertificate []byte,
	HostPrivateKey []byte,
	HostCertificate []byte,
	RootPrivateKey []byte,
	RootCertificate []byte,
	EscrowBag []byte,
	WiFiMACAddress string,
	HostID string,
	SystemBUID string) []byte {
	result := savePairRecordData{DeviceCertificate, HostPrivateKey, HostCertificate, RootPrivateKey, RootCertificate, EscrowBag, WiFiMACAddress, HostID, SystemBUID}
	bytes := []byte(ToPlist(result))
	return bytes
}

func (muxConn *UsbMuxConnection) savePair(udid string, DeviceCertificate []byte,
	HostPrivateKey []byte,
	HostCertificate []byte,
	RootPrivateKey []byte,
	RootCertificate []byte,
	EscrowBag []byte,
	WiFiMACAddress string,
	HostID string,
	SystemBUID string) (bool, error) {
	bytes := newSavePairRecordData(DeviceCertificate, HostPrivateKey, HostCertificate, RootPrivateKey, RootCertificate, EscrowBag, WiFiMACAddress, HostID, SystemBUID)
	err := muxConn.Send(newSavePair(udid, bytes))
	if err != nil {
		return false, err
	}
	resp, err := muxConn.ReadMessage()
	if err != nil {
		return false, err
	}
	muxresponse := MuxResponsefromBytes(resp.Payload)
	return muxresponse.IsSuccessFull(), nil
}
