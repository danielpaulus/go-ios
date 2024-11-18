//go:build windows

package tunnel

import (
	"fmt"
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type tunWrapper struct {
	device Device

	buffer [][]byte
}

func initTUNwrapper(device Device) *tunWrapper {
	t := &tunWrapper{}

	t.device = device
	if device.BatchSize() != 1 {
		panic("batch size not 1")
	}

	mtu, _ := t.device.MTU()
	log.Infof("batch size: %d mtu:%d", device.BatchSize(), mtu)

	t.buffer = make([][]byte, 1)
	t.buffer[0] = make([]byte, mtu)
	go func() {
		for {
			e := <-device.Events()
			log.Infof("event: %v", e)
		}
	}()
	return t
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

	sizes := make([]int, 1)
	_, err := t.device.Read(t.buffer, sizes, 0)

	if err != nil {
		return 0, err
	}

	buf := t.buffer[0]
	size := sizes[0]
	copy(p, buf[:size])
	return size, err

}

func setupWindowsTUN(tunnelInfo tunnelParameters) (io.ReadWriteCloser, error) {
	name := "tun0"

	tunDevice, err := CreateTUN(name, int(tunnelInfo.ClientParameters.Mtu))
	if err != nil {
		fmt.Println("Error creating TUN device:", err)
		return &tunWrapper{}, err
	}
	tunname, err := tunDevice.Name()

	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to get interface name: %w", err)
	}
	const prefixLength = 64
	setIpAddr := exec.Command("netsh", "interface", "ipv6", "set", "address", tunname, fmt.Sprintf("%s/%d", tunnelInfo.ClientParameters.Address, prefixLength))
	err = runCmd(setIpAddr)
	if err != nil {
		return nil, fmt.Errorf("setupTunnelInterface: failed to set IP address for interface: %w", err)
	}
	log.Info("windows cmd")
	log.Info(setIpAddr.String())

	return initTUNwrapper(tunDevice), nil
}
