package dtx_test

// Regression tests for the dtx_codec security hardening round.
// Each test feeds an exact malformed/hostile input that previously panicked,
// over-allocated, or blocked forever, and asserts the code now errors, ignores
// it safely, or returns without panicking/hanging. Happy-path decodes are
// pinned so the guards do not alter behavior for valid DTX streams.

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"os"
	"testing"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers for building raw DTX bytes
// ---------------------------------------------------------------------------

// dtxHeader builds the 32 byte DTX message header.
func dtxHeader(fragIndex, fragments uint16, messageLength, identifier, convIndex, channelCode uint32, expectsReply bool) []byte {
	h := make([]byte, 32)
	binary.BigEndian.PutUint32(h[0:], dtx.DtxMessageMagic)
	binary.LittleEndian.PutUint32(h[4:], dtx.DtxMessageHeaderLength)
	binary.LittleEndian.PutUint16(h[8:], fragIndex)
	binary.LittleEndian.PutUint16(h[10:], fragments)
	binary.LittleEndian.PutUint32(h[12:], messageLength)
	binary.LittleEndian.PutUint32(h[16:], identifier)
	binary.LittleEndian.PutUint32(h[20:], convIndex)
	binary.LittleEndian.PutUint32(h[24:], channelCode)
	if expectsReply {
		binary.LittleEndian.PutUint32(h[28:], 1)
	}
	return h
}

// payloadHeader builds the 16 byte payload header.
func payloadHeader(messageType, auxLength, totalPayloadLength, flags uint32) []byte {
	h := make([]byte, 16)
	binary.LittleEndian.PutUint32(h[0:], messageType)
	binary.LittleEndian.PutUint32(h[4:], auxLength)
	binary.LittleEndian.PutUint32(h[8:], totalPayloadLength)
	binary.LittleEndian.PutUint32(h[12:], flags)
	return h
}

// auxHeader builds the 16 byte auxiliary header (only AuxiliarySize @8 matters).
func auxHeader(auxSize uint32) []byte {
	h := make([]byte, 16)
	binary.LittleEndian.PutUint32(h[8:], auxSize)
	return h
}

// runWithTimeout runs fn and fails the test if it does not return within d.
func runWithTimeout(t *testing.T, d time.Duration, name string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("%s did not return within %s (likely blocked/hung)", name, d)
	}
}

// ---------------------------------------------------------------------------
// DTX-01: empty (non-nil) payload must not panic on Payload[0] indexing
// ---------------------------------------------------------------------------

// TestGlobalDispatch_EmptyPayload_NoPanic feeds a message with a non-nil but
// EMPTY payload. Before the fix, `if msg.Payload != nil` passed and
// `msg.Payload[0]` panicked with index out of range.
func TestGlobalDispatch_EmptyPayload_NoPanic(t *testing.T) {
	dispatcher := dtx.NewGlobalDispatcher(make(chan dtx.Message, 1), nil)
	msg := dtx.Message{Payload: []interface{}{}} // non-nil, length 0

	assert.NotPanics(t, func() {
		dispatcher.Dispatch(msg)
	})
}

// TestGlobalDispatch_ErrorEmptyPayload_NoPanic exercises the HasError() branch
// that logged msg.Payload[0] with an empty payload.
func TestGlobalDispatch_ErrorEmptyPayload_NoPanic(t *testing.T) {
	dispatcher := dtx.NewGlobalDispatcher(make(chan dtx.Message, 1), nil)
	msg := dtx.Message{
		Payload:       []interface{}{},
		PayloadHeader: dtx.PayloadHeader{MessageType: dtx.DtxTypeError},
	}
	assert.NotPanics(t, func() {
		dispatcher.Dispatch(msg)
	})
}

// TestGlobalDispatch_NonStringPayload_NoPanic verifies the string comparisons
// no longer assume Payload[0] is a string.
func TestGlobalDispatch_NonStringPayload_NoPanic(t *testing.T) {
	dispatcher := dtx.NewGlobalDispatcher(make(chan dtx.Message, 1), nil)
	msg := dtx.Message{Payload: []interface{}{uint32(42)}}
	assert.NotPanics(t, func() {
		dispatcher.Dispatch(msg)
	})
}

// ---------------------------------------------------------------------------
// DTX-02: aux dictionary length must not wrap in uint32
// ---------------------------------------------------------------------------

// TestDecodeAuxiliary_LengthOverflow_NoPanic feeds a bytearray entry whose
// declared length is 0xFFFFFFFF. Before the fix, 8+length wrapped to 7 in
// uint32, passed the bounds check, then sliced auxBytes[8:8+length] out of
// range and panicked.
func TestDecodeAuxiliary_LengthOverflow_NoPanic(t *testing.T) {
	// entry: [type=bytearray(0x02)][length=0xFFFFFFFF] with no data following
	aux := make([]byte, 8)
	binary.LittleEndian.PutUint32(aux[0:], 0x02)       // t_bytearray
	binary.LittleEndian.PutUint32(aux[4:], 0xFFFFFFFF) // huge length

	assert.NotPanics(t, func() {
		// DecodeAuxiliary swallows the readEntry error and stops; the point is
		// that it does not panic with a wrapped slice bound.
		d := dtx.DecodeAuxiliary(aux)
		assert.Equal(t, 0, len(d.GetArguments()))
	})
}

// ---------------------------------------------------------------------------
// DTX-04: lz4 uncompressed size must be capped, not blindly allocated
// ---------------------------------------------------------------------------

// TestDecompress_HugeUncompressedSize_Errors builds a valid-looking bv41 frame
// whose declared totalUncompressedSize is 0xFFFFFFFF. Before the fix this drove
// make([]byte, totalUncompressedSize+100) — a ~4GiB allocation from a handful
// of bytes (and the +100 could wrap near the top of the range).
func TestDecompress_HugeUncompressedSize_Errors(t *testing.T) {
	buf := make([]byte, 0)
	// totalUncompressedSize = 0xFFFFFFFF
	sz := make([]byte, 4)
	binary.LittleEndian.PutUint32(sz, 0xFFFFFFFF)
	buf = append(buf, sz...)
	// no bv41 magic follows, so the frame loop is skipped and we go straight to
	// the (now guarded) allocation.
	buf = append(buf, 0x00, 0x00, 0x00, 0x00)

	_, err := dtx.Decompress(buf)
	assert.Error(t, err, "oversized uncompressed size must be rejected before allocation")
}

// ---------------------------------------------------------------------------
// DTX-03: ReadMessage length fields must be bounded before make()
// ---------------------------------------------------------------------------

// TestReadMessage_HugeAuxiliarySize_Errors sets AuxiliarySize to ~4GiB. Before
// the fix ReadMessage did make([]byte, AuxiliarySize) unconditionally.
func TestReadMessage_HugeAuxiliarySize_Errors(t *testing.T) {
	var b []byte
	b = append(b, dtxHeader(0, 1, 0, 1, 0, 0, false)...)
	// payload header: messageType=Methodinvocation, auxLength>0 so HasAuxiliary,
	// totalPayloadLength keeps PayloadLength()==0.
	b = append(b, payloadHeader(uint32(dtx.Methodinvocation), 16, 16, 0)...)
	b = append(b, auxHeader(0xFFFFFFF0)...) // huge aux size

	reader := bufio.NewReader(bytes.NewReader(b))
	_, err := dtx.ReadMessage(reader)
	assert.Error(t, err, "huge AuxiliarySize must be rejected before allocation")
}

// TestReadMessage_HugeFragmentLength_Errors sets a subsequent-fragment message
// length to ~4GiB. Before the fix ReadMessage did make([]byte, MessageLength).
func TestReadMessage_HugeFragmentLength_Errors(t *testing.T) {
	// fragments>1 and fragmentIndex>0 => subsequent fragment path.
	b := dtxHeader(1, 3, 0xFFFFFFF0, 1, 0, 0, false)
	reader := bufio.NewReader(bytes.NewReader(b))
	_, err := dtx.ReadMessage(reader)
	assert.Error(t, err, "huge fragment MessageLength must be rejected before allocation")
}

// TestReadMessage_PayloadLengthUnderflow_Errors sets AuxiliaryLength >
// TotalPayloadLength. Before the fix PayloadLength() underflowed the uint32 to
// a huge value used for make([]byte, PayloadLength()). Now HasPayload() is
// false (clamped to 0), so ReadMessage returns cleanly with no payload rather
// than trying a ~4GiB allocation.
func TestReadMessage_PayloadLengthUnderflow_NoHugeAlloc(t *testing.T) {
	var b []byte
	b = append(b, dtxHeader(0, 1, 0, 1, 0, 0, false)...)
	// auxLength=16 (has aux), totalPayloadLength=8 -> PayloadLength underflow.
	b = append(b, payloadHeader(uint32(dtx.Methodinvocation), 16, 8, 0)...)
	b = append(b, auxHeader(0)...) // aux size 0, nothing to read
	// aux length is 16 but AuxiliarySize header says 0, so no aux bytes read.

	reader := bufio.NewReader(bytes.NewReader(b))
	runWithTimeout(t, 5*time.Second, "ReadMessage underflow", func() {
		msg, err := dtx.ReadMessage(reader)
		// Either way it must not attempt a 4GiB payload read; with the clamp,
		// HasPayload is false so it returns with no error and no payload.
		assert.NoError(t, err)
		assert.False(t, msg.HasPayload())
		assert.Equal(t, uint32(0), msg.PayloadLength())
	})
}

// DTX-05 (channel Dispatch selector guard, nil response-waiter non-blocking
// send) is covered by the internal white-box tests in
// hardening_channel_internal_test.go, since a *Channel cannot be constructed
// device-free through the exported API.

// ---------------------------------------------------------------------------
// DTX-06: fragment decoder must not write out of bounds
// ---------------------------------------------------------------------------

// TestFragmentDecoder_OutOfRangeIndex_NoPanic builds a first fragment declaring
// 3 fragments (fragments slice length 2) and then adds a fragment claiming
// FragmentIndex far beyond the slice. Before the fix AddFragment wrote
// f.fragments[FragmentIndex-1] and panicked with index out of range.
func TestFragmentDecoder_OutOfRangeIndex_NoPanic(t *testing.T) {
	// first fragment: fragments=3, index=0 (IsFirstFragment true)
	firstHeader := dtxHeader(0, 3, 7, 5, 0, 0, false)
	first, _, err := dtx.DecodeNonBlocking(firstHeader)
	require.NoError(t, err)
	require.True(t, first.IsFirstFragment())

	decoder := dtx.NewFragmentDecoder(first)

	// hostile fragment: same identifier/fragments, but a wild FragmentIndex.
	hostile := dtx.Message{
		Identifier:    5,
		Fragments:     3,
		FragmentIndex: 60000, // way past the 2-element fragments slice
	}
	assert.NotPanics(t, func() {
		added := decoder.AddFragment(hostile)
		assert.False(t, added, "out-of-range fragment must be rejected, not stored")
	})
}

// ---------------------------------------------------------------------------
// Happy path pins: valid streams must still decode byte-for-byte identically.
// ---------------------------------------------------------------------------

// TestHappyPath_NotifyCapabilities pins a real valid message decode so the
// guards do not change behavior for legitimate input.
func TestHappyPath_NotifyCapabilities(t *testing.T) {
	dat, err := os.ReadFile("fixtures/notifyOfPublishedCapabilites")
	require.NoError(t, err)

	reader := bufio.NewReader(bytes.NewReader(dat))
	msg, err := dtx.ReadMessage(reader)
	require.NoError(t, err)
	assert.Equal(t, uint16(1), msg.Fragments)
	assert.Equal(t, uint16(0), msg.FragmentIndex)
	assert.Equal(t, 612, msg.MessageLength)
	assert.Equal(t, 0, msg.ChannelCode)
	assert.Equal(t, 2, msg.Identifier)
	assert.Equal(t, dtx.MessageType(2), msg.PayloadHeader.MessageType)
	assert.Equal(t, uint32(425), msg.PayloadHeader.AuxiliaryLength)
	assert.Equal(t, uint32(596), msg.PayloadHeader.TotalPayloadLength)
	assert.True(t, msg.HasPayload())
}

// TestHappyPath_ValidAuxiliaryRoundTrip pins that a well-formed auxiliary
// dictionary still decodes to the expected arguments.
func TestHappyPath_ValidAuxiliaryRoundTrip(t *testing.T) {
	aux := dtx.NewPrimitiveDictionary()
	aux.AddInt32(7)
	raw, err := aux.ToBytes()
	require.NoError(t, err)

	decoded := dtx.DecodeAuxiliary(raw)
	args := decoded.GetArguments()
	require.Equal(t, 1, len(args))
	assert.Equal(t, uint32(7), args[0])
}

// TestHappyPath_LZ4RoundTrip pins that a valid lz4 frame still decompresses.
func TestHappyPath_LZ4RoundTrip(t *testing.T) {
	dat, err := os.ReadFile("fixtures/instruments-metrics-dtx.bin")
	require.NoError(t, err)
	// decode the fixture; it contains an lz4 compressed message and must still
	// decode without error after the size cap was added.
	_, _, err = dtx.DecodeNonBlocking(dat)
	assert.NoError(t, err)
}
