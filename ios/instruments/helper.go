package instruments

import (
	"fmt"
	"reflect"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

const logModule = "go-ios/instruments"

const (
	serviceName      string = "com.apple.instruments.remoteserver"
	serviceNameiOS14 string = "com.apple.instruments.remoteserver.DVTSecureSocketProxy"
	serviceNameRsd   string = "com.apple.instruments.dtservicehub"
)

type loggingDispatcher struct {
	conn *dtx.Connection
}

func (p loggingDispatcher) Dispatch(m dtx.Message) {
	dtx.SendAckIfNeeded(p.conn, m)
	golog.Debug("dispatch message", "module", logModule, "message", m)
}

func connectInstrumentsWithMsgDispatcher(device ios.DeviceEntry, dispatcher dtx.Dispatcher) (*dtx.Connection, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	dtxConn.MessageDispatcher = dispatcher
	golog.Debug("msg dispatcher attached to instruments connection", "module", logModule, "udid", device.Properties.SerialNumber, "dispatcher", reflect.TypeOf(dispatcher))

	return dtxConn, nil
}

func connectInstruments(device ios.DeviceEntry) (*dtx.Connection, error) {
	if device.SupportsRsd() {
		golog.Debug("connecting to service", "module", logModule, "udid", device.Properties.SerialNumber, "service", serviceNameRsd)
		dtxConn, err := dtx.NewTunnelConnection(device, serviceNameRsd)
		if err != nil {
			return nil, instrumentsUnavailableErr(device, serviceNameRsd, err)
		}
		return dtxConn, nil
	}
	dtxConn, err := dtx.NewUsbmuxdConnection(device, serviceName)
	if err != nil {
		golog.Debug("failed connecting to service, trying fallback", "module", logModule, "udid", device.Properties.SerialNumber, "service", serviceName, "fallback", serviceNameiOS14)
		dtxConn, err = dtx.NewUsbmuxdConnection(device, serviceNameiOS14)
		if err != nil {
			return nil, instrumentsUnavailableErr(device, serviceNameiOS14, err)
		}
	}
	return dtxConn, nil
}

// instrumentsUnavailableErr wraps a failed instruments/DTX connection with a
// remedy tailored to the device's iOS version: the service is gated behind a
// mounted Developer Disk Image on iOS <17, and behind an active tunnel on
// iOS 17+. The underlying cause is kept wrapped for debugging.
func instrumentsUnavailableErr(device ios.DeviceEntry, service string, cause error) error {
	remedy := "the Developer Disk Image must be mounted — run `ios image auto`"
	if v, err := ios.GetProductVersion(device); err == nil && v != nil && !v.LessThan(ios.IOS17()) {
		remedy = "this device runs iOS " + v.String() + " and needs an active tunnel — run `ios tunnel start`"
	}
	return fmt.Errorf("instruments service %q unavailable on %s: %s (cause: %w)", service, device.Properties.SerialNumber, remedy, cause)
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

func extractMapPayload(message dtx.Message) (map[string]interface{}, error) {
	if len(message.Payload) != 1 {
		return map[string]interface{}{}, fmt.Errorf("payload of message should have only one element: %+v", message)
	}
	response, ok := message.Payload[0].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, fmt.Errorf("payload type of message should be map[string]interface{}: %+v", message)
	}
	return response, nil
}
