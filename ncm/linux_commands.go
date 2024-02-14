package ncm

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

// works on ubuntu
func SetInterfaceUp(interfaceName string) (string, error) {
	b, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("ip link set dev %s up", interfaceName)).CombinedOutput()
	return string(b), err
}

func AddInterface(interfaceName string, ipv6 string) (string, error) {
	cmd := fmt.Sprintf("ip -6 addr add %s dev %s", ipv6, interfaceName)
	b, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	return string(b), err
}

func InterfaceHasIP(interfaceName string) (bool, string) {
	cmd := fmt.Sprintf("sudo ip -6 addr show dev %s", interfaceName)
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

const usbmuxVersion = "usbmuxd 1.1.1-56-g360619c"

var supportedVersion = semver.MustParse(strings.Replace(usbmuxVersion, "usbmuxd ", "", -1))

func GetUSBMUXVersion() (*semver.Version, error) {
	b, err := exec.Command("/bin/sh", "-c", "usbmuxd --version").CombinedOutput()
	if err != nil {
		return &semver.Version{}, err
	}
	version := strings.Replace(string(b), "usbmuxd ", "", -1)
	version = strings.TrimSpace(version)
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
