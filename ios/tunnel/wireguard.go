package tunnel

import (
	"fmt"
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/tun"
)

type tunWrapper struct {
	device      tun.Device
	readPackets [][]byte
	sizes       []int
}

func (t *tunWrapper) Close() error {
	return t.device.Close()
}

func (t *tunWrapper) Write(p []byte) (int, error) {
	bufs := [][]byte{p}                     // Create a slice of one byte slice
	written, err := t.device.Write(bufs, 0) // Use offset 0
	if written > 0 {
		return len(p), err // Assume the entire slice was written
	}
	return 0, err
}

func (t *tunWrapper) Read(p []byte) (int, error) {
	if len(t.readPackets) == 0 {
		bufs := make([][]byte, 100)
		for i := range bufs {
			mtu, _ := t.device.MTU()

			bufs[i] = make([]byte, mtu)
		}
		sizes := make([]int, 100)
		n, err := t.device.Read(bufs, sizes, 0)
		if n == 0 {
			return 0, err
		}
		if n > 1 {
			t.readPackets = bufs[1:]
			t.sizes = sizes[1:]
		}
		copy(p, bufs[0][:sizes[0]])
		return sizes[0], err
	}
	size := t.sizes[0]
	copy(p, t.readPackets[0][:size])
	t.readPackets = t.readPackets[1:]
	t.sizes = t.sizes[1:]
	return size, nil
}

func setupWindowsTUN(tunnelInfo tunnelParameters) (io.ReadWriteCloser, error) {
	name := "tun0"

	tunDevice, err := tun.CreateTUN(name, int(tunnelInfo.ClientParameters.Mtu))
	if err != nil {
		fmt.Println("Error creating TUN device:", err)
		return &tunWrapper{}, err
	}
	tunname, err := tunDevice.Name()

	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to get interface name: %w", err)
	}
	const prefixLength = 64
	setIpAddr := exec.Command("netsh", "interface", "ipv6", "add", "address", tunname, fmt.Sprintf("%s/%d", tunnelInfo.ClientParameters.Address, prefixLength))
	err = runCmd(setIpAddr)
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to set IP address for interface: %w", err)
	}
	log.Info("windows cmd")
	log.Info(setIpAddr.String())

	return &tunWrapper{device: tunDevice, readPackets: [][]byte{}, sizes: []int{}}, nil
}
