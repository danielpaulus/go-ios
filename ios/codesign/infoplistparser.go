package codesign

import (
	"fmt"
	"os"
	"path"

	"howett.net/plist"
)

const bundleIdentifierKey = "CFBundleIdentifier"
const infoPlist = "Info.plist"

// GetBundleIdentifier takes a directory where it can find a Info.plist, reads the info.plist and will return the Bundle Identifer of the app
func GetBundleIdentifier(binDir string) (string, error) {
	plistBytes, err := os.ReadFile(path.Join(binDir, infoPlist))
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	_, err = plist.Unmarshal(plistBytes, &data)
	if err != nil {
		return "", err
	}
	if val, ok := data[bundleIdentifierKey]; ok {
		return val.(string), nil
	}
	return "", fmt.Errorf("%s not in Info.plist: %+v", bundleIdentifierKey, data)
}
