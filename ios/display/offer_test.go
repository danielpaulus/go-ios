package display

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"howett.net/plist"
)

// capturedVideoMediaBlob is the uncompressed media blob from an offer Xcode sent
// to a device, with capturedVideoSessionID as the session id. dtuhidd rejects
// malformed offers with an opaque "Invalid Parameter", so this fixture is the
// only cheap way to catch encoder drift.
//
// The capture had LTRP enabled and FEC disabled; the default here is the inverse
// (see videoBlobParams), so the test builds with the captured flags.
const capturedVideoMediaBlobHex = "080110012a7f088182bae90810001a3f087b120a0801100118c387032000120a" +
	"0801100218c387032000120a0801100118c387032000120a0801100218c38703" +
	"20001a09464c533b53573a313b20011a2e0864120a0801100118c38703200012" +
	"0a0801100218c3870320001a10464c533b565241453a303b53573a313b200e38" +
	"01403f6001320d56696365726f7920312e372e3040004a0908ea1f1000188080" +
	"014a0b080010c0d1e123188080204a0a08001080b489131880604a0508101084" +
	"204a0b08001080dac409188080064a05080410e4324a0b080010809bee021880" +
	"80084a0b08001080c2d72f188080404a0b080010808ece1c188080104a050801" +
	"10ab026880c0dd87d2a0c0e9ed017002800100900101"

const capturedVideoSessionID = uint32(2368635137)

// TestBuildVideoMediaBlobMatchesCapture is the contract test for the whole
// encoder: every protobuf field number, the 5-byte padded session id, the codec
// banks, the bitrate tier table and its order.
func TestBuildVideoMediaBlobMatchesCapture(t *testing.T) {
	captured, err := hex.DecodeString(capturedVideoMediaBlobHex)
	require.NoError(t, err)

	params := defaultVideoBlobParams(capturedVideoSessionID)
	params.ltrpEnabled = true // the capture had LTRP on
	params.fecEnabled = false // ... and FEC off

	built := buildVideoMediaBlob(params)

	require.Equal(t, len(captured), len(built), "media blob length drifted from the captured offer")
	assert.Equal(t, hex.EncodeToString(captured), hex.EncodeToString(built))
}

func TestVarintPadded(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
		width int
		want  string
	}{
		{name: "zero pads to width", value: 0, width: 5, want: "8080808000"},
		{name: "small value pads", value: 1, width: 3, want: "818000"},
		{name: "exact width unchanged", value: 300, width: 2, want: "ac02"},
		{name: "captured session id", value: uint64(capturedVideoSessionID), width: 5, want: "8182bae908"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := varintPadded(tt.value, tt.width)
			assert.Len(t, got, tt.width)
			assert.Equal(t, tt.want, hex.EncodeToString(got))
		})
	}
}

// TestVarintPaddedIsValueStable verifies the padding does not change the decoded
// value: redundant continuation bytes must be semantically invisible.
func TestVarintPaddedIsValueStable(t *testing.T) {
	for _, v := range []uint64{0, 1, 127, 128, 300, uint64(capturedVideoSessionID)} {
		padded := varintPadded(v, 5)
		decoded, n := decodeVarint(t, padded)
		assert.Equal(t, v, decoded, "padded varint decoded to a different value")
		assert.Equal(t, 5, n, "padded varint should consume exactly the padded width")
	}
}

func decodeVarint(t *testing.T, b []byte) (uint64, int) {
	t.Helper()
	value, n := protowire.ConsumeVarint(b)
	if n < 0 {
		t.Fatalf("varint %x is truncated", b)
	}
	return value, n
}

// TestBuildVideoNegotiatorOfferEnvelope checks the plist envelope the daemon
// parses: the four keys, the video negotiator mode, an upper-case call id, and a
// media blob that is zlib data round-tripping to the uncompressed protobuf.
func TestBuildVideoNegotiatorOfferEnvelope(t *testing.T) {
	callID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	data, err := buildVideoNegotiatorOffer(callID, capturedVideoSessionID)
	require.NoError(t, err)

	var decoded struct {
		RemoteEndpointInfo []byte `plist:"avcMediaStreamOptionRemoteEndpointInfo"`
		NegotiatorMode     int    `plist:"avcMediaStreamNegotiatorMode"`
		MediaBlob          []byte `plist:"avcMediaStreamNegotiatorMediaBlob"`
		CallID             string `plist:"avcMediaStreamOptionCallID"`
	}
	_, err = plist.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, negotiatorModeVideo, decoded.NegotiatorMode)
	assert.Equal(t, "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE", decoded.CallID,
		"the device expects an upper-case call id")
	assert.NotEmpty(t, decoded.RemoteEndpointInfo)

	r, err := zlib.NewReader(bytes.NewReader(decoded.MediaBlob))
	require.NoError(t, err, "media blob must be zlib data")
	inflated, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	assert.Equal(t, buildVideoMediaBlob(defaultVideoBlobParams(capturedVideoSessionID)), inflated)
}

// TestCompressMediaBlobUsesBestCompression pins the zlib level: the device
// rejects offers compressed at any other level with "Invalid Parameter".
func TestCompressMediaBlobUsesBestCompression(t *testing.T) {
	blob := buildVideoMediaBlob(defaultVideoBlobParams(capturedVideoSessionID))

	got, err := compressMediaBlob(blob)
	require.NoError(t, err)

	var want bytes.Buffer
	w, err := zlib.NewWriterLevel(&want, zlib.BestCompression)
	require.NoError(t, err)
	_, err = w.Write(blob)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	assert.Equal(t, want.Bytes(), got)
}

func TestBuildRemoteEndpointInfoCarriesHostIdentity(t *testing.T) {
	info := buildRemoteEndpointInfo()

	assert.Contains(t, string(info), hostIdentity.Model)
	assert.Contains(t, string(info), hostIdentity.OSVersion)
	assert.Contains(t, string(info), hostIdentity.Build)
}
