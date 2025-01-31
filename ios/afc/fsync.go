package afc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

const serviceName = "com.apple.afc"

type Connection struct {
	deviceConn    ios.DeviceConnectionInterface
	packageNumber uint64
}

type statInfo struct {
	stSize       int64
	stBlocks     int64
	stCtime      int64
	stMtime      int64
	stNlink      string
	stIfmt       string
	stLinktarget string
}

func (s *statInfo) IsDir() bool {
	return s.stIfmt == "S_IFDIR"
}

func (s *statInfo) IsLink() bool {
	return s.stIfmt == "S_IFLNK"
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return nil, err
	}
	return &Connection{deviceConn: deviceConn}, nil
}

func NewContainer(device ios.DeviceEntry, bundleID string) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, "com.apple.mobile.house_arrest")
	if err != nil {
		return nil, err
	}
	err = VendContainer(deviceConn, bundleID)
	if err != nil {
		return nil, err
	}
	return &Connection{deviceConn: deviceConn}, nil
}

func VendContainer(deviceConn ios.DeviceConnectionInterface, bundleID string) error {
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

// NewFromConn allows to use AFC on a DeviceConnectionInterface, see crashreport for an example
func NewFromConn(deviceConn ios.DeviceConnectionInterface) *Connection {
	return &Connection{deviceConn: deviceConn}
}

func (conn *Connection) sendAfcPacketAndAwaitResponse(packet AfcPacket) (AfcPacket, error) {
	err := Encode(packet, conn.deviceConn.Writer())
	if err != nil {
		return AfcPacket{}, err
	}
	return Decode(conn.deviceConn.Reader())
}

func (conn *Connection) checkOperationStatus(packet AfcPacket) error {
	if packet.Header.Operation == Afc_operation_status {
		errorCode := binary.LittleEndian.Uint64(packet.HeaderPayload)
		if errorCode != Afc_Err_Success {
			return getError(errorCode)
		}
	}
	return nil
}

func (conn *Connection) Remove(path string) error {
	headerPayload := []byte(path)
	headerLength := uint64(len(headerPayload))
	thisLength := Afc_header_size + headerLength

	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_remove_path, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return fmt.Errorf("remove: unexpected afc status: %v", err)
	}
	return nil
}

func (conn *Connection) RemovePathAndContents(path string) error {
	headerPayload := []byte(path)
	headerPayload = append(headerPayload, 0)
	headerLength := uint64(len(headerPayload))
	thisLength := Afc_header_size + headerLength

	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_remove_path_and_contents, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return fmt.Errorf("remove: unexpected afc status: %v", err)
	}
	return nil
}

func (conn *Connection) RemoveAll(srcPath string) error {
	fileInfo, err := conn.Stat(srcPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		fileList, err := conn.listDir(srcPath)
		if err != nil {
			return err
		}
		for _, v := range fileList {
			sp := path.Join(srcPath, v)
			err = conn.RemoveAll(sp)
			if err != nil {
				return err
			}
		}
	}
	return conn.Remove(srcPath)
}

func (conn *Connection) MkDir(path string) error {
	headerPayload := []byte(path)
	headerPayload = append(headerPayload, 0)
	headerLength := uint64(len(headerPayload))
	thisLength := Afc_header_size + headerLength

	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_make_dir, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return fmt.Errorf("mkdir: unexpected afc status: %v", err)
	}
	return nil
}

func (conn *Connection) Stat(path string) (*statInfo, error) {
	headerPayload := []byte(path)
	headerLength := uint64(len(headerPayload))
	thisLength := Afc_header_size + headerLength

	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_file_info, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return nil, err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return nil, fmt.Errorf("stat: unexpected afc status: %v", err)
	}
	ret := bytes.Split(response.Payload, []byte{0})
	retLen := len(ret)
	if retLen%2 != 0 {
		retLen = retLen - 1
	}
	statInfoMap := make(map[string]string)
	for i := 0; i <= retLen-2; i = i + 2 {
		k := string(ret[i])
		v := string(ret[i+1])
		statInfoMap[k] = v
	}

	var si statInfo
	si.stSize, _ = strconv.ParseInt(statInfoMap["st_size"], 10, 64)
	si.stBlocks, _ = strconv.ParseInt(statInfoMap["st_blocks"], 10, 64)
	si.stCtime, _ = strconv.ParseInt(statInfoMap["st_birthtime"], 10, 64)
	si.stMtime, _ = strconv.ParseInt(statInfoMap["st_mtime"], 10, 64)
	si.stNlink = statInfoMap["st_nlink"]
	si.stIfmt = statInfoMap["st_ifmt"]
	si.stLinktarget = statInfoMap["st_linktarget"]
	return &si, nil
}

func (conn *Connection) listDir(path string) ([]string, error) {
	headerPayload := []byte(path)
	headerLength := uint64(len(headerPayload))
	thisLength := Afc_header_size + headerLength

	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_read_dir, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return nil, err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return nil, fmt.Errorf("list dir: unexpected afc status: %v", err)
	}
	ret := bytes.Split(response.Payload, []byte{0})
	var fileList []string
	for _, v := range ret {
		if string(v) != "." && string(v) != ".." && string(v) != "" {
			fileList = append(fileList, string(v))
		}
	}
	return fileList, nil
}

func (conn *Connection) GetSpaceInfo() (*AFCDeviceInfo, error) {
	thisLength := Afc_header_size
	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_device_info, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: nil, Payload: nil}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return nil, err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return nil, fmt.Errorf("mkdir: unexpected afc status: %v", err)
	}

	bs := bytes.Split(response.Payload, []byte{0})
	strs := make([]string, len(bs)-1)
	for i := 0; i < len(strs); i++ {
		strs[i] = string(bs[i])
	}
	m := make(map[string]string)
	if strs != nil {
		for i := 0; i < len(strs); i += 2 {
			m[strs[i]] = strs[i+1]
		}
	}

	totalBytes, err := strconv.ParseUint(m["FSTotalBytes"], 10, 64)
	if err != nil {
		return nil, err
	}
	freeBytes, err := strconv.ParseUint(m["FSFreeBytes"], 10, 64)
	if err != nil {
		return nil, err
	}
	blockSize, err := strconv.ParseUint(m["FSBlockSize"], 10, 64)
	if err != nil {
		return nil, err
	}

	return &AFCDeviceInfo{
		Model:      m["Model"],
		TotalBytes: totalBytes,
		FreeBytes:  freeBytes,
		BlockSize:  blockSize,
	}, nil
}

// ListFiles returns all files in the given directory, matching the pattern.
// Example: ListFiles(".", "*") returns all files and dirs in the current path the afc connection is in
func (conn *Connection) ListFiles(cwd string, matchPattern string) ([]string, error) {
	headerPayload := []byte(cwd)
	headerLength := uint64(len(headerPayload))

	thisLength := Afc_header_size + headerLength
	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_read_dir, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return nil, err
	}
	fileList := string(response.Payload)
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

func (conn *Connection) TreeView(dpath string, prefix string, treePoint bool) error {
	fileInfo, err := conn.Stat(dpath)
	if err != nil {
		return err
	}
	namePrefix := "`--"
	if !treePoint {
		namePrefix = "|--"
	}
	tPrefix := prefix + namePrefix
	if fileInfo.IsDir() {
		fmt.Printf("%s %s/\n", tPrefix, filepath.Base(dpath))
		fileList, err := conn.listDir(dpath)
		if err != nil {
			return err
		}
		for i, v := range fileList {
			tp := false
			if i == len(fileList)-1 {
				tp = true
			}
			rp := prefix + "    "
			if !treePoint {
				rp = prefix + "|   "
			}
			nPath := path.Join(dpath, v)
			err = conn.TreeView(nPath, rp, tp)
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("%s %s\n", tPrefix, filepath.Base(dpath))
	}
	return nil
}

func (conn *Connection) OpenFile(path string, mode uint64) (uint64, error) {
	pathBytes := []byte(path)
	pathBytes = append(pathBytes, 0)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, mode)
	copy(headerPayload[8:], pathBytes)
	thisLength := Afc_header_size + headerLength
	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_file_open, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}

	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return 0, err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return 0, fmt.Errorf("open file: unexpected afc status: %v", err)
	}
	fd := binary.LittleEndian.Uint64(response.HeaderPayload)
	if fd == 0 {
		return 0, fmt.Errorf("file descriptor should not be zero")
	}

	return fd, nil
}

func (conn *Connection) CloseFile(fd uint64) error {
	headerPayload := make([]byte, 8)
	binary.LittleEndian.PutUint64(headerPayload, fd)
	thisLength := 8 + Afc_header_size
	header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_file_close, This_length: thisLength, Entire_length: thisLength}
	conn.packageNumber++
	packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
	response, err := conn.sendAfcPacketAndAwaitResponse(packet)
	if err != nil {
		return err
	}
	if err = conn.checkOperationStatus(response); err != nil {
		return fmt.Errorf("close file: unexpected afc status: %v", err)
	}
	return nil
}

func (conn *Connection) PullSingleFile(srcPath, dstPath string) error {
	fileInfo, err := conn.Stat(srcPath)
	if err != nil {
		return err
	}
	if fileInfo.IsLink() {
		srcPath = fileInfo.stLinktarget
	}
	fd, err := conn.OpenFile(srcPath, Afc_Mode_RDONLY)
	if err != nil {
		return err
	}
	defer conn.CloseFile(fd)

	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	leftSize := fileInfo.stSize
	maxReadSize := 64 * 1024
	for leftSize > 0 {
		headerPayload := make([]byte, 16)
		binary.LittleEndian.PutUint64(headerPayload, fd)
		thisLength := Afc_header_size + 16
		binary.LittleEndian.PutUint64(headerPayload[8:], uint64(maxReadSize))
		header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_file_read, This_length: thisLength, Entire_length: thisLength}
		conn.packageNumber++
		packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: make([]byte, 0)}
		response, err := conn.sendAfcPacketAndAwaitResponse(packet)
		if err != nil {
			return err
		}
		if err = conn.checkOperationStatus(response); err != nil {
			return fmt.Errorf("read file: unexpected afc status: %v", err)
		}
		leftSize = leftSize - int64(len(response.Payload))
		f.Write(response.Payload)
	}
	return nil
}

func (conn *Connection) Pull(srcPath, dstPath string) error {
	fileInfo, err := conn.Stat(srcPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		ret, _ := ios.PathExists(dstPath)
		if !ret {
			err = os.MkdirAll(dstPath, os.ModePerm)
			if err != nil {
				return err
			}
		}
		fileList, err := conn.listDir(srcPath)
		if err != nil {
			return err
		}
		for _, v := range fileList {
			sp := path.Join(srcPath, v)
			dp := path.Join(dstPath, v)
			err = conn.Pull(sp, dp)
			if err != nil {
				return err
			}
		}
	} else {
		return conn.PullSingleFile(srcPath, dstPath)
	}
	return nil
}

func (conn *Connection) Push(srcPath, dstPath string) error {
	ret, _ := ios.PathExists(srcPath)
	if !ret {
		return fmt.Errorf("%s: no such file.", srcPath)
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if fileInfo, _ := conn.Stat(dstPath); fileInfo != nil {
		if fileInfo.IsDir() {
			dstPath = path.Join(dstPath, filepath.Base(srcPath))
		}
	}

	return conn.WriteToFile(f, dstPath)
}

func (conn *Connection) WriteToFile(reader io.Reader, dstPath string) error {
	if fileInfo, _ := conn.Stat(dstPath); fileInfo != nil {
		if fileInfo.IsDir() {
			return fmt.Errorf("%s is a directory, cannot write to it as file", dstPath)
		}
	}

	fd, err := conn.OpenFile(dstPath, Afc_Mode_WR)
	if err != nil {
		return err
	}
	defer conn.CloseFile(fd)

	maxWriteSize := 64 * 1024
	chunk := make([]byte, maxWriteSize)
	for {
		n, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		bytesRead := chunk[:n]
		headerPayload := make([]byte, 8)
		binary.LittleEndian.PutUint64(headerPayload, fd)
		thisLength := Afc_header_size + 8
		header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_file_write, This_length: thisLength, Entire_length: thisLength + uint64(n)}
		conn.packageNumber++
		packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: bytesRead}
		response, err := conn.sendAfcPacketAndAwaitResponse(packet)
		if err != nil {
			return err
		}
		if err = conn.checkOperationStatus(response); err != nil {
			return fmt.Errorf("write file: unexpected afc status: %v", err)
		}

	}
	return nil
}

func (conn *Connection) Close() {
	conn.deviceConn.Close()
}
