package sniffer

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
)

type connection struct {
	id      connectionId
	w       payloadWriter
	outPath string
	inPath  string
}

func newConnection(id connectionId, p string) *connection {
	inPath := path.Join(p, "incoming")
	incoming, err := os.OpenFile(inPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	outPath := path.Join(p, "outgoing")
	outgoing, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	pw := payloadWriter{
		incoming: incoming,
		outgoing: outgoing,
	}
	return &connection{
		id:      id,
		w:       pw,
		outPath: outPath,
		inPath:  inPath,
	}
}

func (c connection) handlePacket(p gopacket.Packet, ip *layers.IPv6, tcp *layers.TCP) {
	if tcp.SYN && tcp.SrcPort == c.id.localPort {
		logrus.Infof("new connection %s", c.id)
	}
	if len(tcp.Payload) > 0 {
		c.w.Write(c.direction(tcp), tcp.Payload)
	}
	if tcp.RST || tcp.FIN {
		c.Close()
	}
}

func (c connection) direction(tcp *layers.TCP) direction {
	if c.id.localPort == tcp.SrcPort {
		return outgoing
	} else {
		return incoming
	}
}

func (c connection) Close() error {
	_ = c.w.Close()

	parseConnectionData(c.outPath, c.inPath)
	return nil
}

func (c connectionId) String() string {
	return fmt.Sprintf("%d-%d", c.localPort, c.remotePort)
}

func parseConnectionData(outgoing string, incoming string) error {
	//dir := path.Dir(outgoing)

	outFile, err := os.OpenFile(outgoing, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer outFile.Close()
	inFile, err := os.OpenFile(incoming, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer inFile.Close()

	t := detectType(outFile)

	switch t {
	case http2:
		decodeHttp2(io.Discard, outFile, true)
		decodeHttp2(io.Discard, inFile, true)
	default:
		return fmt.Errorf("unknown content type")
	}
	return nil
}
