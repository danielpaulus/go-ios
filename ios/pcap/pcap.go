package pcap

import (
	"encoding/hex"
	"log"

	"github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
)

func Start(device ios.DeviceEntry) {
	intf, err := ios.ConnectToService(device, "com.apple.pcapd")
	if err != nil {
		log.Fatal(err)
	}
	plistCodec := ios.NewPlistCodec()
	for {
		b, _ := plistCodec.Decode(intf.Reader())

		println(hex.Dump(fromBytes(b)))
	}
}

func fromBytes(data []byte) []byte {
	var result []byte
	plist.Unmarshal(data, &result)
	return result
}
