package ios

import (
	log "github.com/sirupsen/logrus"
	"net"
)

//Lockdownport is the port of the always running lockdownd on the iOS device.
const Lockdownport uint16 = 32498

//LockDownConnection allows you to interact with the Lockdown service on the phone.
//You can use this to grab basic info from the device and start other services on the phone.
type LockDownConnection struct {
	deviceConnection DeviceConnectionInterface
	sessionID        string
	plistCodec       PlistCodec
}

//NewLockDownConnection creates a new LockDownConnection with empty sessionId and a PlistCodec.
func NewLockDownConnection(dev DeviceConnectionInterface) *LockDownConnection {
	return &LockDownConnection{deviceConnection: dev, plistCodec: NewPlistCodec()}
}

//Close closes the underlying DeviceConnection
func (lockDownConn *LockDownConnection) Close() {
	lockDownConn.StopSession()
	lockDownConn.deviceConnection.Close()
}

//DisableSessionSSL see documentation in DeviceConnection
func (lockDownConn LockDownConnection) DisableSessionSSL() {
	lockDownConn.deviceConnection.DisableSessionSSL()
}

//EnableSessionSsl see documentation in DeviceConnection
func (lockDownConn LockDownConnection) EnableSessionSsl(pairRecord PairRecord) error {
	return lockDownConn.deviceConnection.EnableSessionSsl(pairRecord)
}

//EnableSessionSslServerMode see documentation in DeviceConnection
func (lockDownConn LockDownConnection) EnableSessionSslServerMode(pairRecord PairRecord) {
	lockDownConn.deviceConnection.EnableSessionSslServerMode(pairRecord)

}

//Send takes a go struct, converts it to a PLIST and sends it with a 4 byte length field.
func (lockDownConn LockDownConnection) Send(msg interface{}) error {
	bytes, err := lockDownConn.plistCodec.Encode(msg)
	if err != nil {
		log.Error("failed lockdown send")
		return err
	}
	return lockDownConn.deviceConnection.Send(bytes)
}

//ReadMessage grabs the next LockDown Message using the PlistDecoder from the underlying
//DeviceConnection and returns the Plist as a byte slice.
func (lockDownConn *LockDownConnection) ReadMessage() ([]byte, error) {
	reader := lockDownConn.deviceConnection.Reader()
	resp, err := lockDownConn.plistCodec.Decode(reader)
	if err != nil {
		return make([]byte, 0), err
	}
	return resp, err
}

func (lockDownConn *LockDownConnection) Conn() net.Conn {
	return lockDownConn.deviceConnection.Conn()
}
