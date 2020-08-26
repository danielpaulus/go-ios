package installationproxy

import (
	"bytes"

	"github.com/danielpaulus/go-ios/usbmux"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

const serviceName = "com.apple.mobile.installation_proxy"

type Connection struct {
	deviceConn usbmux.DeviceConnectionInterface
	plistCodec *usbmux.PlistCodec
}

func New(deviceID int, udid string, pairRecord usbmux.PairRecord) (*Connection, error) {
	deviceConn, err := usbmux.ConnectToService(deviceID, udid, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn, plistCodec: usbmux.NewPlistCodec()}, nil
}

func (conn *Connection) AllValues() (interface{}, error) {
	reader := conn.deviceConn.Reader()
	bytes, err := conn.plistCodec.Encode(browseUserApps())
	if err != nil {
		return nil, err
	}
	conn.deviceConn.Send(bytes)
	response, err := conn.plistCodec.Decode(reader)
	ifa, err := plistFromBytes(response)
	log.Info(ifa)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func plistFromBytes(plistBytes []byte) (interface{}, error) {
	var test interface{}
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))

	err := decoder.Decode(&test)
	if err != nil {
		return test, err
	}
	return test, nil
}
func browseSystemApps() map[string]interface{} {
	returnAttributes := []string{
		"ApplicationDSID",
		"ApplicationType",
		"CFBundleDisplayName",
		"CFBundleExecutable",
		"CFBundleIdentifier",
		"CFBundleName",
		"CFBundleShortVersionString",
		"CFBundleVersion",
		"Container",
		"Entitlements",
		"EnvironmentVariables",
		"MinimumOSVersion",
		"Path",
		"ProfileValidated",
		"SBAppTags",
		"SignerIdentity",
		"UIDeviceFamily",
		"UIRequiredDeviceCapabilities",
	}
	clientOptions := map[string]interface{}{
		"ApplicationType":  "System",
		"ReturnAttributes": returnAttributes,
	}
	return map[string]interface{}{"ClientOptions": clientOptions, "Command": "Browse"}
}

func browseUserApps() map[string]interface{} {
	returnAttributes := []string{
		"ApplicationDSID",
		"ApplicationType",
		"CFBundleDisplayName",
		"CFBundleExecutable",
		"CFBundleIdentifier",
		"CFBundleName",
		"CFBundleShortVersionString",
		"CFBundleVersion",
		"Container",
		"Entitlements",
		"EnvironmentVariables",
		"MinimumOSVersion",
		"Path",
		"ProfileValidated",
		"SBAppTags",
		"SignerIdentity",
		"UIDeviceFamily",
		"UIRequiredDeviceCapabilities",
	}
	clientOptions := map[string]interface{}{
		"ApplicationType":          "User",
		"ReturnAttributes":         returnAttributes,
		"ShowLaunchProhibitedApps": true,
	}
	return map[string]interface{}{"ClientOptions": clientOptions, "Command": "Browse"}
}
