package xpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

const bodyVersion = uint32(0x00000005)

const (
	wrapperMagic = uint32(0x29b00b92)
	objectMagic  = uint32(0x42133742)
)

type xpcType uint32

// TODO: there are more types available and need to be added still when observed
const (
	nullType         = xpcType(0x00001000)
	boolType         = xpcType(0x00002000)
	int64Type        = xpcType(0x00003000)
	uint64Type       = xpcType(0x00004000)
	doubleType       = xpcType(0x00005000)
	dateType         = xpcType(0x00007000)
	dataType         = xpcType(0x00008000)
	stringType       = xpcType(0x00009000)
	uuidType         = xpcType(0x0000a000)
	arrayType        = xpcType(0x0000e000)
	dictionaryType   = xpcType(0x0000f000)
	fileTransferType = xpcType(0x0001a000)
)

const (
	AlwaysSetFlag        = uint32(0x00000001)
	DataFlag             = uint32(0x00000100)
	HeartbeatRequestFlag = uint32(0x00010000)
	HeartbeatReplyFlag   = uint32(0x00020000)
	FileOpenFlag         = uint32(0x00100000)
	InitHandshakeFlag    = uint32(0x00400000)
)

type wrapperHeader struct {
	Flags   uint32
	BodyLen uint64
	MsgId   uint64
}

type Message struct {
	Flags uint32
	Body  map[string]interface{}
	Id    uint64
}

func (m Message) IsFileOpen() bool {
	return m.Flags&FileOpenFlag > 0
}

type FileTransfer struct {
	MsgId        uint64
	TransferSize uint64
}

// DecodeMessage expects a full RemoteXPC message and decodes the message body into a map
func DecodeMessage(r io.Reader) (Message, error) {
	var magic uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return Message{}, fmt.Errorf("DecodeMessage: failed to read magic number: %w", err)
	}
	if magic != wrapperMagic {
		return Message{}, fmt.Errorf("DecodeMessage: wrong magic number 0x%x", magic)
	}
	wrapper, err := decodeWrapper(r)
	if err != nil {
		return Message{}, fmt.Errorf("DecodeMessage: failed to decode wrapper: %w", err)
	}
	return wrapper, nil
}

// EncodeMessage creates a RemoteXPC message encoded with the body and flags provided
func EncodeMessage(w io.Writer, message Message) error {
	if message.Body == nil {
		wrapper := struct {
			magic uint32
			h     wrapperHeader
		}{
			magic: wrapperMagic,
			h: wrapperHeader{
				Flags:   message.Flags,
				BodyLen: 0,
				MsgId:   message.Id,
			},
		}

		err := binary.Write(w, binary.LittleEndian, wrapper)
		if err != nil {
			return fmt.Errorf("EncodeMessage: failed to write empty message: %w", err)
		}
		return nil
	}
	buf := bytes.NewBuffer(nil)
	err := encodeDictionary(buf, message.Body)
	if err != nil {
		return fmt.Errorf("EncodeMessage: failed to encode dictionary: %w", err)
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
			Flags:   message.Flags,
			BodyLen: uint64(buf.Len() + 8),
			MsgId:   message.Id,
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
		return fmt.Errorf("EncodeMessage: failed to write xpc wrapper: %w", err)
	}

	_, err = io.Copy(w, buf)
	if err != nil {
		return fmt.Errorf("EncodeMessage: failed to write message body: %w", err)
	}
	return nil

}

func decodeWrapper(r io.Reader) (Message, error) {
	var h wrapperHeader
	err := binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return Message{}, fmt.Errorf("decodeWrapper: failed to decode header wrapper: %w", err)
	}
	if h.BodyLen == 0 {
		return Message{
			Flags: h.Flags,
		}, nil
	}
	body, err := decodeBody(r, h)
	if err != nil {
		return Message{}, fmt.Errorf("decodeWrapper: failed to decode body: %w", err)
	}
	return Message{
		Flags: h.Flags,
		Body:  body,
	}, nil
}

func decodeBody(r io.Reader, h wrapperHeader) (map[string]interface{}, error) {
	bodyHeader := struct {
		Magic   uint32
		Version uint32
	}{}
	if err := binary.Read(r, binary.LittleEndian, &bodyHeader); err != nil {
		return nil, fmt.Errorf("decodeBody: failed to decode header: %w", err)
	}
	if bodyHeader.Magic != objectMagic {
		return nil, fmt.Errorf("decodeBody: invalid object magic number 0x%x", bodyHeader.Magic)
	}
	if bodyHeader.Version != bodyVersion {
		return nil, fmt.Errorf("decodeBody: expected version 0x%x but got 0x%x", bodyVersion, bodyHeader.Version)
	}
	bodyPayloadLength := h.BodyLen - 8
	body := make([]byte, bodyPayloadLength)
	n, err := r.Read(body)
	if err != nil {
		return nil, fmt.Errorf("decodeBody:: failed to read body data: %w", err)
	}
	if uint64(n) != bodyPayloadLength {
		return nil, fmt.Errorf("decodeBody: could not read full body. only %d instead of %d were read", n, bodyPayloadLength)
	}
	bodyBuf := bytes.NewReader(body)
	res, err := decodeObject(bodyBuf)
	if err != nil {
		return nil, fmt.Errorf("decodeBody: failed to decode body: %w", err)
	}
	return res.(map[string]interface{}), nil
}

func decodeObject(r io.Reader) (interface{}, error) {
	var t xpcType
	err := binary.Read(r, binary.LittleEndian, &t)
	if err != nil {
		return nil, fmt.Errorf("decodeObject: could not read type: %w", err)
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
	case doubleType:
		return decodeDouble(r)
	case dateType:
		return decodeDate(r)
	case dataType:
		return decodeData(r)
	case stringType:
		return decodeString(r)
	case uuidType:
		return decodeUuid(r)
	case arrayType:
		return decodeArray(r)
	case dictionaryType:
		return decodeDictionary(r)
	case fileTransferType:
		return decodeFileTransfer(r)
	default:
		return nil, fmt.Errorf("decodeObject: can't handle unknown type 0x%08x", t)
	}
}

func decodeUuid(r io.Reader) (uuid.UUID, error) {
	b := make([]byte, 16)
	_, err := r.Read(b)
	if err != nil {
		return [16]byte{}, fmt.Errorf("decodeUuid: failed to read data: %w", err)
	}
	u, err := uuid.FromBytes(b)
	if err != nil {
		return [16]byte{}, fmt.Errorf("decodeUuid: failed to parse UUID: %w", err)
	}
	return u, nil
}

func decodeFileTransfer(r io.Reader) (FileTransfer, error) {
	header := struct {
		MsgId uint64 // always 1
	}{}
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return FileTransfer{}, fmt.Errorf("decodeFileTransfer: failed to read data: %w", err)
	}
	d, err := decodeObject(r)
	if err != nil {
		return FileTransfer{}, fmt.Errorf("decodeFileTransfer: failed to decode object: %w", err)
	}
	if dict, ok := d.(map[string]interface{}); ok {
		// the transfer length is always stored in a property 's'
		if transferLen, ok := dict["s"].(uint64); ok {
			return FileTransfer{
				MsgId:        header.MsgId,
				TransferSize: transferLen,
			}, nil
		} else {
			return FileTransfer{}, fmt.Errorf("decodeFileTransfer: expected uint64 for transfer length")
		}
	} else {
		return FileTransfer{}, fmt.Errorf("decodeFileTransfer: expected a dictionary but got %T", d)
	}
}

func decodeDictionary(r io.Reader) (map[string]interface{}, error) {
	var l, numEntries uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, fmt.Errorf("decodeDictionary: failed to read data: %w", err)
	}
	err = binary.Read(r, binary.LittleEndian, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("decodeDictionary: failed to read number of entries: %w", err)
	}
	dict := make(map[string]interface{})
	for i := uint32(0); i < numEntries; i++ {
		key, err := readDictionaryKey(r)
		if err != nil {
			return nil, fmt.Errorf("decodeDictionary: failed to read dictionary key: %w", err)
		}
		dict[key], err = decodeObject(r)
		if err != nil {
			return nil, fmt.Errorf("decodeDictionary: failed to decode object for key '%s': %w", key, err)
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
			return "", fmt.Errorf("readDictionaryKey: failed to read character: %w", err)
		}
		if buf[0] == 0 {
			s := b.String()
			toSkip := calcPadding(len(s) + 1)
			_, err := io.CopyN(io.Discard, r, toSkip)
			if err != nil {
				return "", fmt.Errorf("readDictionaryKey: failed to discard padding: %w", err)
			}
			return s, nil
		}
		b.Write(buf)
	}
}

func decodeArray(r io.Reader) ([]interface{}, error) {
	var l, numEntries uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, fmt.Errorf("decodeArray: failed to read payload length: %w", err)
	}
	err = binary.Read(r, binary.LittleEndian, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("decodeArray: failed to read number of entries: %w", err)
	}
	arr := make([]interface{}, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		arr[i], err = decodeObject(r)
		if err != nil {
			return nil, fmt.Errorf("decodeArray: failed to decode object at index %d: %w", i, err)
		}
	}
	return arr, nil
}

func decodeString(r io.Reader) (string, error) {
	var l uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return "", fmt.Errorf("decodeString: failed to read string length: %w", err)
	}
	s := make([]byte, l)
	_, err = r.Read(s)
	if err != nil {
		return "", fmt.Errorf("decodeString: failed to read string: %w", err)
	}
	res := strings.Trim(string(s), "\000")
	toSkip := calcPadding(int(l))
	_, err = io.CopyN(io.Discard, r, toSkip)
	if err != nil {
		return "", fmt.Errorf("decodeString: faile to skip padding bytes: %w", err)
	}
	return res, nil
}

func decodeData(r io.Reader) ([]byte, error) {
	var l uint32
	err := binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, fmt.Errorf("decodeData: failed to read payload length: %w", err)
	}
	b := make([]byte, l)
	_, err = r.Read(b)
	if err != nil {
		return nil, fmt.Errorf("decodeData: failed to read payload: %w", err)
	}
	toSkip := calcPadding(int(l))
	_, _ = io.CopyN(io.Discard, r, toSkip)
	return b, nil
}

func decodeDouble(r io.Reader) (interface{}, error) {
	var d float64
	err := binary.Read(r, binary.LittleEndian, &d)
	if err != nil {
		return 0, fmt.Errorf("decodeDouble: failed to read data: %w", err)
	}
	return d, nil
}

func decodeUint64(r io.Reader) (uint64, error) {
	var i uint64
	err := binary.Read(r, binary.LittleEndian, &i)
	if err != nil {
		return 0, fmt.Errorf("decodeUint64: failed to read data: %w", err)
	}
	return i, nil
}

func decodeInt64(r io.Reader) (int64, error) {
	var i int64
	err := binary.Read(r, binary.LittleEndian, &i)
	if err != nil {
		return 0, fmt.Errorf("decodeInt64: failed to read data: %w", err)
	}
	return i, nil
}

func decodeBool(r io.Reader) (bool, error) {
	var b bool
	err := binary.Read(r, binary.LittleEndian, &b)
	if err != nil {
		return false, fmt.Errorf("decodeBool: failed to read data: %w", err)
	}
	_, _ = io.CopyN(io.Discard, r, 3)
	return b, nil
}

func decodeDate(r io.Reader) (time.Time, error) {
	var i int64
	err := binary.Read(r, binary.LittleEndian, &i)
	if err != nil {
		return time.Time{}, fmt.Errorf("decodeDate: failed to read date payload: %w", err)
	}
	t := time.Unix(0, i)
	return t, nil
}

func calcPadding(l int) int64 {
	c := int(math.Ceil(float64(l) / 4.0))
	return int64(c*4 - l)
}

func encodeDictionary(w io.Writer, v map[string]interface{}) error {
	buf := bytes.NewBuffer(nil)

	err := binary.Write(buf, binary.LittleEndian, uint32(len(v)))
	if err != nil {
		return fmt.Errorf("encodeDictionary: failed to write number of dictionary entries: %w", err)
	}

	for k, e := range v {
		err := encodeDictionaryKey(buf, k)
		if err != nil {
			return fmt.Errorf("encodeDictionary: failed to encode dictionary key '%s': %w", k, err)
		}
		err = encodeObject(buf, e)
		if err != nil {
			return fmt.Errorf("encodeDictionary: failed to encode object: %w", err)
		}
	}

	err = binary.Write(w, binary.LittleEndian, dictionaryType)
	if err != nil {
		return fmt.Errorf("encodeDictionary: failed to write dictionary type: %w", err)
	}
	err = binary.Write(w, binary.LittleEndian, uint32(buf.Len()))
	if err != nil {
		return fmt.Errorf("encodeDictionary: failed to write payload length: %w", err)
	}
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("encodeDictionary: failed to write ")
	}
	return nil
}

func encodeObject(w io.Writer, e interface{}) error {
	if e == nil {
		if err := binary.Write(w, binary.LittleEndian, nullType); err != nil {
			return fmt.Errorf("encodeObject: failed to encode null objecdt: %w", err)
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
	case float64:
		if err := encodeDouble(w, e.(float64)); err != nil {
			return err
		}
	case string:
		if err := encodeString(w, e.(string)); err != nil {
			return err
		}
	case uuid.UUID:
		if err := encodeUuid(w, e.(uuid.UUID)); err != nil {
			return err
		}
	case time.Time:
		if err := encodeDate(w, e.(time.Time)); err != nil {
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

func encodeUuid(w io.Writer, u uuid.UUID) error {
	out := struct {
		t xpcType
		u uuid.UUID
	}{uuidType, u}
	err := binary.Write(w, binary.LittleEndian, out)
	if err != nil {
		return fmt.Errorf("encodeUuid: failed to write UUID payload: %w", err)
	}
	return nil
}

func encodeArray(w io.Writer, slice []interface{}) error {
	buf := bytes.NewBuffer(nil)
	for i, e := range slice {
		if err := encodeObject(buf, e); err != nil {
			return fmt.Errorf("encodeArray: failed to encode array object at index %d: %w", i, err)
		}
	}

	header := struct {
		t          xpcType
		l          uint32
		numObjects uint32
	}{arrayType, uint32(buf.Len()), uint32(len(slice))}
	if err := binary.Write(w, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("encodeArray: failed to write array header: %w", err)
	}
	if _, err := io.Copy(w, buf); err != nil {
		return fmt.Errorf("encodeArray: failed to copy array payload: %w", err)
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
		return fmt.Errorf("encodeString: failed to write header: %w", err)
	}

	toPad := calcPadding(int(header.l))
	padded := make([]byte, len(s)+int(toPad)+1)
	copy(padded, s)
	_, err = w.Write(padded)
	if err != nil {
		return fmt.Errorf("encodeString: failed to write string payload: %w", err)
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
		return fmt.Errorf("encodeData: failed to write data length: %w", err)
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("encodeData: failed to write data: %w", err)
	}
	toPad := calcPadding(int(header.l))
	_, err = w.Write(make([]byte, toPad))
	if err != nil {
		return fmt.Errorf("encodeData: failed to write padding: %w", err)
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
		return fmt.Errorf("encodeUint64: failed to write data: %w", err)
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
		return fmt.Errorf("encodeInt64: failed to write data: %w", err)
	}
	return nil
}

func encodeDouble(w io.Writer, d float64) error {
	out := struct {
		t xpcType
		d float64
	}{doubleType, d}
	err := binary.Write(w, binary.LittleEndian, out)
	if err != nil {
		return fmt.Errorf("encodeDouble: failed to write data: %w", err)
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
		return fmt.Errorf("encodeBool: failed to write data: %w", err)
	}
	return nil
}

func encodeDate(w io.Writer, t time.Time) error {
	out := struct {
		t xpcType
		i int64
	}{dateType, t.UnixNano()}
	err := binary.Write(w, binary.LittleEndian, out)
	if err != nil {
		return fmt.Errorf("encodeDate: failed to write data: %w", err)
	}
	return nil
}

func encodeDictionaryKey(w io.Writer, k string) error {
	strLen := len(k) + 1
	toPad := calcPadding(strLen)
	content := make([]byte, strLen+int(toPad))
	copy(content, k)
	_, err := w.Write(content)
	if err != nil {
		return fmt.Errorf("encodeDictionaryKey: failed to write data: %w", err)
	}
	return nil
}
