package instruments

import (
	"fmt"
	"io"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
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
	for {
		select {
		case msg := <-dispatcher.messageChannel:
			selector, result, err := toMap(msg)
			if "applicationStateNotification:" == selector && err == nil {
				return result, nil
			}
			if err != nil {
				log.Debugf("error extracting message %+v, %v", msg, err)
			}
		case <-dispatcher.closeChannel:
			return map[string]interface{}{}, io.EOF
		}
	}
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
