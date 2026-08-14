package hid

import (
	"bytes"
	"compress/zlib"
	"fmt"

	plist "howett.net/plist"
)

// This file ports pymobiledevice3's media_stream_offer.py: it builds the
// negotiatorOffer binary-plist that
// com.apple.coredevice.action.mediastreamstart requires. Only the video path is
// ported — that's all the HID auth gate needs. The protobuf field layout and
// the captured constants come straight from the pymd3 source (reverse
// engineered from AVCMediaStreamNegotiator).

const (
	negotiatorModeVideo = 5

	defaultDecoderName = "Viceroy 1.7.0"
	resEntryCodecCapID = 50115
	// Host identity Xcode's mirror uses (Mac15,9 = M2 Air, macOS build 25F80).
	hostModel     = "Mac15,9"
	hostOSVersion = "2205.3.1"
	hostBuild     = "25F80"
	hevcFeatures  = "FLS;SW:1;"
	avcFeatures   = "FLS;VRAE:0;SW:1;"
	// Timestamp baked into Apple's captured offer; the daemon doesn't validate it.
	capturedVideoTimestamp = uint64(17137042128614416384)
)

// videoBitrateTier is one f9 entry in the top-level MediaBlob. present3 marks
// whether the optional f3 (buffer cap) is included.
type videoBitrateTier struct {
	f1, f2, f3 uint64
	present3   bool
}

// defaultVideoBitrateTiers is Apple's canonical bitrate-tier table for video.
var defaultVideoBitrateTiers = []videoBitrateTier{
	{4074, 0, 16384, true},
	{0, 75_000_000, 524288, true},
	{0, 40_000_000, 12288, true},
	{16, 4100, 0, false},
	{0, 20_000_000, 98304, true},
	{4, 6500, 0, false},
	{0, 6_000_000, 131072, true},
	{0, 100_000_000, 1048576, true},
	{0, 60_000_000, 262144, true},
	{1, 299, 0, false},
}

// --- protobuf helpers ---

func pbVarint(v uint64) []byte {
	var out []byte
	for {
		b := byte(v & 0x7F)
		v >>= 7
		if v != 0 {
			out = append(out, b|0x80)
		} else {
			out = append(out, b)
			return out
		}
	}
}

// pbVarintPadded encodes v as a varint padded to exactly width bytes using
// redundant continuation bytes (protobuf tolerates these). Apple's offer uses
// this for the 5-byte session_id slot.
func pbVarintPadded(v uint64, width int) []byte {
	raw := pbVarint(v)
	for len(raw) < width {
		raw[len(raw)-1] |= 0x80
		raw = append(raw, 0x00)
	}
	return raw
}

func pbTag(field, wire int) []byte { return pbVarint(uint64(field<<3) | uint64(wire)) }

func pbFVarint(field int, v uint64) []byte { return append(pbTag(field, 0), pbVarint(v)...) }

func pbFBytes(field int, v []byte) []byte {
	out := pbTag(field, 2)
	out = append(out, pbVarint(uint64(len(v)))...)
	return append(out, v...)
}

func pbFString(field int, v string) []byte { return pbFBytes(field, []byte(v)) }

// --- mediaBlob piece builders ---

func resEntry(pairIndex int) []byte {
	out := pbFVarint(1, 1)
	out = append(out, pbFVarint(2, uint64(pairIndex))...)
	out = append(out, pbFVarint(3, resEntryCodecCapID)...)
	out = append(out, pbFVarint(4, 0)...)
	return out
}

func codecBank(payloadType int, features string, f4 int, resPairCount int) []byte {
	body := pbFVarint(1, uint64(payloadType))
	for i := 0; i < resPairCount; i++ {
		body = append(body, pbFBytes(2, resEntry(1+(i%2)))...)
	}
	body = append(body, pbFString(3, features)...)
	body = append(body, pbFVarint(4, uint64(f4))...)
	return body
}

// buildMediaBlobVideo builds the (uncompressed) video mediaBlob protobuf with
// pymd3's shipped defaults (ltrpEnabled off, fecEnabled on).
func buildMediaBlobVideo(sessionID uint32) []byte {
	return buildMediaBlobVideoOpts(sessionID, false, true)
}

// buildMediaBlobVideoOpts exposes the ltrp/fec knobs so the byte-equivalence
// test can reproduce Apple's captured template (ltrp on, fec off).
func buildMediaBlobVideoOpts(sessionID uint32, ltrpEnabled, fecEnabled bool) []byte {
	videoSettings := pbTag(1, 0)
	videoSettings = append(videoSettings, pbVarintPadded(uint64(sessionID), 5)...)
	videoSettings = append(videoSettings, pbFVarint(2, 0)...) // allowRTCPFB off
	videoSettings = append(videoSettings, pbFBytes(3, codecBank(123, hevcFeatures, 1, 4))...)
	videoSettings = append(videoSettings, pbFBytes(3, codecBank(100, avcFeatures, 14, 2))...)
	if ltrpEnabled {
		videoSettings = append(videoSettings, pbFVarint(7, 1)...)
	} else {
		videoSettings = append(videoSettings, pbFVarint(7, 0)...)
	}
	videoSettings = append(videoSettings, pbFVarint(8, 63)...) // pixelFormats
	if fecEnabled {
		videoSettings = append(videoSettings, pbFVarint(10, 1)...)
	}
	videoSettings = append(videoSettings, pbFVarint(12, 1)...)

	return buildMediaBlobTopLevel(5, videoSettings, defaultVideoBitrateTiers)
}

func buildMediaBlobTopLevel(settingsField int, settingsBody []byte, tiers []videoBitrateTier) []byte {
	var f9s []byte
	for _, t := range tiers {
		body := pbFVarint(1, t.f1)
		body = append(body, pbFVarint(2, t.f2)...)
		if t.present3 {
			body = append(body, pbFVarint(3, t.f3)...)
		}
		f9s = append(f9s, pbFBytes(9, body)...)
	}
	out := pbFVarint(1, 1)
	out = append(out, pbFVarint(2, 1)...)
	out = append(out, pbFBytes(settingsField, settingsBody)...)
	out = append(out, pbFString(6, defaultDecoderName)...)
	out = append(out, pbFVarint(8, 0)...)
	out = append(out, f9s...)
	out = append(out, pbFVarint(13, capturedVideoTimestamp)...)
	out = append(out, pbFVarint(14, 2)...)
	out = append(out, pbFVarint(16, 0)...)
	out = append(out, pbFVarint(18, 1)...)
	return out
}

// buildRemoteEndpointInfo builds the avcMediaStreamOptionRemoteEndpointInfo
// protobuf describing the host.
func buildRemoteEndpointInfo(model, osVersion, build string) []byte {
	out := pbFVarint(1, 0)
	out = append(out, pbFVarint(2, 1)...)
	out = append(out, pbFString(3, model)...)
	out = append(out, pbFString(4, osVersion)...)
	out = append(out, pbFString(5, build)...)
	return out
}

// zlibCompress compresses b with zlib at best compression (level 9), matching
// pymd3's zlib.compress(..., level=9). Apple's daemon rejects other levels.
func zlibCompress(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(b); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildNegotiatorOfferVideo builds the full negotiatorOffer binary plist for a
// video stream. callID is a fresh UUID string; sessionID a random uint32.
func buildNegotiatorOfferVideo(callID string, sessionID uint32) ([]byte, error) {
	endpointInfo := buildRemoteEndpointInfo(hostModel, hostOSVersion, hostBuild)
	mediaBlob := buildMediaBlobVideo(sessionID)
	compressed, err := zlibCompress(mediaBlob)
	if err != nil {
		return nil, fmt.Errorf("buildNegotiatorOfferVideo: zlib compress: %w", err)
	}
	offer := map[string]interface{}{
		"avcMediaStreamOptionRemoteEndpointInfo": endpointInfo,
		"avcMediaStreamNegotiatorMode":           negotiatorModeVideo,
		"avcMediaStreamNegotiatorMediaBlob":      compressed,
		"avcMediaStreamOptionCallID":             callID,
	}
	data, err := plist.Marshal(offer, plist.BinaryFormat)
	if err != nil {
		return nil, fmt.Errorf("buildNegotiatorOfferVideo: plist marshal: %w", err)
	}
	return data, nil
}
