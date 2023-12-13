//go:build darwin

package utun

import (
	"context"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
	"sync/atomic"
)

type session struct {
	localAddr         net.IP
	packetSrc         chan gopacket.Packet
	activeConnections connections
	dumpDir           string
	connectionNum     atomic.Uint32
	rsdProvider       ios.RsdPortProvider
}

func newSession(packets chan gopacket.Packet, addr net.IP, provider ios.RsdPortProvider, dumpDir string) session {
	return session{
		localAddr:         addr,
		packetSrc:         packets,
		activeConnections: connections{},
		dumpDir:           dumpDir,
		rsdProvider:       provider,
	}
}

func (s *session) readPackets(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("context cancelled. closing all connections")
			for _, c := range s.activeConnections {
				c.Close()
			}
			return
		case packet := <-s.packetSrc:
			err := s.handlePacket(packet)
			if err != nil {
				log.Warnf("failed to handle packet: %s", packet.Dump())
			}
		}
	}
}

func (s *session) handlePacket(p gopacket.Packet) error {
	ip, ok := p.NetworkLayer().(*layers.IPv6)
	if !ok {
		return fmt.Errorf("only ipv6 is supported")
	}
	tcp, ok := p.TransportLayer().(*layers.TCP)
	if !ok {
		return fmt.Errorf("only tcp is supported")
	}
	id := s.connectionIdentifier(ip, tcp)
	conn := s.getOrCreateConnection(id)
	conn.handlePacket(p, ip, tcp)
	if tcp.RST || tcp.FIN {
		conn.Close()
		delete(s.activeConnections, id)
	}
	return nil
}

func (s *session) getOrCreateConnection(id connectionId) *connection {
	c, ok := s.activeConnections[id]
	if ok {
		return c
	} else {
		service := s.rsdProvider.GetService(int(id.remotePort))
		if service == "" {
			service = "unknown"
		}
		log.Infof("connection to service %s (%d)", service, id.remotePort)
		p := path.Join(s.dumpDir, fmt.Sprintf("%04d-%s", s.connectionNum.Add(1), service))
		err := os.MkdirAll(p, os.ModePerm)
		if err != nil {
			panic(err)
		}

		conn := newConnection(id, p, service)
		s.activeConnections[id] = conn
		return conn
	}
}
