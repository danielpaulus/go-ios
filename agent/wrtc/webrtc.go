package wrtc

import (
	"io"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/pion/webrtc/v3"
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
	dc, err := CreateNewDataChannelConnection(c.webrtcConn, c.Serial)
	if err != nil {
		return nil, err
	}
	return wrapDataChannel(dc)
}
