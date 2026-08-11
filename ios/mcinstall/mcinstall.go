package mcinstall

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"strings"

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

// ErrPasscodeSet reports that the device already has a passcode, so the keybag refuses
// to mint an unlock token. Remove the passcode before running fetch-unlock-token.
var ErrPasscodeSet = errors.New("device has a passcode set; remove it before fetching an unlock token")

const (
	keybagErrorDomain    = "DMCKeybagErrorDomain"
	keybagErrPasscodeSet = 37002
)

// errorCode extracts an ErrorChain entry's numeric code. plist decoding yields
// different integer types depending on how the device encoded it.
func errorCode(entry map[string]interface{}) int64 {
	switch code := entry["ErrorCode"].(type) {
	case int64:
		return code
	case uint64:
		return int64(code)
	case int:
		return int64(code)
	case float64:
		return int64(code)
	}
	return 0
}

// describeFailure renders the actionable part of a rejected MCInstall response: the
// status plus each ErrorChain entry's domain, code and description.
//
// It deliberately never formats the whole response. RequestUnlockToken responses carry
// the unlock token, and ClearPasscode responses commonly echo the UnlockToken that was
// sent, so dumping them would write the token to every log sink the error reaches --
// defeating the 0o600 permissions the token file is otherwise protected with.
func describeFailure(response map[string]interface{}) string {
	status, _ := response["Status"].(string)
	if status == "" {
		status = "unknown"
	}
	parts := []string{"Status=" + status}
	if chain, ok := response["ErrorChain"].([]interface{}); ok {
		for _, entry := range chain {
			errorMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			domain, _ := errorMap["ErrorDomain"].(string)
			description, _ := errorMap["LocalizedDescription"].(string)
			parts = append(parts, fmt.Sprintf("%s(%d): %s", domain, errorCode(errorMap), description))
		}
	}
	return strings.Join(parts, "; ")
}

// hasErrorCode reports whether the response's ErrorChain contains the given domain and code.
func hasErrorCode(response map[string]interface{}, domain string, code int64) bool {
	chain, ok := response["ErrorChain"].([]interface{})
	if !ok {
		return false
	}
	for _, entry := range chain {
		errorMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if entryDomain, _ := errorMap["ErrorDomain"].(string); entryDomain == domain && errorCode(errorMap) == code {
			return true
		}
	}
	return false
}

// sendChecked sends request and verifies the device acknowledged it. On rejection the
// returned error carries only the redacted description, never the raw response.
func (mcInstallConn *Connection) sendChecked(request map[string]interface{}, name string) (map[string]interface{}, error) {
	response, err := mcInstallConn.sendAndReceive(request)
	if err != nil {
		return nil, err
	}
	if !checkStatus(response) {
		return response, fmt.Errorf("%s failed: %s", name, describeFailure(response))
	}
	return response, nil
}

// FetchUnlockToken retrieves the passcode unlock token from an already-escalated
// MCInstall connection. The device must have no passcode set: when one is active the
// keybag rejects the request and this returns an error wrapping ErrPasscodeSet
// (DMCKeybagErrorDomain code 37002).
func (mcInstallConn *Connection) FetchUnlockToken() ([]byte, error) {
	response, err := mcInstallConn.sendChecked(request("RequestUnlockToken"), "RequestUnlockToken")
	if err != nil {
		if hasErrorCode(response, keybagErrorDomain, keybagErrPasscodeSet) {
			return nil, fmt.Errorf("%w: %v", ErrPasscodeSet, err)
		}
		return nil, err
	}
	token, ok := response["UnlockToken"].([]byte)
	if !ok {
		return nil, fmt.Errorf("UnlockToken missing or wrong type in response")
	}
	return token, nil
}

// SecurityInfo returns the device's security status from an already-escalated MCInstall
// connection: passcode presence and policy compliance, lock grace periods, hardware
// encryption capabilities and management status. It is a read-only query and performs no
// keybag operation.
func (mcInstallConn *Connection) SecurityInfo() (map[string]interface{}, error) {
	response, err := mcInstallConn.sendChecked(request("SecurityInfo"), "SecurityInfo")
	if err != nil {
		return nil, err
	}
	info, ok := response["SecurityInfo"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("SecurityInfo missing or wrong type in response")
	}
	return info, nil
}

// PasscodePresent reports whether a passcode is currently set on the device, using the
// SecurityInfo query on an already-escalated MCInstall connection.
//
// This is the reliable way to answer "does this device have a passcode?". Lockdown's
// PasswordProtected value answers a different question — whether the device is currently
// locked and demanding a passcode — and reads false on an unlocked device that has one.
// Probing with FetchUnlockToken works too (it fails with ErrPasscodeSet) but mints a
// throwaway token and touches the keybag; this does neither.
func (mcInstallConn *Connection) PasscodePresent() (bool, error) {
	info, err := mcInstallConn.SecurityInfo()
	if err != nil {
		return false, err
	}
	present, ok := info["PasscodePresent"].(bool)
	if !ok {
		return false, fmt.Errorf("PasscodePresent missing or wrong type in SecurityInfo response")
	}
	return present, nil
}

// ClearPasscode removes the device lock passcode using a previously saved unlock
// token on an already-escalated MCInstall connection.
func (mcInstallConn *Connection) ClearPasscode(unlockToken []byte) error {
	_, err := mcInstallConn.sendChecked(map[string]interface{}{
		"RequestType": "ClearPasscode",
		"UnlockToken": unlockToken,
	}, "ClearPasscode")
	return err
}

// ClearScreenTimePassword clears the Screen Time restrictions passcode on an
// already-escalated MCInstall connection. No unlock token is required.
func (mcInstallConn *Connection) ClearScreenTimePassword() error {
	_, err := mcInstallConn.sendChecked(request("ClearRestrictionsPassword"), "ClearRestrictionsPassword")
	return err
}
