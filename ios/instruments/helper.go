package instruments

import (
	"fmt"
	"reflect"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

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
	log.Debug(m)
}

func connectInstrumentsWithMsgDispatcher(device ios.DeviceEntry, dispatcher dtx.Dispatcher) (*dtx.Connection, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	dtxConn.MessageDispatcher = dispatcher
	log.Debugf("msg dispatcher: %v attached to instruments connection", reflect.TypeOf(dispatcher))

	return dtxConn, nil
}

func connectInstruments(device ios.DeviceEntry) (*dtx.Connection, error) {
	if device.SupportsRsd() {
		log.Debugf("Connecting to %s", serviceNameRsd)
		return dtx.NewTunnelConnection(device, serviceNameRsd)
	}
	dtxConn, err := dtx.NewUsbmuxdConnection(device, serviceName)
	if err != nil {
		log.Debugf("Failed connecting to %s, trying %s", serviceName, serviceNameiOS14)
		dtxConn, err = dtx.NewUsbmuxdConnection(device, serviceNameiOS14)
		if err != nil {
			return nil, err
		}
	}
	return dtxConn, nil
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
