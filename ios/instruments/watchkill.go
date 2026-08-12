package instruments

import (
	"context"
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
)

// runningState is the state_description value ListenAppStateNotifications
// reports when a process transitions to actively running (observed on a cold
// launch straight to foreground; background/suspended launches report
// "Suspended" instead, and termination reports "Terminated"/"Exited").
// WatchKill only reacts to this transition — it's the earliest point a
// launch is both unambiguous and already visible to the user, which is what
// we're racing.
const runningState = "Running"

// WatchKillEvent describes one action WatchKill took (or tried to take) in
// response to a blocked process starting.
type WatchKillEvent struct {
	ProcessName string
	Pid         uint64
	// Err is set when the kill attempt itself failed; the event is still
	// emitted so callers can alert on it instead of losing the failure
	// silently.
	Err error
}

// WatchKill watches device for any of blockedProcessNames entering the
// running state and kills it immediately, using the push-based
// application-state notification channel (see ListenAppStateNotifications)
// rather than polling — the device notifies us on launch instead of us
// having to catch it in a poll interval, so the launch->kill window is wire
// latency, not a poll period. blockedProcessNames are process/executable
// names as reported by the device (ProcessInfo.Name / installationproxy's
// CFBundleExecutable), not bundle IDs — resolving bundle IDs to process
// names is the caller's job, same as the existing `ios kill <bundleID>`
// command does.
//
// WatchKill opens one shared DTX connection for both the notification
// channel and the process-control (kill) channel rather than composing
// ListenAppStateNotifications and NewProcessControl, which would each open
// their own — the device's instruments service only tolerates a single
// connection per client, so a second concurrent one fails outright with an
// "unavailable ... needs an active tunnel" error even though the tunnel is
// already up and the first connection is healthy.
//
// WatchKill runs until ctx is cancelled or the underlying connection errors,
// and closes the returned channel when it stops. The caller must keep
// draining the channel; a slow consumer delays the next kill.
func WatchKill(ctx context.Context, device ios.DeviceEntry, blockedProcessNames []string) (<-chan WatchKillEvent, error) {
	blocked := make(map[string]bool, len(blockedProcessNames))
	for _, name := range blockedProcessNames {
		blocked[name] = true
	}

	dispatcher := channelDispatcher{messageChannel: make(chan dtx.Message, notificationBacklog), closeChannel: make(chan struct{})}
	conn, err := connectInstrumentsWithMsgDispatcher(device, dispatcher)
	if err != nil {
		return nil, err
	}

	notifChannel := conn.RequestChannelIdentifier(mobileNotificationsChannel, loggingDispatcher{conn})
	if _, err := notifChannel.MethodCall("setApplicationStateNotificationsEnabled:", true); err != nil {
		conn.Close()
		return nil, fmt.Errorf("setApplicationStateNotificationsEnabled: %w", err)
	}
	if _, err := notifChannel.MethodCall("setMemoryNotificationsEnabled:", true); err != nil {
		conn.Close()
		return nil, fmt.Errorf("setMemoryNotificationsEnabled: %w", err)
	}

	pControl := ProcessControl{
		processControlChannel: conn.RequestChannelIdentifier(procControlChannel, loggingDispatcher{conn}),
		conn:                  conn,
	}

	events := make(chan WatchKillEvent)
	go func() {
		<-ctx.Done()
		if err := dispatcher.Close(); err != nil {
			golog.Debug("watchkill: closing notification listener", "module", logModule, "udid", device.Properties.SerialNumber, "error", err)
		}
		conn.Close()
	}()
	go func() {
		defer close(events)
		for {
			notification, err := dispatcher.Receive()
			if err != nil {
				golog.Debug("watchkill: notification listener stopped", "module", logModule, "udid", device.Properties.SerialNumber, "error", err)
				return
			}
			name, pid, ok := shouldKill(notification, blocked)
			if !ok {
				continue
			}
			golog.Info("watchkill: killing blocked app", "module", logModule, "udid", device.Properties.SerialNumber, "process", name, "pid", pid)
			killErr := pControl.KillProcess(pid)
			select {
			case events <- WatchKillEvent{ProcessName: name, Pid: pid, Err: killErr}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, nil
}

// shouldKill reports whether notification is a blocked process transitioning
// to runningState, returning its process name and pid when it is. It is the
// pure decision at the core of WatchKill, split out so it can be
// unit-tested without a device connection.
func shouldKill(notification map[string]interface{}, blocked map[string]bool) (name string, pid uint64, ok bool) {
	if state, _ := notification["state_description"].(string); state != runningState {
		return "", 0, false
	}
	name, _ = notification["appName"].(string)
	if name == "" || !blocked[name] {
		return "", 0, false
	}
	pid, ok = toUint64(notification["pid"])
	if !ok {
		return "", 0, false
	}
	return name, pid, true
}
