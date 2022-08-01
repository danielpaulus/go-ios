package instruments

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type channelDispatcher struct {
	messageChannel chan dtx.Message
	closeChannel   chan struct{}
}

func ListenAppStateNotifications(device ios.DeviceEntry) (func() (map[string]interface{}, error), func() error, error) {
	conn, err := connectInstruments(device)
	if err != nil {
		return nil, nil, err
	}
	dispatcher := channelDispatcher{messageChannel: make(chan dtx.Message), closeChannel: make(chan struct{})}
	conn.AddDefaultChannelReceiver(dispatcher)
	channel := conn.RequestChannelIdentifier(mobileNotificationsChannel, channelDispatcher{})
	resp, err := channel.MethodCall("setApplicationStateNotificationsEnabled:", true)
	if err != nil {
		log.Errorf("resp:%+v, %+v", resp, resp.Payload[0])
		return nil, nil, err
	}
	log.Debugf("appstatenotifications enabled successfully: %+v", resp)
	resp, err = channel.MethodCall("setMemoryNotificationsEnabled:", true)
	if err != nil {
		log.Errorf("resp:%+v, %+v", resp, resp.Payload[0])
		return nil, nil, err
	}
	log.Debugf("memory notifications enabled: %+v", resp)

	return dispatcher.Receive, dispatcher.Close, nil
}

func (dispatcher channelDispatcher) Receive() (map[string]interface{}, error) {

	var msg dtx.Message
L:
	for {
		select {
		case msg = <-dispatcher.messageChannel:
			if "applicationStateNotification:" == msg.Payload[0].(string) {
				break L
			}
		case <-dispatcher.closeChannel:
			return map[string]interface{}{}, io.EOF
		}
	}

	data, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	if err != nil {
		return map[string]interface{}{}, err
	}
	resp := data[0]
	return resp.(map[string]interface{}), nil

}

func (dispatcher *channelDispatcher) Close() error {
	select {
	case dispatcher.closeChannel <- struct{}{}:
		return nil
	case <-time.After(time.Second * 5):
		return fmt.Errorf("timeout")
	}
}

func (dispatcher channelDispatcher) Dispatch(msg dtx.Message) {
	dispatcher.messageChannel <- msg
}
