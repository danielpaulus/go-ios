package crashreport

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	log "github.com/sirupsen/logrus"
)

const (
	crashReportMoverService      = "com.apple.crashreportmover"
	crashReportCopyMobileService = "com.apple.crashreportcopymobile"
)

// DownloadReports gets all crashreports based on the provided file pattern and writes them to targetdir.
// Directories will be recursively added without applying the pattern recursively.
// pattern can be typical filepattern, if you want all files use "*"
func DownloadReports(device ios.DeviceEntry, pattern string, targetdir string) error {
	if pattern == "" {
		return fmt.Errorf("empty pattern not ok, just use *")
	}
	err := moveReports(device)
	if err != nil {
		return err
	}
	deviceConn, err := ios.ConnectToService(device, crashReportCopyMobileService)
	if err != nil {
		return err
	}
	afcConn := afc.NewFromConn(deviceConn)
	err = afcConn.WalkDir(".", func(p string, info afc.FileInfo, err error) error {
		if info.Type == afc.S_IFDIR {
			return nil
		}
		if ok, _ := filepath.Match(pattern, filepath.Base(p)); !ok {
			return nil
		}
		return afcConn.PullSingleFile(p, path.Join(targetdir, filepath.Base(p)))
	})
	return err
}

func RemoveReports(device ios.DeviceEntry, cwd string, pattern string) error {
	if pattern == "" {
		return fmt.Errorf("empty pattern not ok, just use *")
	}
	log.WithFields(log.Fields{"cwd": cwd, "pattern": pattern}).Info("deleting")
	err := moveReports(device)
	if err != nil {
		return err
	}
	deviceConn, err := ios.ConnectToService(device, crashReportCopyMobileService)
	if err != nil {
		return err
	}
	afcClient := afc.NewFromConn(deviceConn)
	return afcClient.WalkDir(cwd, func(path string, info afc.FileInfo, err error) error {
		if info.Type == afc.S_IFDIR {
			return nil
		}
		if ok, _ := filepath.Match(pattern, filepath.Base(path)); !ok {
			return nil
		}
		return afcClient.Remove(path)
	})
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
	afcClient := afc.NewFromConn(deviceConn)

	var files []string
	err = afcClient.WalkDir(".", func(path string, info afc.FileInfo, err error) error {
		if info.Type == afc.S_IFDIR {
			return nil
		}
		if ok, _ := filepath.Match(pattern, filepath.Base(path)); !ok {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return files, nil
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

// NewWithHouseArrest returns a new ZipConduit Connection for the given DeviceID and Udid
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
