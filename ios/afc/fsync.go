package afc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

//NewFromConn allows to use AFC on a DeviceConnectionInterface, see crashreport for an example
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

func (conn *Connection) MkDir(path string) error {
	headerPayload := []byte(path)
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

//ListFiles returns all files in the given directory, matching the pattern.
//Example: ListFiles(".", "*") returns all files and dirs in the current path the afc connection is in
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

func (conn *Connection) openFile(path string, mode uint64) (byte, error) {
	pathBytes := []byte(path)
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
	return response.HeaderPayload[0], nil
}

func (conn *Connection) closeFile(handle byte) error {
	headerPayload := make([]byte, 8)
	headerPayload[0] = handle
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
	fd, err := conn.openFile(srcPath, Afc_Mode_RDONLY)
	if err != nil {
		return err
	}
	defer conn.closeFile(fd)

	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	leftSize := fileInfo.stSize
	maxReadSize := 64 * 1024
	for leftSize > 0 {
		headerPayload := make([]byte, 16)
		headerPayload[0] = fd
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

	fd, err := conn.openFile(dstPath, Afc_Mode_WR)
	if err != nil {
		return err
	}
	defer conn.closeFile(fd)

	maxWriteSize := 64 * 1024
	chunk := make([]byte, maxWriteSize)
	for {
		n, err := f.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		headerPayload := make([]byte, 8)
		headerPayload[0] = fd
		thisLength := Afc_header_size + 8
		header := AfcPacketHeader{Magic: Afc_magic, Packet_num: conn.packageNumber, Operation: Afc_operation_file_write, This_length: thisLength, Entire_length: thisLength + uint64(n)}
		conn.packageNumber++
		packet := AfcPacket{Header: header, HeaderPayload: headerPayload, Payload: chunk}
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
