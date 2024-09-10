package instruments

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const sysmontapName = "com.apple.instruments.server.services.sysmontap"

type SysmontapService struct {
	channel *dtx.Channel
	conn    *dtx.Connection
}

func NewSysmontapService(device ios.DeviceEntry) (*SysmontapService, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(sysmontapName, loggingDispatcher{dtxConn})
	return &SysmontapService{channel: processControlChannel, conn: dtxConn}, nil
}

// Close closes up the DTX connection
func (d *SysmontapService) Close() {
	d.conn.Close()
}

func (s SysmontapService) Start() ([]interface{}, error) {
	fetchDataNow, err := s.channel.MethodCall("start")

	if err != nil {
		return nil, err
	}

	log.Info(fetchDataNow)

	return nil, nil
}

func (s SysmontapService) SetConfig(sysAttrs []interface{}) error {
	config := make(map[string]interface{})
	config["ur"] = 1000
	config["bm"] = 0
	config["cpuUsage"] = true
	config["physFootprint"] = true
	config["sampleInterval"] = 1000 * 1000000
	config["sysAttrs"] = sysAttrs

	_, err := s.channel.MethodCall("setConfig:", config)

	if err != nil {
		return err
	}

	return nil
}
