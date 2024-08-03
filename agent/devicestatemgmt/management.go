package devicestatemgmt

import (
	"time"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/utils"
	log "github.com/sirupsen/logrus"
)

var deviceList *DeviceList

func StartDeviceStateManager(currentDeviceList *DeviceList, disableConfiguration bool) func() error {
	deviceList = currentDeviceList
	closer := make(chan bool)
	closer1 := make(chan bool)
	go sessionTimeoutChecker(closer)
	go func() {
		for {
			discoverPhysicalConnection(deviceList)
			err := models.UpdateDeviceInfo(deviceList.GetCurrentInfo())
			if err != nil {
				log.Warn("error updating devicelist to local db", err)
			}
			if !disableConfiguration {
				updateState(deviceList)
			}
			select {
			case <-time.After(time.Second * utils.DEVICE_DISCOVERY_INTERVAL_SEC):
				break
			case <-closer1:
				return
			}
		}
	}()
	return func() error {
		closer <- true
		closer1 <- true
		return nil
	}

}
