package ncm

import (
	"fmt"
	"github.com/Masterminds/semver"
	"os/exec"
)

// works on ubuntu
func SetInterfaceUp(interfaceName string) (string, error) {
	b, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo ip link set dev %s up", interfaceName)).CombinedOutput()
	return string(b), err
}

func GetUSBMUXVersion() (*semver.Version, error) {
	b, err := exec.Command("/bin/sh", "-c", "usbmuxd --version").CombinedOutput()
	if err != nil {
		return &semver.Version{}, err
	}
	v, err := semver.NewVersion(string(b))

	return v, err
}
