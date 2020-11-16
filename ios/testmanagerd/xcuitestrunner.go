package testmanagerd

import (
	"fmt"
	"path"
	"time"

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
	testConfig             nskeyedarchiver.XCTestConfiguration
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

func testRunnerReadyWithCapabilitiesConfig(testConfig nskeyedarchiver.XCTestConfiguration) dtx.MethodWithResponse {
	return func(msg dtx.Message) (interface{}, error) {

		//protocolVersion, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		//nskeyedarchiver.XCTCapabilities
		response := testConfig
		//caps := protocolVersion[0].(nskeyedarchiver.XCTCapabilities)

		return response, nil
	}
}

func (xdc XCTestManager_DaemonConnectionInterface) startExecutingTestPlanWithProtocolVersion(channel *dtx.Channel, version uint64) error {
	return channel.MethodCallAsync("_IDE_startExecutingTestPlanWithProtocolVersion:", version)
}

func (xdc XCTestManager_DaemonConnectionInterface) authorizeTestSessionWithProcessID(pid uint64) (bool, error) {
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_authorizeTestSessionWithProcessID:", pid)

	returnValue := rply.Payload[0]
	var val bool
	var ok bool
	if val, ok = returnValue.(bool); !ok {
		return val, fmt.Errorf("_IDE_authorizeTestSessionWithProcessID: got wrong returnvalue: %s", rply.Payload)
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("_IDE_authorizeTestSessionWithProcessID: reply")

	return val, err
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateSessionWithIdentifierAndCaps(uuid uuid.UUID, caps nskeyedarchiver.XCTCapabilities) (nskeyedarchiver.XCTCapabilities, error) {
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_initiateSessionWithIdentifier:capabilities:", nskeyedarchiver.NewNSUUID(uuid), caps)

	returnValue := rply.Payload[0]
	var val nskeyedarchiver.XCTCapabilities
	var ok bool
	if val, ok = returnValue.(nskeyedarchiver.XCTCapabilities); !ok {
		return val, fmt.Errorf("_IDE_initiateSessionWithIdentifier:capabilities: got wrong returnvalue: %s", rply.Payload)
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("_IDE_initiateSessionWithIdentifier:capabilities: reply")

	return val, err
}
func (xdc XCTestManager_DaemonConnectionInterface) initiateControlSessionWithCapabilities(caps nskeyedarchiver.XCTCapabilities) (nskeyedarchiver.XCTCapabilities, error) {
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_initiateControlSessionWithCapabilities:", caps)

	returnValue := rply.Payload[0]
	var val nskeyedarchiver.XCTCapabilities
	var ok bool
	if val, ok = returnValue.(nskeyedarchiver.XCTCapabilities); !ok {
		return val, fmt.Errorf("_IDE_initiateControlSessionWithCapabilities got wrong returnvalue: %s", rply.Payload)
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("_IDE_initiateControlSessionWithCapabilities reply")

	return val, err
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
	proxyDispatcher  ProxyDispatcher
}

type ProxyDispatcher struct {
	testBundleReadyChannel          chan dtx.Message
	testRunnerReadyWithCapabilities dtx.MethodWithResponse
	dtxConnection                   *dtx.Connection
	id                              string
}

func (p ProxyDispatcher) Dispatch(m dtx.Message) {
	shouldAck := true
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)
		switch method {
		case "_XCT_testBundleReadyWithProtocolVersion:minimumVersion:":
			p.testBundleReadyChannel <- m
			return
		case "_XCT_logDebugMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)
			log.Info(data)
		case "_XCT_testRunnerReadyWithCapabilities:":
			shouldAck = false
			log.Info("received")
			resp, _ := p.testRunnerReadyWithCapabilities(m)
			payload, _ := nskeyedarchiver.ArchiveBin(resp)
			messageBytes, _ := dtx.Encode2(m.Identifier, 1, m.ChannelCode, false, dtx.ResponseWithReturnValueInPayload, payload, dtx.NewPrimitiveDictionary())
			log.Info("sending response")
			p.dtxConnection.Send(messageBytes)

		default:
			log.Warnf("Method invocation not implement for selector:%s", method)
		}
	}
	if shouldAck {
		dtx.SendAckIfNeeded(p.dtxConnection, m)
	}
	log.Debugf("dispatcher received: %s", m.Payload[0])
}

func newDtxProxy(dtxConnection *dtx.Connection) dtxproxy {
	testBundleReadyChannel := make(chan dtx.Message, 1)
	//(xide XCTestManager_IDEInterface)
	proxyDispatcher := ProxyDispatcher{testBundleReadyChannel: testBundleReadyChannel, dtxConnection: dtxConnection}
	IDEDaemonProxy := dtxConnection.RequestChannelIdentifier(ideToDaemonProxyChannelName, proxyDispatcher)
	ideInterface := XCTestManager_IDEInterface{IDEDaemonProxy: IDEDaemonProxy, testBundleReadyChannel: testBundleReadyChannel}

	return dtxproxy{IDEDaemonProxy: IDEDaemonProxy,
		ideInterface:     ideInterface,
		daemonConnection: XCTestManager_DaemonConnectionInterface{IDEDaemonProxy},
		dtxConnection:    dtxConnection,
		proxyDispatcher:  proxyDispatcher,
	}
}

func newDtxProxyWithConfig(dtxConnection *dtx.Connection, testConfig nskeyedarchiver.XCTestConfiguration) dtxproxy {
	testBundleReadyChannel := make(chan dtx.Message, 1)
	//(xide XCTestManager_IDEInterface)
	proxyDispatcher := ProxyDispatcher{testBundleReadyChannel: testBundleReadyChannel, dtxConnection: dtxConnection, testRunnerReadyWithCapabilities: testRunnerReadyWithCapabilitiesConfig(testConfig)}
	IDEDaemonProxy := dtxConnection.RequestChannelIdentifier(ideToDaemonProxyChannelName, proxyDispatcher)
	ideInterface := XCTestManager_IDEInterface{IDEDaemonProxy: IDEDaemonProxy, testConfig: testConfig, testBundleReadyChannel: testBundleReadyChannel}

	return dtxproxy{IDEDaemonProxy: IDEDaemonProxy,
		ideInterface:     ideInterface,
		daemonConnection: XCTestManager_DaemonConnectionInterface{IDEDaemonProxy},
		dtxConnection:    dtxConnection,
		proxyDispatcher:  proxyDispatcher,
	}
}

const testmanagerd = "com.apple.testmanagerd.lockdown"
const testmanagerdiOS14 = "com.apple.testmanagerd.lockdown.secure"

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

func runXUITestWithBundleIdsXcode12(bundleID string, testRunnerBundleID string, xctestConfigFileName string, device ios.DeviceEntry, conn *dtx.Connection) error {
	testSessionId, _, xctestConfigPath, testConfig, testInfo, err := setupXcuiTest(device, bundleID, testRunnerBundleID, xctestConfigFileName)
	defer conn.Close()
	ideDaemonProxy := newDtxProxyWithConfig(conn, testConfig)

	conn2, _ := dtx.NewConnection(device, testmanagerdiOS14)
	defer conn2.Close()
	ideDaemonProxy2 := newDtxProxyWithConfig(conn2, testConfig)
	ideDaemonProxy2.ideInterface.testConfig = testConfig
	caps, err := ideDaemonProxy.daemonConnection.initiateControlSessionWithCapabilities(nskeyedarchiver.XCTCapabilities{})
	log.Info(caps)
	localCaps := nskeyedarchiver.XCTCapabilities{CapabilitiesDictionary: map[string]interface{}{
		"XCTIssue capability":     uint64(1),
		"skipped test capability": uint64(1),
		"test timeout capability": uint64(1),
	}}

	caps2, err := ideDaemonProxy2.daemonConnection.initiateSessionWithIdentifierAndCaps(testSessionId, localCaps)
	log.Info(caps2)
	pControl, err := instruments.NewProcessControl(device)
	defer pControl.Close()
	if err != nil {
		return err
	}
	pid, err := startTestRunner12(pControl, xctestConfigPath, testRunnerBundleID, testSessionId.String(), testInfo.testrunnerAppPath+"/PlugIns/WebDriverAgentRunner.xctest")
	if err != nil {
		return err
	}
	log.Infof("Runner started with pid:%d, waiting for testBundleReady", pid)

	ideInterfaceChannel := ideDaemonProxy2.dtxConnection.ForChannelRequest(ProxyDispatcher{id: "emty"})
	//caps3 := ideDaemonProxy2.ideInterface.testRunnerReadyWithCapabilities()
	//log.Info(caps3)
	time.Sleep(time.Second)

	success, _ := ideDaemonProxy.daemonConnection.authorizeTestSessionWithProcessID(pid)
	log.Info(success)
	err = ideDaemonProxy2.daemonConnection.startExecutingTestPlanWithProtocolVersion(ideInterfaceChannel, 36)
	if err != nil {
		log.Error(err)
	}
	cs := make(chan string)
	<-cs
	return nil
}

func runXCUIWithBundleIds(bundleID string, testRunnerBundleID string, xctestConfigFileName string, device ios.DeviceEntry) error {

	conn, err := dtx.NewConnection(device, testmanagerdiOS14)
	if err == nil {
		return runXUITestWithBundleIdsXcode12(bundleID, testRunnerBundleID, xctestConfigFileName, device, conn)
	}
	log.Debugf("Failed connecting to %s with %v, trying %s", testmanagerdiOS14, err, testmanagerd)

	testSessionId, _, xctestConfigPath, _, _, err := setupXcuiTest(device, bundleID, testRunnerBundleID, xctestConfigFileName)
	if err != nil {
		return err
	}
	conn, err = dtx.NewConnection(device, testmanagerd)
	if err != nil {
		return err
	}
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

func startTestRunner12(pControl *instruments.ProcessControl, xctestConfigPath string, bundleID string, sessionIdentifier string, testBundlePath string) (uint64, error) {
	args := []interface{}{
		"-NSTreatUnknownArgumentsAsOpen", "NO", "-ApplePersistenceIgnoreState", "YES",
	}
	env := map[string]interface{}{

		"CA_ASSERT_MAIN_THREAD_TRANSACTIONS": "0",
		"CA_DEBUG_TRANSACTIONS":              "0",
		"DYLD_INSERT_LIBRARIES":              "/Developer/usr/lib/libMainThreadChecker.dylib",

		"MTC_CRASH_ON_REPORT":             "1",
		"NSUnbufferedIO":                  "YES",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"XCTestBundlePath":                testBundlePath,
		"XCTestConfigurationFilePath":     "",
		"XCTestSessionIdentifier":         sessionIdentifier,
	}
	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"ActivateSuspended": uint64(1),
	}

	return pControl.StartProcess(bundleID, env, args, opts)

}

func setupXcuiTest(device ios.DeviceEntry, bundleID string, testRunnerBundleID string, xctestConfigFileName string) (uuid.UUID, semver.Version, string, nskeyedarchiver.XCTestConfiguration, testInfo, error) {
	version := ios.GetValues(device).Value.ProductVersion
	testSessionID := uuid.New()

	v, err := semver.NewVersion(version)
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	installationProxy, err := installationproxy.New(device)
	defer installationProxy.Close()
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	apps, err := installationProxy.BrowseUserApps()
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	info, err := getAppInfos(bundleID, testRunnerBundleID, apps)
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	houseArrestService, err := house_arrest.New(device, testRunnerBundleID)
	defer houseArrestService.Close()
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	testConfigPath, testConfig, err := createTestConfigOnDevice(testSessionID, info, houseArrestService, xctestConfigFileName)
	if err != nil {
		return uuid.UUID{}, semver.Version{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}

	return testSessionID, *v, testConfigPath, testConfig, info, nil
}

func createTestConfigOnDevice(testSessionID uuid.UUID, info testInfo, houseArrestService *house_arrest.Connection, xctestConfigFileName string) (string, nskeyedarchiver.XCTestConfiguration, error) {
	relativeXcTestConfigPath := path.Join("tmp", testSessionID.String()+".xctestconfiguration")
	xctestConfigPath := path.Join(info.testRunnerHomePath, relativeXcTestConfigPath)

	if xctestConfigFileName == "" {
		xctestConfigFileName = info.targetAppBundleName + "UITests.xctest"
	}
	testBundleURL := path.Join(info.testrunnerAppPath, "PlugIns", xctestConfigFileName)

	config := nskeyedarchiver.NewXCTestConfiguration(info.targetAppBundleName, testSessionID, info.targetAppBundleID, info.targetAppPath, testBundleURL)
	result, err := nskeyedarchiver.ArchiveXML(config)
	if err != nil {
		return "", nskeyedarchiver.XCTestConfiguration{}, err
	}

	err = houseArrestService.SendFile([]byte(result), relativeXcTestConfigPath)
	if err != nil {
		return "", nskeyedarchiver.XCTestConfiguration{}, err
	}
	return xctestConfigPath, nskeyedarchiver.NewXCTestConfiguration(info.targetAppBundleName, testSessionID, info.targetAppBundleID, info.targetAppPath, testBundleURL), nil
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
