package display

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"howett.net/plist"
)

// This file builds the negotiatorOffer that
// com.apple.coredevice.action.mediastreamstart requires. The offer is a binary
// plist with four keys:
//
//	avcMediaStreamOptionRemoteEndpointInfo: protobuf describing the host
//	avcMediaStreamNegotiatorMode:           5 for video, 6 for audio
//	avcMediaStreamNegotiatorMediaBlob:      zlib(level 9) compressed protobuf
//	                                       carrying codec parameters
//	avcMediaStreamOptionCallID:             upper-case UUID string
//
// The mediaBlob layout was reverse engineered from -[AVCMediaStreamNegotiator
// createOffer]; field numbers come from the __objc_methname table of
// VCMediaNegotiationBlobVideoSettings:
//
//	MediaBlob {
//	  1: int = 1, 2: int = 1
//	  5: VideoSettings
//	  6: string  decoder identity
//	  8: int = 0
//	  9: BitrateTier[]  repeated
//	  13: int  host timestamp (not validated by the daemon)
//	  14: int = 2, 16: int = 0, 18: int = 1
//	}
//
//	VideoSettings {
//	  1: int  SSRC / session id, encoded as a 5-byte padded varint
//	  2: int  allowRTCPFB
//	  3: CodecBank[]  HEVC then AVC
//	  7: int  ltrpEnabled
//	  8: int = 63  pixel formats
//	  10: int  fecEnabled (omitted when false)
//	  12: int = 1  blackFrameOnClearScreen
//	}
//
//	CodecBank { 1: payload type, 2: ResEntry[], 3: feature string, 4: int }
//	ResEntry  { 1: int = 1, 2: pair index, 3: int = 50115, 4: int = 0 }
//	BitrateTier { 1: tier kind, 2: bps, 3: buffer cap (optional) }
//
// The daemon rejects any deviation with an opaque "Invalid Parameter", so
// offer_test.go pins the encoder against a captured Xcode offer byte for byte.

const (
	// negotiatorModeVideo is the avcMediaStreamNegotiatorMode value for a video
	// stream. Audio (6) is not implemented: go-ios only needs a video stream to
	// hold dtuhidd's touch auth gate open.
	negotiatorModeVideo = 5

	// defaultDecoderName is matched by the device to pick a compatible decoder.
	defaultDecoderName = "Viceroy 1.7.0"

	// resEntryCodecCapID is a fixed AVConference codec-capability id that
	// appears unchanged across all captured offers.
	resEntryCodecCapID = 50115

	// hevcFeatures and avcFeatures declare per-bank capability flags. "FLS;" is
	// the AVConference framing marker.
	hevcFeatures = "FLS;SW:1;"
	avcFeatures  = "FLS;VRAE:0;SW:1;"

	hevcPayloadType = 123
	avcPayloadType  = 100

	// capturedVideoTimestamp is the host clock value from the captured offer.
	// The daemon does not appear to validate it, and keeping the captured value
	// lets offer_test.go assert byte equality.
	capturedVideoTimestamp = uint64(17137042128614416384)
)

// hostIdentity describes the host to the device. The device may pick encoder
// parameters based on it: under a Mac16,11 identity captured sessions showed
// frequent encoder stalls, while Mac15,9 (what Xcode reports) is stable.
var hostIdentity = struct {
	Model     string
	OSVersion string
	Build     string
}{
	Model:     "Mac15,9",
	OSVersion: "2205.3.1",
	Build:     "25F80",
}

// bitrateTier is one f9 entry of the media blob. kind 0 is a tier with a
// network bitrate cap, 4074 a header marker, and 16/4/1 codec-specific markers.
// bufferCap is only present for the cap-carrying kinds.
type bitrateTier struct {
	kind      uint64
	bps       uint64
	bufferCap uint64
	hasBuffer bool
}

// defaultVideoBitrateTiers is Apple's canonical tier table, in the order the
// captured offer listed them. The order is part of the byte-equality contract.
var defaultVideoBitrateTiers = []bitrateTier{
	{kind: 4074, bps: 0, bufferCap: 16384, hasBuffer: true},
	{kind: 0, bps: 75_000_000, bufferCap: 524288, hasBuffer: true},
	{kind: 0, bps: 40_000_000, bufferCap: 12288, hasBuffer: true},
	{kind: 16, bps: 4100},
	{kind: 0, bps: 20_000_000, bufferCap: 98304, hasBuffer: true},
	{kind: 4, bps: 6500},
	{kind: 0, bps: 6_000_000, bufferCap: 131072, hasBuffer: true},
	{kind: 0, bps: 100_000_000, bufferCap: 1048576, hasBuffer: true},
	{kind: 0, bps: 60_000_000, bufferCap: 262144, hasBuffer: true},
	{kind: 1, bps: 299},
}

// videoBlobParams are the tunable flags of the video media blob. The zero value
// is what go-ios sends: LTRP off (it eliminates mid-stream tearing under UDP
// loss and the device honours the request), FEC on, one tile per frame.
type videoBlobParams struct {
	sessionID     uint32
	allowRTCPFB   bool
	ltrpEnabled   bool
	fecEnabled    bool
	tilesPerFrame uint64
	timestamp     uint64
}

func defaultVideoBlobParams(sessionID uint32) videoBlobParams {
	return videoBlobParams{
		sessionID:     sessionID,
		fecEnabled:    true,
		tilesPerFrame: 1,
		timestamp:     capturedVideoTimestamp,
	}
}

// --- protobuf primitives -----------------------------------------------------

func varint(v uint64) []byte {
	var out []byte
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			out = append(out, b|0x80)
			continue
		}
		return append(out, b)
	}
}

// varintPadded encodes v as a varint padded to exactly width bytes. Protobuf
// tolerates redundant continuation bytes, and Apple's offers use this for the
// 5-byte session id slot so the field can be rewritten in place.
func varintPadded(v uint64, width int) []byte {
	raw := varint(v)
	if len(raw) > width {
		mask := uint64(1)<<(7*uint(width)) - 1
		raw = varint(v & mask)
	}
	for len(raw) < width {
		raw[len(raw)-1] |= 0x80
		raw = append(raw, 0x00)
	}
	return raw
}

func tag(field, wireType uint64) []byte {
	return varint(field<<3 | wireType)
}

func fieldVarint(field, value uint64) []byte {
	return append(tag(field, 0), varint(value)...)
}

func fieldBytes(field uint64, value []byte) []byte {
	out := append(tag(field, 2), varint(uint64(len(value)))...)
	return append(out, value...)
}

func fieldString(field uint64, value string) []byte {
	return fieldBytes(field, []byte(value))
}

func boolVarint(field uint64, b bool) []byte {
	if b {
		return fieldVarint(field, 1)
	}
	return fieldVarint(field, 0)
}

// --- media blob --------------------------------------------------------------

// resEntry is one codec-capability tier inside a codec bank.
func resEntry(pairIndex uint64) []byte {
	out := fieldVarint(1, 1)
	out = append(out, fieldVarint(2, pairIndex)...)
	out = append(out, fieldVarint(3, resEntryCodecCapID)...)
	return append(out, fieldVarint(4, 0)...)
}

// codecBank describes one codec. resPairCount is 4 for HEVC and 2 for AVC; the
// bank carries alternating pair indices 1, 2 repeated that many times.
func codecBank(payloadType uint64, features string, trailer uint64, resPairCount int) []byte {
	body := fieldVarint(1, payloadType)
	for i := 0; i < resPairCount; i++ {
		body = append(body, fieldBytes(2, resEntry(uint64(1+i%2)))...)
	}
	body = append(body, fieldString(3, features)...)
	return append(body, fieldVarint(4, trailer)...)
}

func videoSettings(p videoBlobParams) []byte {
	out := append(tag(1, 0), varintPadded(uint64(p.sessionID), 5)...)
	out = append(out, boolVarint(2, p.allowRTCPFB)...)
	out = append(out, fieldBytes(3, codecBank(hevcPayloadType, hevcFeatures, 1, 4))...)
	out = append(out, fieldBytes(3, codecBank(avcPayloadType, avcFeatures, 14, 2))...)
	if p.tilesPerFrame != 1 {
		out = append(out, fieldVarint(6, p.tilesPerFrame)...)
	}
	out = append(out, boolVarint(7, p.ltrpEnabled)...)
	out = append(out, fieldVarint(8, 63)...)
	if p.fecEnabled {
		out = append(out, fieldVarint(10, 1)...)
	}
	return append(out, fieldVarint(12, 1)...)
}

// buildVideoMediaBlob assembles the uncompressed video media blob protobuf.
func buildVideoMediaBlob(p videoBlobParams) []byte {
	out := fieldVarint(1, 1)
	out = append(out, fieldVarint(2, 1)...)
	out = append(out, fieldBytes(5, videoSettings(p))...)
	out = append(out, fieldString(6, defaultDecoderName)...)
	out = append(out, fieldVarint(8, 0)...)
	for _, t := range defaultVideoBitrateTiers {
		body := fieldVarint(1, t.kind)
		body = append(body, fieldVarint(2, t.bps)...)
		if t.hasBuffer {
			body = append(body, fieldVarint(3, t.bufferCap)...)
		}
		out = append(out, fieldBytes(9, body)...)
	}
	out = append(out, fieldVarint(13, p.timestamp)...)
	out = append(out, fieldVarint(14, 2)...)
	out = append(out, fieldVarint(16, 0)...)
	return append(out, fieldVarint(18, 1)...)
}

// buildRemoteEndpointInfo describes the host in the
// avcMediaStreamOptionRemoteEndpointInfo protobuf.
func buildRemoteEndpointInfo() []byte {
	out := fieldVarint(1, 0)
	out = append(out, fieldVarint(2, 1)...)
	out = append(out, fieldString(3, hostIdentity.Model)...)
	out = append(out, fieldString(4, hostIdentity.OSVersion)...)
	return append(out, fieldString(5, hostIdentity.Build)...)
}

// --- offer -------------------------------------------------------------------

// negotiatorOffer is the binary plist the device expects. Field order in the
// struct is irrelevant: plist dictionaries are keyed.
type negotiatorOffer struct {
	RemoteEndpointInfo []byte `plist:"avcMediaStreamOptionRemoteEndpointInfo"`
	NegotiatorMode     int    `plist:"avcMediaStreamNegotiatorMode"`
	MediaBlob          []byte `plist:"avcMediaStreamNegotiatorMediaBlob"`
	CallID             string `plist:"avcMediaStreamOptionCallID"`
}

// buildVideoNegotiatorOffer encodes the full offer plist for a video stream.
// callID identifies the call and sessionID the media session inside the blob;
// both are generated per stream by the caller.
func buildVideoNegotiatorOffer(callID uuid.UUID, sessionID uint32) ([]byte, error) {
	blob, err := compressMediaBlob(buildVideoMediaBlob(defaultVideoBlobParams(sessionID)))
	if err != nil {
		return nil, err
	}
	offer := negotiatorOffer{
		RemoteEndpointInfo: buildRemoteEndpointInfo(),
		NegotiatorMode:     negotiatorModeVideo,
		MediaBlob:          blob,
		CallID:             strings.ToUpper(callID.String()),
	}
	data, err := plist.Marshal(offer, plist.BinaryFormat)
	if err != nil {
		return nil, fmt.Errorf("buildVideoNegotiatorOffer: failed to marshal offer plist: %w", err)
	}
	return data, nil
}

// compressMediaBlob zlib-compresses the blob. The level matters: Apple uses
// best compression and the device rejects a stream built with any other level
// with "Invalid Parameter".
func compressMediaBlob(blob []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("compressMediaBlob: failed to create writer: %w", err)
	}
	if _, err := w.Write(blob); err != nil {
		return nil, fmt.Errorf("compressMediaBlob: failed to write blob: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("compressMediaBlob: failed to flush blob: %w", err)
	}
	return buf.Bytes(), nil
}
