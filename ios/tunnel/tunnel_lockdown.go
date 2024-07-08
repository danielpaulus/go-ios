package tunnel

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/sirupsen/logrus"
)

const coreDeviceProxy = "com.apple.internal.devicecompute.CoreDeviceProxy"

func ConnectTunnelLockdown(device ios.DeviceEntry) (Tunnel, error) {
	conn, err := ios.ConnectToService(device, coreDeviceProxy)
	if err != nil {
		return Tunnel{}, err
	}
	return connectToTunnelLockdown(context.TODO(), device, conn)
}

func connectToTunnelLockdown(ctx context.Context, device ios.DeviceEntry, connToDevice io.ReadWriteCloser) (Tunnel, error) {
	logrus.Info("connect to lockdown tunnel endpoint on device")

	tunnelInfo, err := exchangeCoreTunnelParameters(connToDevice)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not exchange tunnel parameters. %w", err)
	}

	utunIface, err := setupTunnelInterface(tunnelInfo)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not setup tunnel interface. %w", err)
	}

	// we want a copy of the parent ctx here, but it shouldn't time out/be cancelled at the same time.
	// doing it like this allows us to have a context with a timeout for the tunnel creation, but the tunnel itself
	tunnelCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	go func() {
		err := forwardTCPToInterface(tunnelCtx, tunnelInfo.ClientParameters.Mtu, connToDevice, utunIface)
		if err != nil {
			logrus.WithError(err).Error("failed to forward data to tunnel interface")
		}
	}()

	go func() {
		err := forwardTUNToDevice(tunnelCtx, tunnelInfo.ClientParameters.Mtu, utunIface, connToDevice)
		if err != nil {
			logrus.WithError(err).Error("failed to forward data to the device")
		}
	}()

	closeFunc := func() error {
		cancel()
		return errors.Join(utunIface.Close(), connToDevice.Close())
	}
	return Tunnel{
		Address: tunnelInfo.ServerAddress,
		RsdPort: int(tunnelInfo.ServerRSDPort),
		Udid:    device.Properties.SerialNumber,
		closer:  closeFunc,
	}, nil
}

func forwardTUNToDevice(ctx context.Context, mtu uint64, tun io.Reader, deviceConn io.Writer) error {
	packet := make([]byte, mtu)
	for {

		select {
		case <-ctx.Done():
			return nil
		default:

			n, err := tun.Read(packet)

			if err != nil {
				return fmt.Errorf("could not read packet. %w", err)
			}

			_, err = deviceConn.Write(packet[:n])
			if err != nil {
				return fmt.Errorf("could not write packet. %w", err)
			}
		}

	}
}

func forwardTCPToInterface(ctx context.Context, mtu uint64, deviceConn io.Reader, tun io.Writer) error {
	payload := make([]byte, mtu)
	ip6Header := make([]byte, 40)

	br := bufio.NewReader(deviceConn)
	bw := bufio.NewWriter(tun)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, err := io.ReadFull(br, ip6Header)
			if err != nil {
				return fmt.Errorf("failed to read IPv6 header: %w", err)
			}

			if ip6Header[0] != 0x60 {
				return fmt.Errorf("not an IPv6 packet. expected 0x60, but got 0x%02x", ip6Header[0])
			}
			payloadLength := binary.BigEndian.Uint16(ip6Header[4:6])
			_, err = io.ReadFull(br, payload[:payloadLength])
			if err != nil {
				return fmt.Errorf("failed to read payload of length %d: %w", payloadLength, err)
			}

			// we don't need to check all errors here as `Flush` will return the error from a previous write as well
			_, _ = bw.Write(ip6Header)
			_, _ = bw.Write(payload[:payloadLength])
			err = bw.Flush()
			if err != nil {
				return fmt.Errorf("could not flush packet: %w", err)
			}
		}

	}
}
