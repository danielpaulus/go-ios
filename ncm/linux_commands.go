package ncm

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

// SetInterfaceUp uses the ubuntu command line to activate an ethernet device with interfaceName
func SetInterfaceUp(interfaceName string) (string, error) {
	b, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("ip link set dev %s up", interfaceName)).CombinedOutput()
	return string(b), err
}

// AddInterface adds an ipv6 address to an interface using an ubuntu cmd line invocation
func AddInterface(interfaceName string, ipv6 string) (string, error) {
	cmd := fmt.Sprintf("ip -6 addr add %s dev %s", ipv6, interfaceName)
	b, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	return string(b), err
}

// InterfaceHasIP uses 'ip -6 addr show dev' to check existin ips
func InterfaceHasIP(interfaceName string) (bool, string) {
	cmd := fmt.Sprintf("ip -6 addr show dev %s", interfaceName)
	b, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return false, ""
	}
	output := string(b)
	output = strings.TrimSpace(output)
	if output == "" {
		return false, output
	}
	return true, output
}

const lowestSupportedVersionString = "usbmuxd 1.1.1-56-g360619c"

var lowestSupportedVersion = semver.MustParse(strings.Replace(lowestSupportedVersionString, "usbmuxd ", "", -1))

// CheckUSBMUXVersion runs usbmuxd --version to make sure it is newer or equal to usbmuxd 1.1.1-56-g360619c
func CheckUSBMUXVersion() (*semver.Version, error) {
	b, err := exec.Command("/bin/sh", "-c", "usbmuxd --version").CombinedOutput()
	if err != nil {
		return &semver.Version{}, err
	}
	version := strings.Replace(string(b), "usbmuxd ", "", -1)
	version = strings.TrimSpace(version)
	v, err := semver.NewVersion(version)
	if err != nil {
		return &semver.Version{}, fmt.Errorf("CheckUSBMUXVersion: could not parse usbmuxd version: %s from '%s'", err.Error(), string(b))
	}
	ok := v.Equal(lowestSupportedVersion) || v.GreaterThan(lowestSupportedVersion)

	if !ok {
		return v, fmt.Errorf("usbmuxd version %s is not supported. Please use at least %s", version, lowestSupportedVersion)
	}

	return v, nil
}
