package usbmux

import (
	"bytes"

	plist "howett.net/plist"
)

//ToPlist converts a given struct to a Plist using the awesome
//github.com/DHowett/go-plist library.
//It returns a string containing the plist.
func ToPlist(data interface{}) string {
	buf := &bytes.Buffer{}
	encoder := plist.NewEncoder(buf)
	encoder.Encode(data)
	return buf.String()
}
