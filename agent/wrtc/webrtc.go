package wrtc

import (
	"encoding/json"
	"io"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

type RTCConnection struct {
	webrtcConn *webrtc.PeerConnection
	Serial     string
}

func Connect(device ios.DeviceEntry) (*RTCConnection, error) {
	conn, err := getOrCreatePeerConnection(device.Properties.SerialNumber)
	if err != nil {
		return &RTCConnection{}, err
	}
	rtcconn := RTCConnection{
		webrtcConn: conn,
		Serial:     device.Properties.SerialNumber,
	}

	return &rtcconn, nil
}

func (c *RTCConnection) RequestResponse(args ...string) (string, error) {
	return "", nil
}

func (c *RTCConnection) StreamingResponse(args ...string) (io.ReadWriteCloser, error) {
	log.Info("new streaming response")
	dc, err := CreateNewDataChannelConnection(c.webrtcConn, c.Serial)
	if err != nil {
		return nil, err
	}
	log.Info("datachannel created")
	cmd := map[string]interface{}{}
	cmd["cmd"] = args[0]
	cmd["serial"] = c.Serial
	cmdBytes, err := json.Marshal(cmd)
	log.Infof("sending command %s", string(cmdBytes))
	dc.Send(cmdBytes)

	return wrapDataChannel(dc)
}
