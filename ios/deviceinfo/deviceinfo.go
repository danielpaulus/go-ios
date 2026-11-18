package deviceinfo

import (
	"fmt"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/coredevice"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

type Connection struct {
	conn     *xpc.Connection
	deviceId string
}

func NewDeviceInfo(deviceEntry ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(deviceEntry, "com.apple.coredevice.deviceinfo")
	if err != nil {
		return nil, fmt.Errorf("NewDeviceInfo: failed to connect to XPC service: %w", err)
	}
	return &Connection{conn: xpcConn, deviceId: uuid.New().String()}, nil
}

func (c Connection) Close() error {
	return c.conn.Close()
}

func (c Connection) GetDisplayInfo() (map[string]interface{}, error) {
	request := coredevice.BuildRequest(c.deviceId, "com.apple.coredevice.feature.getdisplayinfo", map[string]interface{}{})
	err := c.conn.Send(request, xpc.HeartbeatRequestFlag)
	if err != nil {
		return nil, fmt.Errorf("GetDisplayInfo: failed to send 'getdisplayinfo' request: %w", err)
	}

	response, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return nil, fmt.Errorf("GetDisplayInfo: failed to receive response: %w", err)
	}

	if output, ok := response["CoreDevice.output"].(map[string]interface{}); ok {
		return output, nil
	}

	return nil, fmt.Errorf("GetDisplayInfo: could not parse response %+v", response)
}

// hingeAngleTimeout is how long GetHingeAngle waits for the hinge to report.
const hingeAngleTimeout = 5 * time.Second

// GetHingeAngle reads the angle a foldable's hinge is at, in degrees from 0
// (closed) to 180 (open flat). It is the angle the device goes by: the one the
// hinge was last set to, or folded to by hand, whichever came last.
//
// It subscribes to the hinge updates the device hub follows, and the first is
// the angle now. A device that sends none within 5 seconds, such as one
// without a hinge, gets an error. It uses a connection of its own, as the
// subscription cannot be ended.
func GetHingeAngle(device ios.DeviceEntry) (float64, error) {
	c, err := NewDeviceInfo(device)
	if err != nil {
		return 0, fmt.Errorf("GetHingeAngle: %w", err)
	}
	defer c.Close()

	sideChannel := uuid.New()
	request := coredevice.BuildRequestWithAction(c.deviceId, "com.apple.coredevice.feature.monitormotion",
		"com.apple.coredevice.action.streamhingeangle", map[string]interface{}{
			"actualInput": map[string]interface{}{
				"changeThreshold": map[string]interface{}{
					"unit": map[string]interface{}{
						"converter": map[string]interface{}{"coefficient": float64(1), "constant": float64(0)},
						"symbol":    "°",
					},
					"value": float64(0.01),
				},
				// 50 ms, as seconds and attoseconds, as the device hub
				// asks for it.
				"updateInterval": []interface{}{int64(0), uint64(50000000000000000)},
			},
			"streamProxy": map[string]interface{}{"sideChannel": sideChannel},
		})
	if err := c.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		return 0, fmt.Errorf("GetHingeAngle: failed to send 'streamhingeangle' request: %w", err)
	}

	type result struct {
		degrees float64
		err     error
	}
	read := make(chan result, 1)
	go func() {
		// The updates come on the client-server stream. They are not answered:
		// an answer makes the device reset the stream.
		for {
			msg, err := c.conn.ReceiveOnClientServerStream()
			if err != nil {
				read <- result{err: err}
				return
			}
			degrees, ours, err := hingeAngleIn(msg, sideChannel.String())
			if ours || err != nil {
				read <- result{degrees, err}
				return
			}
		}
	}()
	select {
	case r := <-read:
		if r.err != nil {
			return 0, fmt.Errorf("GetHingeAngle: %w", r.err)
		}
		return r.degrees, nil
	case <-time.After(hingeAngleTimeout):
		return 0, fmt.Errorf("GetHingeAngle: the device did not report its hinge within %s", hingeAngleTimeout)
	}
}

// hingeAngleIn is the angle in a hinge update for sideChannel, and whether the
// message is one. It is the mechanical angle, which is where the hinge is; the
// other one, "angle", trails it by a step while it moves.
func hingeAngleIn(msg map[string]interface{}, sideChannel string) (float64, bool, error) {
	if id, _ := msg["XPCSideChannel.uniqueIdentifier"].(string); !strings.EqualFold(id, sideChannel) {
		return 0, false, nil
	}
	status, _ := msg["CoreDevice.XPCMessageKey.sideChannelStatus"].(map[string]interface{})
	pushing, _ := status["pushing"].(map[string]interface{})
	elements, _ := pushing["elements"].([]interface{})
	if len(elements) == 0 {
		return 0, true, fmt.Errorf("no hinge angle in %v", msg)
	}
	update, _ := elements[len(elements)-1].(map[string]interface{})
	if valid, _ := update["isAngleValid"].(bool); !valid {
		return 0, true, fmt.Errorf("the device reports no valid hinge angle")
	}
	angle, _ := update["mechanicalAngle"].(map[string]interface{})
	switch v := angle["value"].(type) {
	case float64:
		return v, true, nil
	case int64:
		return float64(v), true, nil
	case uint64:
		return float64(v), true, nil
	}
	return 0, true, fmt.Errorf("no hinge angle in %v", update)
}
