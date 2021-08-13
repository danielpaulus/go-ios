package ios

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	plist "howett.net/plist"
)

func PairSupervised(device DeviceEntry, p12 []byte) error {

	return nil
}

//Pair tries to pair with a device. The first time usually
//fails because the user has to accept a trust pop up on the iOS device.
// What you have to do to pair is:
// 1. run the Pair() function
// 2. accept the trust pop up on the device
// 3. run the Pair() function a second time
func Pair(device DeviceEntry) error {
	usbmuxConn, err := NewUsbMuxConnectionSimple()
	if err != nil {
		return err
	}
	defer usbmuxConn.Close()
	buid, err := usbmuxConn.ReadBuid()
	if err != nil {
		return err
	}
	lockdown, err := usbmuxConn.ConnectLockdown(device.DeviceID)
	if err != nil {
		return err
	}
	publicKey, err := lockdown.GetValue("DevicePublicKey")
	if err != nil {
		return err
	}
	wifiMac, err := lockdown.GetValue("WiFiAddress")
	if err != nil {
		return err
	}
	rootCert, hostCert, deviceCert, rootPrivateKey, hostPrivateKey, err := createRootCertificate(publicKey.([]byte))
	if err != nil {
		return fmt.Errorf("Failed creating pair record with error: %v", err)
	}

	pairRecordData := newFullPairRecordData(buid, hostCert, rootCert, deviceCert)
	request := newLockDownPairRequest(pairRecordData)

	err = lockdown.Send(request)
	if err != nil {
		return err
	}
	resp, err := lockdown.ReadMessage()
	if err != nil {
		return err
	}
	response := getLockdownPairResponsefromBytes(resp)
	if isPairingDialogOpen(response) {
		return fmt.Errorf("Please accept the PairingDialog on the device and run pairing again!")
	}
	if response.Error != "" {
		return fmt.Errorf("Lockdown error: %s", response.Error)
	}
	usbmuxConn, err = NewUsbMuxConnectionSimple()
	defer usbmuxConn.Close()
	if err != nil {
		return err
	}
	success, err := usbmuxConn.savePair(device.Properties.SerialNumber, deviceCert, hostPrivateKey, hostCert, rootPrivateKey, rootCert, response.EscrowBag, wifiMac.(string), pairRecordData.HostID, buid)
	if !success || err != nil {
		return errors.New("Saving the PairRecord to usbmux failed")
	}
	return nil
}

type FullPairRecordData struct {
	DeviceCertificate []byte
	HostCertificate   []byte
	RootCertificate   []byte
	SystemBUID        string
	HostID            string
}

type PairingOptions struct {
	ExtendedPairingErrors bool
}

type LockDownPairRequest struct {
	Label           string
	PairRecord      FullPairRecordData
	Request         string
	ProtocolVersion string
	PairingOptions  PairingOptions
}

type LockdownPairResponse struct {
	Error     string
	Request   string
	EscrowBag []byte
}

func getLockdownPairResponsefromBytes(plistBytes []byte) *LockdownPairResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data LockdownPairResponse
	_ = decoder.Decode(&data)
	return &data
}

func isPairingDialogOpen(resp *LockdownPairResponse) bool {
	return resp.Error == "PairingDialogResponsePending"
}

func newLockDownPairRequest(pairRecord FullPairRecordData) LockDownPairRequest {
	var req LockDownPairRequest
	req.Label = "go-ios"
	req.PairingOptions = PairingOptions{true}
	req.Request = "Pair"
	req.ProtocolVersion = "2"
	req.PairRecord = (pairRecord)
	return req
}

func newFullPairRecordData(systemBuid string, hostCert []byte, rootCert []byte, deviceCert []byte) FullPairRecordData {
	var data FullPairRecordData
	data.SystemBUID = systemBuid
	data.HostID = strings.ToUpper(uuid.New().String())
	data.RootCertificate = rootCert
	data.HostCertificate = hostCert
	data.DeviceCertificate = deviceCert

	return data
}
