package ios

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	log "github.com/sirupsen/logrus"
)

type SDResponse struct {
	HandshakeResponse RsdHandshakeResponse
	InterfaceName     string
	ServiceEntry      *zeroconf.ServiceEntry
	Ipv6              net.IP
	Err               error
}

func (r SDResponse) Addr() string {
	return fmt.Sprintf("%s%%%s", r.Ipv6.String(), r.InterfaceName)
}

// FindDeviceInterfaceAddress tries to find the address of the device by browsing through all network interfaces.
// It uses mDNS to discover  the "_remoted._tcp" service on the local. domain. Then tries to connect to the RemoteServiceDiscovery
// and checks if the udid of the device matches the udid of the device we are looking for.
func FindDeviceInterfaceAddress(ctx context.Context, device DeviceEntry) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("FindDeviceInterfaceAddress: failed to get network interfaces: %w", err)
	}

	result := make(chan SDResponse)
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
		resolver.Browse(ctx, "_remoted._tcp", "local.", entries)
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go checkEntry(ctx, iface.Name, entries, result, wg)

	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-result:
		log.WithField("udid", device.Properties.SerialNumber).WithField("address", r).Debug("found device address")
		addr := fmt.Sprintf("%s%%%s", r.Ipv6.String(), r.InterfaceName)
		return addr, nil
	}
}

const RemoteDLocal = "_remoted._tcp"

func FindDevicesForService(ctx context.Context, service string) ([]SDResponse, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("FindDeviceInterfaceAddress: failed to get network interfaces: %w", err)
	}

	result := make(chan SDResponse)
	defer close(result)
	wg := sync.WaitGroup{}
	for _, iface := range ifaces {
		resolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces([]net.Interface{iface}), zeroconf.SelectIPTraffic(zeroconf.IPv6))
		if err != nil {
			log.WithField("interface", iface.Name).
				WithField("err", err).
				Debug("failed to initialize resolver")
			continue
		}
		entries := make(chan *zeroconf.ServiceEntry)
		resolver.Browse(ctx, service, "local.", entries)
		wg.Add(1)
		go checkEntry(ctx, iface.Name, entries, result, &wg)

	}
	results := []SDResponse{}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-result:
				if !ok {
					return
				}
				results = append(results, r)
				//log.WithField("udid", device.Properties.SerialNumber).WithField("address", r).Debug("found device address")

			}
		}
	}()

	wg.Wait()
	return results, nil
}

// checkEntry connects to all remote service discoveries and tests which one belongs to this device' udid.
func checkEntry(ctx context.Context, interfaceName string, entries chan *zeroconf.ServiceEntry, result chan<- SDResponse, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-time.After(time.Second):
			return
		case <-ctx.Done():
			return
		case entry := <-entries:
			if entry == nil {
				continue
			}
			for _, ip6 := range entry.AddrIPv6 {
				tryHandshake(ip6, entry, interfaceName, result, wg)
			}
		}
	}

}

var NewTCP func(string, int) (RsdService, error)

func tryHandshake(ip6 net.IP, svcEntry *zeroconf.ServiceEntry, interfaceName string, result chan<- SDResponse, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	resp := SDResponse{
		Err:           nil,
		InterfaceName: interfaceName,
		Ipv6:          ip6,
		ServiceEntry:  svcEntry,
	}
	addr := fmt.Sprintf("%s%%%s", ip6.String(), interfaceName)
	port := svcEntry.Port
	ff := NewTCP
	if ff == nil {
		ff = NewWithAddrPort
	}
	s, err := ff(addr, port)
	if err != nil {
		resp.Err = err
		result <- resp
		return
	}
	defer s.Close()
	h, err := s.Handshake()
	if err != nil {
		resp.Err = err
		result <- resp
		return
	}
	resp.HandshakeResponse = h
	result <- resp
}
