package ios

type BatteryInfo struct {
	BatteryCurrentCapacity uint64
	BatteryIsCharging      bool
	ExternalChargeCapable  bool
	ExternalConnected      bool
	FullyCharged           bool
	GasGaugeCapability     bool
	HasBattery             bool
}

const batteryDomain = "com.apple.mobile.battery"

type Connection struct {
	deviceConn DeviceConnectionInterface
	plistCodec PlistCodec
}

func GetBatteryDiagnostics(device DeviceEntry) (BatteryInfo, error) {
	batteryConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return BatteryInfo{}, err
	}
	defer batteryConn.Close()

	capacityResp, err := batteryConn.GetValueForDomain("BatteryCurrentCapacity", batteryDomain)
	if err != nil {
		return BatteryInfo{}, err
	}
	chargingResp, err := batteryConn.GetValueForDomain("BatteryIsCharging", batteryDomain)
	if err != nil {
		return BatteryInfo{}, err
	}
	externalChargeResp, err := batteryConn.GetValueForDomain("ExternalChargeCapable", batteryDomain)
	if err != nil {
		return BatteryInfo{}, err
	}
	externalResp, err := batteryConn.GetValueForDomain("ExternalConnected", batteryDomain)
	if err != nil {
		return BatteryInfo{}, err
	}
	chargedResp, err := batteryConn.GetValueForDomain("FullyCharged", batteryDomain)
	if err != nil {
		return BatteryInfo{}, err
	}
	gasGaugeResp, err := batteryConn.GetValueForDomain("GasGaugeCapability", batteryDomain)
	if err != nil {
		return BatteryInfo{}, err
	}
	hasBatteryResp, err := batteryConn.GetValueForDomain("HasBattery", batteryDomain)
	if err != nil {
		return BatteryInfo{}, err
	}

	return BatteryInfo{BatteryCurrentCapacity: capacityResp.(uint64), BatteryIsCharging: chargingResp.(bool), ExternalChargeCapable: externalChargeResp.(bool), ExternalConnected: externalResp.(bool), FullyCharged: chargedResp.(bool), GasGaugeCapability: gasGaugeResp.(bool), HasBattery: hasBatteryResp.(bool)}, nil
}
