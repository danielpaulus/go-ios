package instruments

import (
	"fmt"
	"io"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
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
		golog.Error("setApplicationStateNotificationsEnabled failed", "module", logModule, "udid", device.Properties.SerialNumber, "response", resp)
		return nil, nil, err
	}
	golog.Debug("appstatenotifications enabled successfully", "module", logModule, "udid", device.Properties.SerialNumber, "response", resp)
	resp, err = channel.MethodCall("setMemoryNotificationsEnabled:", true)
	if err != nil {
		golog.Error("setMemoryNotificationsEnabled failed", "module", logModule, "udid", device.Properties.SerialNumber, "response", resp)
		return nil, nil, err
	}
	golog.Debug("memory notifications enabled", "module", logModule, "udid", device.Properties.SerialNumber, "response", resp)

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
				golog.Debug("error extracting message", "module", logModule, "message", msg, "error", err)
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
