package debugproxy

import (
	"encoding/hex"
	"io"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type DumpingConn struct {
	conn       net.Conn
	fileHandle *os.File
}

func NewDumpingConn(filePath string, conn net.Conn) *DumpingConn {
	fileHandle, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Println(err)
	}
	dc := DumpingConn{fileHandle: fileHandle, conn: conn}
	return &dc
}

func (d DumpingConn) Read(b []byte) (n int, err error) {
	n, err = d.conn.Read(b)
	if err != nil {
		io.WriteString(d.fileHandle, "\n\nError Reading"+err.Error())
	}
	io.WriteString(d.fileHandle, "\n\nReading------------->\n")
	d.fileHandle.Write([]byte(hex.Dump(b)))
	return n, err
}

func (d DumpingConn) Write(b []byte) (n int, err error) {
	n, err = d.conn.Write(b)
	if err != nil {
		io.WriteString(d.fileHandle, "\n\nError Sending"+err.Error())
	}
	io.WriteString(d.fileHandle, "\n\nSending------------->\n")
	d.fileHandle.Write([]byte(hex.Dump(b)))
	return n, err
}

func (d DumpingConn) Close() error {
	err := d.fileHandle.Close()
	if err != nil {
		log.Warn("failed closing bin file handle", err)
	}
	return d.conn.Close()
}

func (d DumpingConn) LocalAddr() net.Addr {
	return d.conn.LocalAddr()
}

func (d DumpingConn) RemoteAddr() net.Addr {
	return d.conn.RemoteAddr()
}

func (d DumpingConn) SetDeadline(t time.Time) error {
	return d.conn.SetDeadline(t)
}

func (d DumpingConn) SetReadDeadline(t time.Time) error {
	return d.conn.SetReadDeadline(t)
}

func (d DumpingConn) SetWriteDeadline(t time.Time) error {
	return d.conn.SetWriteDeadline(t)
}
