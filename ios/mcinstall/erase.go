package mcinstall

import "github.com/danielpaulus/go-ios/ios"

func Erase(device ios.DeviceEntry) error {
	conn, err := New(device)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
