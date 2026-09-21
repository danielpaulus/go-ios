package display

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protowire"
	"howett.net/plist"
)

// The daemon rejects any deviation with an opaque "Invalid Parameter" naming no
// field, so offer_test.go pins this encoder against a captured offer.

const (
	// negotiatorModeVideo is the avcMediaStreamNegotiatorMode value for video.
	// Audio is 6 and is not implemented.
	negotiatorModeVideo = 5

	// defaultDecoderName is Apple's decoder string. The device matches on it and
	// rejects an unrecognised one.
	defaultDecoderName = "Viceroy 1.7.0"

	// resEntryCodecCapID is an AVConference codec-capability id, the same in every
	// captured offer.
	resEntryCodecCapID = 50115

	// hevcFeatures and avcFeatures declare per-bank capability flags. "FLS;" is
	// the AVConference framing marker.
	hevcFeatures = "FLS;SW:1;"
	avcFeatures  = "FLS;VRAE:0;SW:1;"

	// RTP payload types for the two codec banks we advertise. The device picks
	// one and reports it back as RxPayloadType; on iOS 27 it chooses AVC.
	hevcPayloadType = 123
	avcPayloadType  = 100

	// capturedVideoTimestamp is a host clock value. The daemon does not appear to
	// validate it, and keeping the captured one lets the test assert byte equality.
	capturedVideoTimestamp = uint64(17137042128614416384)
)

// What the device is told the host is, and may pick encoder parameters from.
// Hardcoded because the host is often not a Mac; these are Xcode's values.
var hostIdentity = struct {
	Model     string
	OSVersion string
	Build     string
}{
	Model:     "Mac15,9",
	OSVersion: "2205.3.1",
	Build:     "25F80",
}

// bitrateTier is one entry of the media blob's tier table. kind 0 carries a
// bitrate cap; the others are markers we did not identify.
type bitrateTier struct {
	kind      uint64
	bps       uint64
	bufferCap uint64
	hasBuffer bool
}

// defaultVideoBitrateTiers is Apple's tier table. The order is significant.
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

// videoBlobParams are the media blob's tunable flags. Only the defaults are sent:
// long-term reference frames off (under UDP loss they leave the picture torn).
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

// The offer is a protobuf with no schema, so the field numbers below are Apple's.
// protowire does the encoding; only varintPadded is not minimal-form protobuf.

func fieldVarint(out []byte, field, value uint64) []byte {
	out = protowire.AppendTag(out, protowire.Number(field), protowire.VarintType)
	return protowire.AppendVarint(out, value)
}

func fieldBool(out []byte, field uint64, value bool) []byte {
	if value {
		return fieldVarint(out, field, 1)
	}
	return fieldVarint(out, field, 0)
}

func fieldBytes(out []byte, field uint64, value []byte) []byte {
	out = protowire.AppendTag(out, protowire.Number(field), protowire.BytesType)
	return protowire.AppendBytes(out, value)
}

func fieldString(out []byte, field uint64, value string) []byte {
	out = protowire.AppendTag(out, protowire.Number(field), protowire.BytesType)
	return protowire.AppendString(out, value)
}

// A varint padded to a fixed width with continuation bytes that add nothing, so
// the session id can be rewritten in place. protowire only emits minimal form.
func fieldVarintPadded(out []byte, field, value uint64, width int) []byte {
	out = protowire.AppendTag(out, protowire.Number(field), protowire.VarintType)
	return append(out, varintPadded(value, width)...)
}

// value must fit in 7*width bits.
func varintPadded(value uint64, width int) []byte {
	encoded := protowire.AppendVarint(nil, value)
	for len(encoded) < width {
		encoded[len(encoded)-1] |= 0x80
		encoded = append(encoded, 0x00)
	}
	return encoded
}

// resEntry is one codec-capability tier inside a codec bank.
func resEntry(pairIndex uint64) []byte {
	entry := fieldVarint(nil, 1, 1)
	entry = fieldVarint(entry, 2, pairIndex)
	entry = fieldVarint(entry, 3, resEntryCodecCapID)
	return fieldVarint(entry, 4, 0)
}

// codecBank describes one codec. resPairCount is 4 for HEVC and 2 for AVC; the
// bank carries alternating pair indices 1, 2 repeated that many times.
func codecBank(payloadType uint64, features string, trailer uint64, resPairCount int) []byte {
	bank := fieldVarint(nil, 1, payloadType)
	for i := 0; i < resPairCount; i++ {
		bank = fieldBytes(bank, 2, resEntry(uint64(1+i%2)))
	}
	bank = fieldString(bank, 3, features)
	return fieldVarint(bank, 4, trailer)
}

func videoSettings(params videoBlobParams) []byte {
	settings := fieldVarintPadded(nil, 1, uint64(params.sessionID), 5)
	settings = fieldBool(settings, 2, params.allowRTCPFB)
	settings = fieldBytes(settings, 3, codecBank(hevcPayloadType, hevcFeatures, 1, 4))
	settings = fieldBytes(settings, 3, codecBank(avcPayloadType, avcFeatures, 14, 2))
	if params.tilesPerFrame != 1 {
		settings = fieldVarint(settings, 6, params.tilesPerFrame)
	}
	settings = fieldBool(settings, 7, params.ltrpEnabled)
	settings = fieldVarint(settings, 8, 63)
	if params.fecEnabled {
		settings = fieldVarint(settings, 10, 1)
	}
	return fieldVarint(settings, 12, 1)
}

// buildVideoMediaBlob assembles the media blob, before compression.
func buildVideoMediaBlob(params videoBlobParams) []byte {
	blob := fieldVarint(nil, 1, 1)
	blob = fieldVarint(blob, 2, 1)
	blob = fieldBytes(blob, 5, videoSettings(params))
	blob = fieldString(blob, 6, defaultDecoderName)
	blob = fieldVarint(blob, 8, 0)
	for _, tier := range defaultVideoBitrateTiers {
		entry := fieldVarint(nil, 1, tier.kind)
		entry = fieldVarint(entry, 2, tier.bps)
		if tier.hasBuffer {
			entry = fieldVarint(entry, 3, tier.bufferCap)
		}
		blob = fieldBytes(blob, 9, entry)
	}
	blob = fieldVarint(blob, 13, params.timestamp)
	blob = fieldVarint(blob, 14, 2)
	blob = fieldVarint(blob, 16, 0)
	return fieldVarint(blob, 18, 1)
}

// buildRemoteEndpointInfo describes the host in the
// avcMediaStreamOptionRemoteEndpointInfo protobuf.
func buildRemoteEndpointInfo() []byte {
	info := fieldVarint(nil, 1, 0)
	info = fieldVarint(info, 2, 1)
	info = fieldString(info, 3, hostIdentity.Model)
	info = fieldString(info, 4, hostIdentity.OSVersion)
	return fieldString(info, 5, hostIdentity.Build)
}

// negotiatorOffer is the binary plist the device expects.
type negotiatorOffer struct {
	RemoteEndpointInfo []byte `plist:"avcMediaStreamOptionRemoteEndpointInfo"`
	NegotiatorMode     int    `plist:"avcMediaStreamNegotiatorMode"`
	MediaBlob          []byte `plist:"avcMediaStreamNegotiatorMediaBlob"`
	CallID             string `plist:"avcMediaStreamOptionCallID"`
}

// buildVideoNegotiatorOffer encodes the offer plist. callID identifies the call
// and sessionID the media session; both are generated per stream by the caller.
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

// compressMediaBlob zlib-compresses the blob. The level matters: the device
// rejects anything but best compression.
func compressMediaBlob(blob []byte) ([]byte, error) {
	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("compressMediaBlob: failed to create writer: %w", err)
	}
	if _, err := writer.Write(blob); err != nil {
		return nil, fmt.Errorf("compressMediaBlob: failed to write blob: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("compressMediaBlob: failed to flush blob: %w", err)
	}
	return compressed.Bytes(), nil
}
