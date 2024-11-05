package ios

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/grandcat/zeroconf"
	log "github.com/sirupsen/logrus"
)

// FindDeviceInterfaceAddress tries to find the address of the device by browsing through all network interfaces.
// It uses mDNS to discover  the "_remoted._tcp" service on the local. domain. Then tries to connect to the RemoteServiceDiscovery
// and checks if the udid of the device matches the udid of the device we are looking for.
func FindDeviceInterfaceAddress(ctx context.Context, device DeviceEntry) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("FindDeviceInterfaceAddress: failed to get network interfaces: %w", err)
	}

	result := make(chan string)

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	for _, iface := range ifaces {
		resolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces([]net.Interface{iface}), zeroconf.SelectIPTraffic(zeroconf.IPv6))
		if err != nil {
			log.WithField("interface", iface.Name).
				WithField("err", err).
				Debug("failed to initialize resolver")
			continue
		}
		entries := make(chan *zeroconf.ServiceEntry)
		resolver.Browse(ctx, "_remoted._tcp", "local.", entries)
		go checkEntry(ctx, device, iface.Name, entries, result)

	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-result:
		log.WithField("udid", device.Properties.SerialNumber).WithField("address", r).Debug("found device address")
		return r, nil
	}
}

// checkEntry connects to all remote service discoveries and tests which one belongs to this device' udid.
func checkEntry(ctx context.Context, device DeviceEntry, interfaceName string, entries chan *zeroconf.ServiceEntry, result chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-entries:
			if entry == nil {
				continue
			}
			print(entry.ServiceInstanceName())
			for _, ip6 := range entry.AddrIPv6 {
				tryHandshake(ctx, ip6, entry.Port, interfaceName, device, result)
			}
		}
	}
}

func tryHandshake(ctx context.Context, ip6 net.IP, port int, interfaceName string, device DeviceEntry, result chan<- string) {
	addr := fmt.Sprintf("%s%%%s", ip6.String(), interfaceName)
	s, err := NewWithAddrPortDevice(addr, port, device)
	udid := device.Properties.SerialNumber
	if err != nil {
		slog.Error("failed to connect to remote service discovery", "error", err, "address", addr)
		return
	}
	defer s.Close()
	h, err := s.Handshake()
	if err != nil {
		return
	}
	if udid == h.Udid {
		select {
		case <-ctx.Done():
			slog.Error("failed sending handshake result", "error", ctx.Err())
		case result <- addr:
		}
	}
}
