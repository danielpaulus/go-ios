package instruments

import (
	"time"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

func IterOpenglData(device ios.DeviceEntry) (func() (map[string]interface{}, error), func() error, error) {
	conn, err := connectInstruments(device)
	if err != nil {
		return nil, nil, err
	}
	channel := conn.RequestChannelIdentifier(GraphicsOpenGlChannel, channelDispatcher{})

	resp, err := channel.MethodCall("startSamplingAtTimeInterval:", 0)
	if err != nil {
		log.Errorf("resp:%+v", resp)
		return nil, nil, err
	}
	time.Sleep(time.Duration(5) * time.Second)
	channel.MethodCall("stopSampling:")
	conn.Close()
	// return dispatcher.Receive, dispatcher.Close, nil
	return nil, nil, nil

}
