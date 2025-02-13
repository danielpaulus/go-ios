package codesign_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/danielpaulus/go-ios/ios/codesign"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIpaFile(t *testing.T) {
	ipa := readBytes("fixtures/wda.ipa")

	readerAt := bytes.NewReader(ipa)
	_, directory, err := codesign.ExtractIpa(readerAt, int64(len(ipa)))
	if err != nil {
		log.Fatalf("failed extracting: %+v", err)
	}
	defer os.RemoveAll(directory)
	appdir, _ := codesign.FindAppFolder(directory)
	var expectedBundleId = "com.facebook.WebDriverAgentRunner.xctrunner"

	extractedBundleId, err := codesign.GetBundleIdentifier(appdir)

	assert.NoError(t, err)
	assert.Equal(t, expectedBundleId, extractedBundleId)
}
