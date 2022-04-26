package simlocation

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

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

func (locationConn *Connection) SetLocationGPX(filePath string) error {
	gpxFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer gpxFile.Close()

	byteData, err := ioutil.ReadAll(gpxFile)
	if err != nil {
		return err
	}

	var gpx Gpx

	err = xml.Unmarshal(byteData, &gpx)
	if err != nil {
		return err
	}

	var lastPointTime time.Time

	for _, track := range gpx.Tracks {
		for _, segment := range track.TrackSegments {
			for _, point := range segment.TrackPoints {
				currentPointTime, err := time.Parse(time.RFC3339, point.PointTime)
				if !lastPointTime.IsZero() {
					fmt.Println("we are inside lastpoint non zero\n")
					fmt.Printf("%v\n", lastPointTime)
					//layout := "2022-01-01T01:01:01.000Z"
					fmt.Printf("this is before attempting to parse: %v\n", point.PointTime)
					//pointTime, err := time.Parse(time.RFC3339, point.PointTime)
					if err != nil {
						return err
					}
					fmt.Println("are we waiting\n")
					fmt.Printf("pointTime is %v\n", currentPointTime)
					duration := currentPointTime.Unix() - lastPointTime.Unix()
					fmt.Printf("the duration calculated is: %v\n", duration)
					if duration >= 0 {
						time.Sleep(time.Duration(duration) * time.Second)
					}
				}
				fmt.Println("we are outside lastpoint non zero\n")
				fmt.Printf("%v\n", lastPointTime)
				lastPointTime = currentPointTime
				pointLon := point.PointLongtitude
				pointLat := point.PointLatitude

				latitude, err := strconv.ParseFloat(pointLat, 64)
				if err != nil {
					return err
				}

				longtitude, err := strconv.ParseFloat(pointLon, 64)
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
			}
		}
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

// GPX parsing

type Gpx struct {
	XMLName xml.Name `xml:"gpx"`
	Tracks  []Track  `xml:"trk"`
}

type Track struct {
	XMLName       xml.Name       `xml:"trk"`
	TrackSegments []TrackSegment `xml:"trkseg"`
	Name          string         `xml:"name"`
}

type TrackSegment struct {
	XMLName     xml.Name     `xml:"trkseg"`
	TrackPoints []TrackPoint `xml:"trkpt"`
}

type TrackPoint struct {
	XMLName         xml.Name `xml:"trkpt"`
	PointLongtitude string   `xml:"lon,attr"`
	PointLatitude   string   `xml:"lat,attr"`
	PointTime       string   `xml:"time"`
}
