package hid

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

// DisplayServiceName is the RemoteXPC service that gates HID authentication via
// an active media (screen) stream.
const DisplayServiceName = "com.apple.coredevice.displayservice"

const (
	featureStartMediaStream = "com.apple.coredevice.feature.startmediastream"
	featureStopMediaStream  = "com.apple.coredevice.feature.stopmediastream"
	actionMediaStreamStart  = "com.apple.coredevice.action.mediastreamstart"
	actionMediaStreamStop   = "com.apple.coredevice.action.mediastreamstop"

	// clientSupportedFeatures is a bitmask captured from devicectl.
	clientSupportedFeatures = uint64(140)
	accessNetworkType       = int64(1)
	transportProtocolType   = int64(2)
	mediaStreamTimeout      = uint64(20)

	mediaStreamReplyTimeout = 15 * time.Second
	// gateSettleDelay gives backboardd a moment to re-match the HID surfaces
	// against the newly authenticated stream before touches are dispatched.
	gateSettleDelay = 400 * time.Millisecond
)

// TouchSession holds a CoreDevice media stream open so dtuhidd routes touch
// reports all the way through to UIKit. It ports pymobiledevice3's
// touch_session context manager. The RTP payload the device pushes is drained
// and discarded — the session just needs to exist.
type TouchSession struct {
	device          ios.DeviceEntry
	display         xpcConn
	udp             *net.UDPConn
	clientSessionID uuid.UUID
	stopDrain       chan struct{}
	drainDone       chan struct{}
	drainStarted    bool
	closed          bool
}

// StartTouchSession opens the media-stream auth gate and returns a live
// TouchSession. Close it when done to stop the stream.
func StartTouchSession(device ios.DeviceEntry) (*TouchSession, error) {
	if !device.SupportsRsd() {
		return nil, fmt.Errorf("StartTouchSession: requires iOS 17+ with tunnel")
	}
	udid := device.Properties.SerialNumber

	// Bind a UDP receiver on the host's tunnel-side address so the device can
	// reach it. Learn that address by dialing the device over the (kernel)
	// tunnel and reading the local source address the kernel picks.
	receiverIP, err := hostTunnelAddress(device)
	if err != nil {
		return nil, fmt.Errorf("StartTouchSession: could not determine host tunnel address: %w", err)
	}
	udp, err := net.ListenUDP("udp6", &net.UDPAddr{IP: net.ParseIP(receiverIP)})
	if err != nil {
		return nil, fmt.Errorf("StartTouchSession: failed to bind RTP receiver on %s: %w", receiverIP, err)
	}
	receiverPort := udp.LocalAddr().(*net.UDPAddr).Port

	displayConn, err := ios.ConnectToXpcServiceTunnelIface(device, DisplayServiceName)
	if err != nil {
		_ = udp.Close()
		return nil, fmt.Errorf("StartTouchSession: failed to connect to %s: %w", DisplayServiceName, err)
	}

	clientSessionID := uuid.New()
	callID := uuid.New().String()
	sessionID := rand.Uint32()
	offer, err := buildNegotiatorOfferVideo(callID, sessionID)
	if err != nil {
		_ = displayConn.Close()
		_ = udp.Close()
		return nil, fmt.Errorf("StartTouchSession: %w", err)
	}

	senderIP := device.Address
	request := buildStartMediaStreamRequest(offer, receiverIP, receiverPort, senderIP, clientSessionID)

	golog.Info("starting media stream to open HID auth gate", "module", logModule, "udid", udid,
		"receiverIP", receiverIP, "receiverPort", receiverPort, "senderIP", senderIP)

	ts := &TouchSession{
		device:          device,
		display:         displayConn,
		udp:             udp,
		clientSessionID: clientSessionID,
		stopDrain:       make(chan struct{}),
		drainDone:       make(chan struct{}),
	}

	if err := ts.invokeWithReply(request); err != nil {
		ts.Close()
		return nil, fmt.Errorf("StartTouchSession: startmediastream failed (device media daemon may be wedged — reboot the device and retry): %w", err)
	}

	// Drain the RTP the device now pushes; we discard it.
	ts.drainStarted = true
	go ts.drain()

	// Let backboardd re-authenticate the HID surfaces before touches flow.
	time.Sleep(gateSettleDelay)
	golog.Info("media stream established, HID auth gate open", "module", logModule, "udid", udid)
	return ts, nil
}

// invokeWithReply sends a CoreDevice invoke and waits for CoreDevice.output,
// bounded by mediaStreamReplyTimeout. On timeout it closes the connection to
// unblock the orphaned read.
func (ts *TouchSession) invokeWithReply(request map[string]interface{}) error {
	type result struct {
		reply map[string]interface{}
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		// Send runs in the goroutine too: a blocked HTTP/2 write (device not
		// consuming the request) must be bounded by the same timeout, otherwise
		// it would hang before the timer starts.
		if err := ts.display.Send(request, xpc.HeartbeatRequestFlag); err != nil {
			ch <- result{nil, fmt.Errorf("failed to send request: %w", err)}
			return
		}
		golog.Debug("startmediastream request sent, awaiting reply", "module", logModule, "udid", ts.device.Properties.SerialNumber)
		for {
			reply, err := ts.display.ReceiveOnServerClientStream()
			if err != nil {
				ch <- result{nil, err}
				return
			}
			if len(reply) == 0 {
				continue
			}
			ch <- result{reply, nil}
			return
		}
	}()
	timer := time.NewTimer(mediaStreamReplyTimeout)
	defer timer.Stop()
	select {
	case r := <-ch:
		if r.err != nil {
			return fmt.Errorf("failed to receive reply: %w", r.err)
		}
		if _, ok := r.reply["CoreDevice.output"]; !ok {
			return fmt.Errorf("no CoreDevice.output in reply: %+v", r.reply)
		}
		return nil
	case <-timer.C:
		_ = ts.display.Close()
		return fmt.Errorf("timed out after %s waiting for the media stream to start", mediaStreamReplyTimeout)
	}
}

// drain reads and discards RTP packets until the session is closed.
func (ts *TouchSession) drain() {
	defer close(ts.drainDone)
	buf := make([]byte, 65536)
	for {
		select {
		case <-ts.stopDrain:
			return
		default:
		}
		_ = ts.udp.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, _, err := ts.udp.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return
		}
	}
}

// Close stops the media stream and releases resources. Best-effort: the device
// routinely yanks the RemoteXPC channel mid-stop, which is not an error.
func (ts *TouchSession) Close() error {
	if ts.closed {
		return nil
	}
	ts.closed = true
	if ts.drainStarted && ts.stopDrain != nil {
		close(ts.stopDrain)
		<-ts.drainDone
		ts.stopDrain = nil
	}
	// Best-effort stop request; ignore errors (channel is often already gone).
	stop := buildStopMediaStreamRequest(ts.clientSessionID)
	_ = ts.display.Send(stop, xpc.HeartbeatRequestFlag)
	if ts.udp != nil {
		_ = ts.udp.Close()
	}
	return ts.display.Close()
}

// hostTunnelAddress returns the host's own IPv6 address on the tunnel that
// reaches the device. It dials a UDP socket to the device address (no packets
// are sent) and reads the local source address the kernel selects.
func hostTunnelAddress(device ios.DeviceEntry) (string, error) {
	// Any RSD port works for route selection; use the device address with a
	// throwaway port.
	remote := net.JoinHostPort(device.Address, "1")
	conn, err := net.Dial("udp6", remote)
	if err != nil {
		return "", fmt.Errorf("dial device tunnel address %s: %w", device.Address, err)
	}
	defer conn.Close()
	local := conn.LocalAddr().(*net.UDPAddr)
	ip := local.IP.String()
	// Preserve the IPv6 zone (e.g. fe80::1%utun3) if the kernel selected a
	// scoped source address; without it a link-local receiverIP is unroutable.
	if local.Zone != "" {
		ip += "%" + local.Zone
	}
	return ip, nil
}

// buildStartMediaStreamRequest wraps the negotiatorOffer in the CoreDevice
// displayservice invoke envelope (mediastreamstart).
func buildStartMediaStreamRequest(offer []byte, receiverIP string, receiverPort int, senderIP string, clientSessionID uuid.UUID) map[string]interface{} {
	input := map[string]interface{}{
		"clientSupportedFeatures": clientSupportedFeatures,
		"direction":               "output",
		"negotiatorOffer":         offer,
		"options": map[string]interface{}{
			"AVCMediaStreamNegotiatorAccessNetworkType":     map[string]interface{}{"int": accessNetworkType},
			"AVCMediaStreamNegotiatorTransportProtocolType": map[string]interface{}{"int": transportProtocolType},
			"CoreDeviceVideoDisplayMode":                    map[string]interface{}{"string": "DisplayByID"},
			"VideoStreamForDisplayID":                       map[string]interface{}{"int": int64(1)},
			"avcMediaStreamOptionClientSessionID":           map[string]interface{}{"uuid": clientSessionID},
		},
		"receiverIP":   receiverIP,
		"receiverPort": uint64(receiverPort),
		"senderIP":     senderIP,
		"timeout":      mediaStreamTimeout,
		"type":         "video",
	}
	return buildCoreDeviceInvoke(featureStartMediaStream, actionMediaStreamStart, input)
}

// buildStopMediaStreamRequest builds the mediastreamstop invoke.
func buildStopMediaStreamRequest(clientSessionID uuid.UUID) map[string]interface{} {
	input := map[string]interface{}{
		"avcMediaStreamOptionClientSessionID": map[string]interface{}{"uuid": clientSessionID},
	}
	return buildCoreDeviceInvoke(featureStopMediaStream, actionMediaStreamStop, input)
}

// buildCoreDeviceInvoke builds the CoreDevice invoke envelope displayservice
// expects (DDIProtocolVersion 2, coreDeviceVersion 629.3, plus actionIdentifier
// — distinct from ios/coredevice.BuildRequest which targets other services).
func buildCoreDeviceInvoke(feature, action string, input map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"CoreDevice.CoreDeviceDDIProtocolVersion": int64(2),
		"CoreDevice.coreDeviceVersion":            coreDeviceVersion("629.3"),
		"CoreDevice.deviceIdentifier":             uuid.New().String(),
		"CoreDevice.input":                        input,
		"CoreDevice.invocationIdentifier":         uuid.New().String(),
		"CoreDevice.featureIdentifier":            feature,
		"CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.actionIdentifier":             action,
	}
}

// coreDeviceVersion builds the CoreDevice.coreDeviceVersion dict for a dotted
// version string.
func coreDeviceVersion(version string) map[string]interface{} {
	parts := strings.Split(version, ".")
	components := make([]interface{}, len(parts))
	for i, p := range parts {
		n, _ := strconv.ParseUint(p, 10, 64)
		components[i] = n
	}
	return map[string]interface{}{
		"components":              components,
		"originalComponentsCount": int64(len(parts)),
		"stringValue":             version,
	}
}
