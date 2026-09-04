// Package display starts RTP video streams over
// com.apple.coredevice.displayservice. A stream also authenticates HID input.
package display

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/coredevice"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

const logModule = "go-ios/display"

const (
	serviceName = "com.apple.coredevice.displayservice"

	featureStartMediaStream = "com.apple.coredevice.feature.startmediastream"
	featureStopMediaStream  = "com.apple.coredevice.feature.stopmediastream"

	actionMediaStreamStart = "com.apple.coredevice.action.mediastreamstart"
	actionMediaStreamStop  = "com.apple.coredevice.action.mediastreamstop"

	// Host capabilities and link type. The values Xcode sends over a tunnel, with
	// no documented meaning to derive them from.
	clientSupportedFeatures = uint64(140)
	accessNetworkType       = int64(1)
	transportProtocolType   = int64(2)

	// The deadline the device applies to its own side. Keep it under the caller's:
	// the device giving up first answers with an error and keeps the connection.
	negotiationTimeoutSeconds = uint64(10)

	DefaultDisplayID = 1
)

// Service is a connection to the device's display service. One request runs at a
// time: nothing in the XPC framing pairs a reply with its request (see invoke).
type Service struct {
	conn     xpcConn
	deviceID string
	mutex    sync.Mutex
}

// xpcConn is the subset of *xpc.Connection used here, so the request/reply logic
// can be exercised with a fake.
type xpcConn interface {
	Send(data map[string]interface{}, flags ...uint32) error
	ReceiveOnServerClientStream() (map[string]interface{}, error)
	Close() error
}

// New connects to the display service. The Developer Disk Image must be
// mounted, otherwise the service is not published and connecting fails.
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
// The receiver must already be bound: the device sends as soon as it answers.
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
	// Output is the raw CoreDevice output, kept so callers can inspect the
	// negotiated configuration without this package modelling the whole answer.
	Output map[string]interface{}
}

// StartVideoStream starts an RTP video stream of one display. iOS 27+, earlier
// fails 9021. Pass a deadline: the daemon can stop answering, so close on timeout.
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
		// The id comes back even on failure: the device may have started the
		// stream anyway, and stopping it needs this id. A failure closes this
		// Service's connection, so the stop has to go out on a new one.
		return StreamAnswer{ClientSessionID: clientSessionID}, err
	}
	return StreamAnswer{ClientSessionID: clientSessionID, Output: output}, nil
}

// StopMediaStream stops this client's stream. stopAll is a key the payload is
// expected to carry, and must be true, though a client only ever holds one
// stream. The device often drops the channel mid-stop, which means it worked.
func (s *Service) StopMediaStream(ctx context.Context, clientSessionID uuid.UUID) error {
	input := map[string]interface{}{
		"avcMediaStreamOptionClientSessionID": map[string]interface{}{"uuid": clientSessionID},
		"stopAll":                             true,
	}
	_, err := s.invoke(ctx, featureStopMediaStream, actionMediaStreamStop, input)
	return err
}

// invoke sends one CoreDevice request and waits for its response, honouring ctx.
//
// An interrupted exchange may have consumed part of a message, and nothing
// reports whether it did, so the connection is closed rather than reused: there
// is no way to resynchronise and no id to catch a reply meant for someone else.
func (s *Service) invoke(ctx context.Context, feature, action string, input map[string]interface{}) (map[string]interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Neither a blocked read nor a blocked write can be cancelled, so closing the
	// connection is what ends them. It is discarded either way, so nothing is lost
	// by being blunt.
	stop := context.AfterFunc(ctx, func() { _ = s.conn.Close() })
	defer stop()

	request := coredevice.BuildRequestWithAction(s.deviceID, feature, action, input)
	if err := s.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		_ = s.conn.Close()
		return nil, fmt.Errorf("display: failed to send %s: %w", feature, err)
	}

	response, err := s.conn.ReceiveOnServerClientStream()
	if err != nil {
		_ = s.conn.Close()
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("display: %s did not complete: %w (the device's mediastream daemon may be wedged; rebooting the device clears it)", feature, ctxErr)
		}
		return nil, fmt.Errorf("display: failed to receive the response to %s: %w", feature, err)
	}

	output, ok := response["CoreDevice.output"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("display: unexpected response to %s: %+v", feature, response)
	}
	return output, nil
}
