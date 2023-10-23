package xpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
)

const bodyVersion = uint32(0x00000005)

const (
	wrapperMagic = uint32(0x29b00b92)
	objectMagic  = uint32(0x42133742)
)

type xpcType uint32

// TODO: there are more types available and need to be added still when observed
const (
	nullType       = xpcType(0x00001000)
	boolType       = xpcType(0x00002000)
	int64Type      = xpcType(0x00003000)
	uint64Type     = xpcType(0x00004000)
	dataType       = xpcType(0x00008000)
	stringType     = xpcType(0x00009000)
	arrayType      = xpcType(0x0000e000)
	dictionaryType = xpcType(0x0000f000)
)

const (
	alwaysSetFlag        = uint32(0x00000001)
	dataFlag             = uint32(0x00000100)
	heartbeatRequestFlag = uint32(0x00010000)
	heartbeatReplyFlag   = uint32(0x00020000)
)

type wrapperHeader struct {
	Flags   uint32
	BodyLen uint64
	MsgId   uint64
}

type Message struct {
	Flags uint32
	Body  map[string]interface{}
}

// DecodeMessage expects a full RemoteXPC message and decodes the message body into a map
func DecodeMessage(r io.Reader) (Message, error) {
	var magic uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return Message{}, err
	}
	if magic != wrapperMagic {
		return Message{}, fmt.Errorf("wrong magic number")
	}
	wrapper, err := decodeWrapper(r)
	return wrapper, err
}

// EncodeData creates a RemoteXPC message with the data flag set, if data is present (an empty dictionary is considered
// to be no data)
func EncodeData(w io.Writer, body map[string]interface{}) error {
	if body == nil {
		return encodeMessageWithoutBody(w)
	}
	buf := bytes.NewBuffer(nil)
	err := encodeDictionary(buf, body)
	if err != nil {
		return err
	}

	flags := alwaysSetFlag
	if len(body) > 0 {
		flags |= dataFlag
	}

	wrapper := struct {
		magic uint32
		h     wrapperHeader
		body  struct {
			magic   uint32
			version uint32
		}
	}{
		magic: wrapperMagic,
		h: wrapperHeader{
			Flags:   flags,
			BodyLen: uint64(buf.Len() + 8),
			MsgId:   0,
		},
		body: struct {
			magic   uint32
			version uint32
		}{
			magic:   objectMagic,
			version: bodyVersion,
		},
	}

	err = binary.Write(w, binary.LittleEndian, wrapper)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, buf)
	return err
}

func decodeWrapper(r io.Reader) (Message, error) {
	var h wrapperHeader
	err := binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return Message{}, err
	}
	if h.BodyLen == 0 {
		return Message{
			Flags: h.Flags,
		}, nil
	}
	body, err := decodeBody(r, h)
	return Message{
		Flags: h.Flags,
		Body:  body,
	}, err
}

func decodeBody(r io.Reader, h wrapperHeader) (map[string]interface{}, error) {
	bodyHeader := struct {
		Magic   uint32
		Version uint32
	}{}
	if err := binary.Read(r, binary.LittleEndian, &bodyHeader); err != nil {
		return nil, err
	}
	if bodyHeader.Magic != objectMagic {
		return nil, fmt.Errorf("cant decode")
	}
	if bodyHeader.Version != bodyVersion {
		return nil, fmt.Errorf("expected version 0x%x but got 0x%x", bodyVersion, bodyHeader.Version)
	}
	body := make([]byte, h.BodyLen-8)
	n, err := r.Read(body)
	if err != nil {
		return nil, err
	}
	if uint64(n) != (h.BodyLen - 8) {
		return nil, fmt.Errorf("not enough data")
	}
	bodyBuf := bytes.NewReader(body)
	res, err := decodeObject(bodyBuf)
	if err != nil {
		return nil, err
	}
	return res.(map[string]interface{}), nil
}

func decodeObject(r io.Reader) (interface{}, error) {
	var t xpcType
	err := binary.Read(r, binary.LittleEndian, &t)
	if err != nil {
		return nil, err
	}
	switch t {
	case nullType:
		return nil, nil
	case boolType:
		return decodeBool(r)
	case int64Type:
		return decodeInt64(r)
	case uint64Type:
		return decodeUint64(r)
	case dataType:
		return decodeData(r)
	case stringType:
		return decodeString(r)
	case arrayType:
		return decodeArray(r)
	case dictionaryType:
		return decodeDictionary(r)
	default:
		return nil, fmt.Errorf("can't handle unknown type 0x%08x", t)
	}
}

func decodeDictionary(r io.Reader) (map[string]interface{}, error) {
	var l, numEntries uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &numEntries)
	if err != nil {
		return nil, err
	}
	dict := make(map[string]interface{})
	for i := uint32(0); i < numEntries; i++ {
		key, err := readDictionaryKey(r)
		if err != nil {
			return nil, err
		}
		dict[key], err = decodeObject(r)
		if err != nil {
			return nil, err
		}
	}
	return dict, nil
}

func readDictionaryKey(r io.Reader) (string, error) {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		_, err := r.Read(buf)
		if err != nil {
			return "", err
		}
		if buf[0] == 0 {
			s := b.String()
			toSkip := calcPadding(len(s) + 1)
			_, err := io.CopyN(io.Discard, r, toSkip)
			return s, err
		}
		b.Write(buf)
	}
}

func decodeArray(r io.Reader) ([]interface{}, error) {
	var l, numEntries uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &numEntries)
	if err != nil {
		return nil, err
	}
	arr := make([]interface{}, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		arr[i], err = decodeObject(r)
		if err != nil {
			return nil, err
		}
	}
	return arr, nil
}

func decodeString(r io.Reader) (string, error) {
	var l uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return "", err
	}
	s := make([]byte, l)
	_, err = r.Read(s)
	if err != nil {
		return "", err
	}
	res := strings.Trim(string(s), "\000")
	toSkip := calcPadding(int(l))
	_, _ = io.CopyN(io.Discard, r, toSkip)
	return res, nil
}

func decodeData(r io.Reader) ([]byte, error) {
	var l uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, err
	}
	b := make([]byte, l)
	_, err = r.Read(b)
	if err != nil {
		return nil, err
	}
	toSkip := calcPadding(int(l))
	_, _ = io.CopyN(io.Discard, r, toSkip)
	return b, nil
}

func decodeUint64(r io.Reader) (uint64, error) {
	var i uint64
	err := binary.Read(r, binary.LittleEndian, &i)
	return i, err
}

func decodeInt64(r io.Reader) (int64, error) {
	var i int64
	err := binary.Read(r, binary.LittleEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func decodeBool(r io.Reader) (bool, error) {
	var b bool
	err := binary.Read(r, binary.LittleEndian, &b)
	if err != nil {
		return false, err
	}
	_, _ = io.CopyN(io.Discard, r, 3)
	return b, nil
}

func calcPadding(l int) int64 {
	c := int(math.Ceil(float64(l) / 4.0))
	return int64(c*4 - l)
}

func encodeDictionary(w io.Writer, v map[string]interface{}) error {
	buf := bytes.NewBuffer(nil)

	for k, e := range v {
		err := encodeDictionaryKey(buf, k)
		if err != nil {
			return err
		}
		err2 := encodeObject(buf, e)
		if err2 != nil {
			return err2
		}
	}

	err := binary.Write(w, binary.LittleEndian, dictionaryType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, uint32(buf.Len()))
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, uint32(len(v)))
	if err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	return err
}

func encodeObject(w io.Writer, e interface{}) error {
	if e == nil {
		if err := binary.Write(w, binary.LittleEndian, nullType); err != nil {
			return err
		}
		return nil
	}
	if v := reflect.ValueOf(e); v.Kind() == reflect.Slice {
		if b, ok := e.([]byte); ok {
			return encodeData(w, b)
		}
		r := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			r[i] = v.Index(i).Interface()
		}
		if err := encodeArray(w, r); err != nil {
			return err
		}
		return nil
	}
	switch t := e.(type) {
	case bool:
		if err := encodeBool(w, e.(bool)); err != nil {
			return err
		}
	case int64:
		if err := encodeInt64(w, e.(int64)); err != nil {
			return err
		}
	case uint64:
		if err := encodeUint64(w, e.(uint64)); err != nil {
			return err
		}
	case string:
		if err := encodeString(w, e.(string)); err != nil {
			return err
		}
	case map[string]interface{}:
		if err := encodeDictionary(w, e.(map[string]interface{})); err != nil {
			return err
		}
	default:
		return fmt.Errorf("can not encode type %v", t)
	}
	return nil
}

func encodeArray(w io.Writer, slice []interface{}) error {
	buf := bytes.NewBuffer(nil)
	for _, e := range slice {
		if err := encodeObject(buf, e); err != nil {
			return err
		}
	}

	header := struct {
		t          xpcType
		l          uint32
		numObjects uint32
	}{arrayType, uint32(buf.Len()), uint32(len(slice))}
	if err := binary.Write(w, binary.LittleEndian, header); err != nil {
		return err
	}
	if _, err := io.Copy(w, buf); err != nil {
		return err
	}
	return nil
}

func encodeString(w io.Writer, s string) error {
	header := struct {
		t xpcType
		l uint32
	}{stringType, uint32(len(s) + 1)}
	err := binary.Write(w, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}
	toPad := calcPadding(int(header.l))
	_, err = w.Write(make([]byte, toPad+1))
	if err != nil {
		return err
	}
	return nil
}

func encodeData(w io.Writer, b []byte) error {
	header := struct {
		t xpcType
		l uint32
	}{dataType, uint32(len(b))}
	err := binary.Write(w, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	toPad := calcPadding(int(header.l))
	_, err = w.Write(make([]byte, toPad))
	if err != nil {
		return err
	}
	return nil
}

func encodeUint64(w io.Writer, i uint64) error {
	out := struct {
		t xpcType
		i uint64
	}{uint64Type, i}
	err := binary.Write(w, binary.LittleEndian, out)
	if err != nil {
		return err
	}
	return nil
}

func encodeInt64(w io.Writer, i int64) error {
	out := struct {
		t xpcType
		i int64
	}{int64Type, i}
	err := binary.Write(w, binary.LittleEndian, out)
	if err != nil {
		return err
	}
	return nil
}

func encodeBool(w io.Writer, b bool) error {
	out := struct {
		t   xpcType
		b   bool
		pad [3]byte
	}{
		t: boolType,
		b: b,
	}
	err := binary.Write(w, binary.LittleEndian, out)
	if err != nil {
		return err
	}
	return nil
}

func encodeDictionaryKey(w io.Writer, k string) error {
	toPad := calcPadding(len(k) + 1)
	_, err := w.Write(append([]byte(k), 0x0))
	if err != nil {
		return err
	}
	pad := make([]byte, toPad)
	_, err = w.Write(pad)
	return err
}

func encodeMessageWithoutBody(w io.Writer) error {
	wrapper := struct {
		magic uint32
		h     wrapperHeader
	}{
		magic: wrapperMagic,
		h: wrapperHeader{
			Flags:   alwaysSetFlag,
			BodyLen: 0,
			MsgId:   0,
		},
	}
	err := binary.Write(w, binary.LittleEndian, wrapper)
	return err
}
