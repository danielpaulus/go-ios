// Package fetchsymbols downloads the dyld shared cache from iOS 17+ devices using the
// com.apple.dt.remoteFetchSymbols RemoteXPC service. This is what Xcode does when it shows
// "Preparing debugger support" and enables local symbolication without a device connection.
// The protocol mirrors pymobiledevice3's RemoteFetchSymbolsService: request the list of
// shared cache files with a DSCFilePaths message, receive one metadata message per file and
// stream each file's raw content over a dedicated HTTP2 side stream.
package fetchsymbols

import (
	"fmt"
	"io"
	"math"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/http"
	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
)

const logModule = "go-ios/fetchsymbols"

const serviceName = "com.apple.dt.remoteFetchSymbols"

// maxFileCount bounds the device-declared number of shared cache files so a malformed
// response cannot drive an unbounded receive loop. Real devices report a handful of files.
const maxFileCount = 256

// FileInfo describes one dyld shared cache file offered by the device.
type FileInfo struct {
	// FilePath is the absolute on-device path of the file.
	FilePath string
	// Size is the expected raw byte length of the file.
	Size uint64
}

// Connection is a client for the remoteFetchSymbols service. It is not safe for
// concurrent use.
type Connection struct {
	h    *http.HttpConnection
	xpc  *xpc.Connection
	udid string
}

// New connects to the remoteFetchSymbols service on an iOS 17+ device. It requires a
// running tunnel.
func New(device ios.DeviceEntry) (*Connection, error) {
	if !device.SupportsRsd() {
		return nil, fmt.Errorf("New: cannot connect to %s, missing tunnel address and RSD port. To start the tunnel, run `ios tunnel start`", serviceName)
	}
	port, err := ios.RsdPortForService(device.Rsd, serviceName)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}
	conn, err := ios.ConnectTUNDevice(device.Address, port, device)
	if err != nil {
		return nil, fmt.Errorf("New: failed to dial: %w", err)
	}
	h, err := http.NewHttpConnection(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("New: failed to connect to http2: %w", err)
	}
	xpcConn, err := ios.CreateXpcConnection(h)
	if err != nil {
		h.Close()
		return nil, fmt.Errorf("New: failed to create xpc connection: %w", err)
	}
	return &Connection{h: h, xpc: xpcConn, udid: device.Properties.SerialNumber}, nil
}

// Close closes the connection to the service.
func (c *Connection) Close() error {
	return c.xpc.Close()
}

// ListFiles asks the device which dyld shared cache files it offers for download. The
// returned order matters: the index of a file is needed to download it.
func (c *Connection) ListFiles() ([]FileInfo, error) {
	req := map[string]interface{}{
		"XPCDictionary_sideChannel": uuid.New(),
		"DSCFilePaths":              []interface{}{},
	}
	if err := c.xpc.Send(req, xpc.HeartbeatRequestFlag); err != nil {
		return nil, fmt.Errorf("ListFiles: failed to send request: %w", err)
	}
	resp, err := c.xpc.ReceiveOnServerClientStream()
	if err != nil {
		return nil, fmt.Errorf("ListFiles: failed to receive file count: %w", err)
	}
	count, err := parseFileCount(resp)
	if err != nil {
		return nil, fmt.Errorf("ListFiles: %w", err)
	}
	files := make([]FileInfo, 0, count)
	for i := 0; i < count; i++ {
		resp, err := c.xpc.ReceiveOnServerClientStream()
		if err != nil {
			return nil, fmt.Errorf("ListFiles: failed to receive metadata of file %d: %w", i, err)
		}
		info, err := parseFileInfo(resp)
		if err != nil {
			return nil, fmt.Errorf("ListFiles: file %d: %w", i, err)
		}
		files = append(files, info)
	}
	golog.Info("listed dyld shared cache files", "module", logModule, "udid", c.udid, "count", len(files))
	return files, nil
}

// DownloadFile streams the raw content of the file at fileIndex (its position in the
// ListFiles result) into w. It opens the HTTP2 side stream the device expects for the
// transfer and copies exactly info.Size bytes.
func (c *Connection) DownloadFile(w io.Writer, fileIndex int, info FileInfo) error {
	streamId := fileStreamId(fileIndex)
	golog.Info("downloading dyld shared cache file", "module", logModule, "udid", c.udid,
		"path", info.FilePath, "bytes", info.Size, "streamId", streamId)
	stream, err := http.NewFileStreamReadWriter(c.h, streamId)
	if err != nil {
		return fmt.Errorf("DownloadFile: %w", err)
	}
	err = xpc.EncodeMessage(stream, xpc.Message{
		Flags: xpc.FileTransferStreamResponseFlag | xpc.AlwaysSetFlag,
		Body:  nil,
		Id:    0,
	})
	if err != nil {
		return fmt.Errorf("DownloadFile: failed to open file stream %d: %w", streamId, err)
	}
	if err := copyFileChunks(w, stream, info.Size); err != nil {
		return fmt.Errorf("DownloadFile: %s: %w", info.FilePath, err)
	}
	return nil
}

// fileStreamId maps a file index to the HTTP2 stream id the device streams its content
// on: the n-th file uses the n-th even stream id.
func fileStreamId(fileIndex int) uint32 {
	return uint32(fileIndex+1) * 2
}

// parseFileCount extracts the number of offered files from the reply to the initial
// DSCFilePaths request.
func parseFileCount(resp map[string]interface{}) (int, error) {
	var count int64
	switch v := resp["DSCFilePaths"].(type) {
	case uint64:
		if v > math.MaxInt64 {
			return 0, fmt.Errorf("parseFileCount: implausible file count %d", v)
		}
		count = int64(v)
	case int64:
		count = v
	default:
		return 0, fmt.Errorf("parseFileCount: unexpected type %T for DSCFilePaths in %+v", resp["DSCFilePaths"], resp)
	}
	if count < 0 || count > maxFileCount {
		return 0, fmt.Errorf("parseFileCount: implausible file count %d", count)
	}
	return int(count), nil
}

// parseFileInfo extracts path and expected length from a per-file metadata message.
func parseFileInfo(resp map[string]interface{}) (FileInfo, error) {
	entry, ok := resp["DSCFilePaths"].(map[string]interface{})
	if !ok {
		return FileInfo{}, fmt.Errorf("parseFileInfo: missing DSCFilePaths dictionary in %+v", resp)
	}
	filePath, ok := entry["filePath"].(string)
	if !ok || filePath == "" {
		return FileInfo{}, fmt.Errorf("parseFileInfo: missing filePath in %+v", entry)
	}
	transfer, ok := entry["fileTransfer"].(xpc.FileTransfer)
	if !ok {
		return FileInfo{}, fmt.Errorf("parseFileInfo: missing fileTransfer in %+v", entry)
	}
	return FileInfo{FilePath: filePath, Size: transfer.TransferSize}, nil
}

// copyFileChunks copies exactly size bytes from the file stream to w, reassembling the
// chunked HTTP2 DATA frames the device sends.
func copyFileChunks(w io.Writer, r io.Reader, size uint64) error {
	if size > math.MaxInt64 {
		return fmt.Errorf("copyFileChunks: implausible file size %d", size)
	}
	n, err := io.Copy(w, io.LimitReader(r, int64(size)))
	if err != nil {
		return fmt.Errorf("copyFileChunks: transfer failed after %d of %d bytes: %w", n, size, err)
	}
	if uint64(n) != size {
		return fmt.Errorf("copyFileChunks: device ended stream after %d of %d bytes", n, size)
	}
	return nil
}
