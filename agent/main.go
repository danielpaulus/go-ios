package agent

import (
	"os"
	"os/signal"

	"github.com/danielpaulus/go-ios/agent/devicestatemgmt"
	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/orchestratorclient"
	"github.com/danielpaulus/go-ios/agent/restapi"
	log "github.com/sirupsen/logrus"
)

func Main() {
	devicePool, err := models.LoadOrInitPool()
	if err != nil {
		log.Errorf("error loading poolconfg:%v", err)
		return
	}
	log.Infof("Pool id: %s on host: %s starting..", devicePool.ID, devicePool.Hostname)
	log.Infof("Pool on http://%s:%s", devicePool.Ip, devicePool.Port)
	log.Infof("Last Known Devices: %v", devicePool.Devices)

	list := devicestatemgmt.NewDeviceList(devicePool.Devices)

	deviceConfig := os.Getenv("DEVICE_CONFIG")
	disable := "NONE" == deviceConfig

	managerCloseFunc := devicestatemgmt.StartDeviceStateManager(list, disable)
	defer managerCloseFunc()

	go restapi.StartApi(list)

	orchestratorclient.StartUpdatingOrchestrator()
	stopSignal := make(chan interface{})
	waitForSigInt(stopSignal)
	<-stopSignal

}

func waitForSigInt(stopSignalChannel chan interface{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Debugf("Signal received: %s", sig)
			var stopSignal interface{}
			stopSignalChannel <- stopSignal
		}
	}()
}
