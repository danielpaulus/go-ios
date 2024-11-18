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
	pattern := `(?P<Timestamp>[A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2}) (?P<Device>\S+) (?P<Process>[^\[]+)\[(?P<PID>\d+)\] <(?P<Level>\w+)>: (?P<Message>.+)`
	regexp := regexp.MustCompile(pattern)

	return func(log string) (*LogEntry, error) {
		// Match the log message against the regex pattern
		match := regexp.FindStringSubmatch(log)
		if match == nil {
			return nil, fmt.Errorf("failed to parse syslog message: %s", log)
		}

		// Create a map of named capture groups
		result := make(map[string]string)
		for i, name := range regexp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		// Parse the original timestamp
		originalTimestamp := result["Timestamp"]
		parsedTime, err := time.Parse("Jan 2 15:04:05", originalTimestamp)
		// Set the year to the current year from the system (this might cause friction at year end)
		parsedTime = parsedTime.AddDate(time.Now().Year()-parsedTime.Year(), 0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to parse syslog timestamp: %s", log)
		}

		// Convert to ISO 8601 format
		isoTimestamp := parsedTime.Format("2006-01-02T15:04:05")

		// Populate the LogEntry struct
		entry := &LogEntry{
			Timestamp: isoTimestamp,
			Device:    result["Device"],
			Process:   strings.TrimSpace(result["Process"]),
			PID:       result["PID"],
			Level:     result["Level"],
			Message:   result["Message"],
		}

		return entry, nil
	}
}

// Close closes the underlying UsbMuxConnection
func (sysLogConn *Connection) Close() error {
	return sysLogConn.closer.Close()
}
