package ios

import (
	"bytes"
	"errors"
	"log"
	"strings"

	plist "github.com/DHowett/go-plist"
	uuid "github.com/satori/go.uuid"
)

func Pair(device DeviceEntry) error {
	usbmuxConn := NewUsbMuxConnection()
	defer usbmuxConn.Close()
	buid := usbmuxConn.ReadBuid()
	lockdown, err := usbmuxConn.ConnectLockdown(device.DeviceID)
	if err != nil {
		return err
	}
	publicKey := lockdown.GetValue("DevicePublicKey").([]byte)
	wifiMac := lockdown.GetValue("WiFiAddress").(string)
	rootCert, hostCert, deviceCert, rootPrivateKey, hostPrivateKey, error := createRootCertificate(publicKey)
	if error != nil {
		log.Fatal("Failed creating Pair Record")
	}

	pairRecordData := newFullPairRecordData(buid, hostCert, rootCert, deviceCert)
	request := newLockDownPairRequest(pairRecordData)

	lockdown.deviceConnection.send((request))
	resp := <-lockdown.ResponseChannel
	response := getLockdownPairResponsefromBytes(resp)
	if isPairingDialogOpen(response) {
		log.Fatal("Please accept the PairingDialog on the device and run pairing again!")
	}
	if response.Error != "" {
		log.Fatal(response.Error)
	}
	usbmuxConn = NewUsbMuxConnection()
	defer usbmuxConn.Close()
	success := usbmuxConn.savePair(device.Properties.SerialNumber, deviceCert, hostPrivateKey, hostCert, rootPrivateKey, rootCert, response.EscrowBag, wifiMac, pairRecordData.HostID, buid)
	if !success {
		return errors.New("saving the PairRecord to usbmux failed")
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
	data.HostID = strings.ToUpper(uuid.Must(uuid.NewV4()).String())
	data.RootCertificate = rootCert
	data.HostCertificate = hostCert
	data.DeviceCertificate = deviceCert

	return data
}
