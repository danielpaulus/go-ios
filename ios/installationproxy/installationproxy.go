package installationproxy

import (
	"bytes"

	ios "github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
)

const serviceName = "com.apple.mobile.installation_proxy"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

func (c *Connection) Close() {
	c.deviceConn.Close()
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device.DeviceID, device.Properties.SerialNumber, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn, plistCodec: ios.NewPlistCodec()}, nil
}
func (conn *Connection) BrowseUserApps() ([]AppInfo, error) {
	return conn.browseApps(browseUserApps())
}

func (conn *Connection) BrowseSystemApps() ([]AppInfo, error) {
	return conn.browseApps(browseSystemApps())
}

func (conn *Connection) browseApps(request interface{}) ([]AppInfo, error) {
	reader := conn.deviceConn.Reader()
	bytes, err := conn.plistCodec.Encode(request)
	if err != nil {
		return make([]AppInfo, 0), err
	}
	conn.deviceConn.Send(bytes)
	stillReceiving := true
	responses := make([]BrowseResponse, 0)
	size := uint64(0)
	for stillReceiving {
		response, err := conn.plistCodec.Decode(reader)
		ifa, err := plistFromBytes(response)
		stillReceiving = "Complete" != ifa.Status
		if err != nil {
			return make([]AppInfo, 0), err
		}
		size += ifa.CurrentAmount
		responses = append(responses, ifa)
	}
	appinfos := make([]AppInfo, size)

	for _, v := range responses {
		copy(appinfos[v.CurrentIndex:], v.CurrentList)

	}
	return appinfos, nil
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
