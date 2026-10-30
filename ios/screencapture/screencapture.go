// Package screencapture takes PNG screenshots of a chosen display over
// com.apple.coredevice.screencaptureservice, the service behind
// `devicectl device capture screenshot --display-unique-id`. Unlike the
// instruments screenshot service it can capture a display other than the main
// one, e.g. the inner display of a foldable while the main display is off.
//
// The service is provided by the developer disk image of iOS 27 and later.
// Replies carry the image as one large data object, which the device only sends
// to a host that announced its RemoteXPC version flags in the RSD handshake.
package screencapture

import (
	"context"
	"fmt"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/coredevice"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

const (
	serviceName = "com.apple.coredevice.screencaptureservice"

	featureCaptureScreenshot = "com.apple.coredevice.feature.capturescreenshot"
	actionCaptureScreenshot  = "com.apple.coredevice.action.capturescreenshot"

	// formatPNG is the only ScreenshotImageFormat the service knows.
	formatPNG = "png"
)

// Service is a connection to the device's screen capture service. One request
// runs at a time: nothing in the XPC framing pairs a reply with its request.
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

// New connects to the screen capture service. The developer disk image must be
// mounted, otherwise the service is not published and connecting fails.
func New(device ios.DeviceEntry) (*Service, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, serviceName)
	if err != nil {
		return nil, fmt.Errorf("screencapture.New: failed to connect to %s: %w", serviceName, err)
	}
	return &Service{conn: conn, deviceID: uuid.New().String()}, nil
}

// Close closes the connection to the screen capture service.
func (s *Service) Close() error {
	return s.conn.Close()
}

// TakeScreenshot returns a PNG of the display with the given unique id, as
// listed by deviceinfo's GetDisplayInfo ("uniqueId"). An empty id captures the
// main display.
//
// A request that fails in transport closes the connection, since the stream
// position is unknown afterwards; create a new Service to retry. An error
// reported by the device keeps it open.
func (s *Service) TakeScreenshot(ctx context.Context, displayUniqueID string) ([]byte, error) {
	input := map[string]interface{}{"requestedFormat": formatPNG}
	if displayUniqueID != "" {
		input["displayUniqueID"] = displayUniqueID
	}

	output, err := s.invoke(ctx, input)
	if err != nil {
		return nil, err
	}
	image, ok := output["image"].([]byte)
	if !ok || len(image) == 0 {
		return nil, fmt.Errorf("TakeScreenshot: no image in the response (image is %T)", output["image"])
	}
	return image, nil
}

func (s *Service) invoke(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Neither a blocked read nor a blocked write can be cancelled, so closing the
	// connection is what ends them.
	stop := context.AfterFunc(ctx, func() { _ = s.conn.Close() })
	defer stop()

	request := coredevice.BuildRequestWithAction(s.deviceID, featureCaptureScreenshot, actionCaptureScreenshot, input)
	if err := s.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		_ = s.conn.Close()
		return nil, fmt.Errorf("screencapture: failed to send the request: %w", err)
	}

	response, err := s.conn.ReceiveOnServerClientStream()
	if err != nil {
		_ = s.conn.Close()
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("screencapture: the request did not complete: %w", ctxErr)
		}
		// The daemon traps instead of replying when the host did not announce its
		// RemoteXPC version flags, which shows up here as the connection closing.
		return nil, fmt.Errorf("screencapture: failed to receive the response: %w", err)
	}

	// A complete error reply leaves the connection usable.
	if e, ok := response["CoreDevice.error"].(map[string]interface{}); ok {
		return nil, fmt.Errorf("screencapture: the device returned an error: %s", describeError(e))
	}
	output, ok := response["CoreDevice.output"].(map[string]interface{})
	if !ok {
		_ = s.conn.Close()
		return nil, fmt.Errorf("screencapture: unexpected response: %v", keys(response))
	}
	return output, nil
}

func describeError(e map[string]interface{}) string {
	if info, ok := e["userInfo"].(map[string]interface{}); ok {
		if d, ok := info["NSDebugDescription"].(string); ok {
			return fmt.Sprintf("%s (domain %v, code %v)", d, e["domain"], e["code"])
		}
	}
	return fmt.Sprintf("domain %v, code %v", e["domain"], e["code"])
}

// keys lists the keys of a response, so an error never dumps an image.
func keys(m map[string]interface{}) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}
