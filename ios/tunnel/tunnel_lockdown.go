package tunnel

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

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

const UDP = 0x11

// https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
func iPv6Parser(stream io.Reader) string {
	buf := make([]byte, 66000)
	// magic header and flags
	stream.Read(buf[:40])
	fmt.Printf("magic header and flags:%x\n", buf[:4])
	if (buf[0] & 0xF0) == 4 {
		print("dropping ipv4")
		length := binary.BigEndian.Uint16(buf[2:4])
		print(length)
		print("\n")
		stream.Read(buf[:length-4])
	}
	// length

	length := binary.BigEndian.Uint16(buf[4:6])

	// Combine the bytes into a single 32-bit value.
	combined := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	trafficClass := (combined >> 20) & 0xFF

	fmt.Printf("Traffic Class: %X\n", trafficClass)
	// Mask out the first 12 bits (version and traffic class) and keep the last 20 bits (flow label).
	flowLabel := combined & 0x000FFFFF

	fmt.Printf("Flow Label: %X\n", flowLabel)
	//protocol, like TCP 0x06
	fmt.Printf("next header %x\n", buf[6])

	// TTL, can be anything
	fmt.Printf("hop limit %x\n", buf[7])
	// next header, hop limit

	sourceAddressB := buf[8:24]
	destAddressB := buf[24:40]

	var hexStrings []string
	for _, b := range sourceAddressB {
		hexStrings = append(hexStrings, fmt.Sprintf("%02X", b))
	}

	sourceIP := strings.Join(hexStrings, ":")

	var hexStrings1 []string
	for _, b := range destAddressB {
		hexStrings1 = append(hexStrings1, fmt.Sprintf("%02X", b))
	}
	destIP := strings.Join(hexStrings1, ":")

	protocol := buf[6]
	prot := ""
	if protocol == UDP {
		prot = "UDP"
	} else {
		prot = fmt.Sprintf("PROTOCOL:%d", protocol)
	}
	stream.Read(buf[:length])
	return fmt.Sprintf("IP len:%d transport:%s source:%s dest:%s", length, prot, sourceIP, destIP)
}

func forwardTUNToDevice(ctx context.Context, mtu uint64, tun io.Reader, deviceConn io.Writer) error {
	reader, writer := io.Pipe()
	go func() {
		for {
			fmt.Println(iPv6Parser(reader))

		}
	}()
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
			_, err = writer.Write(packet[:n])
			if err != nil {
				slog.Error("failed to write to binary file", "error", err)
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
