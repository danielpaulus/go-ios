package ios

import (
	"bytes"
	"crypto"
	"errors"
	"fmt"
	"github.com/fullsailor/pkcs7"
	"github.com/google/uuid"
	"golang.org/x/crypto/pkcs12"
	plist "howett.net/plist"
	"strings"
)

//PairSupervised uses an organization id from apple configurator so you can pair
//a supervised device without the need for user interaction (the trust popup)
//Arguments are the device, the p12 files raw contents and the password used for the p12
//file.
func PairSupervised(device DeviceEntry, p12bytes []byte, p12Password string) error {
	supervisedPrivateKey, cert, err := pkcs12.Decode(p12bytes, p12Password)
	if err != nil {
		return err
	}
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

	options := map[string]interface{}{"SupervisorCertificate": cert.Raw, "ExtendedPairingErrors": true}
	request := map[string]interface{}{"Label": "go-ios", "ProtocolVersion": "2", "Request": "Pair", "PairRecord": pairRecordData, "PairingOptions": options}

	err = lockdown.Send(request)
	if err != nil {
		return err
	}
	resp, err := lockdown.ReadMessage()
	if err != nil {
		return err
	}

	challengeBytes, err := extractPairingChallenge(resp)
	if err != nil {
		return err
	}
	sd, err := pkcs7.NewSignedData(challengeBytes)

	if err != nil {
		return err
	}
	err = sd.AddSigner(cert, supervisedPrivateKey.(crypto.Signer), pkcs7.SignerInfoConfig{})
	if err != nil {
		return err
	}
	der, err := sd.Finish()
	if err != nil {
		return err
	}
	options2 := map[string]interface{}{"ChallengeResponse": der}
	request = map[string]interface{}{"Label": "go-ios", "ProtocolVersion": "2", "Request": "Pair", "PairRecord": pairRecordData, "PairingOptions": options2}
	err = lockdown.Send(request)
	if err != nil {
		return err
	}
	resp, err = lockdown.ReadMessage()
	if err != nil {
		return err
	}
	respMap, err := ParsePlist(resp)
	if err != nil {
		return err
	}
	escrow := respMap["EscrowBag"].([]byte)

	usbmuxConn, err = NewUsbMuxConnectionSimple()
	defer usbmuxConn.Close()
	if err != nil {
		return err
	}

	success, err := usbmuxConn.savePair(device.Properties.SerialNumber, deviceCert, hostPrivateKey, hostCert, rootPrivateKey, rootCert, escrow, wifiMac.(string), pairRecordData.HostID, buid)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("pairing failed unexpectedly")
	}
	return nil
}

func extractPairingChallenge(resp []byte) ([]byte, error) {
	respPlist, err := ParsePlist(resp)
	if err != nil {
		return []byte{}, err
	}
	errormsgintf, ok := respPlist["Error"]
	if !ok {
		return []byte{}, fmt.Errorf("the response is missign the Error key: %+v", respPlist)
	}
	errormsg, ok := errormsgintf.(string)
	if !ok {
		return []byte{}, fmt.Errorf("error should have been a string: %+v", respPlist)
	}
	if "MCChallengeRequired" != errormsg {
		return []byte{},
		fmt.Errorf("received wrong error message '%s' error message should have been 'McChallengeRequired' : %+v",errormsg, respPlist)
	}
	respdictintf, ok := respPlist["ExtendedResponse"]
	if !ok {
		return []byte{}, fmt.Errorf("ExtendedResponse key was missing from: %+v", respPlist)
	}
	respdict, ok := respdictintf.(map[string]interface{})
	if !ok {
		return []byte{}, fmt.Errorf("ExtendedResponse should have been a map[string]innterface{}: %+v", respPlist)
	}

	challengeintf, ok := respdict["PairingChallenge"]
	if !ok {
		return []byte{}, fmt.Errorf("PairingChallenge key is missing: %+v", respPlist)
	}
	challenge, ok := challengeintf.([]byte)
	if !ok {
		return []byte{}, fmt.Errorf("PairingChallenge should have been a byte array: %+v", respPlist)
	}
	return challenge, nil

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

func getLockdownPairResponsefromBytes(plistBytes []byte) LockdownPairResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data LockdownPairResponse
	_ = decoder.Decode(&data)
	return data
}

func isPairingDialogOpen(resp LockdownPairResponse) bool {
	return resp.Error == "PairingDialogResponsePending"
}

func newLockDownPairRequest(pairRecord FullPairRecordData) LockDownPairRequest {
	var req LockDownPairRequest
	req.Label = "go-ios"
	req.PairingOptions = PairingOptions{true}
	req.Request = "Pair"
	req.ProtocolVersion = "2"
	req.PairRecord = pairRecord
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
