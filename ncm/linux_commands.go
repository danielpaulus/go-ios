package ncm

import (
	"fmt"
	"github.com/Masterminds/semver"
	"os/exec"
	"strings"
)

// works on ubuntu
func SetInterfaceUp(interfaceName string) (string, error) {
	b, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo ip link set dev %s up", interfaceName)).CombinedOutput()
	return string(b), err
}

func AddInterface(interfaceName string) (string, error) {
	//TODO: figure out if this is actually needed, if so, generate a random IP address
	// and add this command somewhere
	// sudo ip addr add FF:02:00:00:00:00:00:00:00:00:00:00:00:00:00:FB dev iphone
	return "", nil
}

const usbmuxVersion = "usbmuxd 1.1.1-56-g360619c"

var supportedVersion = semver.MustParse(strings.Replace(usbmuxVersion, "usbmuxd ", "", -1))

func GetUSBMUXVersion() (*semver.Version, error) {
	b, err := exec.Command("/bin/sh", "-c", "usbmuxd --version").CombinedOutput()
	if err != nil {
		return &semver.Version{}, err
	}
	version := strings.Replace(string(b), "usbmuxd ", "", -1)
	v, err := semver.NewVersion(version)
	if err != nil {
		return &semver.Version{}, fmt.Errorf("GetUSBMUXVersion: could not parse usbmuxd version: %s from '%s'", err.Error(), string(b))
	}
	ok := v.Equal(supportedVersion) || v.GreaterThan(supportedVersion)

	if !ok {
		return v, fmt.Errorf("usbmuxd version %s is not supported. Please use at least %s", version, supportedVersion)
	}

	return v, nil
}
