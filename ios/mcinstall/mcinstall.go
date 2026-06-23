package mcinstall

import (
	"crypto/x509"
	"fmt"
	"io"

	"golang.org/x/crypto/pkcs12"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

const logModule = "go-ios/mcinstall"

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
	golog.Error("received add response", "module", logModule, "installcmd", installcmd, "response", plist)
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
	golog.Error("received remove response", "module", logModule, "identifier", identifier, "response", plist)
	return fmt.Errorf("remove failed")
}

func (mcInstallConn *Connection) AddProfileSupervised(profileFileBytes []byte, p12fileBytes []byte, password string) error {
	err := mcInstallConn.Escalate(p12fileBytes, password)
	if err != nil {
		return err
	}
	return mcInstallConn.addProfile(profileFileBytes, "InstallProfileSilent")
}

// WallpaperScreen identifies which screen(s) to apply a wallpaper to. Values
// match the MDM "Wallpaper" Settings command's "Where" field.
type WallpaperScreen int

const (
	WallpaperLockScreen WallpaperScreen = 1
	WallpaperHomeScreen WallpaperScreen = 2
	WallpaperBoth       WallpaperScreen = 3
)

// ParseWallpaperScreen maps "lock", "home" or "both" to the corresponding
// WallpaperScreen value. Matches the cfgutil --screen flag.
func ParseWallpaperScreen(s string) (WallpaperScreen, error) {
	switch s {
	case "lock":
		return WallpaperLockScreen, nil
	case "home":
		return WallpaperHomeScreen, nil
	case "both":
		return WallpaperBoth, nil
	}
	return 0, fmt.Errorf("invalid screen %q (expected lock|home|both)", s)
}

// SetWallpaper sends the "Wallpaper" Settings command on an already-escalated
// MCInstall connection. The image bytes should be a JPEG or PNG that the device
// will accept directly. Use SetWallpaperSupervised if escalation has not been
// performed yet.
//
// Note on `screen`: iOS 16+ unified Lock and Home wallpapers as a paired set,
// so the device applies the image to both screens regardless of the Where
// value. Apple's own cfgutil exhibits the same behavior. The argument is
// preserved for older iOS versions and forward compatibility.
func (mcInstallConn *Connection) SetWallpaper(image []byte, screen WallpaperScreen) error {
	request := map[string]interface{}{
		"RequestType": "Settings",
		"Settings": []interface{}{
			map[string]interface{}{
				"Item":  "Wallpaper",
				"Where": int(screen),
				"Image": image,
			},
		},
	}
	dict, err := mcInstallConn.sendAndReceive(request)
	if err != nil {
		return err
	}
	if !checkStatus(dict) {
		return fmt.Errorf("set-wallpaper response had error %+v", dict)
	}
	settings, ok := dict["Settings"].([]interface{})
	if !ok || len(settings) == 0 {
		return fmt.Errorf("set-wallpaper response missing Settings array %+v", dict)
	}
	inner, ok := settings[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("set-wallpaper Settings[0] not a dict %+v", dict)
	}
	if status, _ := inner["Status"].(string); status != "Acknowledged" {
		return fmt.Errorf("set-wallpaper item status: %+v", inner)
	}
	return nil
}

// SetWallpaperSupervised escalates with the supervisor identity (PKCS#12) and
// then issues the wallpaper command. Setting wallpapers requires a supervised
// device.
func (mcInstallConn *Connection) SetWallpaperSupervised(image []byte, screen WallpaperScreen, p12bytes []byte, p12Password string) error {
	if err := mcInstallConn.Escalate(p12bytes, p12Password); err != nil {
		return err
	}
	return mcInstallConn.SetWallpaper(image, screen)
}

// GetCloudConfiguration retrieves the cloud configuration from the device.
// This includes settings like SkipSetup options, supervision status, and organization info.
func (mcInstallConn *Connection) GetCloudConfiguration() (map[string]interface{}, error) {
	return mcInstallConn.sendAndReceive(request("GetCloudConfiguration"))
}

// FetchUnlockToken retrieves the passcode unlock token from an already-escalated
// MCInstall connection. The device must have no passcode set; it returns an error
// with DMCKeybagErrorDomain code 37002 if a passcode is active.
func (mcInstallConn *Connection) FetchUnlockToken() ([]byte, error) {
	response, err := mcInstallConn.sendAndReceive(request("RequestUnlockToken"))
	if err != nil {
		return nil, err
	}
	if !checkStatus(response) {
		return nil, fmt.Errorf("RequestUnlockToken failed: %+v", response)
	}
	token, ok := response["UnlockToken"].([]byte)
	if !ok {
		return nil, fmt.Errorf("UnlockToken missing or wrong type in response")
	}
	return token, nil
}

// FetchUnlockTokenSupervised escalates with the supervisor PKCS#12 identity and
// fetches the passcode unlock token. The device must have no passcode set.
// Store the returned token; it can be passed to ClearPasscodeSupervised later.
func (mcInstallConn *Connection) FetchUnlockTokenSupervised(p12bytes []byte, p12Password string) ([]byte, error) {
	if err := mcInstallConn.Escalate(p12bytes, p12Password); err != nil {
		return nil, err
	}
	return mcInstallConn.FetchUnlockToken()
}

// ClearPasscodeSupervised escalates with the supervisor PKCS#12 identity and
// clears the device lock passcode using a previously saved unlock token.
func (mcInstallConn *Connection) ClearPasscodeSupervised(p12bytes []byte, p12Password string, unlockToken []byte) error {
	if err := mcInstallConn.Escalate(p12bytes, p12Password); err != nil {
		return err
	}
	response, err := mcInstallConn.sendAndReceive(map[string]interface{}{
		"RequestType": "ClearPasscode",
		"UnlockToken": unlockToken,
	})
	if err != nil {
		return err
	}
	if !checkStatus(response) {
		return fmt.Errorf("ClearPasscode failed: %+v", response)
	}
	return nil
}

// ClearScreenTimePasswordSupervised escalates with the supervisor PKCS#12 identity
// and clears the Screen Time restrictions passcode. No unlock token is required.
func (mcInstallConn *Connection) ClearScreenTimePasswordSupervised(p12bytes []byte, p12Password string) error {
	if err := mcInstallConn.Escalate(p12bytes, p12Password); err != nil {
		return err
	}
	response, err := mcInstallConn.sendAndReceive(request("ClearRestrictionsPassword"))
	if err != nil {
		return err
	}
	if !checkStatus(response) {
		return fmt.Errorf("ClearRestrictionsPassword failed: %+v", response)
	}
	return nil
}
