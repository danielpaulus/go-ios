package afc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/danielpaulus/go-ios/ios"
	"golang.org/x/exp/slices"
)

// WalkFunc is used by [Client.WalkDir] for traversing directories
// This function will be called for each entry in a directory
// Execution can be controlled by returning error values from this function:
//   - [fs.SkipDir] will skip files of the current directory
//   - [fs.SkipAll] will stop traversal and exit without an error
//   - returning any other non-nil error will stop traversal and return this error from [Client.WalkDir]
type WalkFunc func(path string, info FileInfo, err error) error

type Client struct {
	connection io.ReadWriteCloser
	packetNum  atomic.Int64
}

// New creates a connection to the afc service
func New(d ios.DeviceEntry) (*Client, error) {
	deviceConn, err := ios.ConnectToService(d, serviceName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to service '%s': %w", serviceName, err)
	}
	return NewFromConn(deviceConn), nil
}

// NewFromConn establishes a new AFC client connection from an existing device connection
func NewFromConn(d ios.DeviceConnectionInterface) *Client {
	return &Client{
		connection: d,
	}
}

// Close the afc client
func (c *Client) Close() error {
	err := c.connection.Close()
	if err != nil {
		return fmt.Errorf("error closing afc client: %w", err)
	}
	return nil
}

// List all entries of the provided path
func (c *Client) List(p string) ([]string, error) {
	err := c.sendPacket(readDir, []byte(p), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing afc dir: %w", err)
	}
	pack, err := c.readPacket()
	if err != nil {
		return nil, fmt.Errorf("error listing afc dir: %w", err)
	}
	reader := bufio.NewReader(bytes.NewReader(pack.Payload))
	var list []string
	for {
		s, err := reader.ReadString('\x00')
		if err != nil {
			break
		}
		entry := s[:len(s)-1]
		if len(entry) == 0 {
			continue
		}
		if entry == "." || entry == ".." {
			continue
		}
		list = append(list, entry)
	}
	return list, nil
}

// Open opens a file with the specified name in the given mode
func (c *Client) Open(p string, mode Mode) (*File, error) {
	pathBytes := []byte(p)
	pathBytes = append(pathBytes, 0)
	headerLength := 8 + uint64(len(pathBytes))
	headerPayload := make([]byte, headerLength)
	binary.LittleEndian.PutUint64(headerPayload, uint64(mode))
	copy(headerPayload[8:], pathBytes)
	err := c.sendPacket(fileOpen, headerPayload, nil)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	resp, err := c.readPacket()
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	fd := binary.LittleEndian.Uint64(resp.HeaderPayload)

	return &File{
		client: c,
		handle: fd,
	}, nil
}

// MkDir creates a directory at the specified path
func (c *Client) MkDir(p string) error {
	headerPayload := []byte(p)
	headerPayload = append(headerPayload, 0)

	err := c.sendPacket(makeDir, headerPayload, nil)
	if err != nil {
		return fmt.Errorf("error creating dir: %w", err)
	}
	_, err = c.readPacket()
	if err != nil {
		return fmt.Errorf("error creating dir: %w", err)
	}
	return nil
}

// Remove deletes the file at the given path
// If the path is a non-empty directory, an error will be returned
func (c *Client) Remove(p string) error {
	return c.remove(p, false)
}

// RemoveAll deletes the file at the given path
// If the path is a non-empty directory, the directory and its contents will be deleted
func (c *Client) RemoveAll(p string) error {
	return c.remove(p, true)
}

func (c *Client) remove(p string, recursive bool) error {
	headerPayload := []byte(p)
	var opcode = removePath
	if recursive {
		opcode = removePathAndContents
	}
	err := c.sendPacket(opcode, headerPayload, nil)
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}
	_, err = c.readPacket()
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}
	return nil
}

func (c *Client) sendPacket(operation opcode, headerPayload []byte, payload []byte) error {
	num := c.packetNum.Add(1)

	thisLen := headerSize + uint64(len(headerPayload))
	p := packet{
		Header: header{
			Magic:     magic,
			EntireLen: thisLen + uint64(len(payload)),
			ThisLen:   thisLen,
			PacketNum: uint64(num),
			Operation: operation,
		},
		HeaderPayload: headerPayload,
		Payload:       payload,
	}

	err := binary.Write(c.connection, binary.LittleEndian, p.Header)
	if err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}
	if len(headerPayload) > 0 {
		_, err = c.connection.Write(headerPayload)
		if err != nil {
			return fmt.Errorf("error writing header payload: %w", err)
		}
	}
	if len(payload) > 0 {
		_, err = c.connection.Write(payload)
		if err != nil {
			return fmt.Errorf("error writing payload: %w", err)
		}
	}
	return nil
}

func (c *Client) readPacket() (packet, error) {
	var h header
	err := binary.Read(c.connection, binary.LittleEndian, &h)
	if err != nil {
		return packet{}, fmt.Errorf("error reading header: %w", err)
	}
	headerPayloadLen := h.ThisLen - headerSize
	payloadLen := h.EntireLen - h.ThisLen

	headerpayload := make([]byte, headerPayloadLen)
	payload := make([]byte, payloadLen)

	p := packet{
		Header:        h,
		HeaderPayload: headerpayload,
		Payload:       payload,
	}

	if headerPayloadLen > 0 {
		_, err = io.ReadFull(c.connection, headerpayload)
		if err != nil {
			return packet{}, fmt.Errorf("error reading header: %w", err)
		}
	}
	if payloadLen > 0 {
		_, err = io.ReadFull(c.connection, payload)
		if err != nil {
			return packet{}, fmt.Errorf("error reading payload: %w", err)
		}
	}

	if p.Header.Operation == status {
		code := binary.LittleEndian.Uint64(p.HeaderPayload)
		if code == errSuccess {
			return p, nil
		}
		return p, afcError{
			code: int(code),
		}
	}

	return p, nil
}

type FileType string

const (
	// S_IFDIR marks a directory
	S_IFDIR FileType = "S_IFDIR"
	// S_IFDIR marks a regular file
	S_IFMT  FileType = "S_IFMT"
	S_IFLNK FileType = "S_IFLNK"
)

type FileInfo struct {
	Name       string
	Type       FileType
	Mode       uint32
	Size       int64
	LinkTarget string
}

func (f FileInfo) IsDir() bool {
	return f.Type == S_IFDIR
}

func (f FileInfo) IsLink() bool {
	return f.Type == S_IFLNK
}

// Stat retrieves information about a given file path
func (c *Client) Stat(s string) (FileInfo, error) {
	err := c.sendPacket(fileInfo, []byte(s), nil)
	if err != nil {
		return FileInfo{}, fmt.Errorf("error getting file info: %w", err)
	}

	resp, err := c.readPacket()
	if err != nil {
		return FileInfo{}, fmt.Errorf("error getting file info: %w", err)
	}

	reader := bufio.NewReader(bytes.NewReader(resp.Payload))
	info := FileInfo{}

	// Parse the file info response which is a series of null-terminated strings
	// in key-value pairs
	for {
		key, err := reader.ReadString('\x00')
		if err != nil {
			break
		}
		if len(key) <= 1 {
			break
		}
		key = key[:len(key)-1] // Remove null terminator

		value, err := reader.ReadString('\x00')
		if err != nil {
			break
		}
		value = value[:len(value)-1] // Remove null terminator

		switch key {
		case "st_ifmt":
			info.Type = FileType(value)
		case "st_size":
			size, _ := strconv.ParseInt(value, 10, 64)
			info.Size = size
		case "st_mode":
			mode, _ := strconv.ParseUint(value, 8, 32)
			info.Mode = uint32(mode)
		case "st_linktarget":
			info.LinkTarget = value
		}
	}

	// Set the name from the path
	parts := strings.Split(s, "/")
	if len(parts) > 0 {
		info.Name = parts[len(parts)-1]
	}

	return info, nil
}

// WalkDir traverses the filesystem starting at the provided path
// It calls the WalkFunc for each file, and if the file is a directory,
// it recursively traverses the directory
func (c *Client) WalkDir(p string, f WalkFunc) error {
	files, err := c.List(p)
	if err != nil {
		if isPermissionDeniedError(err) {
			return nil
		}
		return err
	}

	slices.Sort(files)
	for _, file := range files {
		info, err := c.Stat(path.Join(p, file))
		if err != nil {
			if isPermissionDeniedError(err) {
				continue
			}
			return err
		}
		fnErr := f(path.Join(p, file), info, nil)
		if fnErr != nil {
			if errors.Is(fnErr, fs.SkipDir) {
				continue
			} else if errors.Is(fnErr, fs.SkipAll) {
				return nil
			} else {
				return fnErr
			}
		}
		if info.Type == S_IFDIR {
			walkErr := c.WalkDir(path.Join(p, file), f)
			if walkErr != nil {
				return walkErr
			}
		}
	}
	return nil
}

// DeviceInfo retrieves information about the filesystem of the device
func (c *Client) DeviceInfo() (DeviceInfo, error) {
	err := c.sendPacket(deviceInfo, nil, nil)
	if err != nil {
		return DeviceInfo{}, fmt.Errorf("error getting device info: %w", err)
	}
	resp, err := c.readPacket()
	if err != nil {
		return DeviceInfo{}, fmt.Errorf("error getting device info: %w", err)
	}

	bs := bytes.Split(resp.Payload, []byte{0})

	var info DeviceInfo
	for i := 0; i+1 < len(bs); i += 2 {
		key := string(bs[i])
		if key == "Model" {
			info.Model = string(bs[i+1])
			continue
		}
		value, intParseErr := strconv.ParseUint(string(bs[i+1]), 10, 64)
		switch key {
		case "FSTotalBytes":
			if intParseErr != nil {
				return DeviceInfo{}, fmt.Errorf("error parsing %s: %w", key, intParseErr)
			}
			info.TotalBytes = value
		case "FSFreeBytes":
			if intParseErr != nil {
				return DeviceInfo{}, fmt.Errorf("error parsing %s: %w", key, intParseErr)
			}
			info.FreeBytes = value
		case "FSBlockSize":
			if intParseErr != nil {
				return DeviceInfo{}, fmt.Errorf("error parsing %s: %w", key, intParseErr)
			}
			info.BlockSize = value
		}
	}
	return info, nil
}

type header struct {
	Magic     uint64
	EntireLen uint64
	ThisLen   uint64
	PacketNum uint64
	Operation opcode
}

type packet struct {
	Header        header
	HeaderPayload []byte
	Payload       []byte
}

// File is a reference to a file on the devices filesystem
type File struct {
	client *Client
	handle uint64
}

func (f *File) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	headerPayload := make([]byte, 16)
	binary.LittleEndian.PutUint64(headerPayload, f.handle)
	binary.LittleEndian.PutUint64(headerPayload[8:], uint64(len(p)))

	err := f.client.sendPacket(fileRead, headerPayload, nil)
	if err != nil {
		return 0, fmt.Errorf("error reading data: %w", err)
	}
	resp, err := f.client.readPacket()
	if err != nil {
		return 0, fmt.Errorf("error reading data: %w", err)
	}
	copy(p, resp.Payload)
	l := len(resp.Payload)
	if l == 0 {
		return 0, io.EOF
	}
	return l, nil
}

func (f *File) Write(p []byte) (int, error) {
	headerPayload := make([]byte, 8)
	binary.LittleEndian.PutUint64(headerPayload, f.handle)
	err := f.client.sendPacket(fileWrite, headerPayload, p)
	if err != nil {
		return 0, fmt.Errorf("error writing data: %w", err)
	}
	_, err = f.client.readPacket()
	if err != nil {
		return 0, fmt.Errorf("error reading data: %w", err)
	}
	return len(p), nil
}

func (f *File) Close() error {
	headerPayload := make([]byte, 8)
	binary.LittleEndian.PutUint64(headerPayload, f.handle)
	err := f.client.sendPacket(fileClose, headerPayload, nil)
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}
	_, err = f.client.readPacket()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}
	return nil
}

type Mode uint64

const (
	READ_ONLY                = Mode(0x00000001)
	READ_WRITE_CREATE        = Mode(0x00000002)
	WRITE_ONLY_CREATE_TRUNC  = Mode(0x00000003)
	READ_WRITE_CREATE_TRUNC  = Mode(0x00000004)
	WRITE_ONLY_CREATE_APPEND = Mode(0x00000005)
	READ_WRITE_CREATE_APPEND = Mode(0x00000006)
)
