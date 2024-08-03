package devicestatemgmt

import (
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type iOSDevice struct {
	udid                    string
	configurationState      models.ConfigurationState
	physicalConnectionState models.PhysicalConnectionState
	sessionState            models.SessionState
	GoIosDeviceEntry        ios.DeviceEntry
	mux                     sync.Mutex
	name                    string
}

func NewIosDevice(udid string, name string) *iOSDevice {
	return &iOSDevice{
		udid:                    udid,
		configurationState:      models.ConfigurationState{},
		physicalConnectionState: models.PhysicalConnectionState{},
		sessionState:            models.SessionState{},
		GoIosDeviceEntry:        ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}},
		name:                    name,
		mux:                     sync.Mutex{},
	}
}

func (d *iOSDevice) Udid() string {
	return d.udid
}

func (d *iOSDevice) CopyState() (models.ConfigurationState, models.SessionState, models.PhysicalConnectionState) {
	d.mux.Lock()
	defer d.mux.Unlock()
	return d.configurationState, d.sessionState, d.physicalConnectionState
}

func discoverPhysicalConnection(deviceList *DeviceList) {
	deviceList.mutex.Lock()
	defer deviceList.mutex.Unlock()
	log.Info("fetching devices")
	log.Info("fetching ios devices")
	updateIos(deviceList)

	log.WithField("devices", deviceList.iosDevices).WithField("count", len(deviceList.iosDevices)).Info("found ios devices")
	log.WithField("devices", deviceList.iosDevices).Info("devices fetched")
}

func updateIos(deviceList *DeviceList) {
	devices, err := ios.ListDevices()
	if err != nil {
		log.WithField("err", err).Error("failed getting iOS devices")
		return
	}
	discoveredDevices := devices.DeviceList

	//check if devices we know are present on the host
	for _, knownDevice := range deviceList.iosDevices {
		deviceEntry, contains := containsDevice(discoveredDevices, knownDevice)
		_, _, oldPhysicalConnectionState := knownDevice.CopyState()
		if !contains {
			oldPhysicalConnectionState.ConnectionState = models.ConnectionStateDisconnected
			knownDevice.UpdatePhysicalConnectionState(oldPhysicalConnectionState)
			continue
		}
		knownDevice.GoIosDeviceEntry = deviceEntry
		oldPhysicalConnectionState.ConnectionState = models.ConnectionStateConnected
		oldPhysicalConnectionState.LastDetected = time.Now()
		knownDevice.UpdatePhysicalConnectionState(oldPhysicalConnectionState)
		knownDevice.mux.Lock()
		if knownDevice.name == "" || knownDevice.name == nameError {
			knownDevice.name = getName(deviceEntry)
		}
		knownDevice.mux.Unlock()
	}
	//create new devices for newly connected devices
	for _, deviceEntry := range discoveredDevices {
		_, isDiscovered := deviceList.FindIosDeviceEntry(deviceEntry)
		if isDiscovered {
			continue
		}
		//set up new device
		physicalConnectionState := models.PhysicalConnectionState{
			ConnectionState: models.ConnectionStateConnected,
			MetaInfo:        map[string]interface{}{},
			LastDetected:    time.Now(),
		}
		newDevice := iOSDevice{udid: deviceEntry.Properties.SerialNumber,
			configurationState:      models.NewDeviceStateIos(),
			GoIosDeviceEntry:        deviceEntry,
			name:                    getName(deviceEntry),
			physicalConnectionState: physicalConnectionState,
			sessionState: models.SessionState{
				SessionState:         models.SessionStateFree,
				SessionStateLastPing: time.Time{},
				MetaInfo:             map[string]interface{}{},
				SessionKey:           uuid.UUID{},
			},
		}

		log.WithField("udid", newDevice.udid).Info("new device detected")
		deviceList.iosDevices = append(deviceList.iosDevices, &newDevice)
	}

}

func containsDevice(devices []ios.DeviceEntry, device *iOSDevice) (ios.DeviceEntry, bool) {
	for _, d := range devices {
		if d.Properties.SerialNumber == device.udid {
			return d, true
		}
	}
	return ios.DeviceEntry{}, false
}

const nameError = "could not load name"

func getName(i ios.DeviceEntry) string {
	allValues, err := ios.GetValues(i)
	if err != nil {
		return nameError
	}

	return allValues.Value.DeviceName
}

func (d *iOSDevice) UpdatePhysicalConnectionState(newState models.PhysicalConnectionState) {
	d.mux.Lock()
	defer d.mux.Unlock()
	if !d.physicalConnectionState.Equals(newState) {
		log.WithField("udid", d.udid).WithField("old", d.physicalConnectionState).WithFields(log.Fields{"new": newState}).Info("device physical connection state change")
		d.physicalConnectionState = newState
	}
}
