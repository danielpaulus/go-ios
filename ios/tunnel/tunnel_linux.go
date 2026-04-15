//go:build linux

package tunnel

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

// createSysProcAttr returns attributes for spawning the go-ios agent in a
// new session so it survives as a standalone daemon.
func createSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// setupTunnelInterface creates a TUN interface and assigns the tunnel's IPv6
// address to it using netlink. No external binaries are required — this
// avoids the deprecated ifconfig tool (missing on many modern distros and
// minimal container images) and the file-capability juggling that shelling
// out to helper binaries requires when running as a non-root user with
// CAP_NET_ADMIN.
//
// MTU is left at the kernel default (1500); the kernel fragments IPv6
// packets down to the path MTU as needed. We cannot set the MTU below
// 1280 (the IPv6 minimum) on Linux anyway.
func setupTunnelInterface(tunnelInfo tunnelParameters) (io.ReadWriteCloser, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed creating TUN device: %w", err)
	}

	const prefixLength = 64 // TODO: could be derived from the netmask provided by the device

	link, err := netlink.LinkByName(ifce.Name())
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: LinkByName(%q): %w", ifce.Name(), err)
	}
	addr, err := netlink.ParseAddr(fmt.Sprintf("%s/%d", tunnelInfo.ClientParameters.Address, prefixLength))
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: ParseAddr: %w", err)
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: AddrAdd %s: %w", addr, err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: LinkSetUp: %w", err)
	}
	return ifce, nil
}

// CheckPermissions verifies the process has sufficient privileges to create
// a kernel TUN interface for `ios tunnel start` (without `--userspace`).
//
// Linux accepts either EUID == 0 or a process that holds CAP_NET_ADMIN in
// its effective capability set. Containers commonly grant CAP_NET_ADMIN to
// unprivileged users; requiring full root there is over-strict.
func CheckPermissions() error {
	if os.Geteuid() == 0 {
		return nil
	}
	if hasCapNetAdmin() {
		return nil
	}
	return fmt.Errorf("this program needs root privileges or CAP_NET_ADMIN. Run with sudo, or grant CAP_NET_ADMIN to the process (e.g. via setcap or container cap_add).")
}

// capNetAdmin is the Linux capability number for CAP_NET_ADMIN.
// See include/uapi/linux/capability.h.
const capNetAdmin = 12

// hasCapNetAdmin reports whether the current process holds CAP_NET_ADMIN in
// its effective capability set. It reads /proc/self/status and returns false
// on any error (e.g. unreadable procfs).
func hasCapNetAdmin() bool {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}
	caps, ok := parseCapEff(string(data))
	if !ok {
		return false
	}
	return caps&(1<<capNetAdmin) != 0
}

// parseCapEff extracts the CapEff hex value from the contents of
// /proc/self/status. Returns the parsed capabilities bitmap and true on
// success, or (0, false) if the line is missing or unparseable.
func parseCapEff(status string) (uint64, bool) {
	for _, line := range strings.Split(status, "\n") {
		rest, ok := strings.CutPrefix(line, "CapEff:")
		if !ok {
			continue
		}
		caps, err := strconv.ParseUint(strings.TrimSpace(rest), 16, 64)
		if err != nil {
			return 0, false
		}
		return caps, true
	}
	return 0, false
}
