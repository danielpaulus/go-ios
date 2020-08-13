package usbmux

import (
	"bytes"
	"encoding/binary"

	plist "howett.net/plist"
)

//ToPlist converts a given struct to a Plist using the
//github.com/DHowett/go-plist library. Make sure your struct is exported.
//It returns a string containing the plist.
func ToPlist(data interface{}) string {
	buf := &bytes.Buffer{}
	encoder := plist.NewEncoder(buf)
	encoder.Encode(data)
	return buf.String()
}

//ntohs is a re-implementation of the C function ntohs.
//it means networkorder to host oder and basically swaps
//the endianness of the given int.
//It returns port converted to little endian.
func ntohs(port uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	return binary.LittleEndian.Uint16(buf)
}
