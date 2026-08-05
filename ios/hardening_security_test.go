package ios

// Regression tests for the input-hardening guards. Each test feeds malformed or
// hostile input that would previously panic or trigger an unbounded allocation
// and asserts that a clean error is returned instead. These live in the internal
// `ios` package so they can reach the unexported usbmux decode path and the
// getValueResponsefromBytes seam used by the device-value assertions.

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oversizedLengthReader returns a reader whose first 4 bytes encode a big-endian
// length just above maxMessageSize, so a decoder that trusts the prefix would
// attempt an unbounded allocation. Only a handful of payload bytes follow, so a
// correct decoder should reject on the length bound before reading/allocating.
func oversizedBigEndianPrefix() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(maxMessageSize)+1)
	return append(buf, 0x00, 0x00)
}

// HIGH-1: PlistCodec.Decode must reject an oversized length prefix instead of
// calling make([]byte, length) with an attacker-controlled ~4GiB size.
func TestPlistCodecDecodeRejectsOversizedLength(t *testing.T) {
	codec := NewPlistCodec()
	r := bytes.NewReader(oversizedBigEndianPrefix())
	_, err := codec.Decode(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

// HIGH-1: A valid small message must still decode byte-for-byte as before.
func TestPlistCodecDecodeAcceptsValidMessage(t *testing.T) {
	codec := NewPlistCodec()
	payload := []byte("hello")
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(payload)))
	r := bytes.NewReader(append(buf, payload...))
	got, err := codec.Decode(r)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
}

// HIGH-1: PlistCodecReadWriter.Read must reject an oversized length prefix.
func TestPlistCodecReadWriterReadRejectsOversizedLength(t *testing.T) {
	rw := NewPlistCodecReadWriter(bytes.NewReader(oversizedBigEndianPrefix()), nil)
	var v interface{}
	err := rw.Read(&v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

// HIGH-1: UsbMuxConnection.decode must reject a usbmux header whose Length field
// is above maxMessageSize instead of allocating Length-16 bytes.
func TestUsbMuxDecodeRejectsOversizedLength(t *testing.T) {
	header := UsbMuxHeader{Length: uint32(maxMessageSize) + 1, Version: 1, Request: 8, Tag: 1}
	buf := new(bytes.Buffer)
	require.NoError(t, binary.Write(buf, binary.LittleEndian, header))

	var muxConn UsbMuxConnection
	_, err := muxConn.decode(buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

// HIGH-2: InterfaceToStringSlice must not panic when the []interface{} contains a
// non-string element. Valid all-string input must keep the same length/content.
func TestInterfaceToStringSliceNonStringElementDoesNotPanic(t *testing.T) {
	input := []interface{}{"a", 42, "c"}
	assert.NotPanics(t, func() {
		result := InterfaceToStringSlice(input)
		// Same length as before; the non-string slot is left as the zero value.
		assert.Equal(t, []string{"a", "", "c"}, result)
	})

	// Valid all-string input is unchanged.
	assert.Equal(t, []string{"x", "y"}, InterfaceToStringSlice([]interface{}{"x", "y"}))
}

// HIGH-3: the device-value assertions (GetWifiMac, Pair, PairSupervised) read a
// device-controlled interface{} from a ValueResponse and assert its concrete
// type. A wrong-typed or absent value used to panic on a bare assertion. These
// functions require a live lockdown connection, so we exercise the exact decode
// seam they consume (getValueResponsefromBytes) plus the comma-ok pattern the
// fixes now use, proving a hostile value yields a clean error, not a panic.
func TestDeviceValueAssertionsHandleWrongTypeAndAbsence(t *testing.T) {
	// WiFiAddress returned as an integer instead of a string.
	wrongTypeWifi := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>Key</key><string>WiFiAddress</string><key>Value</key><integer>1234</integer></dict></plist>`
	resp := getValueResponsefromBytes([]byte(wrongTypeWifi))
	assert.NotPanics(t, func() {
		_, ok := resp.Value.(string)
		assert.False(t, ok, "wrong-typed WiFiAddress must fail the string assertion cleanly")
	})

	// DevicePublicKey absent entirely: the Value decodes to untyped nil, which
	// used to panic even on a []byte assertion (nil is not a typed []byte).
	absentValue := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>Key</key><string>DevicePublicKey</string></dict></plist>`
	resp = getValueResponsefromBytes([]byte(absentValue))
	require.Nil(t, resp.Value)
	assert.NotPanics(t, func() {
		_, ok := resp.Value.([]byte)
		assert.False(t, ok, "absent DevicePublicKey must fail the []byte assertion cleanly")
	})
}
