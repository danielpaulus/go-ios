package testmanagerd

import (
	"fmt"
	"path"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/house_arrest"
	"github.com/danielpaulus/go-ios/usbmux/installationproxy"
	"github.com/danielpaulus/go-ios/usbmux/instruments"
	"github.com/danielpaulus/go-ios/usbmux/nskeyedarchiver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type XCTestManager_IDEInterface interface{}
type XCTestManager_DaemonConnectionInterface interface{}

const ideToDaemonProxyChannelName = "dtxproxy:XCTestManager_IDEInterface:XCTestManager_DaemonConnectionInterface"

type dtxproxy struct {
	ideInterface     XCTestManager_IDEInterface
	daemonConnection XCTestManager_DaemonConnectionInterface
}

const testmanagerd = "com.apple.testmanagerd.lockdown"

const testBundleSuffix = "UITests.xctrunner"

/*func newDtxProxy(conn DtxConnection) dtxproxy {
	conn.requestChannelWithCodeAndIdentifier(1, "")
	return dtxproxy{}
}
*/

func RunXCUITest(bundleID string, device usbmux.DeviceEntry) error {
	v, xctestConfigPath, err := setupXcuiTest(device, bundleID)
	if err != nil {
		return err
	}
	log.Info(xctestConfigPath)

	log.Info(v)
	return nil
}

func startTestRunner(device usbmux.DeviceEntry, xctestConfigPath string, bundleID string) (uint64, error) {
	args := []interface{}{}
	env := map[string]interface{}{
		"XCTestConfigurationFilePath": xctestConfigPath,
	}
	opts := map[string]interface{}{
		"StartSuspendedKey": 0,
		"ActivateSuspended": 1,
	}

	return instruments.LaunchAppWithArgs(bundleID, device, args, env, opts)

}

func setupXcuiTest(device usbmux.DeviceEntry, bundleID string) (semver.Version, string, error) {
	version := usbmux.GetValues(device).Value.ProductVersion
	testSessionID := uuid.New()
	testRunnerBundleID := bundleID + testBundleSuffix
	v, err := semver.NewVersion(version)
	if err != nil {
		return semver.Version{}, "", err
	}
	installationProxy, err := installationproxy.New(device)
	defer installationProxy.Close()
	if err != nil {
		return semver.Version{}, "", err
	}
	apps, err := installationProxy.BrowseUserApps()
	if err != nil {
		return semver.Version{}, "", err
	}
	info, err := getAppInfos(bundleID, testRunnerBundleID, apps)
	if err != nil {
		return semver.Version{}, "", err
	}
	houseArrestService, err := house_arrest.New(device, testRunnerBundleID)
	defer houseArrestService.Close()
	if err != nil {
		return semver.Version{}, "", err
	}
	testConfigPath, err := createTestConfigOnDevice(testSessionID, info, houseArrestService)
	if err != nil {
		return semver.Version{}, "", err
	}

	return *v, testConfigPath, nil
}

func createTestConfigOnDevice(testSessionID uuid.UUID, info testInfo, houseArrestService *house_arrest.Connection) (string, error) {
	relativeXcTestConfigPath := path.Join("tmp", testSessionID.String()+".xctestconfiguration")
	xctestConfigPath := path.Join(info.testRunnerHomePath, relativeXcTestConfigPath)

	testBundleURL := path.Join(info.testrunnerAppPath, "PlugIns", info.targetAppBundleName+".xctest")

	config := nskeyedarchiver.NewXCTestConfiguration(info.targetAppBundleName, testSessionID, info.targetAppBundleID, info.targetAppPath, testBundleURL)
	result, err := nskeyedarchiver.ArchiveBin(config)
	if err != nil {
		return "", err
	}
	err = houseArrestService.SendFile(result, relativeXcTestConfigPath)
	if err != nil {
		return "", err
	}
	return xctestConfigPath, nil
}

type testInfo struct {
	testrunnerAppPath   string
	testRunnerHomePath  string
	targetAppPath       string
	targetAppBundleName string
	targetAppBundleID   string
}

func getAppInfos(bundleID string, testRunnerBundleID string, apps installationproxy.BrowseResponse) (testInfo, error) {
	info := testInfo{}

	for _, app := range apps.CurrentList {
		if app.CFBundleIdentifier == bundleID {
			info.targetAppPath = app.Path
			info.targetAppBundleName = app.CFBundleName
			info.targetAppBundleID = app.CFBundleIdentifier
		}
		if app.CFBundleIdentifier == testRunnerBundleID {
			info.testrunnerAppPath = app.Path
			info.testRunnerHomePath = app.EnvironmentVariables["HOME"].(string)
		}
	}

	if info.targetAppPath == "" {
		return testInfo{}, fmt.Errorf("Did not find AppInfo for '%s' on device. Is it installed?", bundleID)
	}
	if info.testRunnerHomePath == "" || info.testrunnerAppPath == "" {
		return testInfo{}, fmt.Errorf("Did not find AppInfo for '%s' on device. Is it installed?", testRunnerBundleID)
	}
	return info, nil
}
