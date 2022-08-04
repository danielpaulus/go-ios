package dtx

import (
	"encoding/binary"
	"github.com/pierrec/lz4"
	log "github.com/sirupsen/logrus"
)

const bv41 = 0x62763431

func Decompress(data []byte) error {
	//no idea what the first four bytes mean
	totalUncompressedSize := binary.LittleEndian.Uint32(data)
	data = data[4:]

	var magic uint32
	magic = binary.BigEndian.Uint32(data)
	result := make([]byte, totalUncompressedSize)
	for magic == bv41 {
		uncompressedSize := binary.LittleEndian.Uint32(data[4:])
		compressedSize := binary.LittleEndian.Uint32(data[8:])
		chunk := data[12 : 12+compressedSize]
		log.Infof("chunk: %x", chunk)
		data = data[12+compressedSize:]
		uncompressedData := make([]byte, uncompressedSize*2)
		_, err := lz4.UncompressBlock(chunk, uncompressedData)
		if err != nil {
			log.Warn(err)
		}
		result = append(result, uncompressedData...)
	}
	log.Infof("%x", result)
	return nil
}
