package jobs

import (
	"time"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/orchestratorclient"
	"github.com/danielpaulus/go-ios/agent/utils"
	"github.com/danielpaulus/go-ios/agent/wrtc"
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
				err = orchestratorclient.UpdateState(pool)
				if err != nil {
					log.Errorf("failed sending state to Device Orchestrator: %v", err)
				}
				config, err := orchestratorclient.GetCloudconfig()
				if err != nil {
					log.Errorf("failed sending state to Device Orchestrator: %v", err)
				} else {
					models.UpdateConfigFromCloud(config)
				}
				log.Info("pulling SDP records for webRTC")
				sdps, err := orchestratorclient.DownloadSDPs()
				if err != nil {
					log.Errorf("failed pulling SDP records %v", err)
				} else {
					sdpAnswers, err := wrtc.CreateSDPAnswer(sdps)
					if err != nil {
						log.Errorf("failed answering SDP records %v", err)
					}
					if len(sdpAnswers) == 0 {
						continue
					}
					log.Info("pushing sdp answer")
					err = orchestratorclient.PushSDPAnswers(sdpAnswers)
					if err != nil {
						log.Errorf("failed answering SDP records %v", err)
					}
				}
			}
		}
	}()
}

func StopUpdatingOrchestrator() {
	stopSignal <- nil
}
