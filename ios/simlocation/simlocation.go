package simlocation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	ios "github.com/danielpaulus/go-ios/ios"
)

const serviceName string = "com.apple.dt.simulatelocation"

type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

type locationData struct {
	lon float64
	lat float64
}

func New(device ios.DeviceEntry) (*Connection, error) {
	locationConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{deviceConn: locationConn, plistCodec: ios.NewPlistCodec()}, nil
}

func (locationConn *Connection) Close() {
	locationConn.deviceConn.Close()
}

func (locationConn *Connection) SetLocation(lat string, lon string) error {
	if lat == "" || lon == "" {
		return errors.New("Please provide non-empty values for latitude and longtitude")
	}

	latitude, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return err
	}

	longtitude, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return err
	}

	data := new(locationData)
	data.lat = latitude
	data.lon = longtitude

	// Generate the byte data needed by the service to set the location
	locationBytes, err := data.LocationBytes()
	if err != nil {
		return err
	}

	err = locationConn.deviceConn.Send(locationBytes)
	if err != nil {
		return err
	}

	return nil
}

func (locationConn *Connection) ResetLocation() error {
	buf := new(bytes.Buffer)

	// The location service accepts the binary representation of 1 to reset to the original location
	err := binary.Write(buf, binary.BigEndian, uint32(1))
	if err != nil {
		return err
	}

	err = locationConn.deviceConn.Send(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// Create the byte data needed to set a specific location
func (l *locationData) LocationBytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
		return nil, fmt.Errorf("creating location bytes: %w", err)
	}

	latString := fmt.Sprintf("%f", l.lat)
	latBytes := []byte(latString)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(latBytes))); err != nil {
		return nil, fmt.Errorf("creating location bytes: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, latBytes); err != nil {
		return nil, fmt.Errorf("creating location bytes: %w", err)
	}

	lonString := fmt.Sprintf("%f", l.lon)
	lonBytes := []byte(lonString)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(lonBytes))); err != nil {
		return nil, fmt.Errorf("creating location bytes: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, lonBytes); err != nil {
		return nil, fmt.Errorf("creating location bytes: %w", err)
	}

	return buf.Bytes(), nil
}
