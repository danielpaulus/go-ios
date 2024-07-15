package codesign_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/danielpaulus/go-ios/ios/codesign"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPhysicalDevice(t *testing.T) {
	ipa := readBytes("../codesign/fixtures/wda.ipa")

	readerAt := bytes.NewReader(ipa)
	_, directory, err := codesign.ExtractIpa(readerAt, int64(len(ipa)))
	if err != nil {
		log.Fatalf("failed extracting: %+v", err)
	}
	defer os.RemoveAll(directory)
	appdir, _ := codesign.FindAppFolder(directory)
	expectedArchs := []string{"armv7", "armv7s", "arm64"}

	extractedArchs, err := codesign.ExtractArchitectures(appdir)

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedArchs, extractedArchs)
	assert.False(t, codesign.IsSimulatorApp(extractedArchs))
}

func TestLipoCheck(t *testing.T) {
	assert.NoError(t, codesign.CheckLipo())
}

func readBytes(name string) []byte {
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return data
}
