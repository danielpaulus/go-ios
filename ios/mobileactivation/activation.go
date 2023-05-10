package mobileactivation

import (
	"io"
	"net/url"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

const serviceName string = "com.apple.mobileactivationd"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

// New creates a new Connection to com.apple.mobileactivationd
func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	var activationdConn Connection
	activationdConn.deviceConn = deviceConn
	activationdConn.plistCodec = ios.NewPlistCodec()

	return &activationdConn, nil
}

// Close closes the connection to the device.
func (activationdConn *Connection) Close() error {
	return activationdConn.deviceConn.Close()
}

const (
	activationStateKey = "ActivationState"
	unactivated        = "Unactivated"
)

// IsActivated uses lockdown to get the ActivationState of the device. Returns ActivationState != 'Unactivated'
func IsActivated(device ios.DeviceEntry) (bool, error) {
	values, err := ios.GetValuesPlist(device)
	if err != nil {
		return false, err
	}
	val, ok := values[activationStateKey]
	if ok {
		return val != unactivated, nil
	}
	return false, nil
}

// Activate kicks off the activation process for a given device. It returns an error if the activation is unsuccessful. It returns
// nil if the device was activated before or the activation was successful.
// The process gets a sendHandshakeRequest from the device, sends it to the Apple activation server and stores the response on the device.
// This means you have to be online for this to work!
// If the device is already activated, this command does nothing and returns nil. It is safe to run multiple times.
func Activate(device ios.DeviceEntry) error {
	isActivated, err := IsActivated(device)
	if err != nil {
		return err
	}
	if isActivated {
		log.WithField("udid", device.Properties.SerialNumber).Info("the device is already activated")
		return nil
	}
	conn, err := New(device)
	if err != nil {
		return err
	}
	defer conn.Close()
	resp, err := conn.sendAndReceive(map[string]interface{}{"Command": "CreateTunnel1SessionInfoRequest"})
	if err != nil {
		return err
	}
	log.Debugf("CreateTunnel1SessionInfoRequest resp: %v", resp)
	val := resp["Value"].(map[string]interface{})

	handshakeRequestMessage := val["HandshakeRequestMessage"].([]byte)
	log.Debugf("HandshakeRequestMessage: %v", handshakeRequestMessage)
	stringPlist := ios.ToPlist(val)
	log.Infof("sending %d bytes via http to the handshake server..", len(stringPlist))
	header, body, err := sendHandshakeRequest(strings.NewReader(stringPlist))
	var handshakeResponse []byte
	if body != nil {
		handshakeResponse, err = io.ReadAll(body)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	defer body.Close()
	log.Debugf("handshare response headers: %v", header)
	log.Debugf("rcv %d bytes handshake response", len(handshakeResponse))
	log.Infof("ok")
	// get activation info from device

	conn1, err := New(device)
	if err != nil {
		return err
	}
	defer conn1.Close()

	activationInfoResponseResp, err := conn1.sendAndReceive(map[string]interface{}{
		"Command": "CreateActivationInfoRequest", "Value": handshakeResponse,
		"Options": map[string]interface{}{"BasebandWaitCount": 90},
	})
	if err != nil {
		return err
	}
	activationInfoResponseMap := activationInfoResponseResp["Value"].(map[string]interface{})
	activationResponsePlist := ios.ToPlist(activationInfoResponseMap)

	params := url.Values{}
	params.Add("activation-info", activationResponsePlist)
	payload := params.Encode()
	log.Info("sending activation info")

	headers, body, err := sendActivationRequest(strings.NewReader(payload))
	log.Debugf("activation response headers:%v", headers)
	activationHttpResponse := []byte{}

	if body != nil {
		activationHttpResponse, err = io.ReadAll(body)
		if err != nil {
			return err
		}
		log.Debugf("activation http response: %s", activationHttpResponse)
	}
	if err != nil {
		return err
	}
	log.Info("activation response received")

	// Technically HTTP Headers are not a map String, String but a map String, []String because
	// Headers can appear multiple times. F.ex.
	// Content-Type: bla
	// Content-Type: blu
	// is perfectly fine. This results in an array like header so Content-Type: [bla, blu]
	// Of course this is not really useful and the device expects a map String, String, so merge it here
	activationResponseHeaders := map[string]interface{}{}
	for name, values := range headers {
		// Loop over all values for the name.
		for _, value := range values {
			activationResponseHeaders[name] = value
		}
	}

	conn2, err := New(device)
	if err != nil {
		return err
	}
	defer conn2.Close()

	activationResponseMap, err := ios.ParsePlist(activationHttpResponse)
	if err != nil {
		return err
	}
	log.Debugf("activation Response Plist: %v", activationResponseMap)
	log.Info("storing activation response to device")
	resp, err = conn2.sendAndReceive(map[string]interface{}{
		"Command": "HandleActivationInfoWithSessionRequest",
		"Value":   activationHttpResponse, "ActivationResponseHeaders": activationResponseHeaders,
	})
	if err != nil {
		return err
	}
	log.Debugf("HandleActivationInfoWithSessionRequest response: %v", resp)
	log.Info("device successfully activated")
	return nil
}

func (mcInstallConn *Connection) sendAndReceive(request map[string]interface{}) (map[string]interface{}, error) {
	reader := mcInstallConn.deviceConn.Reader()
	requestBytes, err := mcInstallConn.plistCodec.Encode(request)
	if err != nil {
		return map[string]interface{}{}, err
	}
	err = mcInstallConn.deviceConn.Send(requestBytes)
	if err != nil {
		return map[string]interface{}{}, err
	}
	responseBytes, err := mcInstallConn.plistCodec.Decode(reader)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return ios.ParsePlist(responseBytes)
}
