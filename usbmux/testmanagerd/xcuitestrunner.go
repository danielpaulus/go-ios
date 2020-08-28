package testmanagerd

import (
	"fmt"
	"path"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/usbmux"
	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	"github.com/danielpaulus/go-ios/usbmux/house_arrest"
	"github.com/danielpaulus/go-ios/usbmux/installationproxy"
	"github.com/danielpaulus/go-ios/usbmux/instruments"
	"github.com/danielpaulus/go-ios/usbmux/nskeyedarchiver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type XCTestManager_IDEInterface struct {
	IDEDaemonProxy dtx.DtxChannel
}
type XCTestManager_DaemonConnectionInterface struct {
	IDEDaemonProxy dtx.DtxChannel
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateSessionWithIdentifier(sessionIdentifier uuid.UUID, protocolVersion uint64) error {
	const objcMethodName = "_IDE_initiateSessionWithIdentifier:forClient:atPath:protocolVersion:"
	payload, _ := nskeyedarchiver.ArchiveBin(objcMethodName)
	auxiliary := dtx.NewDtxPrimitiveDictionary()

	auxiliary.AddNsKeyedArchivedObject(nskeyedarchiver.NewNSUUID(sessionIdentifier))
	auxiliary.AddNsKeyedArchivedObject("D35B1EB7-7969-40A3-9078-EAB51B743DC9-27146-0000674FDBFF842E")
	auxiliary.AddNsKeyedArchivedObject("/Applications/Xcode.app")
	auxiliary.AddNsKeyedArchivedObject(protocolVersion)
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName}).Info("Launching init test Session")
	rply, err := xdc.IDEDaemonProxy.SendAndAwaitReply(true, dtx.MethodinvocationWithoutExpectedReply, payload, auxiliary)
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Info("init test session reply")

	return err
}

const ideToDaemonProxyChannelName = "dtxproxy:XCTestManager_IDEInterface:XCTestManager_DaemonConnectionInterface"

type dtxproxy struct {
	ideInterface     XCTestManager_IDEInterface
	daemonConnection XCTestManager_DaemonConnectionInterface
	IDEDaemonProxy   dtx.DtxChannel
}

type ProxyDispatcher struct{}

func (p ProxyDispatcher) Dispatch(m dtx.DtxMessage) {
	log.Infof("dispatcher received: %s", m.Payload[0])
}

func newDtxProxy(dtxConnection *dtx.DtxConnection) dtxproxy {
	IDEDaemonProxy := dtxConnection.RequestChannelIdentifier(ideToDaemonProxyChannelName, ProxyDispatcher{})
	return dtxproxy{IDEDaemonProxy: IDEDaemonProxy,
		ideInterface:     XCTestManager_IDEInterface{IDEDaemonProxy},
		daemonConnection: XCTestManager_DaemonConnectionInterface{IDEDaemonProxy},
	}
}

const testmanagerd = "com.apple.testmanagerd.lockdown"

const testBundleSuffix = "UITests.xctrunner"

/*func newDtxProxy(conn DtxConnection) dtxproxy {
	conn.requestChannelWithCodeAndIdentifier(1, "")
	return dtxproxy{}
}
*/

func RunXCUITest(bundleID string, device usbmux.DeviceEntry) error {
	testSessionId, v, xctestConfigPath, err := setupXcuiTest(device, bundleID)
	if err != nil {
		return err
	}
	conn, _ := dtx.NewDtxConnection(device.DeviceID, device.Properties.SerialNumber, testmanagerd)
	defer conn.Close()
	ideDaemonProxy := newDtxProxy(conn)
	err = ideDaemonProxy.daemonConnection.initiateSessionWithIdentifier(testSessionId, 29)
	if err != nil {
		return err
	}
	pid, err := startTestRunner(device, xctestConfigPath, bundleID+testBundleSuffix)
	if err != nil {
		return err
	}
	log.Info("Runner started with pid:%d", pid)
	log.Info(xctestConfigPath)

	log.Info(v)
	for {
	}
	return nil
}

func startTestRunner(device usbmux.DeviceEntry, xctestConfigPath string, bundleID string) (uint64, error) {
	args := []interface{}{}
	env := map[string]interface{}{
		"XCTestConfigurationFilePath": xctestConfigPath,
	}
	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"ActivateSuspended": uint64(1),
	}

	return instruments.LaunchAppWithArgs(bundleID, device, args, env, opts)

}

func setupXcuiTest(device usbmux.DeviceEntry, bundleID string) (uuid.UUID, semver.Version, string, error) {
	version := usbmux.GetValues(device).Value.ProductVersion
	testSessionID := uuid.New()
	testRunnerBundleID := bundleID + testBundleSuffix
	v, err := semver.NewVersion(version)
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", err
	}
	installationProxy, err := installationproxy.New(device)
	defer installationProxy.Close()
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", err
	}
	apps, err := installationProxy.BrowseUserApps()
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", err
	}
	info, err := getAppInfos(bundleID, testRunnerBundleID, apps)
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", err
	}
	houseArrestService, err := house_arrest.New(device, testRunnerBundleID)
	defer houseArrestService.Close()
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", err
	}
	testConfigPath, err := createTestConfigOnDevice(testSessionID, info, houseArrestService)
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", err
	}

	return testSessionID, *v, testConfigPath, nil
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
