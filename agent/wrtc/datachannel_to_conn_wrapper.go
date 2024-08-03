package wrtc

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

var channelCounter uint64

type DataChanReadWriter struct {
	dataChannel *webrtc.DataChannel
	reader      *io.PipeReader
	writer      *io.PipeWriter
}

func (d *DataChanReadWriter) Close() error {
	d.dataChannel.Close()
	d.writer.Close()
	d.reader.Close()
	return nil
}

func (d *DataChanReadWriter) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (d *DataChanReadWriter) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (d *DataChanReadWriter) SetDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataChanReadWriter) SetReadDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataChanReadWriter) SetWriteDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataChanReadWriter) Init() {
	pr, pw := io.Pipe()
	d.reader = pr
	d.writer = pw
	d.dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Tracef("chan: %s got message from remote", d.dataChannel.Label())
		_, err := pw.Write(msg.Data)
		if err != nil {
			log.Errorf("failed writing message that came from datachannel: %v", err)
		}
	})
	d.dataChannel.OnError(func(err error) {
		log.Errorf("chan %s err %v", d.dataChannel.Label(), err)
	})

	d.dataChannel.OnClose(func() {

		log.Debugf("datachan %s closing...", d.dataChannel.Label())
		err := pr.Close()
		if err != nil {
			log.Warnf("datachan %s closed with err %v", d.dataChannel.Label(), err)
		}
		log.Debugf("datachan %s closed", d.dataChannel.Label())
	})

}
func (d *DataChanReadWriter) Read(p []byte) (n int, err error) {
	n, err = d.reader.Read(p)
	log.Tracef("chan: %s err: %v from_remote: %s", d.dataChannel.Label(), err, p[:n])
	return n, err
}

func (d *DataChanReadWriter) Write(p []byte) (n int, err error) {
	r := p
	for {
		l := len(r)
		if l <= 65535 {
			err = d.dataChannel.Send(r)
			break
		}
		s := r[:65535]
		err = d.dataChannel.Send(s)
		r = r[65535:]
	}

	n = len(p)
	//log.Debugf("chan: %s err: %v local2remote: %s", d.dataChannel.Label(), err, p)
	return n, err
	//return len(p), err
}

// https://github.com/pojntfx/weron
func CreateNewDataChannelConnection(peerConnection *webrtc.PeerConnection, serial string) (*webrtc.DataChannel, error) {

	o := true
	var r uint16 = 5
	opts := &webrtc.DataChannelInit{
		Ordered:           &o,
		MaxPacketLifeTime: nil,
		MaxRetransmits:    &r,
		Protocol:          nil,
		Negotiated:        nil,
		ID:                nil,
	}
	datachan, err := peerConnection.CreateDataChannel(fmt.Sprintf("direct_%d_%s", atomic.AddUint64(&channelCounter, 1), serial), opts)
	if err != nil {
		log.Errorf("could create webrtc datachan: %+v", err)
		return nil, err
	}
	waiter2 := make(chan interface{})
	datachan.OnMessage(func(msg webrtc.DataChannelMessage) {
		i := string(msg.Data)
		log.Debugf("chan: %s sent init msg: %s", datachan.Label(), i)
		waiter2 <- struct{}{}
	})

	log.Debugf("waiting for channel %s to open..", datachan.Label())
	waiter := make(chan interface{})
	// Register channel opening handling

	datachan.OnOpen(func() {
		log.Debugf("Data channel '%s' open.%v", datachan.Label(), datachan.Negotiated())

		go func() { waiter <- struct{}{} }()
	})

	<-waiter

	<-waiter2
	log.Debugf("ok channel %s open, starting to proxy", datachan.Label())

	return datachan, nil
}

func wrapDataChannel(datachan *webrtc.DataChannel) (net.Conn, error) {
	adapter := &DataChanReadWriter{
		dataChannel: datachan,
	}

	adapter.Init()
	return adapter, nil
}
