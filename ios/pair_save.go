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

func newSavePair(udid string, savePairRecordData *[]byte) *SavePair {
	data := &SavePair{
		BundleID:            "go.ios.control",
		ClientVersionString: "go-usbmux-0.0.1",
		MessageType:         "ReadPairRecord",
		ProgName:            "go-usbmux",
		LibUSBMuxVersion:    3,
		PairRecordID:        udid,
		PairRecordData:      *savePairRecordData,
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
	SystemBUID string) *[]byte {
	result := savePairRecordData{DeviceCertificate, HostPrivateKey, HostCertificate, RootPrivateKey, RootCertificate, EscrowBag, WiFiMACAddress, HostID, SystemBUID}
	bytes := []byte(ToPlist(result))
	return &bytes
}

func (muxConn *MuxConnection) savePair(udid string, DeviceCertificate []byte,
	HostPrivateKey []byte,
	HostCertificate []byte,
	RootPrivateKey []byte,
	RootCertificate []byte,
	EscrowBag []byte,
	WiFiMACAddress string,
	HostID string,
	SystemBUID string) bool {
	bytes := newSavePairRecordData(DeviceCertificate, HostPrivateKey, HostCertificate, RootPrivateKey, RootCertificate, EscrowBag, WiFiMACAddress, HostID, SystemBUID)
	muxConn.deviceConn.send(newSavePair(udid, bytes))
	resp := <-muxConn.ResponseChannel
	muxresponse := usbMuxResponsefromBytes(resp)
	return muxresponse.IsSuccessFull()
}
