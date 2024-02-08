package ios

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
)

// ToPlist converts a given struct to a Plist using the
// github.com/DHowett/go-plist library. Make sure your struct is exported.
// It returns a string containing the plist.
func ToPlist(data interface{}) string {
	return string(ToPlistBytes(data))
}

// ParsePlist tries to parse the given bytes, which should be a Plist, into a map[string]interface.
// It returns the map or an error if the decoding step fails.
func ParsePlist(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := plist.Unmarshal(data, &result)
	return result, err
}

// ToPlistBytes converts a given struct to a Plist using the
// github.com/DHowett/go-plist library. Make sure your struct is exported.
// It returns a byte slice containing the plist.
func ToPlistBytes(data interface{}) []byte {
	bytes, err := plist.Marshal(data, plist.XMLFormat)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("Failed converting to plist %v error:%v", data, err))
	}
	return bytes
}

func ToBinPlistBytes(data interface{}) []byte {
	bytes, err := plist.Marshal(data, plist.BinaryFormat)
	if err != nil {
		// this should not happen
		panic(fmt.Sprintf("Failed converting to plist %v error:%v", data, err))
	}
	return bytes
}

// Ntohs is a re-implementation of the C function Ntohs.
// it means networkorder to host oder and basically swaps
// the endianness of the given int.
// It returns port converted to little endian.
func Ntohs(port uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	return binary.LittleEndian.Uint16(buf)
}

// GetDevice returns:
// the device for the udid if a valid udid is provided.
// if the env variable 'udid' is specified, the device with that udid
// otherwise it returns the first device in the list.
func GetDevice(udid string) (DeviceEntry, error) {
	return GetDeviceWithAddress(udid, "", nil)
}

func GetDeviceWithAddress(udid string, address string, provider RsdPortProvider) (DeviceEntry, error) {
	if udid == "" {
		udid = os.Getenv("udid")
		if udid != "" {
			log.Info("using udid from env.udid variable")
		}
	}
	log.Debugf("Looking for device '%s'", udid)
	deviceList, err := ListDevices()
	if err != nil {
		return DeviceEntry{}, err
	}
	if udid == "" {
		if len(deviceList.DeviceList) == 0 {
			return DeviceEntry{}, errors.New("no iOS devices are attached to this host")
		}
		device := deviceList.DeviceList[0]
		log.WithFields(log.Fields{"udid": device.Properties.SerialNumber}).
			Info("no udid specified using first device in list")
		device.Address = address
		device.Rsd = provider
		return device, nil
	}
	for _, device := range deviceList.DeviceList {
		if device.Properties.SerialNumber == udid {
			device.Address = address
			device.Rsd = provider
			return device, nil
		}
	}
	return DeviceEntry{}, fmt.Errorf("Device '%s' not found. Is it attached to the machine?", udid)
}

// PathExists is used to determine whether the path folder exists
// True if it exists, false otherwise
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IOS17() *semver.Version {
	return semver.MustParse("17.0")
}

func IOS14() *semver.Version {
	return semver.MustParse("14.0")
}

func IOS12() *semver.Version {
	return semver.MustParse("12.0")
}

func IOS11() *semver.Version {
	return semver.MustParse("11.0")
}

// FixWindowsPaths replaces backslashes with forward slashes and removes the X: style
// windows drive letters
func FixWindowsPaths(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if strings.Contains(path, ":/") {
		path = strings.Split(path, ":/")[1]
	}
	return path
}

func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// InterfaceToStringSlice casts an interface{} to []interface{} and then converts each entry to a string.
// It returns an empty slice in case of an error.
func InterfaceToStringSlice(intfSlice interface{}) []string {
	slice, ok := intfSlice.([]interface{})
	if !ok {
		return []string{}
	}
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

// GenericSliceToType tries to convert a slice of interfaces to a slice of the given type.
// It returns an error if the conversion fails but will not panic.
// Example: var b []bool; b, err = GenericSliceToType[bool]([]interface{}{true, false})
func GenericSliceToType[T any](input []interface{}) ([]T, error) {
	result := make([]T, len(input))
	for i, intf := range input {
		if t, ok := intf.(T); ok {
			result[i] = t
		} else {
			return []T{}, fmt.Errorf("GenericSliceToType: could not convert %v to %T", intf, result[i])
		}
	}
	return result, nil
}
