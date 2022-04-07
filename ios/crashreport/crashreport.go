package crashreport

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/house_arrest"
	log "github.com/sirupsen/logrus"
)

const CRASH_REPORT_MOVER_SERVICE = "com.apple.crashreportmover"
const CRASH_REPORT_COPY_MOBILE_SERVICE = "com.apple.crashreportcopymobile"

func DownloadReports(device ios.DeviceEntry) error {
	err := moveReports(device)
	if err != nil {
		return err
	}
	deviceConn, err := ios.ConnectToService(device, CRASH_REPORT_COPY_MOBILE_SERVICE)
	if err != nil {
		return err
	}
	afc := house_arrest.NewFromConn(deviceConn)
	files, err := afc.ListFiles(".")
	if err != nil {
		return err
	}

	fmt.Printf("files:%+v", files)
	return nil
}

func moveReports(device ios.DeviceEntry) error {
	log.Debug("moving crashreports")
	conn, err := newMover(device)
	if err != nil {
		return err
	}
	log.Debug("connected to mover, awaiting ping")
	ping := make([]byte, 4)
	_, err = conn.deviceConn.Reader().Read(ping)
	if err != nil {
		return err
	}
	if "ping" != string(ping) {
		return fmt.Errorf("did not receive ping from crashreport mover: %x", ping)
	}
	log.Debug("ping received")
	return nil
}

type moverConnection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

//New returns a new ZipConduit Connection for the given DeviceID and Udid
func newMover(device ios.DeviceEntry) (*moverConnection, error) {
	deviceConn, err := ios.ConnectToService(device, CRASH_REPORT_MOVER_SERVICE)
	if err != nil {
		return &moverConnection{}, err
	}

	return &moverConnection{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
	}, nil
}
