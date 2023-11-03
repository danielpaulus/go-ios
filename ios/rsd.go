package ios

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"
)

type RsdPortProvider map[string]service

type service struct {
	Port string
}

func NewRsdPortProvider(input io.Reader) (RsdPortProvider, error) {
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

func (r RsdPortProvider) GetPort(service string) int {
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
