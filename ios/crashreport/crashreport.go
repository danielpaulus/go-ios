package crashreport

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/house_arrest"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

const CRASH_REPORT_MOVER_SERVICE = "com.apple.crashreportmover"
const CRASH_REPORT_COPY_MOBILE_SERVICE = "com.apple.crashreportcopymobile"

func DownloadReports(device ios.DeviceEntry, pattern string) error {
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
	var filteredFiles []string
	for _, f := range files {
		matches, err := filepath.Match(pattern, f)
		if err != nil {
			log.Warn("error while matching pattern", err)
		}
		if matches {
			filteredFiles = append(filteredFiles, f)
		}
	}

	fmt.Printf("files:%+v", filteredFiles)
	/*info, err := afc.GetFileInfo("Analytics-Journal-Never-2021-12-05-010127.0002.ips.ca.synced")
	if err != nil {
		return err
	}
	fmt.Printf("info:%+v\n", info)
	info, err = afc.GetFileInfo("./Analytics-Journal-90Day-2021-12-28-010005.ips.ca.synced")
	if err != nil {
		return err
	}
	fmt.Printf("info:%+v\n", info)
	info, err = afc.GetFileInfo(".")
	if err != nil {
		return err
	}
	fmt.Printf("info:%+v\n", info)*/
	return nil
}

func ListReports(device ios.DeviceEntry, pattern string) ([]string, error) {
	err := moveReports(device)
	if err != nil {
		return []string{}, err
	}
	deviceConn, err := ios.ConnectToService(device, CRASH_REPORT_COPY_MOBILE_SERVICE)
	if err != nil {
		return []string{}, err
	}
	afc := house_arrest.NewFromConn(deviceConn)
	files, err := afc.ListFiles(".")
	if err != nil {
		return []string{}, err
	}
	var filteredFiles []string
	for _, f := range files {
		matches, err := filepath.Match(pattern, f)
		if err != nil {
			log.Warn("error while matching pattern", err)
		}
		if matches {
			filteredFiles = append(filteredFiles, f)
		}
	}

	return filteredFiles, nil
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
