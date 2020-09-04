package testmanagerd

import (
	"fmt"
	"path"

	"github.com/Masterminds/semver"
	ios "github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/house_arrest"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type XCTestManager_IDEInterface struct {
	IDEDaemonProxy         *dtx.Channel
	testBundleReadyChannel chan dtx.Message
}
type XCTestManager_DaemonConnectionInterface struct {
	IDEDaemonProxy *dtx.Channel
}

func (xide XCTestManager_IDEInterface) testBundleReady() (uint64, uint64) {
	msg := <-xide.testBundleReadyChannel
	protocolVersion, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	minimalVersion, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[1].([]byte))
	return protocolVersion[0].(uint64), minimalVersion[0].(uint64)
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateSessionWithIdentifier(sessionIdentifier uuid.UUID, protocolVersion uint64) (uint64, error) {
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName}).Debug("Launching init test Session")
	rply, err := xdc.IDEDaemonProxy.MethodCall(
		"_IDE_initiateSessionWithIdentifier:forClient:atPath:protocolVersion:",
		nskeyedarchiver.NewNSUUID(sessionIdentifier),
		"thephonedoesntcarewhatisendhereitseems",
		"/Applications/Xcode.app",
		protocolVersion)

	returnValue := rply.Payload[0]
	var val uint64
	var ok bool
	if val, ok = returnValue.(uint64); !ok {
		return 0, fmt.Errorf("initiateSessionWithIdentifier got wrong returnvalue: %s", rply.Payload)
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("init test session reply")

	return val, err

}

func (xdc XCTestManager_DaemonConnectionInterface) initiateControlSessionForTestProcessID(pid uint64, protocolVersion uint64) error {
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_initiateControlSessionForTestProcessID:protocolVersion:", pid, protocolVersion)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("initiateControlSessionForTestProcessID reply")
	return nil
}

func startExecutingTestPlanWithProtocolVersion(channel *dtx.Channel, protocolVersion uint64) error {
	rply, err := channel.MethodCall("_IDE_startExecutingTestPlanWithProtocolVersion:", protocolVersion)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("_IDE_startExecutingTestPlanWithProtocolVersion reply")
	return nil
}

const ideToDaemonProxyChannelName = "dtxproxy:XCTestManager_IDEInterface:XCTestManager_DaemonConnectionInterface"

type dtxproxy struct {
	ideInterface     XCTestManager_IDEInterface
	daemonConnection XCTestManager_DaemonConnectionInterface
	IDEDaemonProxy   *dtx.Channel
	dtxConnection    *dtx.Connection
}

type ProxyDispatcher struct {
	testBundleReadyChannel chan dtx.Message
}

func (p ProxyDispatcher) Dispatch(m dtx.Message) {
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)
		switch method {
		case "_XCT_testBundleReadyWithProtocolVersion:minimumVersion:":
			p.testBundleReadyChannel <- m
			return
		default:
			log.Warnf("Method invocation not implement for selector:%s", method)
		}
	}
	log.Debugf("dispatcher received: %s", m.Payload[0])
}

func newDtxProxy(dtxConnection *dtx.Connection) dtxproxy {
	testBundleReadyChannel := make(chan dtx.Message, 1)
	IDEDaemonProxy := dtxConnection.RequestChannelIdentifier(ideToDaemonProxyChannelName, ProxyDispatcher{testBundleReadyChannel})
	return dtxproxy{IDEDaemonProxy: IDEDaemonProxy,
		ideInterface:     XCTestManager_IDEInterface{IDEDaemonProxy, testBundleReadyChannel},
		daemonConnection: XCTestManager_DaemonConnectionInterface{IDEDaemonProxy},
		dtxConnection:    dtxConnection,
	}
}

const testmanagerd = "com.apple.testmanagerd.lockdown"

const testBundleSuffix = "UITests.xctrunner"

/*func newDtxProxy(conn DtxConnection) dtxproxy {
	conn.requestChannelWithCodeAndIdentifier(1, "")
	return dtxproxy{}
}
*/

func RunWDA(device ios.DeviceEntry) error {

	return runXCUIWithBundleIds("com.facebook.WebDriverAgentRunner.xctrunner", "com.facebook.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest", device)
}

func RunXCUITest(bundleID string, device ios.DeviceEntry) error {
	testRunnerBundleID := bundleID + testBundleSuffix
	return runXCUIWithBundleIds(bundleID, testRunnerBundleID, "", device)
}

var closeChan = make(chan interface{})

func runXCUIWithBundleIds(bundleID string, testRunnerBundleID string, xctestConfigFileName string, device ios.DeviceEntry) error {
	testSessionId, _, xctestConfigPath, err := setupXcuiTest(device, bundleID, testRunnerBundleID, xctestConfigFileName)
	if err != nil {
		return err
	}
	conn, _ := dtx.NewConnection(device, testmanagerd)
	defer conn.Close()
	ideDaemonProxy := newDtxProxy(conn)
	protocolVersion, err := ideDaemonProxy.daemonConnection.initiateSessionWithIdentifier(testSessionId, 29)
	log.Infof("ProtocolVersion:%d", protocolVersion)
	if err != nil {
		return err
	}
	pControl, err := instruments.NewProcessControl(device)
	defer pControl.Close()
	if err != nil {
		return err
	}
	pid, err := startTestRunner(pControl, xctestConfigPath, testRunnerBundleID)
	if err != nil {
		return err
	}
	log.Infof("Runner started with pid:%d, waiting for testBundleReady", pid)
	protocolVersion, minimalVersion := ideDaemonProxy.ideInterface.testBundleReady()
	channel := ideDaemonProxy.dtxConnection.ForChannelRequest(ProxyDispatcher{})

	log.Infof("ProtocolVersion:%d MinimalVersion:%d", protocolVersion, minimalVersion)
	conn2, _ := dtx.NewConnection(device, testmanagerd)
	defer conn2.Close()
	ideDaemonProxy2 := newDtxProxy(conn2)
	ideDaemonProxy2.daemonConnection.initiateControlSessionForTestProcessID(pid, protocolVersion)
	startExecutingTestPlanWithProtocolVersion(channel, protocolVersion)
	<-closeChan
	log.Infof("Killing WDA Runner pid %d ...", pid)
	err = pControl.KillProcess(pid)
	if err != nil {
		return err
	}
	log.Info("runner killed with success")
	return nil

}

func CloseXCUITestRunner() {
	var signal interface{}
	closeChan <- signal
}

func startTestRunner(pControl *instruments.ProcessControl, xctestConfigPath string, bundleID string) (uint64, error) {
	args := []interface{}{}
	env := map[string]interface{}{
		"XCTestConfigurationFilePath": xctestConfigPath,
	}
	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"ActivateSuspended": uint64(1),
	}

	return pControl.StartProcess(bundleID, env, args, opts)

}

func setupXcuiTest(device ios.DeviceEntry, bundleID string, testRunnerBundleID string, xctestConfigFileName string) (uuid.UUID, semver.Version, string, error) {
	version := ios.GetValues(device).Value.ProductVersion
	testSessionID := uuid.New()

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
	testConfigPath, err := createTestConfigOnDevice(testSessionID, info, houseArrestService, xctestConfigFileName)
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", err
	}

	return testSessionID, *v, testConfigPath, nil
}

func createTestConfigOnDevice(testSessionID uuid.UUID, info testInfo, houseArrestService *house_arrest.Connection, xctestConfigFileName string) (string, error) {
	relativeXcTestConfigPath := path.Join("tmp", testSessionID.String()+".xctestconfiguration")
	xctestConfigPath := path.Join(info.testRunnerHomePath, relativeXcTestConfigPath)

	if xctestConfigFileName == "" {
		xctestConfigFileName = info.targetAppBundleName + "UITests.xctest"
	}
	testBundleURL := path.Join(info.testrunnerAppPath, "PlugIns", xctestConfigFileName)

	config := nskeyedarchiver.NewXCTestConfiguration(info.targetAppBundleName, testSessionID, info.targetAppBundleID, info.targetAppPath, testBundleURL)
	result, err := nskeyedarchiver.ArchiveXML(config)
	if err != nil {
		return "", err
	}

	err = houseArrestService.SendFile([]byte(result), relativeXcTestConfigPath)
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

func getAppInfos(bundleID string, testRunnerBundleID string, apps []installationproxy.AppInfo) (testInfo, error) {
	info := testInfo{}
	for _, app := range apps {
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
