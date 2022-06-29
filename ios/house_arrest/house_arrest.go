package house_arrest

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/danielpaulus/go-ios/ios/afc"
	"strings"

	"github.com/danielpaulus/go-ios/ios"

	"howett.net/plist"
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
	err = vendContainer(deviceConn, bundleID)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: deviceConn}, nil
}

func vendContainer(deviceConn ios.DeviceConnectionInterface, bundleID string) error {
	plistCodec := ios.NewPlistCodec()
	vendContainer := map[string]interface{}{"Command": "VendContainer", "Identifier": bundleID}
	msg, err := plistCodec.Encode(vendContainer)
	if err != nil {
		return fmt.Errorf("VendContainer Encoding cannot fail unless the encoder is broken: %v", err)
	}
	err = deviceConn.Send(msg)
	if err != nil {
		return err
	}
	reader := deviceConn.Reader()
	response, err := plistCodec.Decode(reader)
	if err != nil {
		return err
	}
	return checkResponse(response)
}

func checkResponse(vendContainerResponseBytes []byte) error {
	response, err := plistFromBytes(vendContainerResponseBytes)
	if err != nil {
		return err
	}
	if "Complete" == response.Status {
		return nil
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return errors.New("unknown error during vendcontainer")
}

func plistFromBytes(plistBytes []byte) (vendContainerResponse, error) {
	var vendResponse vendContainerResponse
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))

	err := decoder.Decode(&vendResponse)
	if err != nil {
		return vendResponse, err
	}
	return vendResponse, nil
}

type vendContainerResponse struct {
	Status string
	Error  string
}

func (c Connection) Close() {
	c.deviceConn.Close()
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
		return 0, fmt.Errorf("Unexpected afc response, expected %x received %x", afc.Afc_operation_status, response.Header.Operation)
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
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afc.Afc_operation_status, response.Header.Operation)
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
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afc.Afc_operation_status, response.Header.Operation)
	}
	return nil
}
