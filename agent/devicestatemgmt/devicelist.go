package devicestatemgmt

import (
	"sync"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/ios"
)

type DeviceList struct {
	iosDevices []*iOSDevice
	mutex      sync.Mutex
}

func NewDeviceList(d []models.DeviceInfo) *DeviceList {
	result := DeviceList{
		iosDevices: []*iOSDevice{},
	}
	for _, d := range d {

		if d.DeviceType == models.DeviceTypeIos {
			result.iosDevices = append(result.iosDevices,
				NewIosDevice(d.Serial, d.Name))
		}
	}
	return &result
}

func (l *DeviceList) FindIosDeviceEntry(d ios.DeviceEntry) (*iOSDevice, bool) {
	return l.FindIosDeviceByUdid(d.Properties.SerialNumber)
}

func (l *DeviceList) FindIosDeviceByUdid(udid string) (*iOSDevice, bool) {
	for _, iosDevice := range l.iosDevices {
		if iosDevice.udid == udid {
			return iosDevice, true
		}
	}
	return &iOSDevice{}, false
}

func (l *DeviceList) GetCopy() DeviceList {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	cIosDevices := make([]*iOSDevice, len(l.iosDevices))
	for i, d := range l.iosDevices {
		cIosDevices[i] = d
	}

	result := DeviceList{
		iosDevices: cIosDevices,
	}
	return result
}

func (l *DeviceList) GetCurrentInfo() []models.DeviceInfo {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	iosDeviceCount := len(l.iosDevices)
	devices := make([]models.DeviceInfo, iosDeviceCount)

	for i, d := range l.iosDevices {
		devices[i] = models.DeviceInfo{
			Serial:                  d.udid,
			SessionState:            d.sessionState,
			PhysicalConnectionState: d.physicalConnectionState,
			ConfigurationState:      d.configurationState,
			Name:                    d.name,
			DeviceType:              models.DeviceTypeIos,
		}
	}

	return devices
}
