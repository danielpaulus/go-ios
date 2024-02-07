package ios

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/danielpaulus/go-ios/ios/xpc"
	log "github.com/sirupsen/logrus"
)

type RsdPortProvider interface {
	GetPort(service string) int
	GetService(p int) string
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
		return nil, err
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
			panic(err)
		}
		if port == int64(p) {
			return name
		}
	}
	return ""
}

func RsdCheckin(rw io.ReadWriter) error {
	req := map[string]interface{}{
		"Label":           "go-ios",
		"ProtocolVersion": "2",
		"Request":         "RSDCheckin",
	}
	codec := NewPlistCodec()
	b, err := codec.Encode(req)
	if err != nil {
		return err
	}
	_, err = rw.Write(b)
	if err != nil {
		return err
	}
	res, err := codec.Decode(rw)
	if err != nil {
		return err
	}
	log.Debugf("got rsd checkin response: %v", res)
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

type RsdHandshakeResponse struct {
	Udid     string
	Services map[string]RsdServiceEntry
}

func (r RsdHandshakeResponse) GetService(p int) string {
	for name, s := range r.Services {
		if s.Port == uint32(p) {
			return name
		}
	}
	return ""
}

func (r RsdHandshakeResponse) GetPort(service string) int {
	if s, ok := r.Services[service]; ok {
		return int(s.Port)
	}
	return 0
}

func NewWithAddr(addr string) (RsdService, error) {
	return NewWithAddrPort(addr, port)
}

func NewWithAddrPort(addr string, port int) (RsdService, error) {
	h, err := ConnectToHttp2WithAddr(addr, port)
	if err != nil {
		return RsdService{}, err
	}

	x, err := CreateXpcConnection(h)

	if err != nil {
		return RsdService{}, err
	}

	return RsdService{
		xpc: x,
		c:   h,
	}, nil
}

func (s RsdService) Handshake() (RsdHandshakeResponse, error) {
	log.Debug("execute handshake")
	m, err := s.xpc.ReceiveOnClientServerStream()
	if err != nil {
		return RsdHandshakeResponse{}, fmt.Errorf("failed to receive handshake response. %w", err)
	}
	udid := ""
	if properties, ok := m["Properties"].(map[string]interface{}); ok {
		if u, ok := properties["UniqueDeviceID"].(string); ok {
			udid = u
		}
	}
	if udid == "" {
		return RsdHandshakeResponse{}, fmt.Errorf("could not read UDID")
	}
	if m["MessageType"] == "Handshake" {
		servicesMap := m["Services"].(map[string]interface{})
		res := make(map[string]RsdServiceEntry)
		for s, m := range servicesMap {
			s2 := m.(map[string]interface{})["Port"].(string)
			p, err := strconv.ParseInt(s2, 10, 32)
			if err != nil {
				panic(err)
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
		return RsdHandshakeResponse{}, fmt.Errorf("unknown response")
	}
}
