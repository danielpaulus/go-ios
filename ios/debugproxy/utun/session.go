package utun

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"sync/atomic"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
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
		return fmt.Errorf("handlePacket: can not handle packet. only IPv6 is supported")
	}
	tcp, ok := p.TransportLayer().(*layers.TCP)
	if !ok {
		return fmt.Errorf("handlePacket: can not handle packet. only TCP is supported")
	}
	id := s.connectionIdentifier(ip, tcp)
	conn, err := s.getOrCreateConnection(id)
	if err != nil {
		return fmt.Errorf("handlePacket: failed to get connection: %w", err)
	}
	conn.handlePacket(tcp)
	if tcp.RST || tcp.FIN {
		_ = conn.Close()
		delete(s.activeConnections, id)
	}
	return nil
}

func (s *session) getOrCreateConnection(id connectionId) (*connection, error) {
	c, ok := s.activeConnections[id]
	if ok {
		return c, nil
	} else {
		service := s.rsdProvider.GetService(int(id.remotePort))
		if service == "" {
			service = "unknown"
		}
		log.Infof("connection to service %s (%d)", service, id.remotePort)
		p := path.Join(s.dumpDir, fmt.Sprintf("%04d-%s", s.connectionNum.Add(1), service))
		err := os.MkdirAll(p, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("getOrCreateConnection: failed to create directory for connection dump: %w", err)
		}

		conn, err := newConnection(id, p, service)
		if err != nil {
			return nil, fmt.Errorf("getOrCreateConnection: failed to create new connection: %w", err)
		}
		s.activeConnections[id] = conn
		return conn, nil
	}
}
