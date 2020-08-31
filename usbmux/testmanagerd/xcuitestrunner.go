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
	IDEDaemonProxy         dtx.DtxChannel
	testBundleReadyChannel chan dtx.DtxMessage
}
type XCTestManager_DaemonConnectionInterface struct {
	IDEDaemonProxy dtx.DtxChannel
}

func (xide XCTestManager_IDEInterface) testBundleReady() (uint64, uint64) {
	msg := <-xide.testBundleReadyChannel
	protocolVersion, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
	minimalVersion, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[1].([]byte))
	return protocolVersion[0].(uint64), minimalVersion[0].(uint64)
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateSessionWithIdentifier(sessionIdentifier uuid.UUID, protocolVersion uint64) (uint64, error) {
	const objcMethodName = "_IDE_initiateSessionWithIdentifier:forClient:atPath:protocolVersion:"
	payload, _ := nskeyedarchiver.ArchiveBin(objcMethodName)
	auxiliary := dtx.NewDtxPrimitiveDictionary()

	auxiliary.AddNsKeyedArchivedObject(nskeyedarchiver.NewNSUUID(sessionIdentifier))
	auxiliary.AddNsKeyedArchivedObject("thephonedoesntcarewhatisendhereitseems")
	auxiliary.AddNsKeyedArchivedObject("/Applications/Xcode.app")
	auxiliary.AddNsKeyedArchivedObject(protocolVersion)
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName}).Info("Launching init test Session")
	rply, err := xdc.IDEDaemonProxy.SendAndAwaitReply(true, dtx.Methodinvocation, payload, auxiliary)
	returnValue := rply.Payload[0]
	if val, ok := returnValue.(uint64); !ok {
		return 0, fmt.Errorf("%s got wrong returnvalue: %s", objcMethodName, rply.Payload)
	} else {
		log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Info("init test session reply")

		return val, err
	}
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateControlSessionForTestProcessID(pid uint64, protocolVersion uint64) error {
	const objcMethodName = "_IDE_initiateControlSessionForTestProcessID:protocolVersion:"
	payload, _ := nskeyedarchiver.ArchiveBin(objcMethodName)
	auxiliary := dtx.NewDtxPrimitiveDictionary()
	auxiliary.AddNsKeyedArchivedObject(pid)
	auxiliary.AddNsKeyedArchivedObject(protocolVersion)
	rply, err := xdc.IDEDaemonProxy.SendAndAwaitReply(true, dtx.Methodinvocation, payload, auxiliary)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Info("initiateControlSessionForTestProcessID reply")
	return nil
}

func startExecutingTestPlanWithProtocolVersion(channel dtx.DtxChannel, protocolVersion uint64) error {
	const objcMethodName = "_IDE_startExecutingTestPlanWithProtocolVersion:"
	payload, _ := nskeyedarchiver.ArchiveBin(objcMethodName)
	auxiliary := dtx.NewDtxPrimitiveDictionary()
	auxiliary.AddNsKeyedArchivedObject(protocolVersion)
	rply, err := channel.SendAndAwaitReply(true, dtx.Methodinvocation, payload, auxiliary)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Info("_IDE_startExecutingTestPlanWithProtocolVersion reply")
	return nil
}

const ideToDaemonProxyChannelName = "dtxproxy:XCTestManager_IDEInterface:XCTestManager_DaemonConnectionInterface"

type dtxproxy struct {
	ideInterface     XCTestManager_IDEInterface
	daemonConnection XCTestManager_DaemonConnectionInterface
	IDEDaemonProxy   dtx.DtxChannel
	dtxConnection    *dtx.DtxConnection
}

type ProxyDispatcher struct {
	testBundleReadyChannel chan dtx.DtxMessage
}

func (p ProxyDispatcher) Dispatch(m dtx.DtxMessage) {
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)
		switch method {
		case "_XCT_testBundleReadyWithProtocolVersion:minimumVersion:":
			p.testBundleReadyChannel <- m
			return
		default:
			log.Infof("Method invocation not implement for selector:%s", method)
		}
	}
	log.Infof("dispatcher received: %s", m.Payload[0])
}

func newDtxProxy(dtxConnection *dtx.DtxConnection) dtxproxy {
	testBundleReadyChannel := make(chan dtx.DtxMessage, 1)
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

func RunWDA(device usbmux.DeviceEntry) error {

	return runXCUIWithBundleIds("com.facebook.WebDriverAgentRunner.xctrunner", "com.facebook.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest", device)
}

func RunXCUITest(bundleID string, device usbmux.DeviceEntry) error {
	testRunnerBundleID := bundleID + testBundleSuffix
	return runXCUIWithBundleIds(bundleID, testRunnerBundleID, "", device)
}

func runXCUIWithBundleIds(bundleID string, testRunnerBundleID string, xctestConfigFileName string, device usbmux.DeviceEntry) error {
	testSessionId, v, xctestConfigPath, err := setupXcuiTest(device, bundleID, testRunnerBundleID, xctestConfigFileName)
	if err != nil {
		return err
	}
	conn, _ := dtx.NewDtxConnection(device.DeviceID, device.Properties.SerialNumber, testmanagerd)
	defer conn.Close()
	ideDaemonProxy := newDtxProxy(conn)
	protocolVersion, err := ideDaemonProxy.daemonConnection.initiateSessionWithIdentifier(testSessionId, 29)
	log.Infof("ProtocolVersion:%d", protocolVersion)
	if err != nil {
		return err
	}
	pid, err := startTestRunner(device, xctestConfigPath, testRunnerBundleID)
	if err != nil {
		return err
	}
	log.Info("Runner started with pid:%d, waiting for testBundleReady", pid)
	protocolVersion, minimalVersion := ideDaemonProxy.ideInterface.testBundleReady()
	channel := ideDaemonProxy.dtxConnection.ForChannelRequest(ProxyDispatcher{})

	log.Infof("ProtocolVersion:%d MinimalVersion:%d", protocolVersion, minimalVersion)
	conn2, _ := dtx.NewDtxConnection(device.DeviceID, device.Properties.SerialNumber, testmanagerd)
	defer conn2.Close()
	ideDaemonProxy2 := newDtxProxy(conn2)
	ideDaemonProxy2.daemonConnection.initiateControlSessionForTestProcessID(pid, protocolVersion)
	startExecutingTestPlanWithProtocolVersion(channel, protocolVersion)
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

func setupXcuiTest(device usbmux.DeviceEntry, bundleID string, testRunnerBundleID string, xctestConfigFileName string) (uuid.UUID, semver.Version, string, error) {
	version := usbmux.GetValues(device).Value.ProductVersion
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
	//println(result)
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
	log.Info(apps)
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
