package sniffer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/danielpaulus/go-ios/ios/xpc"
	log "github.com/sirupsen/logrus"
	http22 "golang.org/x/net/http2"
	"io"
)

type contentType int

const (
	http2 = contentType(iota)
	remoteXpc
	unknown
)

func parse(w io.Writer, r io.ReadSeeker) error {
	t := detectType(r)
	switch t {
	case http2:
		log.Infof("decoding http2 data")
	default:
		return fmt.Errorf("could not decode")
	}
	return nil
}

func detectType(r io.ReadSeeker) contentType {
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return unknown
	}
	defer func() {
		r.Seek(offset, io.SeekStart)
	}()
	b := make([]byte, 4)
	_, err = r.Read(b)
	if err != nil {
		return unknown
	}
	if string(b) == "PRI " {
		return http2
	}
	i := binary.LittleEndian.Uint32(b)
	if i == 0x29b00b92 {
		return remoteXpc
	}

	return unknown
}

func decodeHttp2(w io.Writer, r io.Reader, needSkip bool) error {
	if needSkip {
		_, err := io.CopyN(io.Discard, r, 24)
		if err != nil {
			return err
		}
	}
	framer := http22.NewFramer(io.Discard, r)
	for {
		f, err := framer.ReadFrame()
		if err != nil {
			break
		}
		if f.Header().Type == http22.FrameData {
			dataFrame := f.(*http22.DataFrame)
			if _, err := w.Write(dataFrame.Data()); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeRemoteXpc(w io.Writer, r io.Reader) error {
	for {
		m, err := xpc.DecodeMessage(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer(nil)
		json.Compact(buf, b)
		if _, err := io.Copy(w, buf); err != nil {
			return err
		}
	}
}
