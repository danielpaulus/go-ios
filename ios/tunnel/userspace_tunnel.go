package tunnel

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

var networkStack *stack.Stack

func ConnectUserSpace(localportname string, remoteAddr net.IP, remotePort uint16, rwc io.ReadWriteCloser) error {
	remote := tcpip.FullAddress{
		NIC:  1,
		Addr: tcpip.AddrFromSlice(remoteAddr.To16()),
	}

	remote.Port = remotePort

	var localPort uint16
	if v, err := strconv.Atoi(localportname); err != nil {
		log.Fatalf("Unable to convert port %v: %v", localportname, err)
	} else {
		localPort = uint16(v)
	}

	// Create TCP endpoint.
	var wq waiter.Queue
	ep, e := networkStack.NewEndpoint(tcp.ProtocolNumber, ipv6.ProtocolNumber, &wq)
	if e != nil {
		log.Fatal(e)
	}
	ep.SocketOptions().SetKeepAlive(true)
	o := tcpip.KeepaliveIntervalOption(1 * time.Second)
	ep.SetSockOpt(&o)
	// Bind if a port is specified.
	if localPort != 0 {
		if err := ep.Bind(tcpip.FullAddress{Port: localPort}); err != nil {
			log.Fatal("Bind failed: ", err)
		}
	}
	// Issue connect request and wait for it to complete.
	waitEntry, notifyCh := waiter.NewChannelEntry(waiter.WritableEvents)
	wq.EventRegister(&waitEntry)
	terr := ep.Connect(remote)
	if _, ok := terr.(*tcpip.ErrConnectStarted); ok {
		fmt.Println("Connect is pending...")
		<-notifyCh
		terr = ep.LastError()
	}
	wq.EventUnregister(&waitEntry)

	if terr != nil {
		log.Fatal("Unable to connect: ", terr)
	}

	fmt.Println("Connected")
	c := gonet.NewTCPConn(&wq, ep)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go copyMyShit(c, rwc, wg)
	go copyMyShit(rwc, c, wg)
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

func startuserspacetunnel(mtu uint32, lockdownconn io.ReadWriteCloser, addrName string, prefixLength int) {
	/*	if len(os.Args) != 6 {
			log.Fatal("Usage: ", os.Args[0], " <tun-device> <local-ipv4-address> <local-port> <remote-ipv4-address> <remote-port>")
		}
	*/
	//tunName := os.Args[1]
	//addrName := os.Args[2]

	rand.Seed(time.Now().UnixNano())

	addr := tcpip.AddrFromSlice(net.ParseIP(addrName).To16())
	addrWithPrefix := addr.WithPrefix()
	addrWithPrefix.PrefixLen = prefixLength
	// Create the stack with ipv4 and tcp protocols, then add a tun-based
	// NIC and ipv4 address.
	networkStack = stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv6.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	/*
		fd, err := tun.Open(tunName)
		if err != nil {
			log.Fatal(err)
		}*/

	/*linkEP, err := fdbased.New(&fdbased.Options{FDs: []int{fd}, MTU: mtu})
	if err != nil {
		log.Fatal(err)
	}*/
	/*	linkEP := channel.New(50, mtu, tcpip.GetRandMacAddr())
		if err := stack_orinterfaceithink.CreateNIC(1, sniffer.New(linkEP)); err != nil {
			log.Fatal(err)
		}
	*/
	linkEP, err := RWCEndpointNew(lockdownconn, mtu, 0)
	if err != nil {
		log.Fatal(err)
	}
	/*if debug {
		linkEP = sniffer.New(linkEP)
	}*/
	if err := networkStack.CreateNIC(1, linkEP); err != nil {
		log.Fatal(err)
	}

	protocolAddr := tcpip.ProtocolAddress{
		Protocol:          ipv6.ProtocolNumber,
		AddressWithPrefix: addrWithPrefix,
	}
	if err := networkStack.AddProtocolAddress(1, protocolAddr, stack.AddressProperties{}); err != nil {
		log.Fatalf("AddProtocolAddress(%d, %+v, {}): %s", 1, protocolAddr, err)
	}

	// Add default route.
	networkStack.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv6EmptySubnet,
			NIC:         1,
		},
	})
	log.Print("The stack is now running; press Ctrl-C to stop")

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

	startuserspacetunnel(uint32(tunnelInfo.ClientParameters.Mtu), connToDevice, tunnelInfo.ClientParameters.Address, prefixLength)

	closeFunc := func() error {
		return nil
	}
	go listenToConns()
	return Tunnel{
		Address: tunnelInfo.ServerAddress,
		RsdPort: int(tunnelInfo.ServerRSDPort),
		Udid:    device.Properties.SerialNumber,
		closer:  closeFunc,
	}, nil
}

func listenToConns() error {
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
		go ConnectUserSpace("0", net.IP(remoteAddrBytes), uint16(port), client)
	}
}
