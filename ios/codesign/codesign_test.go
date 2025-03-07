package codesign_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios/codesign"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestCodeSign tests the resigning process end to end.
// The ipa will be extracted, signed, zipped and in case
// the environment variable udid is specified, installed to a device.
func TestCodeSign(t *testing.T) {

	ipa := readBytes("fixtures/wda.ipa")

	workspace, cleanup, err := makeWorkspace()
	if err != nil {
		log.Errorf("failed creating workspace: %+v", err)
		t.Fail()
		return
	}
	defer cleanup()

	readerAt := bytes.NewReader(ipa)
	duration, directory, err := codesign.ExtractIpa(readerAt, int64(len(ipa)))
	if err != nil {
		log.Errorf("failed extracting: %+v", err)
		t.Fail()
		return
	}
	log.Infof("Extraction took:%v", duration)
	defer os.RemoveAll(directory)

	index := 0
	if udid, yes := runOnRealDevice(); yes {
		index, err = findProfile(udid)
		if err != nil {
			log.Errorf("failed finding profile: %+v", err)
			t.Fail()
			return
		}
	}
	signingConfig := workspace.GetConfig(index)

	startSigning := time.Now()
	err = codesign.Sign(directory, signingConfig)
	assert.NoError(t, err)
	durationSigning := time.Since(startSigning)
	log.Infof("signing took: %v", durationSigning)

	b := &bytes.Buffer{}

	assert.NoError(t, codesign.Verify(path.Join(directory, "Payload", "WebDriverAgentRunner-Runner.app")))

	compressStart := time.Now()
	err = codesign.CompressToIpa(directory, b)
	if err != nil {
		log.Errorf("Compression failed with %+v", err)
		t.Fail()
		return
	}
	compressDuration := time.Since(compressStart)
	log.Infof("compressiontook: %v", compressDuration)

	if udid, yes := runOnRealDevice(); yes {
		installOnRealDevice(udid, b.Bytes())
	} else {
		log.Warn("No UDID provided, not running installation on actual device")
	}
}

func runOnRealDevice() (string, bool) {
	udid := os.Getenv("udid")
	return udid, udid != ""
}

func makeWorkspace() (codesign.SigningWorkspace, func(), error) {
	dir, err := os.MkdirTemp("", "sign-test")
	if err != nil {
		return codesign.SigningWorkspace{}, nil, err
	}

	workspace := codesign.NewSigningWorkspace(dir)
	workspace.PrepareProfiles("../provisioningprofiles")
	workspace.PrepareKeychain("test.keychain")

	cleanUp := func() {
		defer os.RemoveAll(dir)
		defer workspace.Close()
	}
	return workspace, cleanUp, nil
}

func findProfile(udid string) (int, error) {
	profiles, err := codesign.ParseProfiles("../provisioningprofiles")
	if err != nil {
		return -1, fmt.Errorf("could not parse profiles %+v", err)
	}
	index := codesign.FindProfileForDevice(udid, profiles)
	if index == -1 {
		return -1, fmt.Errorf("Device: %s is not in profiles", udid)
	}
	return index, nil
}

func installOnRealDevice(udid string, ipa []byte) {
	ipafile, err := os.CreateTemp("", "myname-*.ipa")
	if err != nil {
		log.Error(err)
	}
	defer os.Remove(ipafile.Name())

	ipafile.Write(ipa)
	ipafile.Close()

	installerlogs, err := exec.Command("ios", "install", ipafile.Name(), "--udid="+udid).CombinedOutput()
	if err != nil {
		log.Errorf("failed installing, logs: %s with err %+v", string(installerlogs), err)
	}
	log.Info("Install successful")
}
