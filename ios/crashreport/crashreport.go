package crashreport

import (
	"fmt"
	"os"
	"path"

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
	afc := afc.NewFromConn(deviceConn)
	return copyReports(afc, ".", pattern, targetdir)
}

func copyReports(afc *afc.Connection, cwd string, pattern string, targetDir string) error {
	log.WithFields(log.Fields{"dir": cwd, "pattern": pattern, "to": targetDir}).Info("downloading")
	targetDirInfo, err := os.Stat(targetDir)
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
		targetFilePath := path.Join(targetDir, f)
		log.WithFields(log.Fields{"from": devicePath, "to": targetFilePath}).Info("downloading")
		info, err := afc.Stat(devicePath)
		if err != nil {
			log.Warnf("failed getting info for file: %s, skipping", f)
			continue
		}
		log.Debugf("%+v", info)

		if info.IsDir() {
			err := os.Mkdir(targetFilePath, targetDirInfo.Mode().Perm())
			if err != nil {
				return err
			}
			err = copyReports(afc, devicePath, "*", targetFilePath)
			if err != nil {
				return err
			}
			continue
		}

		err = afc.PullSingleFile(devicePath, targetFilePath)
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{"from": devicePath, "to": targetFilePath}).Info("done")
	}
	return nil
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
	afc := afc.NewFromConn(deviceConn)
	files, err := afc.ListFiles(cwd, pattern)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f == "." || f == ".." {
			continue
		}
		log.WithFields(log.Fields{"path": path.Join(cwd, f)}).Info("delete")
		err := afc.Remove(path.Join(cwd, f))
		if err != nil {
			return err
		}
	}
	log.WithFields(log.Fields{"cwd": cwd, "pattern": pattern}).Info("done deleting")
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
	afc := afc.NewFromConn(deviceConn)
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
