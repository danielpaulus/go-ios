package codesign_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/danielpaulus/go-ios/ios/codesign"
	"github.com/stretchr/testify/assert"
)

func TestFindAppfolder(t *testing.T) {
	tempdir, appPath, notAnAppPath, appStoreAppPath, err := setUpExampleAppDir()
	defer os.RemoveAll(tempdir)
	foundAppPath, err := codesign.FindAppFolder(appPath)
	if assert.NoError(t, err) {
		assert.Contains(t, foundAppPath, tempdir)
	}

	foundAppPath, err = codesign.FindAppFolderVirtualDevice(path.Join(appPath, "Payload"))
	if assert.NoError(t, err) {
		assert.Contains(t, foundAppPath, tempdir)
	}

	_, err = codesign.FindAppFolder(notAnAppPath)
	assert.Error(t, err)

	assert.False(t, codesign.ContainsAppstoreBuild(appPath))
	assert.False(t, codesign.ContainsAppstoreBuild(notAnAppPath))
	assert.True(t, codesign.ContainsAppstoreBuild(appStoreAppPath))

}

func setUpExampleAppDir() (string, string, string, string, error) {
	tempdir, err := os.MkdirTemp("", "goios-findappfolder-test")
	if err != nil {
		return "", "", "", "", err
	}
	appPath := path.Join(tempdir, "app")
	err = os.MkdirAll(path.Join(appPath, "Payload", "test.app"), 0777)
	if err != nil {
		return "", "", "", "", err
	}
	notAnAppPath := path.Join(tempdir, "no-app")
	err = os.MkdirAll(path.Join(notAnAppPath, "ayload", "test.app"), 0777)
	if err != nil {
		return "", "", "", "", err
	}
	appStoreAppPath := path.Join(tempdir, "appstore")
	err = os.MkdirAll(path.Join(appStoreAppPath, "Payload", "test.app"), 0777)
	if err != nil {
		return "", "", "", "", err
	}

	err = ioutil.WriteFile(path.Join(appPath, "Payload", "test.app", codesign.EmbeddedProfileName), []byte("example file"), 777)
	if err != nil {
		return "", "", "", "", err
	}

	return tempdir, appPath, notAnAppPath, appStoreAppPath, nil
}
