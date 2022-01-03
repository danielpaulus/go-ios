//go:build !fast
// +build !fast

package testmanagerd_test

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/danielpaulus/go-ios/ios/zipconduit"
	log "github.com/sirupsen/logrus"
	stdlog "log"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const wdapath = "../../testdata/wda.ipa"
const wdaSignedPath = "../../testdata/wda-signed.ipa"

const signerPath = "../../testdata/app-signer-mac"

const wdaSuccessLogMessage = "ServerURLHere->"

func TestXcuiTest(t *testing.T) {
	patchLogger()
	device, err := ios.GetDevice("")
	if err != nil {
		t.Fatal(err)
	}
	err = imagemounter.FixDevImage(device, ".")
	if err != nil {
		t.Error(err)
		return
	}

	err = signAndInstall(device)
	if err != nil {
		t.Error(err)
		return
	}

	bundleID, testbundleID, xctestconfig := "com.facebook.WebDriverAgentRunner.xctrunner", "com.facebook.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest"
	var wdaargs []string
	var wdaenv []string
	go func() {
		err := testmanagerd.RunXCUIWithBundleIds(bundleID, testbundleID, xctestconfig, device, wdaargs, wdaenv)

		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed running WDA")
		}
	}()
	select {
	case <-time.After(time.Second * 5):
		t.Error("timeout")
		return
	case <-successChannel:
		log.Info("wda started successfully")
	}

	log.Infof("done")

	err = testmanagerd.CloseXCUITestRunner()
	if err != nil {
		log.Error("Failed closing wda-testrunner")
		t.Fail()
	}
}

func signAndInstall(device ios.DeviceEntry) error {
	svc, _ := installationproxy.New(device)
	response, err := svc.BrowseUserApps()
	for _, info := range response {
		if "com.facebook.WebDriverAgentRunner.xctrunner" == info.CFBundleIdentifier {
			log.Info("wda installed, skipping installation")
			return nil
		}
	}

	err = SignWda(device)
	if err != nil {
		return err
	}
	conn, err := zipconduit.New(device)
	if err != nil {
		return err
	}
	return conn.SendFile(wdaSignedPath)
}

func SignWda(device ios.DeviceEntry) error {
	cmd := exec.Command(signerPath,
		fmt.Sprintf("--udid=%s", device.Properties.SerialNumber),
		"--p12password=a",
		"--profilespath=../../testdata",
		fmt.Sprintf("--ipa=%s", wdapath),
		fmt.Sprintf("--output=%s", wdaSignedPath),
	)
	_, err := cmd.CombinedOutput()
	return err
}

func patchLogger() {
	successChannel = make(chan bool, 2)
	//log.SetLevel(log.DebugLevel)
	stdlog.SetOutput(new(LogrusWriter))
}

type LogrusWriter int

var successChannel chan bool

func (LogrusWriter) Write(data []byte) (int, error) {
	logmessage := string(data)
	if strings.Contains(logmessage, wdaSuccessLogMessage) {
		successChannel <- true
		return len(data), nil
	}
	log.Infof("gousb_logs:%s", logmessage)
	return len(data), nil
}
