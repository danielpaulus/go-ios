package opack

import (
	"bytes"
	"fmt"
	"io"
)

func Encode(m map[string]interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := encodeDict(buf, m)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeDict(w io.Writer, d map[string]interface{}) error {
	l := len(d)
	if l > 0xF {
		return fmt.Errorf("%d exceeds max size of 0xF", l)
	}
	b := 0xE0 | uint8(l)
	_, err := w.Write([]byte{b})
	if err != nil {
		return err
	}
	for k, e := range d {
		err := encodeString(w, k)
		if err != nil {
			return err
		}
		switch t := e.(type) {
		case string:
			err := encodeString(w, e.(string))
			if err != nil {
				return err
			}
		case []byte:
			err := encodeData(w, e.([]byte))
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("can't encode type %s", t)
		}
	}
	return nil
}

func encodeString(w io.Writer, s string) error {
	err := writeLengthBasedIdentifier(w, 0x40, len(s))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

func encodeData(w io.Writer, b []byte) error {
	err := writeLengthBasedIdentifier(w, 0x70, len(b))
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func createIdentifierWithLength(t byte, l int) ([]byte, error) {
	if l <= 0xF {
		return []byte{t | byte(l)}, nil
	} else if l < 0x20 {
		inc := t + (1 << 4)
		return []byte{inc | (byte(l) & 0xF)}, nil
	} else if l <= 0xFF {
		inc := (t + (2 << 4)) | 0x1
		return []byte{inc, byte(l)}, nil
	} else {
		return nil, fmt.Errorf("string too long: %d", l)
	}
}

func writeLengthBasedIdentifier(w io.Writer, t byte, l int) error {
	id, err := createIdentifierWithLength(t, l)
	if err != nil {
		return err
	}
	_, err = w.Write(id)
	return err
}
