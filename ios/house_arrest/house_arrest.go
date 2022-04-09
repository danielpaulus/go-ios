package house_arrest

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/danielpaulus/go-ios/ios"

	"howett.net/plist"
)

const serviceName = "com.apple.mobile.house_arrest"

type Connection struct {
	deviceConn    ios.DeviceConnectionInterface
	packageNumber uint64
}

//NewFromConn allows to use AFC on a DeviceConnectionInterface, see crashreport for an example
func NewFromConn(deviceConn ios.DeviceConnectionInterface) *Connection {
	return &Connection{deviceConn: deviceConn}
}

//NewWithHouseArrest gives an AFC connection to an app container using the house arrest service
func NewWithHouseArrest(device ios.DeviceEntry, bundleID string) (*Connection, error) {
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

func (conn Connection) Close() error {
	return conn.deviceConn.Close()
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
	return conn.CloseHandle(handle)
}

func (conn *Connection) ListFiles(filePath string, matchPattern string) ([]string, error) {
	headerPayload := []byte(filePath)
	headerLength := uint64(len(headerPayload))

	this_length := afc_header_size + headerLength
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_read_dir, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return []string{}, err
	}
	fileList := string(response.payload)
	files := strings.Split(fileList, string([]byte{0}))
	var filteredFiles []string
	for _, f := range files {
		if f == "" {
			continue
		}
		matches, err := filepath.Match(matchPattern, f)
		if err != nil {
			log.Warn("error while matching pattern", err)
		}
		if matches {
			filteredFiles = append(filteredFiles, f)
		}
	}
	return filteredFiles, nil
}

func (conn *Connection) openFileForWriting(filePath string) (byte, error) {
	pathBytes := []byte(filePath)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, afc_fopen_wronly)
	copy(headerPayload[8:], pathBytes)
	this_length := afc_header_size + headerLength
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_open, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if response.header.Operation != afc_operation_file_open_result {
		return 0, fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return response.headerPayload[0], nil
}

func (conn *Connection) OpenFileForReading(filePath string) (byte, error) {
	pathBytes := []byte(filePath)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, afc_fopen_readonly)
	copy(headerPayload[8:], pathBytes)
	this_length := afc_header_size + headerLength
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_open, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if response.header.Operation != afc_operation_file_open_result {
		return 0, fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return response.headerPayload[0], nil
}

func (conn *Connection) StreamFile(file string, target io.Writer) error {
	handle, err := conn.OpenFileForReading(file)
	if err != nil {
		return err
	}
	log.Debugf("remote file %s open with handle %d", file, handle)

	totalBytes := 0
	bytesRead := 1
	for bytesRead > 0 {
		data, n, readErr := conn.readBytes(handle, 4096)
		if readErr != nil {
			return readErr
		}
		bytesRead = n
		totalBytes += bytesRead
		_, err = target.Write(data)
		if err != nil {
			return err
		}
	}
	log.Debugf("finished reading %d kb %s", totalBytes/1024, file)
	return conn.CloseHandle(handle)
}

func (conn *Connection) readBytes(handle byte, length int) ([]byte, int, error) {
	headerPayload := make([]byte, 16)
	binary.LittleEndian.PutUint64(headerPayload, uint64(handle))
	binary.LittleEndian.PutUint64(headerPayload[8:], uint64(length))
	payloadLength := uint64(16)
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_read, This_length: payloadLength + afc_header_size, Entire_length: payloadLength + afc_header_size}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return []byte{}, 0, err
	}
	return response.payload, len(response.payload), nil
}

func (conn *Connection) sendAfcPacketAndAwaitResponse(packet AfcPacket) (AfcPacket, error) {
	err := Encode(packet, conn.deviceConn.Writer())
	if err != nil {
		return AfcPacket{}, err
	}
	return Decode(conn.deviceConn.Reader())
}

func (conn *Connection) sendFileContents(fileContents []byte, handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_write, This_length: 8 + afc_header_size, Entire_length: 8 + afc_header_size + uint64(len(fileContents))}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: fileContents}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.header.Operation != afc_operation_status {
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return nil
}

func (conn *Connection) CloseHandle(handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
	this_length := 8 + afc_header_size
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_close, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.header.Operation != afc_operation_status {
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return nil
}

func (conn *Connection) Delete(path string) error {
	log.Debugf("Delete:%s", path)
	headerPayload := []byte(path)
	this_length := afc_header_size + uint64(len(headerPayload))
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: AFC_OP_REMOVE_PATH_AND_CONTENTS, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	log.Debugf("%+v", response)
	return err
}

func (conn *Connection) GetFileInfo(path string) (AFCFileInfo, error) {
	log.Debugf("GetFileInfo:%s", path)
	headerPayload := []byte(path)
	this_length := afc_header_size + uint64(len(headerPayload))
	header := AfcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: AfcOpGetFileInfo, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := AfcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	log.Debugf("%+v", response)
	if err != nil {
		return AFCFileInfo{}, err
	}

	if len(response.payload) != int(response.header.Entire_length-response.header.This_length) {
		return AFCFileInfo{}, fmt.Errorf("not all payload data was received")
	}
	if response.header.Operation != AfcOpData {
		return AFCFileInfo{}, fmt.Errorf("expected AfcOpData got: %d", response.header.Operation)
	}
	res := bytes.Split(response.payload, []byte{0})
	result := AFCFileInfo{}
	for i, b := range res {
		switch string(b) {
		case "st_size":
			result.St_size, _ = strconv.Atoi(string(res[i+1]))
			break
		case "st_blocks":
			result.St_blocks, _ = strconv.Atoi(string(res[i+1]))
			break
		case "st_nlink":
			result.St_nlink, _ = strconv.Atoi(string(res[i+1]))
			break
		case "st_ifmt":
			result.St_ifmt = string(res[i+1])
			break
		case "st_mtime":
			result.St_mtime, _ = strconv.ParseUint(string(res[i+1]), 0, 64)
			break
		case "st_birthtime":
			result.St_birthtime, _ = strconv.ParseUint(string(res[i+1]), 0, 64)
			break
		}
	}
	return result, nil
}

type AFCFileInfo struct {
	St_size      int
	St_blocks    int
	St_nlink     int
	St_ifmt      string
	St_mtime     uint64
	St_birthtime uint64
}

func (info AFCFileInfo) IsDir() bool {
	return info.St_ifmt == "S_IFDIR"
}
