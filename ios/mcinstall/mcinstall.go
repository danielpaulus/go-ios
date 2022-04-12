package mcinstall

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"

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
		return result, fmt.Errorf("keyError PayloadDescription %+v", dict)
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

//Close closes the underlying DeviceConnection
func (mcInstallConn *Connection) Close() error {
	return mcInstallConn.deviceConn.Close()
}

func (mcInstallConn *Connection) AddProfile(profilePlist []byte) error {
request := map[string]interface{}{"RequestType":"InstallProfile", "Payload": profilePlist}
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return err
	}
	respBytes, err := mcInstallConn.plistCodec.Decode(mcInstallConn.deviceConn.Reader())
	if err!=nil{
		return err
	}
	log.Infof("received install response %x", respBytes)
	return nil
}

func (mcInstallConn *Connection) RemoveProfile(identifier string) error {
	request := map[string]interface{}{"RequestType":"RemoveProfile", "ProfileIdentifier": identifier}
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return err
	}
	respBytes, err := mcInstallConn.plistCodec.Decode(mcInstallConn.deviceConn.Reader())
	if err!=nil{
		return err
	}
	log.Infof("received install response %x", respBytes)
	return nil
}