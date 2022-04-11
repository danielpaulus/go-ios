package afc

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

/*
Contains a basic AFC Client. I did not implement support for everything libimobiledevice has.
It only supports files and folders, no symlinks network sockets etc. I think that is usually
not really needed anyway? Let me know if you miss anything or send a PR.
*/
const (
	afcMagic                   uint64 = 0x4141504c36414643
	afcHeaderSize              uint64 = 40
	afcFopenWronly             uint64 = 0x3
	afcFopenReadonly           uint64 = 0x1
	afcOperationStatus         uint64 = 0x1
	afcOpData                  uint64 = 0x2
	afcOperationReadDir        uint64 = 0x3
	afcOperationFileOpen       uint64 = 0x0000000D
	afcOperationFileClose      uint64 = 0x00000014
	afcOperationFileRead       uint64 = 0x0000000F
	afcOperationFileWrite      uint64 = 0x00000010
	afcOperationFileOpenResult uint64 = 0x0000000E
	afcOpRemovePathAndContents uint64 = 0x00000022
	afcOpGetFileInfo           uint64 = 0x0000000A
)

type afcPacketHeader struct {
	Magic        uint64
	EntireLength uint64
	ThisLength   uint64
	PacketNum    uint64
	Operation    uint64
}

type afcPacket struct {
	header        afcPacketHeader
	headerPayload []byte
	payload       []byte
}

//FileInfo a struct containing file info
type FileInfo struct {
	StSize      int
	StBlocks    int
	StNlink     int
	StIfmt      string
	StMtime     uint64
	StBirthtime uint64
}

//IsDir checks if info.StIfmt == "S_IFDIR" which means the FileInfo is for a directory
func (info FileInfo) IsDir() bool {
	return info.StIfmt == "S_IFDIR"
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

	this_length := afcHeaderSize + headerLength
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOperationReadDir, ThisLength: this_length, EntireLength: this_length}
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
	this_length := afcHeaderSize + uint64(len(headerPayload))
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOpRemovePathAndContents, ThisLength: this_length, EntireLength: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	log.Debugf("%+v", response)
	return err
}

//GetFileInfo gets basic info about the path on the device
func (conn *Connection) GetFileInfo(path string) (FileInfo, error) {
	log.Debugf("GetFileInfo:%s", path)
	headerPayload := []byte(path)
	this_length := afcHeaderSize + uint64(len(headerPayload))
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOpGetFileInfo, ThisLength: this_length, EntireLength: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	log.Debugf("%+v", response)
	if err != nil {
		return FileInfo{}, err
	}

	if len(response.payload) != int(response.header.EntireLength-response.header.ThisLength) {
		return FileInfo{}, fmt.Errorf("not all payload data was received")
	}
	if response.header.Operation != afcOpData {
		return FileInfo{}, fmt.Errorf("expected afcOpData got: %d", response.header.Operation)
	}
	res := bytes.Split(response.payload, []byte{0})
	result := FileInfo{}
	for i, b := range res {
		switch string(b) {
		case "st_size":
			result.StSize, _ = strconv.Atoi(string(res[i+1]))
			break
		case "st_blocks":
			result.StBlocks, _ = strconv.Atoi(string(res[i+1]))
			break
		case "st_nlink":
			result.StNlink, _ = strconv.Atoi(string(res[i+1]))
			break
		case "st_ifmt":
			result.StIfmt = string(res[i+1])
			break
		case "st_mtime":
			result.StMtime, _ = strconv.ParseUint(string(res[i+1]), 0, 64)
			break
		case "st_birthtime":
			result.StBirthtime, _ = strconv.ParseUint(string(res[i+1]), 0, 64)
			break
		}
	}
	return result, nil
}

func (conn *Connection) openFileForWriting(filePath string) (byte, error) {
	pathBytes := []byte(filePath)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, afcFopenWronly)
	copy(headerPayload[8:], pathBytes)
	this_length := afcHeaderSize + headerLength
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOperationFileOpen, ThisLength: this_length, EntireLength: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if response.header.Operation != afcOperationFileOpenResult {
		return 0, fmt.Errorf("Unexpected afc response, expected %x received %x", afcOperationStatus, response.header.Operation)
	}
	return response.headerPayload[0], nil
}

func (conn *Connection) openFileForReading(filePath string) (byte, error) {
	pathBytes := []byte(filePath)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, afcFopenReadonly)
	copy(headerPayload[8:], pathBytes)
	this_length := afcHeaderSize + headerLength
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOperationFileOpen, ThisLength: this_length, EntireLength: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if response.header.Operation != afcOperationFileOpenResult {
		return 0, fmt.Errorf("Unexpected afc response, expected %x received %x", afcOperationStatus, response.header.Operation)
	}
	return response.headerPayload[0], nil
}

func (conn *Connection) readBytes(handle byte, length int) ([]byte, int, error) {
	headerPayload := make([]byte, 16)
	binary.LittleEndian.PutUint64(headerPayload, uint64(handle))
	binary.LittleEndian.PutUint64(headerPayload[8:], uint64(length))
	payloadLength := uint64(16)
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOperationFileRead, ThisLength: payloadLength + afcHeaderSize, EntireLength: payloadLength + afcHeaderSize}
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
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOperationFileWrite, ThisLength: 8 + afcHeaderSize, EntireLength: 8 + afcHeaderSize + uint64(len(fileContents))}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: fileContents}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.header.Operation != afcOperationStatus {
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afcOperationStatus, response.header.Operation)
	}
	return nil
}

func (conn *Connection) closeHandle(handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
	this_length := 8 + afcHeaderSize
	header := afcPacketHeader{Magic: afcMagic, PacketNum: conn.packageNumber, Operation: afcOperationFileClose, ThisLength: this_length, EntireLength: this_length}
	conn.packageNumber++
	packet := afcPacket{header: header, headerPayload: headerPayload, payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if response.header.Operation != afcOperationStatus {
		return fmt.Errorf("Unexpected afc response, expected %x received %x", afcOperationStatus, response.header.Operation)
	}
	return nil
}

func decode(reader io.Reader) (afcPacket, error) {
	var header afcPacketHeader
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return afcPacket{}, err
	}
	if header.Magic != afcMagic {
		return afcPacket{}, fmt.Errorf("Wrong magic:%x expected: %x", header.Magic, afcMagic)
	}
	headerPayloadLength := header.ThisLength - afcHeaderSize
	headerPayload := make([]byte, headerPayloadLength)
	_, err = io.ReadFull(reader, headerPayload)
	if err != nil {
		return afcPacket{}, err
	}

	contentPayloadLength := header.EntireLength - header.ThisLength
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
