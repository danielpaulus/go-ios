//go:build darwin

package utun

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/xpc"
	log "github.com/sirupsen/logrus"
	http22 "golang.org/x/net/http2"
)

type contentType int

const (
	http2 = contentType(iota)
	remoteXpc
	remoteDtx
	unknown
)

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
	if string(b[:3]) == "y[=" {
		return remoteDtx
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
			return err
		}
		if f.Header().Type == http22.FrameData {
			dataFrame := f.(*http22.DataFrame)
			if _, err := w.Write(dataFrame.Data()); err != nil {
				return err
			}
		}
	}
}

func decodeHttp2FrameHeaders(w io.Writer, r io.Reader, needSkip bool) error {
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
			return err
		}
		_, err = w.Write(append([]byte(f.Header().String()), '\n'))
		if err != nil {
			return err
		}
	}
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
		if m.IsFileOpen() {
			log.Info("file transfer started, skipping remaining data ")
			return nil
		}
	}
}

func decodeRemoteDtx(w io.Writer, r io.Reader) error {
	for {
		m, err := dtx.ReadMessage(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		buf := bytes.NewBufferString(m.StringDebug() + "\n")
		if _, err := io.Copy(w, buf); err != nil {
			return err
		}
	}
}
