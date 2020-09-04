package ios

import (
	"encoding/binary"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
)

//ToPlist converts a given struct to a Plist using the
//github.com/DHowett/go-plist library. Make sure your struct is exported.
//It returns a string containing the plist.
func ToPlist(data interface{}) string {
	return string(ToPlistBytes(data))
}

//ToPlistBytes converts a given struct to a Plist using the
//github.com/DHowett/go-plist library. Make sure your struct is exported.
//It returns a byte slice containing the plist.
func ToPlistBytes(data interface{}) []byte {
	bytes, err := plist.Marshal(data, plist.XMLFormat)
	if err != nil {
		log.Fatal("Failed converting to plist", data, err)
	}
	return bytes
}

//Ntohs is a re-implementation of the C function Ntohs.
//it means networkorder to host oder and basically swaps
//the endianness of the given int.
//It returns port converted to little endian.
func Ntohs(port uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	return binary.LittleEndian.Uint16(buf)
}

//GetDevice returns the device for udid. If udid equals emptystring, it returns the first device in the list.
func GetDevice(udid string) (DeviceEntry, error) {
	log.Debugf("Looking for device '%s'", udid)
	deviceList := ListDevices()
	if udid == "" {
		if len(deviceList.DeviceList) == 0 {
			return DeviceEntry{}, errors.New("no iOS devices are attached to this host")
		}
		return deviceList.DeviceList[0], nil
	}
	for _, device := range deviceList.DeviceList {
		if device.Properties.SerialNumber == udid {
			return device, nil
		}
	}
	return DeviceEntry{}, fmt.Errorf("Device '%s' not found. Is it attached to the machine?", udid)
}
