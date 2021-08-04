package instruments

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

const conditionInducerChannelName = "com.apple.instruments.server.services.ConditionInducer"

type DeviceStateControl struct {
	controlChannel *dtx.Channel
	conn           *dtx.Connection
}

func NewDeviceStateControl(device ios.DeviceEntry) (*DeviceStateControl, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	conditionInducerChannel := dtxConn.RequestChannelIdentifier(conditionInducerChannelName, loggingDispatcher{dtxConn})
	return &DeviceStateControl{controlChannel: conditionInducerChannel, conn: dtxConn}, nil
}
