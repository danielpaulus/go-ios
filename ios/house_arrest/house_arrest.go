package house_arrest

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/danielpaulus/go-ios/ios/afc"

	"github.com/danielpaulus/go-ios/ios"
)

const serviceName = "com.apple.mobile.house_arrest"

type Connection struct {
	deviceConn    ios.DeviceConnectionInterface
	packageNumber uint64
}

func New(device ios.DeviceEntry, bundleID string) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	err = afc.VendContainer(deviceConn, bundleID)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn}, nil
}

func (c Connection) Close() {
	if c.deviceConn != nil {
		c.deviceConn.Close()
	}
}

func (conn *Connection) SendFile(fileContents []byte, filePath string) error {
	handle, err := conn.openFileForWriting(filePath)
	if err != nil {
		return err
	}
	err = conn.sendFileContents(fileContents, handle)
	if err != nil {
		return err
	}
	return conn.closeHandle(handle)
}

func (conn *Connection) ListFiles(filePath string) ([]string, error) {
	headerPayload := []byte(filePath)
	headerLength := uint64(len(headerPayload))

	this_length := afc.Afc_header_size + headerLength
	header := afc.AfcPacketHeader{Magic: afc.Afc_magic, Packet_num: conn.packageNumber, Operation: afc.Afc_operation_read_dir, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afc.AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return []string{}, err
	}
	fileList := string(response.Payload)
	return strings.Split(fileList, string([]byte{0})), nil
}

func (conn *Connection) openFileForWriting(filePath string) (byte, error) {
	pathBytes := []byte(filePath)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, afc.Afc_Mode_WRONLY)
	copy(headerPayload[8:], pathBytes)
	this_length := afc.Afc_header_size + headerLength
	header := afc.AfcPacketHeader{Magic: afc.Afc_magic, Packet_num: conn.packageNumber, Operation: afc.Afc_operation_file_open, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afc.AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if response.Header.Operation != afc.Afc_operation_file_open_result {
		return 0, fmt.Errorf("unexpected afc response, expected %x received %x", afc.Afc_operation_status, response.Header.Operation)
	}
	return response.HeaderPayload[0], nil
}

func (conn *Connection) sendAfcPacketAndAwaitResponse(packet afc.AfcPacket) (afc.AfcPacket, error) {
	err := afc.Encode(packet, conn.deviceConn.Writer())
	if err != nil {
		return afc.AfcPacket{}, err
	}
	return afc.Decode(conn.deviceConn.Reader())
}

func (conn *Connection) sendFileContents(fileContents []byte, handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
	header := afc.AfcPacketHeader{Magic: afc.Afc_magic, Packet_num: conn.packageNumber, Operation: afc.Afc_operation_file_write, This_length: 8 + afc.Afc_header_size, Entire_length: 8 + afc.Afc_header_size + uint64(len(fileContents))}
	conn.packageNumber++
	packet := afc.AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: fileContents}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.Header.Operation != afc.Afc_operation_status {
		return fmt.Errorf("unexpected afc response, expected %x received %x", afc.Afc_operation_status, response.Header.Operation)
	}
	return nil
}

func (conn *Connection) closeHandle(handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
	this_length := 8 + afc.Afc_header_size
	header := afc.AfcPacketHeader{Magic: afc.Afc_magic, Packet_num: conn.packageNumber, Operation: afc.Afc_operation_file_close, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afc.AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.Header.Operation != afc.Afc_operation_status {
		return fmt.Errorf("unexpected afc response, expected %x received %x", afc.Afc_operation_status, response.Header.Operation)
	}
	return nil
}
