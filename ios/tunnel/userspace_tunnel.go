// Copyright 2018 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This sample creates a stack with TCP and IPv4 protocols on top of a TUN
// device, and connects to a peer. Similar to "nc <address> <port>". While the
// sample is running, attempts to connect to its IPv4 address will result in
// a RST segment.
//
// As an example of how to run it, a TUN device can be created and enabled on
// a linux host as follows (this only needs to be done once per boot):
//
// [sudo] ip tuntap add user <username> mode tun <device-name>
// [sudo] ip link set <device-name> up
// [sudo] ip addr add <ipv4-address>/<mask-length> dev <device-name>
//
// A concrete example:
//
// $ sudo ip tuntap add user wedsonaf mode tun tun0
// $ sudo ip link set tun0 up
// $ sudo ip addr add 192.168.1.1/24 dev tun0
//
// Then one can run tun_tcp_connect as such:
//
// $ ./tun/tun_tcp_connect tun0 192.168.1.2 0 192.168.1.1 1234
//
// This will attempt to connect to the linux host's stack. One can run nc in
// listen mode to accept a connect from tun_tcp_connect and exchange data.
package tunnel

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

// writer reads from standard input and writes to the endpoint until standard
// input is closed. It signals that it's done by closing the provided channel.
func writer(ch chan struct{}, ep tcpip.Endpoint, rwc io.ReadWriteCloser) {
	defer func() {
		ep.Shutdown(tcpip.ShutdownWrite)
		close(ch)
	}()

	if err := func() error {
		bs := make([]byte, 1500)
		for {
			var b bytes.Buffer

			n, err := rwc.Read(bs)
			if err != nil {
				return fmt.Errorf("rwc.Read failed: %s", err)
			}

			b.Write(bs[:n])
			for b.Len() != 0 {
				if _, err := ep.Write(&b, tcpip.WriteOptions{Atomic: true}); err != nil {
					return fmt.Errorf("ep.Write failed: %s", err)
				}
			}
		}
	}(); err != nil {
		fmt.Println(err)
	}
}

var interface_i_think *stack.Stack

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
	ep, e := interface_i_think.NewEndpoint(tcp.ProtocolNumber, ipv6.ProtocolNumber, &wq)
	if e != nil {
		log.Fatal(e)
	}

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

	// Start the writer in its own goroutine.
	writerCompletedCh := make(chan struct{})
	go writer(writerCompletedCh, ep, rwc) // S/R-SAFE: sample code.

	// Read data and write to standard output until the peer closes the
	// connection from its side.
	waitEntry, notifyCh = waiter.NewChannelEntry(waiter.ReadableEvents)
	wq.EventRegister(&waitEntry)
	for {
		_, err := ep.Read(rwc, tcpip.ReadOptions{})
		if err != nil {
			if _, ok := err.(*tcpip.ErrClosedForReceive); ok {
				break
			}

			if _, ok := err.(*tcpip.ErrWouldBlock); ok {
				<-notifyCh
				continue
			}

			log.Fatal("Read() failed:", err)
		}
	}
	wq.EventUnregister(&waitEntry)

	// The reader has completed. Now wait for the writer as well.
	<-writerCompletedCh

	ep.Close()
	return nil
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
	stack_orinterfaceithink := stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv6.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})
	interface_i_think = stack_orinterfaceithink
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
	if err := stack_orinterfaceithink.CreateNIC(1, sniffer.New(linkEP)); err != nil {
		log.Fatal(err)
	}

	protocolAddr := tcpip.ProtocolAddress{
		Protocol:          ipv6.ProtocolNumber,
		AddressWithPrefix: addrWithPrefix,
	}
	if err := stack_orinterfaceithink.AddProtocolAddress(1, protocolAddr, stack.AddressProperties{}); err != nil {
		log.Fatalf("AddProtocolAddress(%d, %+v, {}): %s", 1, protocolAddr, err)
	}

	// Add default route.
	stack_orinterfaceithink.SetRouteTable([]tcpip.Route{
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
