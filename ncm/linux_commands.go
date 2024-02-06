package ncm

import (
	"fmt"
	"os/exec"
)

// works on ubuntu
func SetInterfaceUp(interfaceName string) (string, error) {
	b, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo ip link set dev %s up", interfaceName)).CombinedOutput()
	return string(b), err
}
