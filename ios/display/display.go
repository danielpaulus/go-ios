// Package display drives com.apple.coredevice.displayservice, the CoreDevice
// service behind Xcode's device screen view, on iOS 17+ devices with a mounted
// Developer Disk Image.
//
// go-ios uses it for one purpose: dtuhidd publishes the HID digitizer surfaces
// as unauthenticated until a media stream is running, and backboardd silently
// discards touch reports aimed at an unauthenticated surface. Starting a video
// stream flips the surfaces to authenticated and lets touch input through. The
// RTP payload itself is discarded - see Receiver - so this package deliberately
// implements stream negotiation and nothing about decoding video.
//
// The gate is per client: a stream started by Xcode does not authenticate
// go-ios' reports. Whoever sends the touches must own the stream.
//
// Streams are expensive to negotiate and the device's mediastream daemon has
// been observed to wedge (until a reboot) when many streams are started and
// stopped in quick succession, so hold one stream open for as long as input is
// needed rather than one per gesture. hid.Session does this.
package display

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/coredevice"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

const (
	serviceName = "com.apple.coredevice.displayservice"

	featureStartMediaStream    = "com.apple.coredevice.feature.startmediastream"
	featureStopMediaStream     = "com.apple.coredevice.feature.stopmediastream"
	featureGetMediaSupportInfo = "com.apple.coredevice.feature.getmediasupportinfo"

	actionMediaStreamStart      = "com.apple.coredevice.action.mediastreamstart"
	actionMediaStreamStop       = "com.apple.coredevice.action.mediastreamstop"
	actionMediaStreamGetSupport = "com.apple.coredevice.action.mediastreamgetsupportinfo"

	// clientSupportedFeatures is a bit mask captured from devicectl describing
	// host-side feature support.
	clientSupportedFeatures = uint64(140)

	// The access-network and transport-protocol values captured from a live
	// screen-sharing session.
	accessNetworkType     = int64(1)
	transportProtocolType = int64(2)

	// negotiationTimeoutSeconds is the timeout the device applies to its side of
	// the negotiation.
	negotiationTimeoutSeconds = uint64(20)

	// DefaultDisplayID is the main display.
	DefaultDisplayID = 1
)

// Service is a connection to the device's display service.
type Service struct {
	conn     *xpc.Connection
	deviceID string
}

// New connects to the display service. Requires iOS 17+ with the Developer Disk
// Image mounted; without it the service is not published and connecting fails.
func New(device ios.DeviceEntry) (*Service, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, serviceName)
	if err != nil {
		return nil, fmt.Errorf("display.New: failed to connect to %s: %w", serviceName, err)
	}
	return &Service{conn: conn, deviceID: uuid.New().String()}, nil
}

// Close closes the connection to the display service. Stop any running stream
// first, otherwise the device keeps sending RTP until it times out.
func (s *Service) Close() error {
	return s.conn.Close()
}

// VideoStreamRequest describes where the device should send a display's frames.
// The receiver socket must already be bound: the device starts sending as soon
// as it answers.
type VideoStreamRequest struct {
	// ReceiverIP and ReceiverPort are the host's tunnel address and bound UDP
	// port, as reported by a Receiver.
	ReceiverIP   string
	ReceiverPort int
	// SenderIP is the device's address over the tunnel (ios.DeviceEntry.Address).
	SenderIP string
	// DisplayID selects the display; use DefaultDisplayID for the main one.
	DisplayID int
}

// StreamAnswer is the device's response to a stream start.
type StreamAnswer struct {
	// ClientSessionID identifies the stream and is required to stop it.
	ClientSessionID uuid.UUID
	// Output is the raw CoreDevice output, which also carries the negotiated
	// stream configuration. Kept so callers can inspect what was agreed without
	// this package having to model the whole answer.
	Output map[string]interface{}
}

// StartVideoStream negotiates and starts an RTP video stream of one display.
//
// ctx bounds the negotiation. The device's mediastream daemon can stop answering
// entirely - a state that survives reconnecting and is only cleared by rebooting
// the device - so always pass a ctx with a deadline; on timeout the returned
// error says so and this Service should be closed.
func (s *Service) StartVideoStream(ctx context.Context, req VideoStreamRequest) (StreamAnswer, error) {
	if req.ReceiverIP == "" || req.ReceiverPort == 0 {
		return StreamAnswer{}, fmt.Errorf("StartVideoStream: receiver address is required, bind a Receiver first")
	}
	if req.SenderIP == "" {
		return StreamAnswer{}, fmt.Errorf("StartVideoStream: sender address is required")
	}
	displayID := req.DisplayID
	if displayID == 0 {
		displayID = DefaultDisplayID
	}

	clientSessionID := uuid.New()
	offer, err := buildVideoNegotiatorOffer(uuid.New(), rand.Uint32())
	if err != nil {
		return StreamAnswer{}, err
	}

	input := map[string]interface{}{
		"clientSupportedFeatures": clientSupportedFeatures,
		"direction":               "output",
		"negotiatorOffer":         offer,
		"options": map[string]interface{}{
			"AVCMediaStreamNegotiatorAccessNetworkType":     map[string]interface{}{"int": accessNetworkType},
			"AVCMediaStreamNegotiatorTransportProtocolType": map[string]interface{}{"int": transportProtocolType},
			"CoreDeviceVideoDisplayMode":                    map[string]interface{}{"string": "DisplayByID"},
			"VideoStreamForDisplayID":                       map[string]interface{}{"int": int64(displayID)},
			"avcMediaStreamOptionClientSessionID":           map[string]interface{}{"uuid": clientSessionID},
		},
		"receiverIP":   req.ReceiverIP,
		"receiverPort": uint64(req.ReceiverPort),
		"senderIP":     req.SenderIP,
		"timeout":      negotiationTimeoutSeconds,
		"type":         "video",
	}

	output, err := s.invoke(ctx, featureStartMediaStream, actionMediaStreamStart, input)
	if err != nil {
		return StreamAnswer{}, err
	}
	return StreamAnswer{ClientSessionID: clientSessionID, Output: output}, nil
}

// StopMediaStream stops a running stream.
//
// The device usually tears the channel down while processing the stop, so a
// closed connection here means the stop was carried out, not that it failed.
// Errors are therefore worth logging but not acting on, and this Service should
// be closed afterwards either way.
func (s *Service) StopMediaStream(ctx context.Context, clientSessionID uuid.UUID) error {
	input := map[string]interface{}{
		"avcMediaStreamOptionClientSessionID": map[string]interface{}{"uuid": clientSessionID},
	}
	_, err := s.invoke(ctx, featureStopMediaStream, actionMediaStreamStop, input)
	return err
}

// GetMediaSupportInfo reports the media-stream features and AVC framework
// version the device supports. Useful as a cheap liveness probe of the
// mediastream daemon before negotiating a stream.
func (s *Service) GetMediaSupportInfo(ctx context.Context) (map[string]interface{}, error) {
	return s.invoke(ctx, featureGetMediaSupportInfo, actionMediaStreamGetSupport, map[string]interface{}{})
}

// invoke sends one CoreDevice request and waits for its response, honouring ctx.
//
// The XPC connection has no read deadline, so the exchange runs in a goroutine
// and ctx cancellation abandons it. An abandoned exchange leaves the connection
// unusable, which is why callers close the Service after a timeout.
func (s *Service) invoke(ctx context.Context, feature, action string, input map[string]interface{}) (map[string]interface{}, error) {
	type result struct {
		output map[string]interface{}
		err    error
	}
	done := make(chan result, 1)

	go func() {
		request := coredevice.BuildRequestWithAction(s.deviceID, feature, action, input)
		if err := s.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
			done <- result{err: fmt.Errorf("failed to send %s: %w", feature, err)}
			return
		}
		response, err := s.conn.ReceiveOnServerClientStream()
		if err != nil {
			done <- result{err: fmt.Errorf("failed to receive the response to %s: %w", feature, err)}
			return
		}
		output, ok := response["CoreDevice.output"].(map[string]interface{})
		if !ok {
			done <- result{err: fmt.Errorf("unexpected response to %s: %+v", feature, response)}
			return
		}
		done <- result{output: output}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			return nil, fmt.Errorf("display: %w", r.err)
		}
		return r.output, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("display: %s did not complete: %w (the device's mediastream daemon may be wedged; rebooting the device clears it)", feature, ctx.Err())
	}
}
