package tunnel

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

// ioResourceCloser is a type for closing function.
type ioResourceCloser func()

// createIoCloser returns a ioResourceCloser for closing both writer and together
func createIoCloser(rw1, rw2 io.ReadWriteCloser) ioResourceCloser {

	// Using sync.Once is essential to close writer and reader just once
	var once sync.Once
	return func() {
		once.Do(func() {
			rw1.Close()
			rw2.Close()
		})
	}
}

// UserSpaceTUNInterface uses gVisor's netstack to create a userspace virtual network interface.
// You can use it to connect local tcp connections to remote adresses on the network.
// Set it up with the Init method and provide a io.ReadWriter to a IP/TUN compatible device.
// If EnableSniffer, raw TCP packets will be dumped to the console.
type UserSpaceTUNInterface struct {
	nicID tcpip.NICID
	//If EnableSniffer, raw TCP packets will be dumped to the console.
	EnableSniffer bool
	networkStack  *stack.Stack
}

func (iface *UserSpaceTUNInterface) TunnelRWCThroughInterface(localPort uint16, remoteAddr net.IP, remotePort uint16, rw io.ReadWriteCloser) error {
	defer rw.Close()
	remote := tcpip.FullAddress{
		NIC:  iface.nicID,
		Addr: tcpip.AddrFromSlice(remoteAddr.To16()),
		Port: remotePort,
	}

	// Create TCP endpoint.
	var wq waiter.Queue
	ep, err := iface.networkStack.NewEndpoint(tcp.ProtocolNumber, ipv6.ProtocolNumber, &wq)
	if err != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: NewEndpoint failed: %+v", err)
	}

	ep.SocketOptions().SetKeepAlive(true)
	// Set keep alive idle value more aggresive than the gVisor's 2 hours. NAT and Firewalls can drop the idle connections more aggresive.
	p := tcpip.KeepaliveIdleOption(30 * time.Second)
	ep.SetSockOpt(&p)

	o := tcpip.KeepaliveIntervalOption(1 * time.Second)
	ep.SetSockOpt(&o)

	// Bind if a port is specified.
	if localPort != 0 {
		if err := ep.Bind(tcpip.FullAddress{Port: localPort}); err != nil {
			return fmt.Errorf("TunnelRWCThroughInterface: Bind failed: %+v", err)
		}
	}
	// Issue connect request and wait for it to complete.
	waitEntry, notifyCh := waiter.NewChannelEntry(waiter.WritableEvents)
	wq.EventRegister(&waitEntry)
	err = ep.Connect(remote)
	if _, ok := err.(*tcpip.ErrConnectStarted); ok {
		<-notifyCh
		err = ep.LastError()
	}
	wq.EventUnregister(&waitEntry)
	if err != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: Connect to remote failed: %+v", err)
	}

	slog.Info("Connected to ", "remoteAddr", remoteAddr, "remotePort", remotePort)
	remoteConn := gonet.NewTCPConn(&wq, ep)
	defer remoteConn.Close()
	perr := proxyConns(rw, remoteConn)
	if perr != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: proxyConns failed: %+v", perr)
	}
	return nil
}

func proxyConns(rw1 io.ReadWriteCloser, rw2 io.ReadWriteCloser) error {

	// Use buffered channel for non-blocking send recieve. We use the same single channel 2 times for 2 ioCopyWithErr.
	errCh := make(chan error, 2)

	// Create a IO closing functions to unblock stuck io.Copy() call
	ioCloser := createIoCloser(rw1, rw2)

	// Send same error channel and the io close function
	go ioCopyWithErr(rw1, rw2, errCh, ioCloser)
	go ioCopyWithErr(rw2, rw1, errCh, ioCloser)

	// Read from error channel. As the channel is a FIFO queue first in first out, each <-errCh will read one message and remove it from the channel.
	// Order of messages are not important.
	err1 := <-errCh
	err2 := <-errCh

	return errors.Join(err1, err2)
}

func ioCopyWithErr(w io.Writer, r io.Reader, errCh chan error, ioCloser ioResourceCloser) {
	_, err := io.Copy(w, r)
	errCh <- err

	// Close the writer and reader to notify the second io.Copy() if one part of the connection closed.
	// This is also necessary to avoid resource leaking.
	ioCloser()
}

// Init initializes the virtual network interface.
// The connToTUNIface needs to be connection that understands IP packets to a remote TUN device or sth.
// provide mtu, ip address as a string and the prefix length of the interface.
func (iface *UserSpaceTUNInterface) Init(mtu uint32, connToTUNIface io.ReadWriteCloser, ipAddrString string, prefixLength int) error {
	addr := tcpip.AddrFromSlice(net.ParseIP(ipAddrString).To16())
	addrWithPrefix := addr.WithPrefix()
	addrWithPrefix.PrefixLen = prefixLength

	//Create a new stack, ipv6 is enough for ios devices
	iface.networkStack = stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv6.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	// connToTUNIface needs to be connection that understands IP packets,
	// so we can use it to link it against a virtual network interface
	var linkEP stack.LinkEndpoint
	linkEP, err := RWCEndpointNew(connToTUNIface, mtu, 0)
	if err != nil {
		return fmt.Errorf("initVirtualInterface: RWCEndpointNew failed: %+v", err)
	}

	nicID := tcpip.NICID(iface.networkStack.UniqueID())
	iface.nicID = nicID
	if iface.EnableSniffer {
		linkEP = sniffer.New(linkEP)
	}
	if err := iface.networkStack.CreateNIC(nicID, linkEP); err != nil {
		return fmt.Errorf("initVirtualInterface: CreateNIC failed: %+v", err)
	}

	protocolAddr := tcpip.ProtocolAddress{
		Protocol:          ipv6.ProtocolNumber,
		AddressWithPrefix: addrWithPrefix,
	}
	if err := iface.networkStack.AddProtocolAddress(iface.nicID, protocolAddr, stack.AddressProperties{}); err != nil {
		return fmt.Errorf("initVirtualInterface: AddProtocolAddress(%d, %v, {}): %+v", nicID, protocolAddr, err)
	}

	// Add default route.
	iface.networkStack.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv6EmptySubnet,
			NIC:         nicID,
		},
	})
	return nil
}

func ConnectUserSpaceTunnelLockdown(device ios.DeviceEntry, ifacePort int) (Tunnel, error) {
	conn, err := ios.ConnectToService(device, coreDeviceProxy)
	if err != nil {
		return Tunnel{}, err
	}
	return connectToUserspaceTunnelLockdown(context.TODO(), device, conn, ifacePort)
}

func connectToUserspaceTunnelLockdown(ctx context.Context, device ios.DeviceEntry, connToDevice io.ReadWriteCloser, ifacePort int) (Tunnel, error) {
	slog.Info("connect to lockdown tunnel endpoint on device")
	tunnelInfo, err := exchangeCoreTunnelParameters(connToDevice)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not exchange tunnel parameters. %w", err)
	}
	const prefixLength = 64
	iface := UserSpaceTUNInterface{}
	err = iface.Init(uint32(tunnelInfo.ClientParameters.Mtu), connToDevice, tunnelInfo.ClientParameters.Address, prefixLength)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not setup tunnel interface. %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", ifacePort))
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not setup listener. %w", err)
	}

	listener.Addr()
	go listenToConns(iface, listener)

	closeFunc := func() error {
		iface.networkStack.Close()
		return errors.Join(connToDevice.Close(), listener.Close())
	}
	return Tunnel{
		Address: tunnelInfo.ServerAddress,
		RsdPort: int(tunnelInfo.ServerRSDPort),
		Udid:    device.Properties.SerialNumber,
		closer:  closeFunc,
	}, nil
}

func listenToConns(iface UserSpaceTUNInterface, listener net.Listener) error {
	defer func() {
		slog.Info("Stopped listening for connections")
	}()

	for {
		client, err := listener.Accept()
		if err != nil {
			return err
		}
		slog.Info("Received connection request", "from", client.RemoteAddr(), "to", client.LocalAddr())
		remoteAddrBytes := make([]byte, 16)
		_, err = client.Read(remoteAddrBytes)
		if err != nil {
			return err
		}

		remotePortBytes := make([]byte, 4)
		_, err = client.Read(remotePortBytes)
		port := binary.LittleEndian.Uint32(remotePortBytes)
		slog.Info("Received connection request to device ", "ip", net.IP(remoteAddrBytes), "port", port)
		go iface.TunnelRWCThroughInterface(0, net.IP(remoteAddrBytes), uint16(port), client)
	}
}
