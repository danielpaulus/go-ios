package house_arrest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	afc_magic                      uint64 = 0x4141504c36414643
	afc_header_size                uint64 = 40
	afc_fopen_wronly               uint64 = 0x3
	afc_fopen_readonly             uint64 = 0x1
	afc_operation_status           uint64 = 0x1
	afcOpData                      uint64 = 0x2
	afc_operation_read_dir         uint64 = 0x3
	afc_operation_file_open        uint64 = 0x0000000D
	afc_operation_file_close       uint64 = 0x00000014
	afc_operation_file_read        uint64 = 0x0000000F
	afc_operation_file_write       uint64 = 0x00000010
	afc_operation_file_open_result uint64 = 0x0000000E
	afcOpRemovePathAndContents     uint64 = 0x00000022
	afcOpGetFileInfo               uint64 = 0x0000000A
)

type afcPacketHeader struct {
	Magic         uint64
	Entire_length uint64
	This_length   uint64
	Packet_num    uint64
	Operation     uint64
}

type afcPacket struct {
	header        afcPacketHeader
	headerPayload []byte
	payload       []byte
}

//AFCFileInfo a struct containing file info
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



//SendFile writes fileContents to the given filePath on the device
//FIXME: right now sends the entire byte array, will probably fail for larger files
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

//ListFiles returns all files in the given directory, matching the pattern.
//Example: ListFiles(".", "*") returns all files and dirs in the current path the afc connection is in
func (conn *Connection) ListFiles(cwd string, matchPattern string) ([]string, error) {
	headerPayload := []byte(cwd)
	headerLength := uint64(len(headerPayload))

	this_length := afc_header_size + headerLength
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_read_dir, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

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

//DownloadFile streams filecontents of file from the device to the given writer.
func (conn *Connection) DownloadFile(file string, target io.Writer) error {
	handle, err := conn.openFileForReading(file)
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
	return conn.closeHandle(handle)
}

//Delete removes a given file or directory
func (conn *Connection) Delete(path string) error {
	log.Debugf("Delete:%s", path)
	headerPayload := []byte(path)
	this_length := afc_header_size + uint64(len(headerPayload))
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afcOpRemovePathAndContents, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	log.Debugf("%+v", response)
	return err
}

//GetFileInfo gets basic info about the path on the device
func (conn *Connection) GetFileInfo(path string) (AFCFileInfo, error) {
	log.Debugf("GetFileInfo:%s", path)
	headerPayload := []byte(path)
	this_length := afc_header_size + uint64(len(headerPayload))
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afcOpGetFileInfo, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	log.Debugf("%+v", response)
	if err != nil {
		return AFCFileInfo{}, err
	}

	if len(response.payload) != int(response.header.Entire_length-response.header.This_length) {
		return AFCFileInfo{}, fmt.Errorf("not all payload data was received")
	}
	if response.header.Operation != afcOpData {
		return AFCFileInfo{}, fmt.Errorf("expected afcOpData got: %d", response.header.Operation)
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

func (conn *Connection) openFileForWriting(filePath string) (byte, error) {
	pathBytes := []byte(filePath)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, afc_fopen_wronly)
	copy(headerPayload[8:], pathBytes)
	this_length := afc_header_size + headerLength
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_open, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if response.header.Operation != afc_operation_file_open_result {
		return 0, fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return response.headerPayload[0], nil
}

func (conn *Connection) openFileForReading(filePath string) (byte, error) {
	pathBytes := []byte(filePath)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, afc_fopen_readonly)
	copy(headerPayload[8:], pathBytes)
	this_length := afc_header_size + headerLength
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_open, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if response.header.Operation != afc_operation_file_open_result {
		return 0, fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return response.headerPayload[0], nil
}

func (conn *Connection) readBytes(handle byte, length int) ([]byte, int, error) {
	headerPayload := make([]byte, 16)
	binary.LittleEndian.PutUint64(headerPayload, uint64(handle))
	binary.LittleEndian.PutUint64(headerPayload[8:], uint64(length))
	payloadLength := uint64(16)
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_read, This_length: payloadLength + afc_header_size, Entire_length: payloadLength + afc_header_size}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return []byte{}, 0, err
	}
	return response.payload, len(response.payload), nil
}

func (conn *Connection) sendAfcPacketAndAwaitResponse(packet afcPacket) (afcPacket, error) {
	err := encode(packet, conn.deviceConn.Writer())
	if err != nil {
		return afcPacket{}, err
	}
	return decode(conn.deviceConn.Reader())
}

func (conn *Connection) sendFileContents(fileContents []byte, handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_write, This_length: 8 + afc_header_size, Entire_length: 8 + afc_header_size + uint64(len(fileContents))}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: fileContents}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.header.Operation != afc_operation_status {
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return nil
}

func (conn *Connection) closeHandle(handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
	this_length := 8 + afc_header_size
	header := afcPacketHeader{Magic: afc_magic, Packet_num: conn.packageNumber, Operation: afc_operation_file_close, This_length: this_length, Entire_length: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.header.Operation != afc_operation_status {
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afc_operation_status, response.header.Operation)
	}
	return nil
}

func decode(reader io.Reader) (afcPacket, error) {
	var header afcPacketHeader
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return afcPacket{}, err
	}
	if header.Magic != afc_magic {
		return afcPacket{}, fmt.Errorf("Wrong magic:%x expected: %x", header.Magic, afc_magic)
	}
	headerPayloadLength := header.This_length - afc_header_size
	headerPayload := make([]byte, headerPayloadLength)
	_, err = io.ReadFull(reader, headerPayload)
	if err != nil {
		return afcPacket{}, err
	}

	contentPayloadLength := header.Entire_length - header.This_length
	payload := make([]byte, contentPayloadLength)
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return afcPacket{}, err
	}
	return afcPacket{header, headerPayload, payload}, nil
}

func encode(packet afcPacket, writer io.Writer) error {
	err := binary.Write(writer, binary.LittleEndian, packet.header)
	if err != nil {
		return err
	}
	_, err = writer.Write(packet.headerPayload)
	if err != nil {
		return err
	}

	_, err = writer.Write(packet.payload)
	if err != nil {
		return err
	}
	return nil
}
