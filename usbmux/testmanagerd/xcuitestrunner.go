package testmanagerd

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/house_arrest"
	"github.com/danielpaulus/go-ios/usbmux/installationproxy"
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
	version := usbmux.GetValues(device).Value.ProductVersion
	v, err := semver.NewVersion(version)
	testSessionID := uuid.New()
	testRunnerBundleID := bundleID + testBundleSuffix
	if err != nil {
		return err
	}
	installationProxy, err := installationproxy.New(device)
	defer installationProxy.Close()
	if err != nil {
		return err
	}
	apps, err := installationProxy.BrowseUserApps()
	if err != nil {
		return err
	}
	info, err := getAppInfos(bundleID, testRunnerBundleID, apps)
	if err != nil {
		return err
	}
	houseArrestService, err := house_arrest.New(device, testRunnerBundleID)
	defer houseArrestService.Close()
	if err != nil {
		return err
	}
	testConfigPath, err := createTestConfigOnDevice(testSessionID, info, houseArrestService)
	if err != nil {
		return err
	}

	log.Info(testConfigPath)
	log.Info(info)

	log.Info(v)
	return nil
}

func createTestConfigOnDevice(testSessionID uuid.UUID, info testInfo, houseArrestService *house_arrest.Connection) (string, error) {
	return "", nil
}

type testInfo struct {
	testrunnerAppPath  string
	testRunnerHomePath string
	appPath            string
}

func getAppInfos(bundleID string, testRunnerBundleID string, apps installationproxy.BrowseResponse) (testInfo, error) {
	info := testInfo{}

	for _, app := range apps.CurrentList {
		if app.CFBundleIdentifier == bundleID {
			info.appPath = app.Path
		}
		if app.CFBundleIdentifier == testRunnerBundleID {
			info.testrunnerAppPath = app.Path
			info.testRunnerHomePath = app.EnvironmentVariables["HOME"].(string)
		}
	}

	if info.appPath == "" {
		return testInfo{}, fmt.Errorf("Did not find AppInfo for '%s' on device. Is it installed?", bundleID)
	}
	if info.testRunnerHomePath == "" || info.testrunnerAppPath == "" {
		return testInfo{}, fmt.Errorf("Did not find AppInfo for '%s' on device. Is it installed?", testRunnerBundleID)
	}
	return info, nil
}
