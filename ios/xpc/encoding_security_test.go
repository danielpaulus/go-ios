package xpc_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Wire constants mirrored from the (unexported) encoding.go, so these external
// regression tests can hand-craft malformed RemoteXPC messages.
const (
	wrapperMagic = uint32(0x29b00b92)
	objectMagic  = uint32(0x42133742)
	bodyVersion  = uint32(0x00000005)

	int64Type      = uint32(0x00003000)
	stringType     = uint32(0x00009000)
	arrayType      = uint32(0x0000e000)
	dictionaryType = uint32(0x0000f000)

	alwaysSetFlag = uint32(0x00000001)
)

// buildMessage frames a raw XPC body (the bytes following the body header, i.e.
// the top-level object) into a complete RemoteXPC message with BodyLen set to
// the real payload size (8 + len(bodyObject)).
func buildMessage(bodyObject []byte) []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.LittleEndian, wrapperMagic)
	_ = binary.Write(buf, binary.LittleEndian, alwaysSetFlag)             // Flags
	_ = binary.Write(buf, binary.LittleEndian, uint64(len(bodyObject)+8)) // BodyLen
	_ = binary.Write(buf, binary.LittleEndian, uint64(0))                 // MsgId
	_ = binary.Write(buf, binary.LittleEndian, objectMagic)               // body magic
	_ = binary.Write(buf, binary.LittleEndian, bodyVersion)               // body version
	buf.Write(bodyObject)
	return buf.Bytes()
}

// HIGH-7: BodyLen in [1..7] underflows h.BodyLen-8 to a huge uint64 and today
// panics in make([]byte, ...). With the fix it must return a clean error.
func TestDecodeMessageBodyLenUnderflow(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.LittleEndian, wrapperMagic)
	_ = binary.Write(buf, binary.LittleEndian, alwaysSetFlag) // Flags
	_ = binary.Write(buf, binary.LittleEndian, uint64(4))     // BodyLen = 4 (< 8) -> BodyLen-8 underflows
	_ = binary.Write(buf, binary.LittleEndian, uint64(0))     // MsgId
	// A valid body header must decode first so control reaches the
	// bodyPayloadLength := h.BodyLen - 8 subtraction (the underflow site).
	_ = binary.Write(buf, binary.LittleEndian, objectMagic) // body magic
	_ = binary.Write(buf, binary.LittleEndian, bodyVersion) // body version

	var err error
	require.NotPanics(t, func() {
		_, err = xpc.DecodeMessage(bytes.NewReader(buf.Bytes()))
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "body length")
}

// HIGH-8: a non-dictionary top-level object (here an int64) today panics on the
// bare type assertion res.(map[string]interface{}). With the fix: clean error.
func TestDecodeMessageNonDictTopLevel(t *testing.T) {
	body := &bytes.Buffer{}
	_ = binary.Write(body, binary.LittleEndian, int64Type)
	_ = binary.Write(body, binary.LittleEndian, int64(42))

	msg := buildMessage(body.Bytes())

	require.NotPanics(t, func() {
		_, err := xpc.DecodeMessage(bytes.NewReader(msg))
		assert.Error(t, err)
	})
}

// MED-1 (string): an oversized wire string length must be rejected against the
// bytes actually present, not fed straight into make([]byte, l).
func TestDecodeMessageOversizedStringLength(t *testing.T) {
	body := &bytes.Buffer{}
	_ = binary.Write(body, binary.LittleEndian, stringType)
	_ = binary.Write(body, binary.LittleEndian, uint32(0xFFFFFFF0)) // ~4 GiB claimed
	body.Write([]byte{0x41, 0x00})                                  // only 2 bytes actually present

	msg := buildMessage(body.Bytes())

	require.NotPanics(t, func() {
		_, err := xpc.DecodeMessage(bytes.NewReader(msg))
		assert.Error(t, err)
	})
}

// MED-1 (array): an oversized entry count must be rejected against the bytes
// remaining before make([]interface{}, numEntries) allocates gigabytes. The
// guard error is asserted specifically so the test fails if the bound is
// removed (without the bound, decoding instead attempts the huge allocation and
// then errors much later on EOF — a different message and a resource hazard).
func TestDecodeMessageOversizedArrayCount(t *testing.T) {
	body := &bytes.Buffer{}
	_ = binary.Write(body, binary.LittleEndian, arrayType)
	_ = binary.Write(body, binary.LittleEndian, uint32(0))          // payload length (unused for bound)
	_ = binary.Write(body, binary.LittleEndian, uint32(0x7FFFFFFF)) // ~2 billion entries claimed
	// no actual entry bytes follow

	var err error
	require.NotPanics(t, func() {
		_, err = xpc.DecodeMessage(bytes.NewReader(buildMessage(body.Bytes())))
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decodeArray: entry count")
	assert.Contains(t, err.Error(), "bytes remaining")
}

// MED-1 (dictionary): an oversized entry count must be rejected against the
// bytes remaining before the decode loop is entered. The guard error is
// asserted specifically so the test fails if the bound is removed.
func TestDecodeMessageOversizedDictionaryCount(t *testing.T) {
	body := &bytes.Buffer{}
	_ = binary.Write(body, binary.LittleEndian, dictionaryType)
	_ = binary.Write(body, binary.LittleEndian, uint32(0))          // payload length (unused for bound)
	_ = binary.Write(body, binary.LittleEndian, uint32(0x7FFFFFFF)) // ~2 billion entries claimed
	// no actual entry bytes follow

	var err error
	require.NotPanics(t, func() {
		_, err = xpc.DecodeMessage(bytes.NewReader(buildMessage(body.Bytes())))
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decodeDictionary: entry count")
	assert.Contains(t, err.Error(), "bytes remaining")
}
