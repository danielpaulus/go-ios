package orchestratorclient

import (
	"time"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/utils"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

var stopSignal chan interface{}

func StartUpdatingOrchestrator() {
	log.Infof("starting orchestrator updates with %ds interval", utils.ORCHESTRATOR_UPDATE_FREQUENCY_SECONDS)
	go func() {
		for {
			select {
			case <-stopSignal:
				log.Info("shutting down device orchestrator job")
				return
			case <-time.After(time.Second * utils.ORCHESTRATOR_UPDATE_FREQUENCY_SECONDS):
				pool, err := models.LoadOrInitPool()
				if err != nil {
					log.Errorf("failed loading pool %v", err)
					break
				}
				err = UpdateState(pool)
				if err != nil {
					log.Errorf("failed sending state to Device Orchestrator: %v", err)
				}
				config, err := GetCloudconfig()
				if err != nil {
					log.Errorf("failed sending state to Device Orchestrator: %v", err)
				} else {
					models.UpdateConfigFromCloud(config)
				}
				log.Info("pulling SDP records for webRTC")
				_, err = DownloadSDPs()
				if err != nil {
					log.Errorf("failed pulling SDP records %v", err)
				} else {

				}
			}
		}
	}()
}

func StopUpdatingOrchestrator() {
	stopSignal <- nil
}
