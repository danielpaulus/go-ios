package tunnel

import (
	"context"
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
		err := forwardTCPToInterface(tunnelCtx, connToDevice, utunIface)
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

func forwardTCPToInterface(ctx context.Context, deviceConn io.Reader, tun io.Writer) error {
	b := make([]byte, 20000)
	for {

		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := deviceConn.Read(b)
			if err != nil {
				return fmt.Errorf("failed to read datagram. %w", err)
			}
			_, err = tun.Write(b[:n])
			if err != nil {
				return fmt.Errorf("failed to forward data. %w", err)
			}
		}

	}
}
