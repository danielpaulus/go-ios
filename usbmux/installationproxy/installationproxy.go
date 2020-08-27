package installationproxy

import (
	"bytes"

	"github.com/danielpaulus/go-ios/usbmux"
	"howett.net/plist"
)

const serviceName = "com.apple.mobile.installation_proxy"

type Connection struct {
	deviceConn usbmux.DeviceConnectionInterface
	plistCodec *usbmux.PlistCodec
}

func New(deviceID int, udid string) (*Connection, error) {
	deviceConn, err := usbmux.ConnectToService(deviceID, udid, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn, plistCodec: usbmux.NewPlistCodec()}, nil
}
func (conn *Connection) BrowseUserApps() (BrowseResponse, error) {
	return conn.browseApps(browseUserApps())
}

func (conn *Connection) BrowseSystemApps() (BrowseResponse, error) {
	return conn.browseApps(browseSystemApps())
}

func (conn *Connection) browseApps(request interface{}) (BrowseResponse, error) {
	reader := conn.deviceConn.Reader()
	bytes, err := conn.plistCodec.Encode(request)
	if err != nil {
		return BrowseResponse{}, err
	}
	conn.deviceConn.Send(bytes)
	response, err := conn.plistCodec.Decode(reader)
	ifa, err := plistFromBytes(response)
	if err != nil {
		return BrowseResponse{}, err
	}
	return ifa, nil
}
func plistFromBytes(plistBytes []byte) (BrowseResponse, error) {
	var browseResponse BrowseResponse
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))

	err := decoder.Decode(&browseResponse)
	if err != nil {
		return browseResponse, err
	}
	return browseResponse, nil
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

type BrowseResponse struct {
	CurrentIndex  uint64
	CurrentAmount uint64
	Status        string
	CurrentList   []AppInfo
}
type AppInfo struct {
	ApplicationDSID              int
	ApplicationType              string
	CFBundleDisplayName          string
	CFBundleExecutable           string
	CFBundleIdentifier           string
	CFBundleName                 string
	CFBundleShortVersionString   string
	CFBundleVersion              string
	Container                    string
	Entitlements                 map[string]interface{}
	EnvironmentVariables         map[string]interface{}
	MinimumOSVersion             string
	Path                         string
	ProfileValidated             bool
	SBAppTags                    []string
	SignerIdentity               string
	UIDeviceFamily               []int
	UIRequiredDeviceCapabilities []string
}
