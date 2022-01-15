package ios

import (
	"github.com/gopherjs/gopherjs/js"
	log "github.com/sirupsen/logrus"

	"net"
	"time"
)

type wsconn struct {
	ws  *js.Object
	rdr *ChannelReader
}

var _ net.Conn = (*wsconn)(nil)

func newWSConn(ws *js.Object) *wsconn {
	ws.Set("binaryType", "arraybuffer")
	out := make(chan []byte, 1)
	ws.Call("on", "data", func(evt *js.Object) {
		log.Tracef("rcv: %x", evt)
		out <- toBytes(evt)
	})
	rdr := NewChannelReader(out)
	return &wsconn{
		ws:  ws,
		rdr: rdr,
	}
}

func (c *wsconn) Read(b []byte) (n int, err error) {
	log.Tracef("read")
	n, err = c.rdr.Read(b)
	log.Tracef("read done")
	return n, err
}

func (c *wsconn) Write(b []byte) (n int, err error) {
	log.Tracef("before writing: %x", b)
	c.ws.Call("write", b)
	log.Tracef("after writing")
	return len(b), nil
}

func (c *wsconn) Close() error {
	log.Tracef("closing conn..")
	c.ws.Call("destroy")
	log.Tracef("conn closed")
	return nil
}

func (c *wsconn) LocalAddr() net.Addr {
	return websocketAddress{c.ws.Get("url").String()}
}

func (c *wsconn) RemoteAddr() net.Addr {
	return websocketAddress{c.ws.Get("url").String()}
}

func (c *wsconn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *wsconn) SetReadDeadline(t time.Time) error {
	c.rdr.SetDeadline(t)
	return nil
}

func (c *wsconn) SetWriteDeadline(t time.Time) error {
	log.Println("SetWriteDeadline not implemented")
	return nil
}

func toBytes(obj *js.Object) []byte {
	return js.Global.Get("Uint8Array").New(obj).Interface().([]byte)
}

type websocketAddress struct {
	url string
}

func (wsa websocketAddress) Network() string {
	return "ws"
}

func (wsa websocketAddress) String() string {
	return wsa.url
}
