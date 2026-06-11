package ios

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/grandcat/zeroconf"
)

// RemotedDevice is a device discovered via the "_remoted._tcp" Bonjour service,
// identified by an RSD handshake.
type RemotedDevice struct {
	Udid    string
	Address string // "<ip6>%<iface>"
	RsdPort int
}

// BrowseRemoted discovers all devices reachable over a RemoteServiceDiscovery
// tunnel by browsing the "_remoted._tcp" mDNS service over every IPv6 network
// interface and RSD-handshaking each discovered entry to read its real udid.
// Results are de-duplicated by udid. The whole browse is bounded by ctx (the
// caller supplies a timeout); whatever was found when ctx expires is returned.
// Entries whose handshake fails are logged and skipped (best-effort).
func BrowseRemoted(ctx context.Context) ([]RemotedDevice, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("BrowseRemoted: failed to get network interfaces: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var mu sync.Mutex
	byUdid := map[string]RemotedDevice{}

	var wg sync.WaitGroup
	for _, iface := range ifaces {
		resolver, err := zeroconf.NewResolver(zeroconf.SelectIfaces([]net.Interface{iface}), zeroconf.SelectIPTraffic(zeroconf.IPv6))
		if err != nil {
			golog.Debug("failed to initialize resolver", "module", logModule, "interface", iface.Name, "err", err)
			continue
		}
		entries := make(chan *zeroconf.ServiceEntry)
		if err := resolver.Browse(ctx, "_remoted._tcp", "local.", entries); err != nil {
			golog.Debug("failed to browse remoted service", "module", logModule, "interface", iface.Name, "err", err)
			continue
		}
		wg.Add(1)
		go func(interfaceName string, entries chan *zeroconf.ServiceEntry) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case entry := <-entries:
					if entry == nil {
						continue
					}
					for _, ip6 := range entry.AddrIPv6 {
						dev, ok := handshakeRemoted(ip6, entry.Port, interfaceName)
						if !ok {
							continue
						}
						mu.Lock()
						if _, exists := byUdid[dev.Udid]; !exists {
							byUdid[dev.Udid] = dev
						}
						mu.Unlock()
					}
				}
			}
		}(iface.Name, entries)
	}

	<-ctx.Done()
	wg.Wait()

	result := make([]RemotedDevice, 0, len(byUdid))
	for _, d := range byUdid {
		result = append(result, d)
	}
	return result, nil
}

// handshakeRemoted RSD-handshakes a single discovered remoted address to read
// its udid. It returns false (and logs at debug level) if the handshake fails.
func handshakeRemoted(ip6 net.IP, port int, interfaceName string) (RemotedDevice, bool) {
	addr := fmt.Sprintf("%s%%%s", ip6.String(), interfaceName)
	s, err := NewWithAddrPortDevice(addr, port, DeviceEntry{})
	if err != nil {
		golog.Debug("failed to connect to remote service discovery", "module", logModule, "address", addr, "err", err)
		return RemotedDevice{}, false
	}
	defer s.Close()
	h, err := s.Handshake()
	if err != nil {
		golog.Debug("remote service discovery handshake failed", "module", logModule, "address", addr, "err", err)
		return RemotedDevice{}, false
	}
	if h.Udid == "" {
		return RemotedDevice{}, false
	}
	return RemotedDevice{Udid: h.Udid, Address: addr, RsdPort: port}, true
}

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
			golog.Debug("failed to initialize resolver", "module", logModule, "udid", device.Properties.SerialNumber, "interface", iface.Name, "err", err)
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
		golog.Debug("found device address", "module", logModule, "udid", device.Properties.SerialNumber, "address", r)
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
			fmt.Print(entry.ServiceInstanceName())
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
		golog.Error("failed to connect to remote service discovery", "module", logModule, "udid", udid, "error", err, "address", addr)
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
			golog.Error("failed sending handshake result", "module", logModule, "udid", udid, "error", ctx.Err())
		case result <- addr:
		}
	}
}
