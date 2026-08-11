package dtx

import (
	"encoding/binary"
	"fmt"

	"github.com/pierrec/lz4"
)

const bv41 = 0x62763431

// maxUncompressedSize caps the buffer we allocate for an lz4-decompressed DTX
// message. totalUncompressedSize is an attacker-controlled uint32 from the
// device, so without a cap a hostile frame could request a ~4GiB allocation
// from only a few bytes of input. 256 MiB is far above any legitimate DTX
// message while keeping a malformed frame from exhausting memory.
const maxUncompressedSize = 256 * 1024 * 1024

// https://discuss.appium.io/t/how-to-parse-trace-file-to-get-cpu-performance-usage-data-for-ios-apps/35334/2
func Decompress(data []byte) ([]byte, error) {
	// no idea what the first four bytes mean
	if len(data) < 4 {
		return nil, fmt.Errorf("lz4 decompress: need at least 4 bytes, got %d", len(data))
	}
	totalUncompressedSize := binary.LittleEndian.Uint32(data)
	data = data[4:]

	if len(data) < 4 {
		return nil, fmt.Errorf("lz4 decompress: need at least 4 bytes for magic, got %d", len(data))
	}
	var magic uint32
	magic = binary.BigEndian.Uint32(data)
	compressedAgg := make([]byte, 0)
	for magic == bv41 {
		// each bv41 frame is at least a 12-byte header (magic, uncompressed size, compressed size).
		if len(data) < 12 {
			return nil, fmt.Errorf("lz4 decompress: truncated bv41 header, need 12 bytes, got %d", len(data))
		}
		// uncompressedSize := binary.LittleEndian.Uint32(data[4:])
		compressedSize := binary.LittleEndian.Uint32(data[8:])
		if uint64(len(data)) < 12+uint64(compressedSize) {
			return nil, fmt.Errorf("lz4 decompress: compressed size %d exceeds remaining %d bytes", compressedSize, len(data)-12)
		}
		chunk := data[12 : 12+compressedSize]
		// log.Infof("chunk: %x", chunk)
		data = data[12+compressedSize:]

		compressedAgg = append(compressedAgg, chunk...)
		if len(data) < 4 {
			break
		}
		magic = binary.BigEndian.Uint32(data)
	}
	// Do the +100 slack in uint64 so a totalUncompressedSize near 0xFFFFFFFF
	// cannot wrap around, and reject anything above the sane maximum before
	// allocating.
	allocSize := uint64(totalUncompressedSize) + 100
	if allocSize > maxUncompressedSize {
		return nil, fmt.Errorf("lz4 decompress: uncompressed size %d exceeds maximum %d", totalUncompressedSize, maxUncompressedSize)
	}
	uncompressedData := make([]byte, allocSize)
	n, err := lz4.UncompressBlock(compressedAgg, uncompressedData)
	if err != nil {
		return []byte{}, err
	}
	// log.Infof("uncompressed lz4 data of %d bytes", len(uncompressedData[:n]))
	return uncompressedData[:n], nil
}
