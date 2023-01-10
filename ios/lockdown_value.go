package ios

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/semver"

	plist "howett.net/plist"
)

//BasebandKeyHashInformationType containing some baseband related
//data directly from the ios device
type BasebandKeyHashInformationType struct {
	AKeyStatus int
	SKeyHash   []byte
	SKeyStatus int
}

//NonVolatileRAMType contains some internal device info
//and can be retrieved by getting all values
type NonVolatileRAMType struct {
	AutoBoot              []byte `plist:"auto-boot"`
	BacklightLevel        []byte `plist:"backlight-level"`
	BootArgs              string `plist:"boot-args"`
	Bootdelay             []byte `plist:"bootdelay"`
	ComAppleSystemTz0Size []byte `plist:"com.apple.System.tz0-size"`
	OblitBegins           []byte `plist:"oblit-begins"`
	Obliteration          []byte `plist:"obliteration"`
}

//GetAllValuesResponse just the wrapper for AllValuesType
type GetAllValuesResponse struct {
	Request string
	Value   AllValuesType
}

//AllValuesType contains all possible values that can be requested from
//LockDown
type AllValuesType struct {
	ActivationState                             string
	ActivationStateAcknowledged                 bool
	BasebandActivationTicketVersion             string
	BasebandCertID                              int `plist:"BasebandCertId"`
	BasebandChipID                              int
	BasebandKeyHashInformation                  BasebandKeyHashInformationType
	BasebandMasterKeyHash                       string
	BasebandRegionSKU                           []byte
	BasebandSerialNumber                        []byte
	BasebandStatus                              string
	BasebandVersion                             string
	BluetoothAddress                            string
	BoardID                                     int `plist:"BoardId"`
	BrickState                                  bool
	BuildVersion                                string
	CPUArchitecture                             string
	CarrierBundleInfoArray                      []interface{}
	CertID                                      int
	ChipID                                      int
	ChipSerialNo                                []byte
	DeviceClass                                 string
	DeviceColor                                 string
	DeviceName                                  string
	DieID                                       int
	EthernetAddress                             string
	FirmwareVersion                             string
	FusingStatus                                int
	HardwareModel                               string
	HardwarePlatform                            string
	HasSiDP                                     bool
	HostAttached                                bool
	InternationalMobileEquipmentIdentity        string
	MLBSerialNumber                             string
	MobileEquipmentIdentifier                   string
	MobileSubscriberCountryCode                 string
	MobileSubscriberNetworkCode                 string
	ModelNumber                                 string
	NonVolatileRAM                              NonVolatileRAMType
	PartitionType                               string
	PasswordProtected                           bool
	PkHash                                      []byte
	ProductName                                 string
	ProductType                                 string
	ProductVersion                              string
	ProductionSOC                               bool
	ProtocolVersion                             string
	ProximitySensorCalibration                  []byte
	RegionInfo                                  string
	SBLockdownEverRegisteredKey                 bool
	SIMStatus                                   string
	SIMTrayStatus                               string
	SerialNumber                                string
	SoftwareBehavior                            []byte
	SoftwareBundleVersion                       string
	SupportedDeviceFamilies                     []int
	TelephonyCapability                         bool
	TimeIntervalSince1970                       float64
	TimeZone                                    string
	TimeZoneOffsetFromUTC                       float64
	TrustedHostAttached                         bool
	UniqueChipID                                uint64
	UniqueDeviceID                              string
	UseRaptorCerts                              bool
	Uses24HourClock                             bool
	WiFiAddress                                 string
	WirelessBoardSerialNumber                   string
	KCTPostponementInfoPRIVersion               string `plist:"kCTPostponementInfoPRIVersion"`
	KCTPostponementInfoPRLName                  int    `plist:"kCTPostponementInfoPRLName"`
	KCTPostponementInfoServiceProvisioningState bool   `plist:"kCTPostponementInfoServiceProvisioningState"`
	KCTPostponementStatus                       string `plist:"kCTPostponementStatus"`
}

type valueRequest struct {
	Label   string
	Key     string `plist:"Key,omitempty"`
	Request string
	Domain  string `plist:"Domain,omitempty"`
	Value   interface{} `plist:"Value,omitempty"`
}

func newGetValue(key string) valueRequest {
	data := valueRequest{
		Label:   "go.ios.control",
		Key:     key,
		Request: "GetValue",
	}
	return data
}

func newSetValue(key string, domain string, value interface{}) valueRequest {
	data := valueRequest{
		Label:   "go.ios.control",
		Key:     key,
		Domain:  domain,
		Request: "SetValue",
		Value:   value,
	}
	return data
}

//GetValues retrieves a GetAllValuesResponse containing all values lockdown returns
func (lockDownConn *LockDownConnection) GetValues() (GetAllValuesResponse, error) {
	err := lockDownConn.Send(newGetValue(""))
	if err != nil {
		return GetAllValuesResponse{}, err
	}
	resp, err := lockDownConn.ReadMessage()
	if err != nil {
		return GetAllValuesResponse{}, err
	}
	response := getAllValuesResponseFromBytes(resp)
	return response, nil
}

//GetProductVersion gets the iOS version of a device
func GetProductVersion(device DeviceEntry) (*semver.Version, error) {
	lockdownConnection, err := ConnectLockdownWithSession(device)
	if err != nil {
		return &semver.Version{}, err
	}
	defer lockdownConnection.Close()
	version, err := lockdownConnection.GetProductVersion()
	if err != nil {
		return &semver.Version{}, err
	}
	v, err := semver.NewVersion(version)
	return v, err
}

//GetWifiMac gets the static MAC address of the device WiFi.
//note: this does not report the dynamic MAC if you enable the
//"automatic WiFi address" feature.
func GetWifiMac(device DeviceEntry) (string, error) {
	lockdownConnection, err := ConnectLockdownWithSession(device)
	if err != nil {
		return "", err
	}
	defer lockdownConnection.Close()
	wifiMac, err := lockdownConnection.GetValue("WiFiAddress")
	if err != nil {
		return "", err
	}

	return wifiMac.(string), err
}

//GetProductVersion returns the ProductVersion of the device f.ex. "10.3"
func (lockDownConn *LockDownConnection) GetProductVersion() (string, error) {
	msg, err := lockDownConn.GetValue("ProductVersion")
	if err != nil {
		return "", fmt.Errorf("Failed getting ProductVersion: %v", err)
	}
	result, ok := msg.(string)
	if !ok {
		return "", fmt.Errorf("could not convert response to string: %+v", msg)
	}
	return result, nil
}

//GetValue gets and returns the string value for the lockdown key
func (lockDownConn *LockDownConnection) GetValue(key string) (interface{}, error) {
	lockDownConn.Send(newGetValue(key))
	resp, err := lockDownConn.ReadMessage()
	response := getValueResponsefromBytes(resp)
	return response.Value, err
}

//GetValueForDomain gets and returns the string value for the lockdown key and domain
func (lockDownConn *LockDownConnection) GetValueForDomain(key string, domain string) (interface{}, error) {
	gv := newGetValue(key)
	gv.Domain = domain
	lockDownConn.Send(gv)
	resp, err := lockDownConn.ReadMessage()
	response := getValueResponsefromBytes(resp)
	return response.Value, err
}

//SetValueForDomain sets the string value for the lockdown key and domain. If the device returns an error it will be returned as a go error.
func (lockDownConn *LockDownConnection) SetValueForDomain(key string, domain string, value interface{}) error {
	gv := newSetValue(key, domain, value)
	lockDownConn.Send(gv)
	resp, err := lockDownConn.ReadMessage()
	response := getValueResponsefromBytes(resp)
	if response.Error != "" {
		return fmt.Errorf("Failed setting '%s' to '%s' with err: %s", key, value, response.Error)
	}
	return err
}

//ValueResponse contains the response for a GetValue or SetValue Request
type ValueResponse struct {
	Key     string
	Request string
	Error   string
	Domain  string
	Value   interface{}
}

func getValueResponsefromBytes(plistBytes []byte) ValueResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var getValueResponse ValueResponse
	_ = decoder.Decode(&getValueResponse)
	return getValueResponse
}

func getAllValuesResponseFromBytes(plistBytes []byte) GetAllValuesResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var getAllValuesResponse GetAllValuesResponse
	_ = decoder.Decode(&getAllValuesResponse)
	return getAllValuesResponse
}

//GetValuesPlist returns the full lockdown values response as a map, so it can be converted to JSON easily.
func GetValuesPlist(device DeviceEntry) (map[string]interface{}, error) {
	lockdownConnection, err := ConnectLockdownWithSession(device)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer lockdownConnection.Close()
	err = lockdownConnection.Send(newGetValue(""))
	if err != nil {
		return map[string]interface{}{}, err
	}
	resp, err := lockdownConnection.ReadMessage()
	if err != nil {
		return map[string]interface{}{}, err
	}
	plist, err := ParsePlist(resp)
	if err != nil {
		return map[string]interface{}{}, err
	}
	plist, ok := plist["Value"].(map[string]interface{})
	if !ok {
		return plist, fmt.Errorf("Failed converting lockdown response:%+v", plist)
	}
	return plist, err
}

//GetValues returns all values of deviceInformation from lockdown
func GetValues(device DeviceEntry) (GetAllValuesResponse, error) {
	lockdownConnection, err := ConnectLockdownWithSession(device)
	if err != nil {
		return GetAllValuesResponse{}, err
	}
	defer lockdownConnection.Close()

	allValues, err := lockdownConnection.GetValues()
	if err != nil {
		return GetAllValuesResponse{}, err
	}
	return allValues, nil
}
