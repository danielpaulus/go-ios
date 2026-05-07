package syslog

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
)

const (
	usbmuxdServiceName string = "com.apple.syslog_relay"
	shimServiceName           = "com.apple.syslog_relay.shim.remote"
)

// Connection exposes the LogReader channel which send the LogMessages as strings.
type Connection struct {
	closer         io.Closer
	bufferedReader *bufio.Reader
}

// New returns a new SysLog Connection for the given DeviceID and Udid
// It will create LogReader as a buffered Channel because Syslog is very verbose.
func New(device ios.DeviceEntry) (*Connection, error) {
	if !device.SupportsRsd() {
		return NewWithUsbmuxdConnection(device)
	}
	return NewWithShimConnection(device)
}

// NewWithUsbmuxdConnection connects to the syslog_relay service on the device over the usbmuxd socket
func NewWithUsbmuxdConnection(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, usbmuxdServiceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{
		closer:         deviceConn,
		bufferedReader: bufio.NewReader(deviceConn),
	}, nil
}

// NewWithShimConnection connects to the syslog_relay service over a tunnel interface and the service port
// is obtained from remote service discovery
func NewWithShimConnection(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToShimService(device, shimServiceName)
	if err != nil {
		return nil, err
	}
	return &Connection{
		closer:         deviceConn,
		bufferedReader: bufio.NewReader(deviceConn),
	}, nil
}

// ReadLogMessage this is a blocking function that will return individual log messages received from syslog.
// Call it in an endless for loop in a separate go routine.
func (sysLogConn *Connection) ReadLogMessage() (string, error) {
	logmsg, err := sysLogConn.bufferedReader.ReadString(0)
	if err != nil {
		return "", err
	}
	return logmsg, nil
}

// ReadLogMessageBytes returns the next null-terminated log message as a slice
// into the underlying bufio buffer. The returned slice is only valid until
// the next call on the Connection — callers must finish processing (writing,
// json-marshaling, copying) before reading the next message. Use this for
// allocation-sensitive hot paths; otherwise use ReadLogMessage.
func (sysLogConn *Connection) ReadLogMessageBytes() ([]byte, error) {
	return sysLogConn.bufferedReader.ReadSlice(0)
}

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Device    string `json:"device"`
	Process   string `json:"process"`
	PID       string `json:"pid"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

func Parser() func(log string) (*LogEntry, error) {
	pattern := `(?P<Timestamp>[A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})\s+(?P<Device>\S+)\s+(?P<Process>[^\[]+)\[(?P<PID>\d+)\]\s+<(?P<Level>\w+)>: (?P<Message>.+)`
	re := regexp.MustCompile(pattern)

	// Resolve named-group indexes once. Avoids per-line map[string]string +
	// SubexpNames() walk that the previous implementation did for every line.
	tsIdx := re.SubexpIndex("Timestamp")
	devIdx := re.SubexpIndex("Device")
	procIdx := re.SubexpIndex("Process")
	pidIdx := re.SubexpIndex("PID")
	lvlIdx := re.SubexpIndex("Level")
	msgIdx := re.SubexpIndex("Message")
	// Cache the current year. Syslog timestamps lack a year; we stamp with
	// the year the parser was constructed. This drifts at year boundaries
	// for long-running streams — a known limitation, same as upstream.
	parserYear := time.Now().Year()

	return func(log string) (*LogEntry, error) {
		// FindStringSubmatchIndex returns []int with [start,end) byte ranges.
		// The previous FindStringSubmatch allocated a fresh string per group;
		// here we slice into the input string directly, which is zero-alloc.
		loc := re.FindStringSubmatchIndex(log)
		if loc == nil {
			return nil, fmt.Errorf("failed to parse syslog message: %s", log)
		}

		field := func(idx int) string { return log[loc[2*idx]:loc[2*idx+1]] }

		parsedTime, err := time.Parse("Jan 2 15:04:05", field(tsIdx))
		if err != nil {
			return nil, fmt.Errorf("failed to parse syslog timestamp: %s", log)
		}
		parsedTime = parsedTime.AddDate(parserYear-parsedTime.Year(), 0, 0)

		return &LogEntry{
			Timestamp: parsedTime.Format("2006-01-02T15:04:05"),
			Device:    field(devIdx),
			Process:   strings.TrimSpace(field(procIdx)),
			PID:       field(pidIdx),
			Level:     field(lvlIdx),
			Message:   field(msgIdx),
		}, nil
	}
}

// Close closes the underlying UsbMuxConnection
func (sysLogConn *Connection) Close() error {
	return sysLogConn.closer.Close()
}
