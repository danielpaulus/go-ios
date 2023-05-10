package ios

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetSocketTypeAndAddress(socketAddress string) (string, string) {
	chunks := strings.Split(socketAddress, "://")
	if len(chunks) != 2 {
		panic("Needs scheme://address")
	}
	return chunks[0], chunks[1]
}

func ToUnixSocketPath(socketAddress string) string {
	scheme, name := GetSocketTypeAndAddress(socketAddress)
	if scheme != "unix" {
		panic("Needs a unix socket")
	}
	return name
}

// GetUsbmuxdSocket this is the default socket address for the platform to connect to.
func GetUsbmuxdSocket() string {
	socket_override := os.Getenv("USBMUXD_SOCKET_ADDRESS")
	if socket_override != "" {
		if strings.Contains(socket_override, ":") {
			return "tcp://" + socket_override
		} else {
			return "unix://" + socket_override
		}
	}
	switch runtime.GOOS {
	case "windows":
		return "tcp://127.0.0.1:27015"
	default:
		return "unix:///var/run/usbmuxd"
	}
}

// UsbMuxConnection can send and read messages to the usbmuxd process to manage pairrecors, listen for device changes
// and connect to services on the phone. Usually messages follow a  request-response pattern. there is a tag integer
// in the message header, that is increased with every sent message.
type UsbMuxConnection struct {
	// tag will be incremented for every message, so responses can be correlated to requests
	tag        uint32
	deviceConn DeviceConnectionInterface
}

// NewUsbMuxConnection creates a new UsbMuxConnection from an already initialized DeviceConnectionInterface
func NewUsbMuxConnection(deviceConn DeviceConnectionInterface) *UsbMuxConnection {
	muxConn := &UsbMuxConnection{tag: 0, deviceConn: deviceConn}
	return muxConn
}

// NewUsbMuxConnectionSimple creates a new UsbMuxConnection with a connection to /var/run/usbmuxd
func NewUsbMuxConnectionSimple() (*UsbMuxConnection, error) {
	deviceConn, err := NewDeviceConnection(GetUsbmuxdSocket())
	muxConn := &UsbMuxConnection{tag: 0, deviceConn: deviceConn}
	return muxConn, err
}

// ReleaseDeviceConnection dereferences this UsbMuxConnection from the underlying DeviceConnection and it returns the DeviceConnection for later use.
// This UsbMuxConnection cannot be used after calling this.
func (muxConn *UsbMuxConnection) ReleaseDeviceConnection() DeviceConnectionInterface {
	conn := muxConn.deviceConn
	muxConn.deviceConn = nil
	return conn
}

// Close calls close on the underlying DeviceConnection
func (muxConn *UsbMuxConnection) Close() error {
	return muxConn.deviceConn.Close()
}

// UsbMuxMessage contains header and payload for a message to usbmux
type UsbMuxMessage struct {
	Header  UsbMuxHeader
	Payload []byte
}

// UsbMuxHeader contains the header for plist messages for the usbmux daemon.
type UsbMuxHeader struct {
	Length  uint32
	Version uint32
	Request uint32
	Tag     uint32
}

// Send sends and encodes a Plist using the usbmux Encoder. Increases the connection tag by one.
func (muxConn *UsbMuxConnection) Send(msg interface{}) error {
	if muxConn.deviceConn == nil {
		return io.EOF
	}
	writer := muxConn.deviceConn.Writer()
	muxConn.tag++
	err := muxConn.encode(msg, writer)
	if err != nil {
		log.Error("Error sending mux")
		return err
	}
	return nil
}

// SendMuxMessage serializes and sends a UsbMuxMessage to the underlying DeviceConnection.
// This does not increase the tag on the connection. Is used mainly by the debug proxy to
// forward messages between device and host
func (muxConn *UsbMuxConnection) SendMuxMessage(msg UsbMuxMessage) error {
	if muxConn.deviceConn == nil {
		return io.EOF
	}
	writer := muxConn.deviceConn.Writer()
	err := binary.Write(writer, binary.LittleEndian, msg.Header)
	if err != nil {
		return err
	}
	_, err = writer.Write(msg.Payload)
	return err
}

// ReadMessage blocks until the next muxMessage is available on the underlying DeviceConnection and returns it.
func (muxConn *UsbMuxConnection) ReadMessage() (UsbMuxMessage, error) {
	if muxConn.deviceConn == nil {
		return UsbMuxMessage{}, io.EOF
	}
	reader := muxConn.deviceConn.Reader()
	msg, err := muxConn.decode(reader)
	if err != nil {
		return UsbMuxMessage{}, err
	}
	return msg, nil
}

// encode serializes a MuxMessage struct to a Plist and writes it to the io.Writer.
func (muxConn *UsbMuxConnection) encode(message interface{}, writer io.Writer) error {
	log.Tracef("UsbMux send %v  on  %v", reflect.TypeOf(message), &muxConn.deviceConn)
	mbytes := ToPlistBytes(message)
	err := writeHeader(len(mbytes), muxConn.tag, writer)
	if err != nil {
		return err
	}
	_, err = writer.Write(mbytes)
	return err
}

func writeHeader(length int, tag uint32, writer io.Writer) error {
	header := UsbMuxHeader{Length: 16 + uint32(length), Request: 8, Version: 1, Tag: tag}
	return binary.Write(writer, binary.LittleEndian, header)
}

// decode reads all bytes for the next MuxMessage from r io.Reader and
// returns a UsbMuxMessage
func (muxConn UsbMuxConnection) decode(r io.Reader) (UsbMuxMessage, error) {
	var muxHeader UsbMuxHeader

	err := binary.Read(r, binary.LittleEndian, &muxHeader)
	if err != nil {
		return UsbMuxMessage{}, err
	}

	payloadBytes := make([]byte, muxHeader.Length-16)
	n, err := io.ReadFull(r, payloadBytes)
	if err != nil {
		return UsbMuxMessage{}, fmt.Errorf("Error '%s' while reading usbmux package. Only %d bytes received instead of %d", err.Error(), n, muxHeader.Length-16)
	}
	log.Tracef("UsbMux Receive on %v", &muxConn.deviceConn)

	return UsbMuxMessage{muxHeader, payloadBytes}, nil
}
