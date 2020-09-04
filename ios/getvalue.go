package ios

import (
	"bytes"

	log "github.com/sirupsen/logrus"
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
	BasebandCertId                              int
	BasebandChipID                              int
	BasebandKeyHashInformation                  BasebandKeyHashInformationType
	BasebandMasterKeyHash                       string
	BasebandRegionSKU                           []byte
	BasebandSerialNumber                        []byte
	BasebandStatus                              string
	BasebandVersion                             string
	BluetoothAddress                            string
	BoardId                                     int
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
	TimeZoneOffsetFromUTC                       int
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

type getValue struct {
	Label   string
	Key     string `plist:"Key,omitempty"`
	Request string
}

func newGetValue(key string) getValue {
	data := getValue{
		Label:   "go.ios.control",
		Key:     key,
		Request: "GetValue",
	}
	return data
}

//GetValues retrieves a GetAllValuesResponse containing all values lockdown returns
func (lockDownConn *LockDownConnection) GetValues() (GetAllValuesResponse, error) {
	lockDownConn.Send(newGetValue(""))
	resp, err := lockDownConn.ReadMessage()

	response := getAllValuesResponseFromBytes(resp)
	return response, err
}

//GetProductVersion returns the ProductVersion of the device f.ex. "10.3"
func (lockDownConn *LockDownConnection) GetProductVersion() string {
	msg, err := lockDownConn.GetValue("ProductVersion")
	if err != nil {
		log.Fatal("Failed getting ProductVersion", err)
	}
	return msg.(string)
}

//GetValue gets and returns the string value for the lockdown key
func (lockDownConn *LockDownConnection) GetValue(key string) (interface{}, error) {
	lockDownConn.Send(newGetValue(key))
	resp, err := lockDownConn.ReadMessage()
	response := getValueResponsefromBytes(resp)
	return response.Value, err
}

//GetValueResponse contains the response for a GetValue Request
type GetValueResponse struct {
	Key     string
	Request string
	Value   interface{}
}

func getValueResponsefromBytes(plistBytes []byte) GetValueResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var getValueResponse GetValueResponse
	_ = decoder.Decode(&getValueResponse)
	return getValueResponse
}

func getAllValuesResponseFromBytes(plistBytes []byte) GetAllValuesResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var getAllValuesResponse GetAllValuesResponse
	_ = decoder.Decode(&getAllValuesResponse)
	return getAllValuesResponse
}

//GetValues returns all values of deviceInformation from lockdown
func GetValues(device DeviceEntry) GetAllValuesResponse {
	muxConnection := NewUsbMuxConnection(NewDeviceConnection(DefaultUsbmuxdSocket))
	defer muxConnection.ReleaseDeviceConnection()

	pairRecord := muxConnection.ReadPair(device.Properties.SerialNumber)

	lockdownConnection, err := muxConnection.ConnectLockdown(device.DeviceID)
	if err != nil {
		log.Fatal(err)
	}
	lockdownConnection.StartSession(pairRecord)

	allValues, err := lockdownConnection.GetValues()
	if err != nil {
		log.Fatal(err)
	}
	lockdownConnection.StopSession()
	return allValues
}
