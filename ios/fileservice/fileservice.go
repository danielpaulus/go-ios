// Package fileservice provides functions to pull and push files on iOS 17+ devices using RemoteXPC.
package fileservice

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

const (
	// ControlServiceName is the RemoteXPC service name for file operations control
	ControlServiceName = "com.apple.coredevice.fileservice.control"
	// DataServiceName is the RemoteXPC service name for file data transfer
	DataServiceName = "com.apple.coredevice.fileservice.data"

	// MaxFileSize is the maximum file size we'll process (1GB) to prevent OOM
	MaxFileSize = 1024 * 1024 * 1024
	// MaxInlineDataSize is the maximum file size that can be sent inline in XPC message
	// Files larger than this must use the data service
	MaxInlineDataSize = 500 // bytes - based on Ghidra code checking for 500 bytes

	// defaultReceiveTimeout bounds how long ListDirectory waits for the device
	// to answer a RetrieveDirectoryList request before giving up.
	defaultReceiveTimeout = 30 * time.Second
)

// ErrTimeout is returned when the device does not answer a request within the
// receive timeout. The connection should be closed afterwards, as a response
// may still arrive on one of its streams.
var ErrTimeout = errors.New("timed out waiting for a response from the device")

// DeviceError is an error reported by the device's file service in the
// EncodedError field of a response, e.g. RemoteServices error 11007
// "File paths cannot contain '..'." on iOS 26.5.x App Group containers.
type DeviceError struct {
	// Description is the LocalizedDescription reported by the device, if any.
	Description string
	// Encoded is the raw EncodedError value from the response.
	Encoded interface{}
}

func (e *DeviceError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("device error: %s", e.Description)
	}
	return fmt.Sprintf("device error: %+v", e.Encoded)
}

// Domain represents a file system domain on the iOS device
type Domain uint64

const (
	// DomainAppDataContainer is the app's Documents directory
	DomainAppDataContainer Domain = 1
	// DomainAppGroupDataContainer is the app group shared container
	DomainAppGroupDataContainer Domain = 2
	// DomainTemporary is the temporary directory
	DomainTemporary Domain = 3
	// DomainRootStaging is the root staging directory (no idea what that would be)
	DomainRootStaging Domain = 4
	// DomainSystemCrashLogs is the system crash logs directory
	DomainSystemCrashLogs Domain = 5
)

// controlConnection is the subset of *xpc.Connection used by the file service
// control channel. It exists as a seam so the control protocol can be unit
// tested without a device.
type controlConnection interface {
	Send(data map[string]interface{}, flags ...uint32) error
	ReceiveOnClientServerStream() (map[string]interface{}, error)
	ReceiveOnServerClientStream() (map[string]interface{}, error)
	Close() error
}

// receiveResult carries the outcome of a stream receive across a channel.
type receiveResult struct {
	response map[string]interface{}
	err      error
}

// Connection represents a connection to the file service on an iOS 17+ device.
// It manages file operations like listing, pulling, and pushing files.
// Note: Connection is not safe for concurrent use. Each goroutine should have its own Connection.
type Connection struct {
	mu         sync.Mutex
	conn       controlConnection
	device     ios.DeviceEntry
	sessionID  string
	domain     Domain
	identifier string
	// receiveTimeout bounds how long ListDirectory waits for a response.
	receiveTimeout time.Duration
	// pendingControl holds an in-flight read of the control (server->client)
	// stream started by ListDirectory that no message arrived on yet. It is
	// consumed by the next control receive so the stream never has two readers.
	pendingControl chan receiveResult
}

// New creates a new connection to the file service on the device for iOS 17+.
// The domain parameter specifies which file system domain to access.
// The identifier parameter is typically an app bundle ID (e.g., "com.example.app") for app domains.
// For system domains like DomainSystemCrashLogs, the identifier can be empty.
func New(device ios.DeviceEntry, domain Domain, identifier string) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(device, ControlServiceName)
	if err != nil {
		return nil, fmt.Errorf("New: failed to connect to file service: %w", err)
	}

	c := &Connection{
		conn:           xpcConn,
		device:         device,
		domain:         domain,
		identifier:     identifier,
		receiveTimeout: defaultReceiveTimeout,
	}

	// Create session immediately
	if err := c.createSession(); err != nil {
		xpcConn.Close()
		return nil, fmt.Errorf("New: failed to create session: %w", err)
	}

	return c, nil
}

// createSession sends a CreateSession command to establish a file service session
func (c *Connection) createSession() error {
	request := map[string]interface{}{
		"Cmd":        "CreateSession",
		"Domain":     uint64(c.domain),
		"Identifier": c.identifier,
		"Session":    "",
		"User":       "mobile",
	}

	if err := c.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("createSession: failed to send request: %w", err)
	}

	// Session creation responses come on ServerClientStream
	response, err := c.receiveControl()
	if err != nil {
		return fmt.Errorf("createSession: failed to receive response: %w", err)
	}

	// Check for errors in response
	if err := extractError(response); err != nil {
		return fmt.Errorf("createSession: %w", err)
	}

	// Extract session ID
	sessionID, ok := response["NewSessionID"].(string)
	if !ok {
		return fmt.Errorf("createSession: missing or invalid NewSessionID in response (got: %+v)", response)
	}

	c.sessionID = sessionID
	return nil
}

// receiveControl receives the next message on the control (server->client)
// stream, consuming a pending read left behind by ListDirectory if one exists.
func (c *Connection) receiveControl() (map[string]interface{}, error) {
	if c.pendingControl != nil {
		res := <-c.pendingControl
		c.pendingControl = nil
		return res.response, res.err
	}
	return c.conn.ReceiveOnServerClientStream()
}

// ListDirectory returns a list of file names in the specified directory path.
// The path is relative to the domain root.
//
// The listing itself arrives on the client-server stream, but device-side
// failures (e.g. RemoteServices error 11007 on iOS 26.5.x App Group
// containers, see issue #784) are reported on the control (server->client)
// stream like every other control response. Both streams are therefore
// received concurrently with a timeout, so a failure surfaces as a typed
// error (*DeviceError or ErrTimeout) instead of blocking forever.
// After an error the connection should be closed.
func (c *Connection) ListDirectory(path string) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msgUUID := uuid.New().String()

	request := map[string]interface{}{
		"Cmd":         "RetrieveDirectoryList",
		"MessageUUID": msgUUID,
		"Path":        path,
		"SessionID":   c.sessionID,
	}

	if err := c.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		return nil, fmt.Errorf("ListDirectory: failed to send request: %w", err)
	}

	// Directory list responses come on ClientServerStream
	listCh := make(chan receiveResult, 1)
	go func() {
		response, err := c.conn.ReceiveOnClientServerStream()
		listCh <- receiveResult{response, err}
	}()

	// Watch the control stream for an error response. Reuse a pending read
	// from a previous call so the stream never has two concurrent readers.
	controlCh := c.pendingControl
	if controlCh == nil {
		controlCh = make(chan receiveResult, 1)
		conn := c.conn
		go func() {
			response, err := conn.ReceiveOnServerClientStream()
			controlCh <- receiveResult{response, err}
		}()
	}
	c.pendingControl = controlCh

	timeout := time.NewTimer(c.receiveTimeout)
	defer timeout.Stop()

	select {
	case res := <-listCh:
		if res.err != nil {
			return nil, fmt.Errorf("ListDirectory: failed to receive response: %w", res.err)
		}
		if err := extractError(res.response); err != nil {
			return nil, fmt.Errorf("ListDirectory: %w", err)
		}
		return parseFileList(res.response)
	case res := <-controlCh:
		c.pendingControl = nil
		if res.err != nil {
			return nil, fmt.Errorf("ListDirectory: failed to receive control response: %w", res.err)
		}
		if err := extractError(res.response); err != nil {
			return nil, fmt.Errorf("ListDirectory: %w", err)
		}
		return nil, fmt.Errorf("ListDirectory: unexpected control response without file list (got: %+v)", res.response)
	case <-timeout.C:
		return nil, fmt.Errorf("ListDirectory: no response within %v: %w", c.receiveTimeout, ErrTimeout)
	}
}

// parseFileList extracts the FileList entries from a RetrieveDirectoryList response.
func parseFileList(response map[string]interface{}) ([]string, error) {
	fileListRaw, ok := response["FileList"]
	if !ok {
		return nil, fmt.Errorf("ListDirectory: missing FileList in response")
	}

	fileList, ok := fileListRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("ListDirectory: FileList is not an array")
	}

	// Convert to string slice
	result := make([]string, 0, len(fileList))
	for _, item := range fileList {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}

	return result, nil
}

// PullFile downloads a file from the device by streaming to an io.Writer.
// This is memory-efficient as it streams the file in chunks rather than loading it entirely into memory.
// The path is relative to the domain root.
func (c *Connection) PullFile(path string, writer io.Writer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Step 1: Send RetrieveFile command to control service
	request := map[string]interface{}{
		"Cmd":       "RetrieveFile",
		"Path":      path,
		"SessionID": c.sessionID,
	}

	if err := c.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("PullFile: failed to send request: %w", err)
	}

	// File retrieval control responses come on ServerClientStream (like CreateSession)
	response, err := c.receiveControl()
	if err != nil {
		return fmt.Errorf("PullFile: failed to receive response: %w", err)
	}

	// Check for errors in response
	if err := extractError(response); err != nil {
		return fmt.Errorf("PullFile: %w", err)
	}

	// Extract response token and file ID
	responseToken, ok := response["Response"].(uint64)
	if !ok {
		return fmt.Errorf("PullFile: missing or invalid Response token")
	}

	fileID, ok := response["NewFileID"].(uint64)
	if !ok {
		return fmt.Errorf("PullFile: missing or invalid NewFileID")
	}

	// Step 2: Connect to data service and download file
	if err := c.downloadFileData(responseToken, fileID, writer); err != nil {
		return fmt.Errorf("PullFile: failed to download file data: %w", err)
	}

	return nil
}

// downloadFileData connects to the data service and streams file contents to a writer
func (c *Connection) downloadFileData(responseToken, fileID uint64, writer io.Writer) error {
	// Connect to data service using raw connection (not XPC)
	dataConn, err := ios.ConnectToServiceTunnelIface(c.device, DataServiceName)
	if err != nil {
		return fmt.Errorf("downloadFileData: failed to connect to data service: %w", err)
	}
	defer dataConn.Close()

	// Build the wire protocol request
	// Protocol: magic (8 bytes) + response token (8 bytes BE) + padding (8 bytes) + file ID (8 bytes BE) + padding (8 bytes)
	wireRequest := make([]byte, 40)

	// Magic: "rwb!FILE" (8 bytes)
	copy(wireRequest[0:8], []byte("rwb!FILE"))

	// Response token (8 bytes, big-endian)
	binary.BigEndian.PutUint64(wireRequest[8:16], responseToken)

	// Padding (8 bytes of zeros) - wireRequest is already zero-initialized

	// File ID (8 bytes, big-endian)
	binary.BigEndian.PutUint64(wireRequest[24:32], fileID)

	// Padding (8 bytes of zeros) - wireRequest is already zero-initialized

	// Send the wire protocol request through the raw connection
	if err := sendRawData(dataConn, wireRequest); err != nil {
		return fmt.Errorf("downloadFileData: failed to send wire request: %w", err)
	}

	// Receive and stream file data - Protocol: 36 bytes header + 4 bytes file size (BE) + file data
	if err := receiveFileDataToWriter(dataConn, writer); err != nil {
		return fmt.Errorf("downloadFileData: failed to receive file data: %w", err)
	}

	return nil
}

// PushFile uploads a file to the device by streaming from an io.Reader.
// This is memory-efficient as it streams the file in chunks rather than loading it entirely into memory.
// The path is relative to the domain root.
// permissions should be in octal format (e.g., 0o644 or 420 in decimal)
func (c *Connection) PushFile(path string, reader io.Reader, fileSize int64, permissions int64, uid, gid int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate file size
	if fileSize > MaxFileSize {
		return fmt.Errorf("PushFile: file size %d exceeds maximum allowed size %d", fileSize, MaxFileSize)
	}

	// Get current time for file timestamps
	now := time.Now().Unix()

	// Use ProposeFile for files with data, ProposeEmptyFile for empty files
	cmd := "ProposeFile"
	if fileSize == 0 {
		cmd = "ProposeEmptyFile"
	}

	request := map[string]interface{}{
		"Cmd":                      cmd,
		"FileCreationTime":         now,
		"FileLastModificationTime": now,
		"FilePermissions":          permissions,
		"FileOwnerUserID":          uid,
		"FileOwnerGroupID":         gid,
		"Path":                     path,
		"SessionID":                c.sessionID,
	}

	// For small files (<= 500 bytes), read data and send inline in XPC message
	// For larger files, send via data service after ProposeFile
	if cmd == "ProposeFile" {
		request["FileSize"] = uint64(fileSize)
		if fileSize <= MaxInlineDataSize {
			// Small file: read into memory and include inline
			data := make([]byte, fileSize)
			_, err := io.ReadFull(reader, data)
			if err != nil {
				return fmt.Errorf("PushFile: failed to read small file: %w", err)
			}
			request["FileData"] = data
		}
	}

	if err := c.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("PushFile: failed to send request: %w", err)
	}

	// Propose responses come on ServerClientStream (like other control commands)
	response, err := c.receiveControl()
	if err != nil {
		return fmt.Errorf("PushFile: failed to receive response: %w", err)
	}

	// Check for errors in response
	if err := extractError(response); err != nil {
		return fmt.Errorf("PushFile: %w", err)
	}

	// For large files, stream data via data service
	if cmd == "ProposeFile" && fileSize > MaxInlineDataSize {
		// Extract response token and file ID for data service upload
		responseToken, ok := response["Response"].(uint64)
		if !ok {
			return fmt.Errorf("PushFile: missing or invalid Response token")
		}

		fileID, ok := response["NewFileID"].(uint64)
		if !ok {
			return fmt.Errorf("PushFile: missing or invalid NewFileID")
		}

		if err := c.uploadFileData(responseToken, fileID, reader, fileSize); err != nil {
			return fmt.Errorf("PushFile: failed to upload file data: %w", err)
		}
	}

	return nil
}

// uploadFileData uploads file data via the data service by streaming from an io.Reader
func (c *Connection) uploadFileData(responseToken, fileID uint64, reader io.Reader, fileSize int64) error {
	// Connect to data service using raw connection (not XPC)
	dataConn, err := ios.ConnectToServiceTunnelIface(c.device, DataServiceName)
	if err != nil {
		return fmt.Errorf("uploadFileData: failed to connect to data service: %w", err)
	}
	defer dataConn.Close()

	// Build and send the wire protocol header
	// Based on reverse engineering DTRemoteServices FUN_000221f8:
	// Offset 0: magic "rwb!FILE" (8 bytes)
	// Offset 8: token = 0 (8 bytes big-endian)
	// Offset 16: padding (8 bytes)
	// Offset 24: file ID (8 bytes big-endian)
	// Offset 32: file size (8 bytes big-endian)
	// Offset 40: file data (streamed)
	header := make([]byte, 40)

	// Magic: "rwb!FILE" (8 bytes)
	copy(header[0:8], []byte("rwb!FILE"))

	// Response token (8 bytes, big-endian) - Set to 0 for uploads
	binary.BigEndian.PutUint64(header[8:16], 0)

	// Padding (8 bytes of zeros) - header is already zero-initialized

	// File ID (8 bytes, big-endian)
	binary.BigEndian.PutUint64(header[24:32], fileID)

	// File size (8 bytes, big-endian)
	binary.BigEndian.PutUint64(header[32:40], uint64(fileSize))

	// Send the header
	if err := sendRawData(dataConn, header); err != nil {
		return fmt.Errorf("uploadFileData: failed to send header: %w", err)
	}

	// Stream the file data in chunks
	const chunkSize = 256 * 1024 // 256KB chunks
	buffer := make([]byte, chunkSize)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if err := sendRawData(dataConn, buffer[:n]); err != nil {
				return fmt.Errorf("uploadFileData: failed to send chunk: %w", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("uploadFileData: failed to read chunk: %w", err)
		}
	}

	// Wait for confirmation from data service
	confirmHeader := make([]byte, 32)
	_, err = io.ReadFull(dataConn, confirmHeader)
	if err != nil {
		return fmt.Errorf("uploadFileData: failed to read confirmation: %w", err)
	}

	// Verify confirmation
	confirmMagic := string(confirmHeader[0:8])
	if confirmMagic != "rwb!FILE" {
		return fmt.Errorf("uploadFileData: invalid confirmation magic: %s", confirmMagic)
	}

	return nil
}

// Close closes the connection to the file service
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Close()
}

// extractError checks if the response contains an error and returns it as a *DeviceError
func extractError(response map[string]interface{}) error {
	if encodedError, ok := response["EncodedError"]; ok && encodedError != nil {
		devErr := &DeviceError{Encoded: encodedError}
		// Try to get localized description first
		if errorMap, ok := encodedError.(map[string]interface{}); ok {
			if desc, ok := errorMap["LocalizedDescription"].(string); ok && desc != "" {
				devErr.Description = desc
			}
		}
		return devErr
	}
	return nil
}
