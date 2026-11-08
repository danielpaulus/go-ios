package ncm

import (
	"fmt"
	"net/netip"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

var interfaceNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,15}$`)

func validateInterfaceName(interfaceName string) error {
	if !interfaceNamePattern.MatchString(interfaceName) {
		return fmt.Errorf("invalid network interface name %q", interfaceName)
	}
	return nil
}

// SetInterfaceUp uses the ubuntu command line to activate an ethernet device with interfaceName
func SetInterfaceUp(interfaceName string) (string, error) {
	if err := validateInterfaceName(interfaceName); err != nil {
		return "", err
	}
	b, err := exec.Command("ip", "link", "set", "dev", interfaceName, "up").CombinedOutput()
	return string(b), err
}

// AddInterface adds an ipv6 address to an interface using an ubuntu cmd line invocation
func AddInterface(interfaceName string, ipv6 string) (string, error) {
	if err := validateInterfaceName(interfaceName); err != nil {
		return "", err
	}
	prefix, err := netip.ParsePrefix(ipv6)
	if err != nil || !prefix.Addr().Is6() {
		return "", fmt.Errorf("invalid IPv6 prefix %q", ipv6)
	}
	b, err := exec.Command("ip", "-6", "addr", "add", prefix.String(), "dev", interfaceName).CombinedOutput()
	return string(b), err
}

// InterfaceHasIP uses 'ip -6 addr show dev' to check existin ips
func InterfaceHasIP(interfaceName string) (bool, string) {
	if validateInterfaceName(interfaceName) != nil {
		return false, ""
	}
	b, err := exec.Command("ip", "-6", "addr", "show", "dev", interfaceName).CombinedOutput()
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
	b, err := exec.Command("usbmuxd", "--version").CombinedOutput()
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
