package usbmux

import (
	"encoding/binary"

	log "github.com/sirupsen/logrus"
	plist "howett.net/plist"
)

//ToPlist converts a given struct to a Plist using the
//github.com/DHowett/go-plist library. Make sure your struct is exported.
//It returns a string containing the plist.
func ToPlist(data interface{}) string {
	bytes, err := plist.Marshal(data, plist.XMLFormat)
	if err != nil {
		log.Fatal("Failed converting to plist", data, err)
	}
	return string(bytes)
}

//Ntohs is a re-implementation of the C function Ntohs.
//it means networkorder to host oder and basically swaps
//the endianness of the given int.
//It returns port converted to little endian.
func Ntohs(port uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	return binary.LittleEndian.Uint16(buf)
}
