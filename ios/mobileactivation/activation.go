package mobileactivation

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"io"
	"net/url"
	"strings"
)

const serviceName string = "com.apple.mobileactivationd"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

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
func (activationdConn *Connection) Close() error {
	return activationdConn.deviceConn.Close()
}

func Activate(device ios.DeviceEntry) error {
	conn, err := New(device)
	if err != nil {
		return err
	}
	defer conn.Close()
	resp, err := conn.sendAndReceive(map[string]interface{}{"Command": "CreateTunnel1SessionInfoRequest"})
	if err != nil {
		return err
	}
	//log.Infof("resp: %v", resp)
	val := resp["Value"].(map[string]interface{})
	//collectionBlob := val["CollectionBlob"].([]byte)
	handshareRequestMessage := val["HandshakeRequestMessage"].([]byte)
	log.Infof("%v", handshareRequestMessage)
	//udid := val["UniqueDeviceID"].(string)
	//blobplist, _ := ios.ParsePlist(collectionBlob)
	stringPlist := ios.ToPlist(val)
	log.Infof("sending http post bytes:%d", len(stringPlist))
	header, body, err := request(strings.NewReader(stringPlist))
	var handshakeResponse = []byte{}
	if body != nil {
		bodyb, _ := io.ReadAll(body)
		handshakeResponse = (bodyb)
	}
	if err != nil {
		return err
	}
	defer body.Close()
	log.Infof("headers: %v", header)

	log.Infof("rcv %d bytes handshake response", len(handshakeResponse))

	//get activation info from device
	handshakeResponsePlist, _ := ios.ParsePlist(handshakeResponse)

	valio := ios.ToPlist(handshakeResponsePlist)
	log.Infof("%v", valio)

	conn1, err := New(device)
	if err != nil {
		return err
	}
	defer conn1.Close()

	resp, err = conn1.sendAndReceive(map[string]interface{}{"Command": "CreateActivationInfoRequest", "Value": handshakeResponse,
		"Options": map[string]interface{}{"BasebandWaitCount": 90}})
	if err != nil {
		return err
	}
	val2 := resp["Value"].(map[string]interface{})
	//activationInfo := map[string]interface{}{"activation-info": val2}
	stringPlist = ios.ToPlist(val2)
	//strippedplist := removePlistThings(stringPlist)
	//payload := fmt.Sprintf("%s=%s", "activation-info", stringPlist)
	params := url.Values{}
	params.Add("activation-info", stringPlist)
	payload := params.Encode()
	log.Info("sending")
	//log.Infof("%v", resp)
	log.Infof("sending activation post:")
	headers, body, err := request2(strings.NewReader(payload))
	log.Infof("actresphead:%v", headers)
	var activationHttpResponse = []byte{}

	if body != nil {
		bodyb, _ := io.ReadAll(body)
		activationHttpResponse = (bodyb)
		println(string(activationHttpResponse))
	}

	if err != nil {
		return err
	}

	activationResponseHeaders := map[string]interface{}{}
	for name, values := range headers {
		// Loop over all values for the name.
		for _, value := range values {
			//fmt.Println(name, value)
			activationResponseHeaders[name] = value
		}
	}

	conn2, err := New(device)
	if err != nil {
		return err
	}
	defer conn2.Close()

	ar, _ := ios.ParsePlist(activationHttpResponse)
	log.Infof("%v", ar)
	valsen := ios.ToBinPlistBytes(ar["ActivationRecord"])
	print("here:")
	fmt.Printf("%x", valsen)
	resp, err = conn2.sendAndReceive(map[string]interface{}{"Command": "HandleActivationInfoWithSessionRequest",
		"Value": activationHttpResponse, "ActivationResponseHeaders": activationResponseHeaders})
	if err != nil {
		return err
	}
	log.Infof("%v", resp)

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
