package devicestatemgmt

import (
	"github.com/danielpaulus/go-ios/agent/models"
	log "github.com/sirupsen/logrus"
)

func updateState(deviceList *DeviceList) {
	log.Info("updating device configuration")
	//this can take longer than the interval for device discovery, get a copy of the device list
	currentDevices := deviceList.GetCopy()
	for _, d := range currentDevices.iosDevices {
		if d.physicalConnectionState.ConnectionState == models.ConnectionStateDisconnected {
			continue
		}
		//go configureIosDevice(d)
	}
}

func (d *iOSDevice) UpdateConfigurationState(newState models.ConfigurationState) {
	d.mux.Lock()
	defer d.mux.Unlock()
	if d.configurationState.Equals(newState) {
		log.WithField("udid", d.udid).WithField("old", d.configurationState).WithFields(log.Fields{"new": newState}).Trace("device configurationState change, no changes")
		return
	}
	log.WithField("udid", d.udid).WithField("old", d.configurationState).WithFields(log.Fields{"new": newState}).Info("device configurationState change")
	d.configurationState = newState
}
