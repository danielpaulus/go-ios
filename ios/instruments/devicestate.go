package instruments

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const conditionInducerChannelName = "com.apple.instruments.server.services.ConditionInducer"

type DeviceStateControl struct {
	controlChannel *dtx.Channel
	conn           *dtx.Connection
}

func NewDeviceStateControl(device ios.DeviceEntry) (*DeviceStateControl, error) {
	dtxConn, err := dtx.NewConnection(device, serviceName)
	if err != nil {
		log.Debugf("Failed connecting to %s, trying %s", serviceName, serviceNameiOS14)
		dtxConn, err = dtx.NewConnection(device, serviceNameiOS14)
		if err != nil {
			return nil, err
		}
	}
	conditionInducerChannel := dtxConn.RequestChannelIdentifier(conditionInducerChannelName, loggingDispatcher{dtxConn})
	return &DeviceStateControl{controlChannel: conditionInducerChannel, conn: dtxConn}, nil
}
