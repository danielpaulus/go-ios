//go:build darwin

package utun

import (
	"context"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

type direction uint8

const (
	outgoing = iota
	incoming
)

type connections map[connectionId]*connection

type connectionId struct {
	localPort  layers.TCPPort
	remotePort layers.TCPPort
}

func Live(ctx context.Context, iface string, provider ios.RsdPortProvider, dumpDir string) error {
	addr, err := ifaceAddr(iface)
	if err != nil {
		return err
	}
	log.Infof("Capture traffice for iface %s with address %s", iface, addr)
	if handle, err := pcap.OpenLive(iface, 64*1024, true, pcap.BlockForever); err != nil {
		return fmt.Errorf("failed to connect to iface %s. %w", iface, err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		s := newSession(packetSource.Packets(), addr, provider, dumpDir)
		s.readPackets(ctx)
	}
	return nil
}

func ifaceAddr(name string) (net.IP, error) {
	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Name == name {
			return iface.Addresses[1].IP, nil
		}
	}
	return nil, fmt.Errorf("could not find iface")
}

func (s *session) connectionIdentifier(ip *layers.IPv6, tcp *layers.TCP) connectionId {
	if ip.SrcIP.String() == s.localAddr.String() {
		return connectionId{
			localPort:  tcp.SrcPort,
			remotePort: tcp.DstPort,
		}
	} else {
		return connectionId{
			localPort:  tcp.DstPort,
			remotePort: tcp.SrcPort,
		}
	}
}

type payloadWriter struct {
	incoming io.WriteCloser
	outgoing io.WriteCloser
}

func (p payloadWriter) Close() error {
	p.incoming.Close()
	p.outgoing.Close()
	return nil
}

func (p payloadWriter) Write(d direction, b []byte) (int, error) {
	switch d {
	case outgoing:
		return p.outgoing.Write(b)
	case incoming:
		return p.incoming.Write(b)
	default:
		return 0, fmt.Errorf("unknown direction")
	}
}
