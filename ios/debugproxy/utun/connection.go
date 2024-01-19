//go:build darwin

package utun

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sirupsen/logrus"
)

type connection struct {
	id      connectionId
	w       payloadWriter
	outPath string
	inPath  string
	service string
}

func newConnection(id connectionId, p string, service string) *connection {
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
		service: service,
	}
}

func (c connection) handlePacket(p gopacket.Packet, ip *layers.IPv6, tcp *layers.TCP) {
	if tcp.SYN && tcp.SrcPort == c.id.localPort {
		logrus.Infof("new connection %s", c.id)
	}
	if len(tcp.Payload) > 0 {
		c.w.Write(c.direction(tcp), tcp.Payload)
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
	logrus.WithField("connection", c.id.String()).WithField("service", c.service).Info("closing connection")
	err := parseConnectionData(c.outPath, c.inPath)
	if err != nil {
		logrus.WithField("connection", c.id.String()).
			WithField("service", c.service).
			WithError(err).
			Warn("failed parsing data")
	}
	return nil
}

func (c connectionId) String() string {
	return fmt.Sprintf("%d-%d", c.localPort, c.remotePort)
}

func parseConnectionData(outgoing string, incoming string) error {
	dir := path.Dir(outgoing)

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
		_ = createDecodingFiles(dir, "http.frames", func(outgoing, incoming pair) error {
			outErr := decodeHttp2FrameHeaders(outgoing.w, outFile, true)
			inErr := decodeHttp2FrameHeaders(incoming.w, inFile, false)
			return errors.Join(outErr, inErr)
		})
		_, _ = outFile.Seek(0, io.SeekStart)
		_, _ = inFile.Seek(0, io.SeekStart)
		return createDecodingFiles(dir, "http.bin", func(outgoing, incoming pair) error {
			outErr := decodeHttp2(outgoing.w, outFile, true)
			inErr := decodeHttp2(incoming.w, inFile, false)
			if err := errors.Join(outErr, inErr); err != nil {
				//return err
			}
			return parseConnectionData(outgoing.p, incoming.p)
		})
	case remoteXpc:
		return createDecodingFiles(dir, "xpc.jsonl", func(outgoing, incoming pair) error {
			outErr := decodeRemoteXpc(outgoing.w, outFile)
			inErr := decodeRemoteXpc(incoming.w, inFile)
			return errors.Join(outErr, inErr)
		})
	case remoteDtx:
		return createDecodingFiles(dir, "dtx", func(outgoing, incoming pair) error {
			outErr := decodeRemoteDtx(outgoing.w, outFile)
			inErr := decodeRemoteDtx(incoming.w, inFile)
			return errors.Join(outErr, inErr)
		})
	default:
		stat, _ := os.Stat(outgoing)
		if stat.Size() == 0 {
			return nil
		}
		return fmt.Errorf("unknown content type: %s/%s", outgoing, incoming)
	}
}

func createDecodingFiles(dir, suffix string, consumer func(outgoing, incoming pair) error) error {
	outPath := path.Join(dir, fmt.Sprintf("outgoing.%s", suffix))
	inPath := path.Join(dir, fmt.Sprintf("incoming.%s", suffix))

	outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer outFile.Close()
	inFile, err := os.OpenFile(inPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer inFile.Close()

	return consumer(pair{outPath, outFile}, pair{inPath, inFile})
}

type pair struct {
	p string
	w io.Writer
}
