package usbmux

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
