package ios

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	plist "howett.net/plist"
	"software.sslmate.com/src/go-pkcs12"
)

// ErrDeviceLockedPairingDeferred reports that the device accepted the supervisor
// identity but is locked, so it returned no EscrowBag and no pair record could be
// persisted. Any pre-existing pair record is left untouched and stays usable; if there
// was none, the device must be unlocked once before it can be paired.
var ErrDeviceLockedPairingDeferred = errors.New("device is locked: supervised pairing was accepted but no pair record could be persisted, unlock the device once and pair again")

// PairSupervised uses an organization id from apple configurator so you can pair
// a supervised device without the need for user interaction (the trust popup)
// Arguments are the device, the p12 files raw contents and the password used for the p12
// file.
// I basically got this from cfgutil:
// https://configautomation.com/cfgutil-man-page.html
// here is how to turn a p12 into crt and key:
// openssl pkcs12 -in organization.p12 -out organization.pem -nodes -password pass:a
// openssl x509 -outform DER -out organization.crt -in organization.pem
// openssl rsa -outform DER -out organization.key -in organization.pem
// then you can run:
// cfgutil -K organization.key -C organization.crt pair
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
	der, err := Sign(challengeBytes, cert, supervisedPrivateKey)
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
	// Not all rejections are encoded as a string: this protocol also returns maps
	// (see ExtendedResponse handling elsewhere in this file), so match on presence and
	// format whatever the device sent rather than only the string case. Missing this
	// would let a genuine rejection fall through to the locked-device branch below.
	if errValue, ok := respMap["Error"]; ok {
		return fmt.Errorf("PairSupervised failed: %v", errValue)
	}
	escrow, ok := respMap["EscrowBag"].([]byte)
	if !ok {
		// Device is locked: it accepted the supervisor identity but cannot update its
		// trust store or generate a new escrow bag while the passcode screen is active.
		// Saving a record with freshly generated certificates here would corrupt a
		// working one, so skip savePair -- but this is NOT a successful pairing, and
		// reporting it as one leaves callers believing a record exists when none does.
		return ErrDeviceLockedPairingDeferred
	}

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
			fmt.Errorf("received wrong error message '%s' error message should have been 'McChallengeRequired' : %+v", errormsg, respPlist)
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

// Pair tries to pair with a device. The first time usually
// fails because the user has to accept a trust pop up on the iOS device.
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
