package mcinstall

import (
	"crypto/x509"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pkcs12"

	ios "github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.mobile.MCInstall"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	var mcInstallConn Connection
	mcInstallConn.deviceConn = deviceConn
	mcInstallConn.plistCodec = ios.NewPlistCodec()

	return &mcInstallConn, nil
}

type ProfileInfo struct {
	Identifier string
	Manifest   ProfileManifest
	Metadata   ProfileMetadata
	Status     string
}

type ProfileMetadata struct {
	PayloadDescription       string
	PayloadDisplayName       string
	PayloadRemovalDisallowed bool
	PayloadUUID              string
	PayloadVersion           uint64
}

type ProfileManifest struct {
	Description string
	IsActive    bool
}

func (mcInstallConn *Connection) readExchangeResponse(reader io.Reader) ([]ProfileInfo, error) {
	responseBytes, err := mcInstallConn.plistCodec.Decode(reader)
	if err != nil {
		return []ProfileInfo{}, err
	}

	dict, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return []ProfileInfo{}, err
	}
	identifiersIntf, ok := dict["OrderedIdentifiers"]
	if !ok {
		return []ProfileInfo{}, fmt.Errorf("invalid plist response, missing key 'OrderedIdentifiers' dump: %x", responseBytes)
	}
	identifiers, ok := identifiersIntf.([]interface{})
	if !ok {
		return []ProfileInfo{}, fmt.Errorf("identifiers should be array, dump: %x", responseBytes)
	}
	profiles := make([]ProfileInfo, len(identifiers))
	for i, id := range identifiers {
		idString, ok := id.(string)
		if !ok {
			return []ProfileInfo{}, fmt.Errorf("identifiers should be array of strings, dump: %x", responseBytes)
		}
		profile, err := parseProfile(idString, dict)
		if err != nil {
			return []ProfileInfo{}, err
		}
		profiles[i] = profile

	}

	return profiles, nil
}

func parseProfile(idString string, dict map[string]interface{}) (ProfileInfo, error) {
	result := ProfileInfo{}
	result.Identifier = idString
	manifestIntf, ok := dict["ProfileManifest"]
	if !ok {
		return result, fmt.Errorf("missing key ProfileManifest %+v", dict)
	}
	manifest, ok := manifestIntf.(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("ProfileManifest should be a map %+v", dict)
	}
	manifestIntf, ok = manifest[idString]
	if !ok {
		return result, fmt.Errorf("missing key %s %+v", idString, dict)
	}
	manifest, ok = manifestIntf.(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("%s should be a map %+v", idString, dict)
	}
	result.Manifest.IsActive, ok = manifest["IsActive"].(bool)
	if !ok {
		return result, fmt.Errorf("keyError %+v", dict)
	}
	result.Manifest.Description, ok = manifest["Description"].(string)
	if !ok {
		return result, fmt.Errorf("keyError %+v", dict)
	}
	result.Status, ok = dict["Status"].(string)
	if !ok {
		return result, fmt.Errorf("keyError %+v", dict)
	}

	metadataIntf, ok := dict["ProfileMetadata"]
	if !ok {
		return result, fmt.Errorf("missing key ProfileMetadata %+v", dict)
	}
	metadata, ok := metadataIntf.(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("ProfileMetadata should be a map %+v", dict)
	}
	metadataIntf, ok = metadata[idString]
	if !ok {
		return result, fmt.Errorf("missing key %s %+v", idString, dict)
	}
	metadata, ok = metadataIntf.(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("%s should be a map %+v", idString, dict)
	}

	result.Metadata.PayloadDescription, ok = metadata["PayloadDescription"].(string)
	if !ok {
		result.Metadata.PayloadDescription = ""
	}
	result.Metadata.PayloadDisplayName, ok = metadata["PayloadDisplayName"].(string)
	if !ok {
		return result, fmt.Errorf("keyError PayloadDisplayName %+v", dict)
	}
	result.Metadata.PayloadRemovalDisallowed, ok = metadata["PayloadRemovalDisallowed"].(bool)
	if !ok {
		return result, fmt.Errorf("keyError PayloadRemovalDisallowed %+v", dict)
	}
	result.Metadata.PayloadUUID, ok = metadata["PayloadUUID"].(string)
	if !ok {
		return result, fmt.Errorf("keyError PayloadUUID %+v", dict)
	}
	result.Metadata.PayloadVersion, ok = metadata["PayloadVersion"].(uint64)
	if !ok {
		return result, fmt.Errorf("keyError PayloadVersion %+v", dict)
	}

	return result, nil
}

func (mcInstallConn *Connection) EscalateUnsupervised() error {
	request := map[string]interface{}{
		"RequestType":           "Escalate",
		"SupervisorCertificate": []byte{0},
	}
	dict, err := mcInstallConn.sendAndReceive(request)
	if err != nil {
		return err
	}
	if !checkStatus(dict) {
		return fmt.Errorf("escalate response had error %+v", dict)
	}
	return nil
}

func (mcInstallConn *Connection) EscalateWithCertAndKey(supervisedPrivateKey interface{}, supervisionCert *x509.Certificate) error {
	request := map[string]interface{}{"RequestType": "Escalate", "SupervisorCertificate": supervisionCert.Raw}
	dict, err := mcInstallConn.sendAndReceive(request)
	if err != nil {
		return err
	}
	if !checkStatus(dict) {
		return fmt.Errorf("escalate response had error %+v", dict)
	}
	challengeInt, ok := dict["Challenge"]
	if !ok {
		return fmt.Errorf("missing key Challenge %+v", dict)
	}
	challenge, ok := challengeInt.([]byte)
	signedRequest, err := ios.Sign(challenge, supervisionCert, supervisedPrivateKey)
	if err != nil {
		return err
	}

	request = map[string]interface{}{"RequestType": "EscalateResponse", "SignedRequest": signedRequest}
	dict, err = mcInstallConn.sendAndReceive(request)
	if err != nil {
		return err
	}
	if !checkStatus(dict) {
		return fmt.Errorf("escalateresponse response had error %+v", dict)
	}
	request = map[string]interface{}{"RequestType": "ProceedWithKeybagMigration"}
	dict, err = mcInstallConn.sendAndReceive(request)
	if err != nil {
		return err
	}
	if !checkStatus(dict) {
		return fmt.Errorf("proceedWithKeybagMigration response had error %+v", dict)
	}
	return nil
}

func (mcInstallConn *Connection) Escalate(p12bytes []byte, p12Password string) error {
	supervisedPrivateKey, supervisionCert, err := pkcs12.Decode(p12bytes, p12Password)
	if err != nil {
		return err
	}
	return mcInstallConn.EscalateWithCertAndKey(supervisedPrivateKey, supervisionCert)
}

func checkStatus(response map[string]interface{}) bool {
	statusIntf, ok := response["Status"]
	if !ok {
		return false
	}
	status, ok := statusIntf.(string)
	if !ok {
		return false
	}
	if "Acknowledged" != status {
		return false
	}
	return true
}

func request(requestType string) map[string]interface{} {
	return map[string]interface{}{"RequestType": requestType}
}

func (mcInstallConn *Connection) sendAndReceive(request map[string]interface{}) (map[string]interface{}, error) {
	reader := mcInstallConn.deviceConn.Reader()
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return map[string]interface{}{}, err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return map[string]interface{}{}, err
	}
	responseBytes, err := mcInstallConn.plistCodec.Decode(reader)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return ios.ParsePlist(responseBytes)
}

func (mcInstallConn *Connection) HandleList() ([]ProfileInfo, error) {
	reader := mcInstallConn.deviceConn.Reader()
	request := map[string]interface{}{"RequestType": "GetProfileList"}
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return []ProfileInfo{}, err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return []ProfileInfo{}, err
	}
	return mcInstallConn.readExchangeResponse(reader)
}

// Close closes the underlying DeviceConnection
func (mcInstallConn *Connection) Close() error {
	return mcInstallConn.deviceConn.Close()
}

func (mcInstallConn *Connection) AddProfile(profilePlist []byte) error {
	return mcInstallConn.addProfile(profilePlist, "InstallProfile")
}

func (mcInstallConn *Connection) addProfile(profilePlist []byte, installcmd string) error {
	request := map[string]interface{}{"RequestType": installcmd, "Payload": profilePlist}
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return err
	}
	respBytes, err := mcInstallConn.plistCodec.Decode(mcInstallConn.deviceConn.Reader())
	if err != nil {
		return err
	}
	plist, err := ios.ParsePlist(respBytes)
	if checkStatus(plist) {
		return nil
	}
	log.Errorf("received add response %+v", plist)
	return fmt.Errorf("add failed")
}

func (mcInstallConn *Connection) RemoveProfile(identifier string) error {
	request := map[string]interface{}{"RequestType": "RemoveProfile", "ProfileIdentifier": identifier}
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return err
	}
	respBytes, err := mcInstallConn.plistCodec.Decode(mcInstallConn.deviceConn.Reader())
	if err != nil {
		return err
	}
	plist, err := ios.ParsePlist(respBytes)
	if checkStatus(plist) {
		return nil
	}
	log.Errorf("received remove response %+v", plist)
	return fmt.Errorf("remove failed")
}

func (mcInstallConn *Connection) AddProfileSupervised(profileFileBytes []byte, p12fileBytes []byte, password string) error {
	err := mcInstallConn.Escalate(p12fileBytes, password)
	if err != nil {
		return err
	}
	return mcInstallConn.addProfile(profileFileBytes, "InstallProfileSilent")
}
