package codesign

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// EmbeddedProfileName contains the default name for the
// embedded.mobileprovision profile in all developer apps
const EmbeddedProfileName = "embedded.mobileprovision"

// ContainsAppstoreBuild for a given root dir,
// returns true if /Payload/*.app/embedded.mobileprovision exists and false otherwise.
func ContainsAppstoreBuild(root string) bool {
	appFolder, err := FindAppFolder(root)
	if err != nil {
		return false
	}
	embeddedProfile := path.Join(appFolder, EmbeddedProfileName)
	_, err = os.Stat(embeddedProfile)
	if err != nil {
		return true
	}
	return false
}

// FindAppFolder returns the path of the /Payload/*.app directory
// or an error if there is no .app directory or more than one.
func FindAppFolder(rootDir string) (string, error) {
	appFolders, err := filepath.Glob(path.Join(rootDir, "Payload", "*.app"))
	if err != nil {
		return "", err
	}
	if len(appFolders) != 1 {
		return "", fmt.Errorf("found more or less than exactly one app folder: %+v", appFolders)
	}
	return appFolders[0], nil
}

// FindAppFolderVirtualDevice returns the path of the *.app directory
// which must be in the root of the unzipped file.
// or an error if there is no .app directory or more than one.
func FindAppFolderVirtualDevice(rootDir string) (string, error) {
	appFolders, err := filepath.Glob(path.Join(rootDir, "*.app"))
	if err != nil {
		return "", err
	}
	if len(appFolders) != 1 {
		return "", fmt.Errorf("found more or less than exactly one app folder: %+v", appFolders)
	}
	return appFolders[0], nil
}
