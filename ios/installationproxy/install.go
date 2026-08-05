package installationproxy

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/golog"
)

// stagingPath is the directory relative to the AFC root (/var/mobile/Media)
// where packages are uploaded before installation.
const stagingPath = "PublicStaging"

// InstallIpa uploads the given .ipa file to the PublicStaging directory on the
// device using AFC and then installs it by sending an Install command to the
// installation_proxy service. This is an alternative to the zipconduit based
// installation: the .ipa is transferred as-is (compressed, without unpacking it
// on the host) and unpacked and verified on the device.
func InstallIpa(device ios.DeviceEntry, ipaPath string) error {
	info, err := os.Stat(ipaPath)
	if err != nil {
		return fmt.Errorf("installproxy: could not read package %s: %w", ipaPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("installproxy: %s is a directory, the installproxy method only supports .ipa files. Use the default zipconduit method for .app folders", ipaPath)
	}
	remotePath := path.Join(stagingPath, filepath.Base(ipaPath))

	afcClient, err := afc.New(device)
	if err != nil {
		return fmt.Errorf("installproxy: failed connecting to AFC: %w", err)
	}
	defer afcClient.Close()
	if _, err := afcClient.Stat(stagingPath); err != nil {
		if err := afcClient.MkDir(stagingPath); err != nil {
			return fmt.Errorf("installproxy: failed creating %s: %w", stagingPath, err)
		}
	}
	f, err := os.Open(ipaPath)
	if err != nil {
		return fmt.Errorf("installproxy: failed opening %s: %w", ipaPath, err)
	}
	defer f.Close()
	golog.Info("uploading package to device", "module", logModule, "udid", device.Properties.SerialNumber, "path", ipaPath, "remotePath", remotePath, "size", info.Size())
	err = afcClient.WriteToFile(f, remotePath)
	if err != nil {
		return fmt.Errorf("installproxy: failed uploading %s to %s: %w", ipaPath, remotePath, err)
	}
	golog.Info("upload complete, starting installation", "module", logModule, "udid", device.Properties.SerialNumber, "remotePath", remotePath)

	conn, err := New(device)
	if err != nil {
		return fmt.Errorf("installproxy: failed connecting to installation_proxy: %w", err)
	}
	defer conn.Close()
	return conn.Install(remotePath, nil)
}

// Install sends the Install command for a package that was previously staged on
// the device (packagePath is relative to the AFC root, e.g.
// "PublicStaging/app.ipa") and blocks streaming PercentComplete progress until
// the installation completes or fails.
func (c *Connection) Install(packagePath string, options map[string]interface{}) error {
	b, err := c.plistCodec.Encode(installCommand(packagePath, options))
	if err != nil {
		return err
	}
	err = c.deviceConn.Send(b)
	if err != nil {
		return err
	}
	for {
		response, err := c.plistCodec.Decode(c.deviceConn.Reader())
		if err != nil {
			return err
		}
		dict, err := ios.ParsePlist(response)
		if err != nil {
			return err
		}
		done, percent, status, err := evaluateInstallProgress(dict)
		if err != nil {
			return err
		}
		if done {
			golog.Info("done installing", "module", logModule, "packagePath", packagePath)
			return nil
		}
		golog.Info("install status", "module", logModule, "packagePath", packagePath, "status", status, "percentComplete", percent)
	}
}

// installCommand builds the installation_proxy Install request. ClientOptions
// is always present, an empty dict when no options are given.
func installCommand(packagePath string, options map[string]interface{}) map[string]interface{} {
	if options == nil {
		options = map[string]interface{}{}
	}
	return map[string]interface{}{
		"Command":       "Install",
		"PackagePath":   packagePath,
		"ClientOptions": options,
	}
}

// evaluateInstallProgress parses a single progress update sent by
// installation_proxy while an Install command is running.
func evaluateInstallProgress(dict map[string]interface{}) (done bool, percent int, status string, err error) {
	if errValue, ok := dict["Error"]; ok {
		if description, ok := dict["ErrorDescription"]; ok {
			return false, 0, "", fmt.Errorf("received install error: %v errorDescription: %v", errValue, description)
		}
		return false, 0, "", fmt.Errorf("received install error: %v", errValue)
	}
	statusValue, ok := dict["Status"].(string)
	if !ok {
		return false, 0, "", fmt.Errorf("unknown install status update: %+v", dict)
	}
	if statusValue == "Complete" {
		return true, 100, statusValue, nil
	}
	return false, toPercent(dict["PercentComplete"]), statusValue, nil
}

// toPercent converts a PercentComplete plist value, plists decode integers as
// uint64 or int64 depending on encoding and sign.
func toPercent(v interface{}) int {
	switch n := v.(type) {
	case uint64:
		return int(n)
	case int64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
