package pcap

import (
	"encoding/hex"

	"github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
)

func Start(device ios.DeviceEntry) error {
	intf, err := ios.ConnectToService(device, "com.apple.pcapd")
	if err != nil {
		return err
	}
	plistCodec := ios.NewPlistCodec()
	for {
		b, err := plistCodec.Decode(intf.Reader())
		if err != nil {
			return err
		}
		decodedBytes, err := fromBytes(b)
		if err != nil {
			return err
		}
		println(hex.Dump(decodedBytes))
	}
}

func fromBytes(data []byte) ([]byte, error) {
	var result []byte
	_, err := plist.Unmarshal(data, &result)
	return result, err
}
