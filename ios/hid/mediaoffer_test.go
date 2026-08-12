package hid

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// capturedVideoTemplateHex is the byte-for-byte video mediaBlob captured from a
// live Xcode screen-mirror session (see pymobiledevice3 media_stream_offer.py).
// The capture used ltrp_enabled=true, fec_enabled=false and this session_id.
const (
	capturedVideoTemplateHex = "080110012a7f088182bae90810001a3f087b120a0801100118c387032000120a" +
		"0801100218c387032000120a0801100118c387032000120a0801100218c38703" +
		"20001a09464c533b53573a313b20011a2e0864120a0801100118c38703200012" +
		"0a0801100218c3870320001a10464c533b565241453a303b53573a313b200e38" +
		"01403f6001320d56696365726f7920312e372e3040004a0908ea1f1000188080" +
		"014a0b080010c0d1e123188080204a0a08001080b489131880604a0508101084" +
		"204a0b08001080dac409188080064a05080410e4324a0b080010809bee021880" +
		"80084a0b08001080c2d72f188080404a0b080010808ece1c188080104a050801" +
		"10ab026880c0dd87d2a0c0e9ed017002800100900101"
	capturedVideoSessionID = uint32(2368635137)
)

func TestBuildMediaBlobVideoMatchesXcodeCapture(t *testing.T) {
	want, err := hex.DecodeString(capturedVideoTemplateHex)
	if err != nil {
		t.Fatalf("decode captured template: %v", err)
	}
	got := buildMediaBlobVideoOpts(capturedVideoSessionID, true, false)
	if !bytes.Equal(got, want) {
		t.Fatalf("video mediaBlob drifted from Xcode capture:\n got (%d bytes) %x\nwant (%d bytes) %x",
			len(got), got, len(want), want)
	}
}

func TestPBVarintPaddedFillsWidth(t *testing.T) {
	got := pbVarintPadded(uint64(capturedVideoSessionID), 5)
	if len(got) != 5 {
		t.Fatalf("padded varint length = %d, want 5", len(got))
	}
	// 5-byte padded encoding of 2368635137 from the capture: 81 82 ba e9 08
	want := []byte{0x81, 0x82, 0xba, 0xe9, 0x08}
	if !bytes.Equal(got, want) {
		t.Fatalf("padded varint = %x, want %x", got, want)
	}
}

func TestBuildNegotiatorOfferVideoIsBinaryPlist(t *testing.T) {
	offer, err := buildNegotiatorOfferVideo("A1B2C3D4-0000-0000-0000-000000000000", 1234)
	if err != nil {
		t.Fatalf("buildNegotiatorOfferVideo: %v", err)
	}
	// Binary plist magic header.
	if !bytes.HasPrefix(offer, []byte("bplist00")) {
		t.Fatalf("offer is not a binary plist, prefix = %q", offer[:min(8, len(offer))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
