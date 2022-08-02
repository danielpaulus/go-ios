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

func toMap(msg dtx.Message) (string, map[string]interface{}, error) {
	if len(msg.Payload) != 1 {
		return "", map[string]interface{}{}, fmt.Errorf("error extracting, msg %+v has payload size !=1", msg)
	}
	selector, ok := msg.Payload[0].(string)
	if !ok {
		return "", map[string]interface{}{}, fmt.Errorf("error extracting, msg %+v payload: %+v wasn't a string", msg, msg.Payload[0])
	}
	args := msg.Auxiliary.GetArguments()
	if len(args) == 0 {
		return "", map[string]interface{}{}, fmt.Errorf("error extracting, msg %+v has an empty auxiliary dictionary", msg)
	}

	data, ok := args[0].([]byte)
	if !ok {
		return "", map[string]interface{}{}, fmt.Errorf("error extracting, msg %+v invalid aux", msg)
	}

	unarchived, err := nskeyedarchiver.Unarchive(data)
	if err != nil {
		return "", map[string]interface{}{}, err
	}
	if len(unarchived) == 0 {
		return "", map[string]interface{}{}, fmt.Errorf("error extracting, msg %+v invalid aux", msg)
	}

	aux, ok := unarchived[0].(map[string]interface{})
	if !ok {
		return "", map[string]interface{}{}, fmt.Errorf("error extracting, msg %+v auxiliary: %+v didn't contain a map[string]interface{}", msg, msg.Payload[0])
	}

	return selector, aux, nil
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
