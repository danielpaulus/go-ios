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

// notificationBacklog buffers messageChannel so the device's initial burst of
// applicationStateNotification: pushes (one per already-running process, sent
// as soon as setApplicationStateNotificationsEnabled: takes effect) doesn't
// block the connection's single reader goroutine before the caller has even
// gotten the Receive func back to start draining it. An unbuffered channel
// here deadlocks setApplicationStateNotificationsEnabled/
// setMemoryNotificationsEnabled's own reply, since the reader can't process it
// while stuck delivering the backlog to a channel nobody is reading yet.
const notificationBacklog = 256

func ListenAppStateNotifications(device ios.DeviceEntry) (func() (map[string]interface{}, error), func() error, error) {
	dispatcher := channelDispatcher{messageChannel: make(chan dtx.Message, notificationBacklog), closeChannel: make(chan struct{})}
	// applicationStateNotification:/memoryLevelNotification: arrive as
	// unsolicited pushes on the global channel (channel code 0), not on the
	// channel requested below — AddDefaultChannelReceiver binds channel code
	// -1/4294967295, which these never use, so it never actually delivers
	// them. Registering as the connection's MessageDispatcher is what
	// GlobalDispatcher.Dispatch forwards global-channel pushes to.
	conn, err := connectInstrumentsWithMsgDispatcher(device, dispatcher)
	if err != nil {
		return nil, nil, err
	}
	channel := conn.RequestChannelIdentifier(mobileNotificationsChannel, loggingDispatcher{conn})
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
