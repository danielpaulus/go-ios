package sniffer

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	http22 "golang.org/x/net/http2"
	"io"
)

type contentType int

const (
	http2 = contentType(iota)
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
		log.Infof("decoded frame %s", f.Header().Type)
	}
	return nil
}
