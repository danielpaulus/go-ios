package dtx_test

import (
	"encoding/binary"
	"testing"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/pierrec/lz4"
)

func TestDecompressErrorCases(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty input", data: []byte{}},
		{name: "less than four bytes", data: []byte{0x01, 0x02}},
		{name: "no magic bytes", data: []byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x02}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dtx.Decompress(tt.data)
			if err == nil {
				t.Errorf("Decompress(%v) expected an error, got nil", tt.data)
			}
		})
	}
}

// TestDecompressCompressedSizeTooLarge ensures a bv41 frame claiming more
// compressed bytes than are present errors instead of panicking with an
// index-out-of-range.
func TestDecompressCompressedSizeTooLarge(t *testing.T) {
	buf := make([]byte, 0)
	totalUncompressed := make([]byte, 4)
	binary.LittleEndian.PutUint32(totalUncompressed, 100)
	buf = append(buf, totalUncompressed...)

	// magic
	magic := make([]byte, 4)
	binary.BigEndian.PutUint32(magic, 0x62763431)
	buf = append(buf, magic...)
	// uncompressed size (unused)
	buf = append(buf, 0, 0, 0, 0)
	// compressed size claims 1000 bytes but none follow
	compressedSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(compressedSize, 1000)
	buf = append(buf, compressedSize...)

	_, err := dtx.Decompress(buf)
	if err == nil {
		t.Fatalf("Decompress with oversized compressedSize expected an error, got nil")
	}
}

func TestDecompressRoundTrip(t *testing.T) {
	// highly compressible payload so CompressBlock produces a non-empty block.
	original := make([]byte, 256)
	for i := range original {
		original[i] = byte(i % 4)
	}

	compressed := make([]byte, lz4.CompressBlockBound(len(original)))
	n, err := lz4.CompressBlock(original, compressed, nil)
	if err != nil {
		t.Fatalf("CompressBlock failed: %v", err)
	}
	if n == 0 {
		t.Fatalf("CompressBlock produced an empty (incompressible) block, cannot build frame")
	}
	compressed = compressed[:n]

	buf := make([]byte, 0)
	totalUncompressed := make([]byte, 4)
	binary.LittleEndian.PutUint32(totalUncompressed, uint32(len(original)))
	buf = append(buf, totalUncompressed...)

	magic := make([]byte, 4)
	binary.BigEndian.PutUint32(magic, 0x62763431)
	buf = append(buf, magic...)
	// uncompressed size (unused by Decompress)
	uncompressedSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(uncompressedSize, uint32(len(original)))
	buf = append(buf, uncompressedSize...)
	compressedSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(compressedSize, uint32(len(compressed)))
	buf = append(buf, compressedSize...)
	buf = append(buf, compressed...)
	// trailing bytes that are not the bv41 magic terminate the loop
	buf = append(buf, 0x00, 0x00, 0x00, 0x00)

	got, err := dtx.Decompress(buf)
	if err != nil {
		t.Fatalf("Decompress round trip failed: %v", err)
	}
	if len(got) != len(original) {
		t.Fatalf("Decompress round trip length = %d, want %d", len(got), len(original))
	}
	for i := range original {
		if got[i] != original[i] {
			t.Fatalf("Decompress round trip mismatch at %d: got %d want %d", i, got[i], original[i])
		}
	}
}
