package ostrace

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

const (
	usbmuxdServiceName = "com.apple.os_trace_relay"
	shimServiceName    = "com.apple.os_trace_relay.shim.remote"
)

// LogLevel is the severity byte stored inside each parsed log entry (the
// high byte of diagnosticd's traceid shifted into the binary frame).
type LogLevel uint8

const (
	LogLevelDefault    LogLevel = 0x00 // OS_LOG_TYPE_DEFAULT
	LogLevelInfo       LogLevel = 0x01 // OS_LOG_TYPE_INFO
	LogLevelDebug      LogLevel = 0x02 // OS_LOG_TYPE_DEBUG
	LogLevelUserAction LogLevel = 0x03
	LogLevelError      LogLevel = 0x10 // OS_LOG_TYPE_ERROR
	LogLevelFault      LogLevel = 0x11 // OS_LOG_TYPE_FAULT
)

// MessageFilter selects which RECORD TYPES diagnosticd emits. Bits map to the
// low byte of traceid: bit N enables (traceid & 0xFF) == value. Reverse-
// engineered from diagnosticd's type cascade (FUN_100007040 / 0x10000710c+).
// Bits 4–15 are no-ops in the type gate.
const (
	MessageFilterActivityCreate     uint16 = 1 << 0 // (traceid & 0xFF) == 2
	MessageFilterActivityTransition uint16 = 1 << 1 // (traceid & 0xFF) == 3
	MessageFilterLogMessage         uint16 = 1 << 2 // (traceid & 0xFF) == 4 — os_log entries
	MessageFilterSignpost           uint16 = 1 << 3 // (traceid & 0xFF) == 6
	MessageFilterAll                uint16 = 0xFFFF
)

// StreamFlags gates SEVERITY for record types that carry a level
// (ActivityTransition and LogMessage). Default, Error, and Fault always
// emit unconditionally; Info and Debug require StreamFlagsDebug (0x20).
//
// Info/Debug cannot be separated at the wire level on stock iOS 18.
// The kernel's /dev/oslog_stream decides at the source whether Info/Debug
// records exist — controlled by the arm64 commpage at 0xfffffc104:
//
//	bit 0 = preserve Default, bit 1 = preserve Info, bit 3 = streaming active
//
// When diagnosticd processes StreamFlags it calls
// host_set_atm_diagnostic_flag(mode & ~3), stripping bits 0 and 1 — so the
// StreamFlags value alone can never turn on Info preservation. Setting 0x20
// triggers mode=0xb, which also invokes _os_trace_set_mode(0xb) and flips
// the per-process preservation bits for the streaming session. That's what
// enables Info and Debug to flow together. 0x100 (mode=9) skips that call,
// which is why it's a no-op on stock iOS despite passing diagnosticd's
// userspace gate. Only `log config --mode level:info` (root-entitled) can
// set commpage bit 1 system-wide.
//
// Verified on iOS 18.7.1 by a full 16-bit single-bit sweep plus several
// multi-bit combos: no bit or combo separates Info from Debug. Clients
// wanting only one must filter client-side.
const (
	StreamFlagsDebug uint32 = 0x20 // triggers mode=0xb → enables Info+Debug together
	StreamFlagsAll   uint32 = StreamFlagsDebug
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDefault:
		return "Default"
	case LogLevelInfo:
		return "Info"
	case LogLevelDebug:
		return "Debug"
	case LogLevelUserAction:
		return "UserAction"
	case LogLevelError:
		return "Error"
	case LogLevelFault:
		return "Fault"
	default:
		return fmt.Sprintf("Unknown(%d)", l)
	}
}

// ClientFilter defines client-side filters applied after entries are received
// from the device. These do NOT reduce USB traffic — they only filter output.
type ClientFilter struct {
	Levels    []LogLevel // If non-empty, only entries whose Level is in this set pass
	Subsystem string     // Only show entries where label subsystem contains this string
	Match     string     // Only show entries where message contains this string
	Exclude   string     // Hide entries where message contains this string
}

// Matches returns true if the entry passes all client-side filters.
func (f ClientFilter) Matches(entry LogEntry) bool {
	if len(f.Levels) > 0 {
		matched := false
		for _, l := range f.Levels {
			if entry.Level == l {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if f.Subsystem != "" {
		if entry.Label == nil || !strings.Contains(entry.Label.Subsystem, f.Subsystem) {
			return false
		}
	}
	if f.Match != "" {
		if !strings.Contains(entry.Message, f.Match) {
			return false
		}
	}
	if f.Exclude != "" {
		if strings.Contains(entry.Message, f.Exclude) {
			return false
		}
	}
	return true
}

// LevelFilter bundles the device-side parameters and client-side filter
// needed to deliver only the user-requested log levels.
type LevelFilter struct {
	MessageFilter uint16     // what record types the device emits
	StreamFlags   uint32     // severity gate for record types that carry a level
	ClientLevels  []LogLevel // exact levels the user wants (for client-side filtering)
}

// DefaultLevelFilter returns a filter that streams log messages at every severity.
func DefaultLevelFilter() LevelFilter {
	return LevelFilter{
		MessageFilter: MessageFilterLogMessage,
		StreamFlags:   StreamFlagsAll,
	}
}

// ParseLevelFilter parses a comma-separated list of level names (case-insensitive:
// default, info, debug, error, fault) into a LevelFilter. It computes the minimum
// MessageFilter and StreamFlags needed to make the device emit the requested levels,
// and lists the exact levels for client-side post-filtering (since the device emits
// Default, Error, and Fault together and can't split them).
// Empty input returns DefaultLevelFilter.
func ParseLevelFilter(levels string) (LevelFilter, error) {
	if levels == "" {
		return DefaultLevelFilter(), nil
	}

	// Collect requested levels without duplicates, preserving order.
	seen := make(map[LogLevel]bool)
	var ordered []LogLevel
	for _, raw := range strings.Split(levels, ",") {
		name := strings.TrimSpace(strings.ToLower(raw))
		if name == "" {
			continue
		}
		level, ok := levelNameToValue[name]
		if !ok {
			return LevelFilter{}, fmt.Errorf("unknown log level %q, valid levels: default, info, debug, error, fault", name)
		}
		if !seen[level] {
			seen[level] = true
			ordered = append(ordered, level)
		}
	}

	spec := LevelFilter{
		MessageFilter: MessageFilterLogMessage,
		ClientLevels:  ordered,
	}
	// StreamFlagsDebug is needed for both Info and Debug (the 0x100 bit from
	// reverse-engineering notes doesn't work alone on iOS 18). This means
	// requesting Info also pulls Debug off the wire; the client-side level
	// filter then drops Debug entries if the user didn't ask for them.
	if seen[LogLevelInfo] || seen[LogLevelDebug] {
		spec.StreamFlags |= StreamFlagsDebug
	}
	return spec, nil
}

var levelNameToValue = map[string]LogLevel{
	"default": LogLevelDefault,
	"info":    LogLevelInfo,
	"debug":   LogLevelDebug,
	"error":   LogLevelError,
	"fault":   LogLevelFault,
}

// LogLabel contains the subsystem and category for a structured log entry.
type LogLabel struct {
	Subsystem string `json:"subsystem"`
	Category  string `json:"category"`
}

// LogEntry represents a single parsed os_trace syslog entry.
type LogEntry struct {
	PID         uint32    `json:"pid"`
	Timestamp   time.Time `json:"timestamp"`
	Level       LogLevel  `json:"level"`
	LevelName   string    `json:"levelName"`
	ThreadID    uint32    `json:"threadId"`
	ImageName   string    `json:"imageName"`
	ImageOffset uint32    `json:"imageOffset"`
	Filename    string    `json:"filename"`
	Message     string    `json:"message"`
	Label       *LogLabel `json:"label,omitempty"`
}

// Connection wraps a device connection to the os_trace_relay service.
type Connection struct {
	closer io.Closer
	reader io.Reader
	writer io.Writer
}

// New creates a new os_trace_relay connection, sends the StartActivity
// request with the given pid filter, MessageFilter bitmask, and StreamFlags
// bitmask, and performs the handshake. Use pid=-1 for all processes. See the
// MessageFilter* and StreamFlags* constants, or build a spec with
// ParseLevelFilter.
func New(device ios.DeviceEntry, pid int, messageFilter uint16, streamFlags uint32) (*Connection, error) {
	var deviceConn ios.DeviceConnectionInterface
	var err error

	if device.SupportsRsd() {
		deviceConn, err = ios.ConnectToShimService(device, shimServiceName)
	} else {
		deviceConn, err = ios.ConnectToService(device, usbmuxdServiceName)
	}
	if err != nil {
		return nil, fmt.Errorf("ostrace: failed to connect to service: %w", err)
	}

	conn := &Connection{
		closer: deviceConn,
		reader: deviceConn.Reader(),
		writer: deviceConn.Writer(),
	}

	if err := conn.startActivity(pid, messageFilter, streamFlags); err != nil {
		deviceConn.Close()
		return nil, err
	}

	return conn, nil
}

// startActivity sends the StartActivity plist request and reads the handshake response.
func (c *Connection) startActivity(pid int, messageFilter uint16, streamFlags uint32) error {
	codec := ios.NewPlistCodecReadWriter(c.reader, c.writer)

	request := map[string]interface{}{
		"Request":       "StartActivity",
		"MessageFilter": messageFilter,
		"Pid":           pid,
		"StreamFlags":   streamFlags,
	}

	if err := codec.Write(request); err != nil {
		return fmt.Errorf("ostrace: failed to send StartActivity request: %w", err)
	}

	// Response handshake:
	// 1. Read 4 bytes LE → lengthLength (number of bytes encoding the plist length)
	// 2. Read lengthLength bytes → reverse → parse as big-endian int → plistLength
	// 3. Read plistLength bytes → decode plist → expect {"Status": "RequestSuccessful"}
	var lengthLength uint32
	if err := binary.Read(c.reader, binary.LittleEndian, &lengthLength); err != nil {
		return fmt.Errorf("ostrace: failed to read length-length: %w", err)
	}

	lengthBytes := make([]byte, lengthLength)
	if _, err := io.ReadFull(c.reader, lengthBytes); err != nil {
		return fmt.Errorf("ostrace: failed to read plist length bytes: %w", err)
	}

	// Reverse the bytes and interpret as big-endian integer
	slices.Reverse(lengthBytes)
	var plistLength uint64
	for _, b := range lengthBytes {
		plistLength = (plistLength << 8) | uint64(b)
	}

	plistData := make([]byte, plistLength)
	if _, err := io.ReadFull(c.reader, plistData); err != nil {
		return fmt.Errorf("ostrace: failed to read response plist: %w", err)
	}

	response, err := ios.ParsePlist(plistData)
	if err != nil {
		return fmt.Errorf("ostrace: failed to parse response plist: %w", err)
	}
	status, ok := response["Status"].(string)
	if !ok || status != "RequestSuccessful" {
		return fmt.Errorf("ostrace: StartActivity failed, response: %v", response)
	}

	log.Debug("ostrace: StartActivity handshake successful")
	return nil
}

// ReadEntry reads a single log entry from the stream. This is a blocking call.
func (c *Connection) ReadEntry() (LogEntry, error) {
	// Frame: 1 byte magic (0x02) + 4 bytes LE length + N bytes entry data
	var magic [1]byte
	if _, err := io.ReadFull(c.reader, magic[:]); err != nil {
		return LogEntry{}, fmt.Errorf("ostrace: failed to read magic byte: %w", err)
	}
	if magic[0] != 0x02 {
		return LogEntry{}, fmt.Errorf("ostrace: unexpected magic byte: 0x%02x, expected 0x02", magic[0])
	}

	var length uint32
	if err := binary.Read(c.reader, binary.LittleEndian, &length); err != nil {
		return LogEntry{}, fmt.Errorf("ostrace: failed to read entry length: %w", err)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(c.reader, data); err != nil {
		return LogEntry{}, fmt.Errorf("ostrace: failed to read entry data: %w", err)
	}

	return parseEntry(data)
}

// ReadFilteredEntry reads entries from the stream and returns the next one
// that passes the given client-side filter. Loops internally until a match is
// found or an error occurs.
func (c *Connection) ReadFilteredEntry(filter ClientFilter) (LogEntry, error) {
	for {
		entry, err := c.ReadEntry()
		if err != nil {
			return LogEntry{}, err
		}
		if filter.Matches(entry) {
			return entry, nil
		}
	}
}

// Close closes the underlying device connection.
func (c *Connection) Close() error {
	return c.closer.Close()
}

// parseEntry parses a binary syslog entry into a LogEntry.
//
// Based on the ostrace_packet_header_t struct from libimobiledevice.
// All fields are little-endian, struct is packed (no padding).
//
//	Offset  Size  Type       Field
//	0       1     uint8      marker
//	1       4     uint32     type
//	5       4     uint32     headerSize
//	9       4     uint32     pid
//	13      8     uint64     procid
//	21      16    [16]byte   procuuid
//	37      2     uint16     procpathLen
//	39      8     uint64     activityID
//	47      8     uint64     parentActivityID
//	55      8     uint64     timeSec
//	63      4     uint32     timeUsec
//	67      1     uint8      (unknown)
//	68      1     uint8      level
//	69      6     [6]byte    (unknown)
//	75      8     uint64     timestamp (mach continuous time)
//	83      4     uint32     threadID
//	87      4     uint32     (unknown)
//	91      16    [16]byte   imageuuid
//	107     2     uint16     imagepathLen
//	109     4     uint32     messageLen
//	113     4     uint32     senderImageOffset
//	117     2     uint16     subsystemLen
//	119     2     uint16     (unknown)
//	121     2     uint16     categoryLen
//	123     2     uint16     (unknown)
//	125     4     uint32     (unknown)
//	129+    var   cstring    procpath (procpathLen bytes)
//	~       var   cstring    imagepath (imagepathLen bytes)
//	~       var   cstring    message (messageLen bytes)
//	~       var   cstring    subsystem (subsystemLen bytes, optional)
//	~       var   cstring    category (categoryLen bytes, optional)
func parseEntry(data []byte) (LogEntry, error) {
	if len(data) < 129 {
		return LogEntry{}, fmt.Errorf("ostrace: entry too short: %d bytes", len(data))
	}

	pid := binary.LittleEndian.Uint32(data[9:13])
	if pid > 999999 {
		return LogEntry{}, fmt.Errorf("ostrace: pid %d exceeds sanity limit, possibly corrupted stream", pid)
	}
	procpathLen := binary.LittleEndian.Uint16(data[37:39])
	timeSec := binary.LittleEndian.Uint64(data[55:63])
	timeUsec := binary.LittleEndian.Uint32(data[63:67])
	level := LogLevel(data[68])
	threadID := binary.LittleEndian.Uint32(data[83:87])
	imagepathLen := binary.LittleEndian.Uint16(data[107:109])
	messageLen := binary.LittleEndian.Uint32(data[109:113])
	senderImageOffset := binary.LittleEndian.Uint32(data[113:117])
	subsystemLen := binary.LittleEndian.Uint16(data[117:119])
	categoryLen := binary.LittleEndian.Uint16(data[121:123])

	timestamp := time.Unix(int64(timeSec), int64(timeUsec)*1000)

	// Variable-length fields start at offset 129
	offset := 129

	// procpath (filename): length from procpathLen, includes null terminator
	if procpathLen > 0 {
		if offset+int(procpathLen) > len(data) {
			return LogEntry{}, fmt.Errorf("ostrace: not enough data for procpath")
		}
	}
	filename := cstring(data[offset : offset+int(procpathLen)])
	offset += int(procpathLen)

	// imagepath: imagepathLen bytes (includes null terminator)
	if offset+int(imagepathLen) > len(data) {
		return LogEntry{}, fmt.Errorf("ostrace: not enough data for imagepath")
	}
	imageName := cstring(data[offset : offset+int(imagepathLen)])
	offset += int(imagepathLen)

	// message: messageLen bytes (includes null terminator)
	if offset+int(messageLen) > len(data) {
		return LogEntry{}, fmt.Errorf("ostrace: not enough data for message")
	}
	message := cstring(data[offset : offset+int(messageLen)])
	offset += int(messageLen)

	entry := LogEntry{
		PID:         pid,
		Timestamp:   timestamp,
		Level:       level,
		LevelName:   level.String(),
		ThreadID:    threadID,
		ImageName:   imageName,
		ImageOffset: senderImageOffset,
		Filename:    filename,
		Message:     message,
	}

	// Optional label: subsystem + category
	if subsystemLen > 0 && categoryLen > 0 {
		if offset+int(subsystemLen)+int(categoryLen) > len(data) {
			return entry, nil // return what we have without label
		}
		subsystem := cstring(data[offset : offset+int(subsystemLen)])
		offset += int(subsystemLen)
		category := cstring(data[offset : offset+int(categoryLen)])
		entry.Label = &LogLabel{
			Subsystem: subsystem,
			Category:  category,
		}
	}

	return entry, nil
}

// cstring extracts a string from a byte slice, stripping the null terminator if present.
func cstring(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if data[len(data)-1] == 0 {
		return string(data[:len(data)-1])
	}
	return string(data)
}
