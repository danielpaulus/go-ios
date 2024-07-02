package zipconduit

import (
	"encoding/hex"
	"strings"

	log "github.com/sirupsen/logrus"
)

// sadly apple does not use a standard compliant zip implementation for this
// so I had to hack my own basic pseudo zip format together.
// this is for a directory.
func newZipHeaderDir(name string) (zipHeader, []byte, []byte) {
	return zipHeader{
		signature:              0x04034b50,
		version:                20,
		generalPurposeBitFlags: 0,
		compressionMethod:      0,
		lastModifiedTime:       0xBDEF,
		lastModifiedDate:       0x52EC,
		crc32:                  0,
		compressedSize:         0,
		uncompressedSize:       0,
		fileNameLength:         uint16(len(name)),
		extraFieldLength:       32,
	}, []byte(name), zipExtraBytes
}

// sadly apple does not use a standard compliant zip implementation for this
// so I had to hack my own basic pseudo zip format together.
// this is for a file. It returns the file header, the bytes for the file name and an extra.
func newZipHeader(size uint32, crc32 uint32, name string) (zipHeader, []byte, []byte) {
	// the predefined values are just random ones I grabbed from a hexdump
	// since we only want to get files to a device so it can install an app
	// timestamps and all that don't really matter anyway
	return zipHeader{
		signature:              0x04034b50,
		version:                20,
		generalPurposeBitFlags: 0,
		compressionMethod:      0,
		lastModifiedTime:       0xBDEF,
		lastModifiedDate:       0x52EC,
		crc32:                  crc32,
		compressedSize:         size,
		uncompressedSize:       size,
		fileNameLength:         uint16(len(name)),
		extraFieldLength:       32,
	}, []byte(name), zipExtraBytes
}

// will be set by init()
var zipExtraBytes []byte

func init() {
	/**
	Zip files can carry extra data in their file header fields.
	Those are usually things like timestamps or some unix permissions we don't really care about.
	Mostly XCode sends UT extras
	(https://commons.apache.org/proper/commons-compress/apidocs/org/apache/commons/compress/archivers/zip/X5455_ExtendedTimestamp.html)
	Since we only push data to the device and don't really care about correct timestamps or anything like that,
	I just dumped what XCode generates and always send the same extra.
	In this case I took a 0x5455 "UT" extra. Should it ever break, it'll be easy to fix.
	*/
	s := "55540D00 07F3A2EC 60F6A2EC 60F3A2EC 6075780B 000104F5 01000004 14000000"
	s = strings.ReplaceAll(s, " ", "")

	extra, err := hex.DecodeString(s)
	zipExtraBytes = extra
	if err != nil {
		log.Fatal("this is impossible to break", err)
	}
}

// zipHeader is pretty much the structure of a standard zip file header as can be found
// here f.ex. https://en.wikipedia.org/wiki/ZIP_(file_format)#Local_file_header
type zipHeader struct {
	signature              uint32
	version                uint16
	generalPurposeBitFlags uint16
	compressionMethod      uint16
	lastModifiedTime       uint16
	lastModifiedDate       uint16
	crc32                  uint32
	compressedSize         uint32
	uncompressedSize       uint32
	fileNameLength         uint16
	extraFieldLength       uint16
}

// standard header signature for central directory of a zip file
var centralDirectoryHeader []byte = []byte{0x50, 0x4b, 0x01, 0x02}
