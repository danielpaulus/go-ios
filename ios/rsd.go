package ios

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/danielpaulus/go-ios/ios/http"
	"github.com/danielpaulus/go-ios/ios/xpc"
	log "github.com/sirupsen/logrus"
)

// RsdPortProvider is an interface to get a port for a service, or a service for a port from the Remote Service Discovery on the device.
// Used in iOS17+
type RsdPortProvider interface {
	GetPort(service string) int
	GetService(p int) string
	GetServices() map[string]RsdServiceEntry
}

type RsdPortProviderJson map[string]service

type service struct {
	Port string
}

func NewRsdPortProvider(input io.Reader) (RsdPortProviderJson, error) {
	decoder := json.NewDecoder(input)
	parse := struct {
		Services map[string]service
	}{}

	err := decoder.Decode(&parse)
	if err != nil {
		return nil, fmt.Errorf("NewRsdPortProvider: failed to parse rsd response: %w", err)
	}

	return parse.Services, nil
}

func (r RsdPortProviderJson) GetPort(service string) int {
	p := r[service].Port
	if p == "" {
		shim := fmt.Sprintf("%s.shim.remote", service)
		if r[shim].Port != "" {
			log.Debugf("returning port of '%s'-shim", service)
			return r.GetPort(shim)
		}
	}
	port, err := strconv.ParseInt(p, 10, 64)
	if err != nil {
		return 0
	}
	return int(port)
}

func (r RsdPortProviderJson) GetService(p int) string {
	for name, s := range r {
		port, err := strconv.ParseInt(s.Port, 10, 64)
		if err != nil {
			log.Errorf("GetService: failed to parse port: %v", err)
			return ""
		}
		if port == int64(p) {
			return name
		}
	}
	return ""
}

func (r RsdPortProviderJson) GetServices() (services map[string]RsdServiceEntry) {
	services = make(map[string]RsdServiceEntry, len(r))
	for name, s := range r {
		port, err := strconv.ParseInt(s.Port, 10, 64)
		if err != nil {
			log.Errorf("GetService: failed to parse port: %v", err)
			continue
		}

		services[name] = RsdServiceEntry{Port: uint32(port)}
	}

	return
}

// RsdCheckin sends a plist encoded message with the request 'RSDCheckin' to the device.
// The device is expected to reply with two plist encoded messages. The first message is the response for the
// checkin itself, and the second message contains a 'StartService' request, which does not need any action
// from the host side
func RsdCheckin(rw io.ReadWriter) error {
	req := map[string]interface{}{
		"Label":           "go-ios",
		"ProtocolVersion": "2",
		"Request":         "RSDCheckin",
	}

	prw := NewPlistCodecReadWriter(rw, rw)

	err := prw.Write(req)
	if err != nil {
		return fmt.Errorf("RsdCheckin: failed to send checkin request: %w", err)
	}

	var checkinResponse map[string]any
	err = prw.Read(&checkinResponse)
	if err != nil {
		return fmt.Errorf("RsdCheckin: failed to read checkin response: %w", err)
	}
	var startService map[string]any
	err = prw.Read(&startService)
	if err != nil {
		return fmt.Errorf("RsdCheckin: failed to read start service message: %w", err)
	}
	return nil
}

const port = 58783

type RsdService struct {
	xpc *xpc.Connection
	c   io.Closer
}

func (s RsdService) Close() error {
	return s.c.Close()
}

type RsdServiceEntry struct {
	Port uint32
}

// RsdHandshakeResponse is the response to the RSDCheckin request and contains the UDID
// and the services available on the device.
type RsdHandshakeResponse struct {
	Udid     string
	Services map[string]RsdServiceEntry
}

// GetService returns the service name for the given port.
func (r RsdHandshakeResponse) GetService(p int) string {
	for name, s := range r.Services {
		if s.Port == uint32(p) {
			return name
		}
	}
	return ""
}

// GetPort returns the port for the given service.
func (r RsdHandshakeResponse) GetPort(service string) int {
	if s, ok := r.Services[service]; ok {
		return int(s.Port)
	}
	return 0
}

func (r RsdHandshakeResponse) GetServices() map[string]RsdServiceEntry {
	return r.Services
}

// NewWithAddrPort creates a new RsdService with the given address and port 58783 using a HTTP2 based XPC connection,
// connecting to an operating system level TUN device.
func NewWithAddrPort(addr string, port int) (RsdService, error) {
	conn, err := connectTUN(addr, port)
	if err != nil {
		return RsdService{}, fmt.Errorf("NewWithAddrPort: failed to connect to device: %w", err)
	}
	return newRsdServiceFromTcpConn(conn)
}

// NewWithAddrDevice creates a new RsdService with the given address and port 58783 using a HTTP2 based XPC connection.
func NewWithAddrDevice(addr string, d DeviceEntry) (RsdService, error) {
	return NewWithAddrPortDevice(addr, port, d)
}

// NewWithAddrPortDevice creates a new RsdService with the given address and port using a HTTP2 based XPC connection.
func NewWithAddrPortDevice(addr string, port int, d DeviceEntry) (RsdService, error) {
	conn, err := ConnectTUNDevice(addr, port, d)
	if err != nil {
		return RsdService{}, fmt.Errorf("NewWithAddrPortTUNDevice: failed to connect to device: %w", err)
	}
	return newRsdServiceFromTcpConn(conn)
}

func newRsdServiceFromTcpConn(conn *net.TCPConn) (RsdService, error) {
	h, err := http.NewHttpConnection(conn)
	if err != nil {
		return RsdService{}, fmt.Errorf("newRsdServiceFromTcpConn: failed to connect to http2: %w", err)
	}

	x, err := CreateXpcConnection(h)
	if err != nil {
		return RsdService{}, fmt.Errorf("newRsdServiceFromTcpConn: failed to create xpc connection: %w", err)
	}

	return RsdService{
		xpc: x,
		c:   h,
	}, nil
}

// Handshake sends a handshake request to the device and returns the RsdHandshakeResponse
// which contains the UDID and the services available on the device.
func (s RsdService) Handshake() (RsdHandshakeResponse, error) {
	log.Debug("execute handshake")
	m, err := s.xpc.ReceiveOnClientServerStream()
	if err != nil {
		return RsdHandshakeResponse{}, fmt.Errorf("Handshake: failed to receive handshake response. %w", err)
	}
	udid := ""
	if properties, ok := m["Properties"].(map[string]interface{}); ok {
		if u, ok := properties["UniqueDeviceID"].(string); ok {
			udid = u
		}
	}
	if udid == "" {
		return RsdHandshakeResponse{}, fmt.Errorf("Handshake: could not read UDID")
	}
	if m["MessageType"] == "Handshake" {
		servicesMap := m["Services"].(map[string]interface{})
		res := make(map[string]RsdServiceEntry)
		for s, m := range servicesMap {
			s2 := m.(map[string]interface{})["Port"].(string)
			p, err := strconv.ParseInt(s2, 10, 32)
			if err != nil {
				return RsdHandshakeResponse{}, fmt.Errorf("Handshake: failed to parse port: %w", err)
			}
			res[s] = RsdServiceEntry{
				Port: uint32(p),
			}
		}
		return RsdHandshakeResponse{
			Services: res,
			Udid:     udid,
		}, nil
	} else {
		return RsdHandshakeResponse{}, fmt.Errorf("Handshake: unknown response")
	}
}
