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

// AppStateEvent is a parsed applicationStateNotification: push — a process
// transitioning between states such as "Running", "Suspended", or
// "Terminated"/"Exited".
type AppStateEvent struct {
	ProcessName string
	Pid         uint64
	State       string
}

// parseAppStateEvent extracts an AppStateEvent from a raw notification map
// (as returned by ListenAppStateNotifications' Receive func), reporting
// false if the expected fields aren't present/well-typed.
func parseAppStateEvent(notification map[string]interface{}) (AppStateEvent, bool) {
	name, _ := notification["appName"].(string)
	if name == "" {
		return AppStateEvent{}, false
	}
	pid, ok := toUint64(notification["pid"])
	if !ok {
		return AppStateEvent{}, false
	}
	state, _ := notification["state_description"].(string)
	return AppStateEvent{ProcessName: name, Pid: pid, State: state}, true
}

// ListenAndKill opens one shared DTX connection exposing both the app-state
// notification stream (see ListenAppStateNotifications) and process-kill
// capability (see ProcessControl.KillProcess) — the device's instruments
// service only tolerates a single connection per client, so composing
// ListenAppStateNotifications and NewProcessControl, which would each open
// their own, fails intermittently.
//
// receive blocks until the next AppStateEvent, or returns an error once
// closeFunc has been called or the connection drops. kill kills a pid over
// the same connection. Callers decide what to do with each event — this is
// deliberately policy-free (no blocklist/matching), leaving that to the
// caller.
func ListenAndKill(device ios.DeviceEntry) (receive func() (AppStateEvent, error), kill func(pid uint64) error, closeFunc func() error, err error) {
	dispatcher := channelDispatcher{messageChannel: make(chan dtx.Message, notificationBacklog), closeChannel: make(chan struct{})}
	conn, err := connectInstrumentsWithMsgDispatcher(device, dispatcher)
	if err != nil {
		return nil, nil, nil, err
	}

	notifChannel := conn.RequestChannelIdentifier(mobileNotificationsChannel, loggingDispatcher{conn})
	if _, err := notifChannel.MethodCall("setApplicationStateNotificationsEnabled:", true); err != nil {
		conn.Close()
		return nil, nil, nil, fmt.Errorf("setApplicationStateNotificationsEnabled: %w", err)
	}
	if _, err := notifChannel.MethodCall("setMemoryNotificationsEnabled:", true); err != nil {
		conn.Close()
		return nil, nil, nil, fmt.Errorf("setMemoryNotificationsEnabled: %w", err)
	}

	pControl := ProcessControl{
		processControlChannel: conn.RequestChannelIdentifier(procControlChannel, loggingDispatcher{conn}),
		conn:                  conn,
	}

	receive = func() (AppStateEvent, error) {
		for {
			raw, err := dispatcher.Receive()
			if err != nil {
				return AppStateEvent{}, err
			}
			event, ok := parseAppStateEvent(raw)
			if !ok {
				continue
			}
			return event, nil
		}
	}
	closeFunc = func() error {
		err := dispatcher.Close()
		conn.Close()
		return err
	}

	return receive, pControl.KillProcess, closeFunc, nil
}
