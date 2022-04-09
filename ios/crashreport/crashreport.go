package crashreport

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/house_arrest"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

const crashReportMoverService = "com.apple.crashreportmover"
const crashReportCopyMobileService = "com.apple.crashreportcopymobile"

func DownloadReports(device ios.DeviceEntry, pattern string, targetdir string) error {
	err := moveReports(device)
	if err != nil {
		return err
	}
	deviceConn, err := ios.ConnectToService(device, crashReportCopyMobileService)
	if err != nil {
		return err
	}
	afc := house_arrest.NewFromConn(deviceConn)
	return copyReports(afc,".", pattern, targetdir)
}

func copyReports(afc *house_arrest.Connection, cwd string, pattern string, targetdir string) error {
	targetdirInfo, err := os.Stat(targetdir)
	if err != nil {
		return err
	}
	files, err := afc.ListFiles(cwd, pattern)
	if err != nil {
		return err
	}

	log.Debugf("files:%+v", files)
	for _, f := range files {
		if f == "." || f == ".." {
			continue
		}
		devicePath := path.Join(cwd, f)
		info, err := afc.GetFileInfo(devicePath)
		if err != nil {
			log.Warnf("failed getting info for file: %s", f)
		}
		if info.IsDir() {
			dir := path.Join(targetdir, f)
			err := os.Mkdir(dir, targetdirInfo.Mode().Perm())
			if err != nil {
				return err
			}
			err = copyReports(afc, devicePath,"*", dir)
			if err != nil {
				return err
			}
			continue
		}
		log.Debugf("%+v", info)
		fi, err := os.Create(path.Join(targetdir, f))
		if err != nil {
			panic(err)
		}
		err = afc.StreamFile(devicePath, fi)
		if err != nil {
			return err
		}
		err = fi.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func ListReports(device ios.DeviceEntry, pattern string) ([]string, error) {
	err := moveReports(device)
	if err != nil {
		return []string{}, err
	}
	deviceConn, err := ios.ConnectToService(device, crashReportCopyMobileService)
	if err != nil {
		return []string{}, err
	}
	afc := house_arrest.NewFromConn(deviceConn)
	return afc.ListFiles(".", pattern)
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

//NewWithHouseArrest returns a new ZipConduit Connection for the given DeviceID and Udid
func newMover(device ios.DeviceEntry) (*moverConnection, error) {
	deviceConn, err := ios.ConnectToService(device, crashReportMoverService)
	if err != nil {
		return &moverConnection{}, err
	}

	return &moverConnection{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
	}, nil
}
