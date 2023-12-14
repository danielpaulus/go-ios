package ios

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	log "github.com/sirupsen/logrus"
	"net"
)

func FindDeviceInterfaceAddress(ctx context.Context, device DeviceEntry) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	result := make(chan string)
	defer close(result)

	for _, iface := range ifaces {
		resolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces([]net.Interface{iface}), zeroconf.SelectIPTraffic(zeroconf.IPv6))
		if err != nil {
			log.WithField("interface", iface.Name).
				WithField("err", err).
				Debug("failed to initialize resolver")
			continue
		}
		entries := make(chan *zeroconf.ServiceEntry)
		err = resolver.Browse(ctx, "_remoted._tcp", "local.", entries)
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

func checkEntry(ctx context.Context, device DeviceEntry, interfaceName string, entries chan *zeroconf.ServiceEntry, result chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-entries:
			for _, ip6 := range entry.AddrIPv6 {
				log.WithField("adrr", ip6).WithField("ifce", interfaceName).Info("query addr")
				addr := fmt.Sprintf("%s%%%s", ip6.String(), interfaceName)
				s, err := NewWithAddr(addr)
				if err != nil {
					continue
				}
				defer s.Close()
				h, err := s.Handshake()
				if err != nil {
					continue
				}
				if device.Properties.SerialNumber == h.Udid {
					result <- addr
				}
			}
		}
	}
}
