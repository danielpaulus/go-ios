//go:build darwin

package tunnel

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/songgao/water"
)

// createSysProcAttr returns attributes for spawning the go-ios agent in a
// new session so it survives as a standalone daemon.
func createSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// setupTunnelInterface creates a utun device and assigns the tunnel's IPv6
// address to it via ifconfig. macOS has no netlink equivalent, but ifconfig
// is a system binary always present on Darwin, so shelling out is fine.
//
// The MTU is lowered to 1202 so the OS splits payloads into smaller packets;
// with a larger MTU the QUIC tunnel underneath doesn't transmit correctly.
func setupTunnelInterface(tunnelInfo tunnelParameters) (io.ReadWriteCloser, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed creating TUN device: %w", err)
	}

	const prefixLength = 64 // TODO: could be derived from the netmask provided by the device
	const mtu = 1202

	addr := fmt.Sprintf("%s/%d", tunnelInfo.ClientParameters.Address, prefixLength)
	if err := runCmd(exec.Command("ifconfig", ifce.Name(), "inet6", "add", addr)); err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to set IP address: %w", err)
	}
	if err := runCmd(exec.Command("ifconfig", ifce.Name(), "mtu", fmt.Sprintf("%d", mtu), "up")); err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to configure MTU: %w", err)
	}
	if err := runCmd(exec.Command("ifconfig", ifce.Name(), "up")); err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to enable interface %s: %w", ifce.Name(), err)
	}
	return ifce, nil
}

// CheckPermissions verifies the process has sufficient privileges to create
// a utun interface for `ios tunnel start` (without `--userspace`).
//
// macOS has no capability system; utun creation is a root operation.
func CheckPermissions() error {
	if os.Geteuid() == 0 {
		return nil
	}
	return fmt.Errorf("this program needs root privileges. Run with sudo.")
}
