package tunnel

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

type UserSpaceTUNInterface struct {
	NicID         tcpip.NICID
	EnableSniffer bool
	NetworkStack  *stack.Stack
}

func (iface *UserSpaceTUNInterface) TunnelRWCThroughInterface(localPort uint16, remoteAddr net.IP, remotePort uint16, rw io.ReadWriter) error {
	remote := tcpip.FullAddress{
		NIC:  iface.NicID,
		Addr: tcpip.AddrFromSlice(remoteAddr.To16()),
		Port: remotePort,
	}

	// Create TCP endpoint.
	var wq waiter.Queue
	ep, err := iface.NetworkStack.NewEndpoint(tcp.ProtocolNumber, ipv6.ProtocolNumber, &wq)
	if err != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: NewEndpoint failed: %+v", err)
	}
	ep.SocketOptions().SetKeepAlive(true)
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
		fmt.Println("Connect is pending...")
		<-notifyCh
		err = ep.LastError()
	}
	wq.EventUnregister(&waitEntry)

	if err != nil {
		return fmt.Errorf("TunnelRWCThroughInterface: Connect to remote failed: %+v", err)
	}

	slog.Info("Connected to ", "remoteAddr", remoteAddr, "remotePort", remotePort)
	remoteConn := gonet.NewTCPConn(&wq, ep)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go copyMyShit(remoteConn, rw, wg)
	go copyMyShit(rw, remoteConn, wg)
	wg.Wait()
	return nil
}

func copyMyShit(w io.Writer, r io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := io.Copy(w, r)
	if err != nil {
		log.Print(err)
	}

}

func (iface *UserSpaceTUNInterface) Init(mtu uint32, lockdownconn io.ReadWriteCloser, addrName string, prefixLength int) error {
	addr := tcpip.AddrFromSlice(net.ParseIP(addrName).To16())
	addrWithPrefix := addr.WithPrefix()
	addrWithPrefix.PrefixLen = prefixLength
	// Create the stack with ipv4 and tcp protocols, then add a tun-based
	// NIC and ipv4 address.
	iface.NetworkStack = stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv6.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	var linkEP stack.LinkEndpoint
	linkEP, err := RWCEndpointNew(lockdownconn, mtu, 0)
	if err != nil {
		return fmt.Errorf("initVirtualInterface: RWCEndpointNew failed: %+v", err)
	}

	nicID := tcpip.NICID(iface.NetworkStack.UniqueID())
	iface.NicID = nicID
	if iface.EnableSniffer {
		linkEP = sniffer.New(linkEP)
	}
	if err := iface.NetworkStack.CreateNIC(nicID, linkEP); err != nil {
		return fmt.Errorf("initVirtualInterface: CreateNIC failed: %+v", err)
	}

	protocolAddr := tcpip.ProtocolAddress{
		Protocol:          ipv6.ProtocolNumber,
		AddressWithPrefix: addrWithPrefix,
	}
	if err := iface.NetworkStack.AddProtocolAddress(1, protocolAddr, stack.AddressProperties{}); err != nil {
		return fmt.Errorf("initVirtualInterface: AddProtocolAddress(%d, %v, {}): %+v", nicID, protocolAddr, err)
	}

	// Add default route.
	iface.NetworkStack.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv6EmptySubnet,
			NIC:         nicID,
		},
	})
	return nil
}

func ConnectUserSpaceTunnelLockdown(device ios.DeviceEntry) (Tunnel, error) {
	conn, err := ios.ConnectToService(device, coreDeviceProxy)
	if err != nil {
		return Tunnel{}, err
	}
	return connectToUserspaceTunnelLockdown(context.TODO(), device, conn)
}

func connectToUserspaceTunnelLockdown(ctx context.Context, device ios.DeviceEntry, connToDevice io.ReadWriteCloser) (Tunnel, error) {
	logrus.Info("connect to lockdown tunnel endpoint on device")

	tunnelInfo, err := exchangeCoreTunnelParameters(connToDevice)
	if err != nil {
		return Tunnel{}, fmt.Errorf("could not exchange tunnel parameters. %w", err)
	}
	const prefixLength = 64
	iface := UserSpaceTUNInterface{}
	iface.Init(uint32(tunnelInfo.ClientParameters.Mtu), connToDevice, tunnelInfo.ClientParameters.Address, prefixLength)

	closeFunc := func() error {
		return nil
	}
	go listenToConns(iface)
	return Tunnel{
		Address: tunnelInfo.ServerAddress,
		RsdPort: int(tunnelInfo.ServerRSDPort),
		Udid:    device.Properties.SerialNumber,
		closer:  closeFunc,
	}, nil
}

func listenToConns(iface UserSpaceTUNInterface) error {
	defer func() {
		slog.Info("Stopped listening for connections")
	}()
	listener, err := net.Listen("tcp", "localhost:7779")
	if err != nil {
		return err
	}
	for {

		client, err := listener.Accept()
		if err != nil {
			return err
		}
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
