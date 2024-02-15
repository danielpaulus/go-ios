package testmanagerd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/danielpaulus/go-ios/ios/appservice"

	"github.com/danielpaulus/go-ios/ios/house_arrest"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
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
		// protocolVersion, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		// nskeyedarchiver.XCTCapabilities
		response := testConfig
		// caps := protocolVersion[0].(nskeyedarchiver.XCTCapabilities)

		return response, nil
	}
}

func (xdc XCTestManager_DaemonConnectionInterface) startExecutingTestPlanWithProtocolVersion(channel *dtx.Channel, version uint64) error {
	return channel.MethodCallAsync("_IDE_startExecutingTestPlanWithProtocolVersion:", version)
}

func (xdc XCTestManager_DaemonConnectionInterface) authorizeTestSessionWithProcessID(pid uint64) (bool, error) {
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_authorizeTestSessionWithProcessID:", pid)
	if err != nil {
		log.Errorf("authorizeTestSessionWithProcessID failed: %v, err:%v", pid, err)
		return false, err
	}
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
	var val nskeyedarchiver.XCTCapabilities
	var ok bool
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_initiateSessionWithIdentifier:capabilities:", nskeyedarchiver.NewNSUUID(uuid), caps)
	if err != nil {
		log.Errorf("initiateSessionWithIdentifierAndCaps failed: %v", err)
		return val, err
	}
	returnValue := rply.Payload[0]
	if val, ok = returnValue.(nskeyedarchiver.XCTCapabilities); !ok {
		return val, fmt.Errorf("_IDE_initiateSessionWithIdentifier:capabilities: got wrong returnvalue: %s", rply.Payload)
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("_IDE_initiateSessionWithIdentifier:capabilities: reply")

	return val, err
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateControlSessionWithCapabilities(caps nskeyedarchiver.XCTCapabilities) (nskeyedarchiver.XCTCapabilities, error) {
	var val nskeyedarchiver.XCTCapabilities
	var ok bool
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_initiateControlSessionWithCapabilities:", caps)
	if err != nil {
		log.Errorf("initiateControlSessionWithCapabilities failed: %v", err)
		return val, err
	}
	returnValue := rply.Payload[0]

	if val, ok = returnValue.(nskeyedarchiver.XCTCapabilities); !ok {
		return val, fmt.Errorf("_IDE_initiateControlSessionWithCapabilities got wrong returnvalue: %s", rply.Payload)
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("_IDE_initiateControlSessionWithCapabilities reply")

	return val, err
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateSessionWithIdentifier(sessionIdentifier uuid.UUID, protocolVersion uint64) (uint64, error) {
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName}).Debug("Launching init test Session")
	var val uint64
	var ok bool
	rply, err := xdc.IDEDaemonProxy.MethodCall(
		"_IDE_initiateSessionWithIdentifier:forClient:atPath:protocolVersion:",
		nskeyedarchiver.NewNSUUID(sessionIdentifier),
		"thephonedoesntcarewhatisendhereitseems",
		"/Applications/Xcode.app",
		protocolVersion)
	if err != nil {
		log.Errorf("initiateSessionWithIdentifier failed: %v", err)
		return val, err
	}
	returnValue := rply.Payload[0]
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

func (xdc XCTestManager_DaemonConnectionInterface) initiateControlSessionWithProtocolVersion(protocolVersion uint64) (uint64, error) {
	rply, err := xdc.IDEDaemonProxy.MethodCall("_IDE_initiateControlSessionWithProtocolVersion:", protocolVersion)
	if err != nil {
		return 0, err
	}
	returnValue := rply.Payload[0]
	var val uint64
	var ok bool
	if val, ok = returnValue.(uint64); !ok {
		return val, fmt.Errorf("_IDE_initiateControlSessionWithProtocolVersion got wrong returnvalue: %s", rply.Payload)
	}
	log.WithFields(log.Fields{"channel_id": ideToDaemonProxyChannelName, "reply": rply}).Debug("initiateControlSessionForTestProcessID reply")
	return val, nil
}

func (xdc XCTestManager_DaemonConnectionInterface) initiateControlSession(pid uint64, version uint64) error {
	_, err := xdc.IDEDaemonProxy.MethodCall("_IDE_initiateControlSessionForTestProcessID:protocolVersion:", pid, version)
	return err
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
	proxyDispatcher  proxyDispatcher
}

func newDtxProxy(dtxConnection *dtx.Connection) dtxproxy {
	testBundleReadyChannel := make(chan dtx.Message, 1)
	//(xide XCTestManager_IDEInterface)
	proxyDispatcher := proxyDispatcher{testBundleReadyChannel: testBundleReadyChannel, dtxConnection: dtxConnection}
	IDEDaemonProxy := dtxConnection.RequestChannelIdentifier(ideToDaemonProxyChannelName, proxyDispatcher)
	ideInterface := XCTestManager_IDEInterface{IDEDaemonProxy: IDEDaemonProxy, testBundleReadyChannel: testBundleReadyChannel}

	return dtxproxy{
		IDEDaemonProxy:   IDEDaemonProxy,
		ideInterface:     ideInterface,
		daemonConnection: XCTestManager_DaemonConnectionInterface{IDEDaemonProxy},
		dtxConnection:    dtxConnection,
		proxyDispatcher:  proxyDispatcher,
	}
}

func newDtxProxyWithConfig(dtxConnection *dtx.Connection, testConfig nskeyedarchiver.XCTestConfiguration, testListener *TestListener) dtxproxy {
	testBundleReadyChannel := make(chan dtx.Message, 1)
	//(xide XCTestManager_IDEInterface)
	proxyDispatcher := proxyDispatcher{
		testBundleReadyChannel:          testBundleReadyChannel,
		dtxConnection:                   dtxConnection,
		testRunnerReadyWithCapabilities: testRunnerReadyWithCapabilitiesConfig(testConfig),
		testListener:                    testListener,
	}
	IDEDaemonProxy := dtxConnection.RequestChannelIdentifier(ideToDaemonProxyChannelName, proxyDispatcher)
	ideInterface := XCTestManager_IDEInterface{IDEDaemonProxy: IDEDaemonProxy, testConfig: testConfig, testBundleReadyChannel: testBundleReadyChannel}

	return dtxproxy{
		IDEDaemonProxy:   IDEDaemonProxy,
		ideInterface:     ideInterface,
		daemonConnection: XCTestManager_DaemonConnectionInterface{IDEDaemonProxy},
		dtxConnection:    dtxConnection,
		proxyDispatcher:  proxyDispatcher,
	}
}

const (
	testmanagerd      = "com.apple.testmanagerd.lockdown"
	testmanagerdiOS14 = "com.apple.testmanagerd.lockdown.secure"
	testmanagerdiOS17 = "com.apple.dt.testmanagerd.remote"
)

const testBundleSuffix = "UITests.xctrunner"

func RunXCUITest(bundleID string, testRunnerBundleID string, xctestConfigName string, device ios.DeviceEntry, env []string, testListener *TestListener) (TestSuite, error) {
	// FIXME: this is redundant code, getting the app list twice and creating the appinfos twice
	// just to generate the xctestConfigFileName. Should be cleaned up at some point.
	installationProxy, err := installationproxy.New(device)
	if err != nil {
		return TestSuite{}, fmt.Errorf("RunXCUITest: cannot connect to installation proxy: %w", err)
	}
	defer installationProxy.Close()

	if testRunnerBundleID == "" {
		testRunnerBundleID = bundleID + testBundleSuffix
	}

	apps, err := installationProxy.BrowseUserApps()
	if err != nil {
		return TestSuite{}, fmt.Errorf("RunXCUITest: cannot browse user apps: %w", err)
	}
	info, err := getAppInfos(bundleID, testRunnerBundleID, apps)
	if err != nil {
		return TestSuite{}, fmt.Errorf("RunXCUITest: cannot get app information: %w", err)
	}

	if xctestConfigName == "" {
		xctestConfigName = info.targetAppBundleName + "UITests.xctest"
	}

	return RunXCUIWithBundleIdsCtx(context.TODO(), bundleID, testRunnerBundleID, xctestConfigName, device, nil, env, testListener)
}

func RunXCUIWithBundleIdsCtx(
	ctx context.Context,
	bundleID string,
	testRunnerBundleID string,
	xctestConfigFileName string,
	device ios.DeviceEntry,
	args []string,
	env []string,
	testListener *TestListener,
) (TestSuite, error) {
	version, err := ios.GetProductVersion(device)
	if err != nil {
		return TestSuite{}, fmt.Errorf("RunXCUIWithBundleIdsCtx: cannot determine iOS version: %w", err)
	}

	if version.LessThan(ios.IOS14()) {
		log.Debugf("iOS version: %s detected, running with ios11 support", version)
		return runXCUIWithBundleIdsXcode11Ctx(ctx, bundleID, testRunnerBundleID, xctestConfigFileName, device, args, env, testListener)
	}

	if version.LessThan(ios.IOS17()) {
		log.Debugf("iOS version: %s detected, running with ios14 support", version)
		return runXUITestWithBundleIdsXcode12Ctx(ctx, bundleID, testRunnerBundleID, xctestConfigFileName, device, args, env, testListener)
	}

	log.Debugf("iOS version: %s detected, running with ios17 support", version)
	return runXUITestWithBundleIdsXcode15Ctx(ctx, bundleID, testRunnerBundleID, xctestConfigFileName, device, args, env, testListener)
}

func runXUITestWithBundleIdsXcode15Ctx(
	ctx context.Context,
	bundleID string,
	testRunnerBundleID string,
	xctestConfigFileName string,
	device ios.DeviceEntry,
	args []string,
	env []string,
	testListener *TestListener,
) (TestSuite, error) {
	conn1, err := dtx.NewTunnelConnection(device, testmanagerdiOS17)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot create a tunnel connection to testmanagerd: %w", err)
	}
	defer conn1.Close()

	conn2, err := dtx.NewTunnelConnection(device, testmanagerdiOS17)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot create a tunnel connection to testmanagerd: %w", err)
	}
	defer conn2.Close()

	installationProxy, err := installationproxy.New(device)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot connect to installation proxy: %w", err)
	}
	defer installationProxy.Close()
	apps, err := installationProxy.BrowseUserApps()
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot browse user apps: %w", err)
	}

	info, err := getAppInfos(bundleID, testRunnerBundleID, apps)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot get app information: %w", err)
	}

	testSessionID := uuid.New()
	testconfig := createTestConfig(info, testSessionID, xctestConfigFileName)
	ideDaemonProxy1 := newDtxProxyWithConfig(conn1, testconfig, testListener)

	localCaps := nskeyedarchiver.XCTCapabilities{CapabilitiesDictionary: map[string]interface{}{
		"XCTIssue capability":                      uint64(1),
		"daemon container sandbox extension":       uint64(1),
		"delayed attachment transfer":              uint64(1),
		"expected failure test capability":         uint64(1),
		"request diagnostics for specific devices": uint64(1),
		"skipped test capability":                  uint64(1),
		"test case run configurations":             uint64(1),
		"test iterations":                          uint64(1),
		"test timeout capability":                  uint64(1),
		"ubiquitous test identifiers":              uint64(1),
	}}
	receivedCaps, err := ideDaemonProxy1.daemonConnection.initiateSessionWithIdentifierAndCaps(testSessionID, localCaps)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot initiate a IDE session: %w", err)
	}
	log.WithField("receivedCaps", receivedCaps).Info("got capabilities")

	appserviceConn, err := appservice.New(device)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot connect to app service: %w", err)
	}
	defer appserviceConn.Close()

	pid, err := startTestRunner17(device, appserviceConn, "", testRunnerBundleID, strings.ToUpper(testSessionID.String()), info.testrunnerAppPath+"/PlugIns/"+xctestConfigFileName, args, env)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot start test runner: %w", err)
	}

	ideDaemonProxy2 := newDtxProxyWithConfig(conn2, testconfig, testListener)
	caps, err := ideDaemonProxy2.daemonConnection.initiateControlSessionWithCapabilities(nskeyedarchiver.XCTCapabilities{CapabilitiesDictionary: map[string]interface{}{}})
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot initiate a control session with capabilities: %w", err)
	}
	log.WithField("caps", caps).Info("got capabilities")
	authorized, err := ideDaemonProxy2.daemonConnection.authorizeTestSessionWithProcessID(uint64(pid))
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot authorize test session: %w", err)
	}
	log.WithField("authorized", authorized).Info("authorized")

	ideInterfaceChannel := ideDaemonProxy1.dtxConnection.ForChannelRequest(proxyDispatcher{id: "dtxproxy:XCTestDriverInterface:XCTestManager_IDEInterface"})

	proto := uint64(36)
	err = ideDaemonProxy1.daemonConnection.startExecutingTestPlanWithProtocolVersion(ideInterfaceChannel, proto)
	if err != nil {
		return TestSuite{}, fmt.Errorf("runXUITestWithBundleIdsXcode15Ctx: cannot start executing test plan: %w", err)
	}

	select {
	case <-testListener.Done():
		break
	case <-ctx.Done():
		log.Infof("Killing test runner with pid %d ...", pid)
		err = killTestRunner(appserviceConn, pid)
		if err != nil {
			log.Infof("Nothing to kill, process with pid %d is already dead", pid)
		} else {
			log.Info("Test runner killed with success")
		}
	}

	log.Debugf("Done running test")

	return *testListener.TestSuite, testListener.err
}

type processKiller interface {
	KillProcess(pid int) error
}

func killTestRunner(killer processKiller, pid int) error {
	log.Infof("Killing test runner with pid %d ...", pid)
	err := killer.KillProcess(pid)
	if err != nil {
		return err
	}
	log.Info("Test runner killed with success")

	return nil
}

func startTestRunner17(device ios.DeviceEntry, appserviceConn *appservice.Connection, xctestConfigPath string, bundleID string,
	sessionIdentifier string, testBundlePath string, testArgs []string, testEnv []string,
) (int, error) {
	args := []interface{}{}
	for _, arg := range testArgs {
		args = append(args, arg)
	}

	env := map[string]interface{}{
		"CA_ASSERT_MAIN_THREAD_TRANSACTIONS": "0",
		"CA_DEBUG_TRANSACTIONS":              "0",
		"DYLD_INSERT_LIBRARIES":              "/Developer/usr/lib/libMainThreadChecker.dylib",
		"DYLD_FRAMEWORK_PATH":                "/System/Developer/Library/Frameworks",
		"DYLD_LIBRARY_PATH":                  "/System/Developer/usr/lib",

		"MTC_CRASH_ON_REPORT":             "1",
		"NSUnbufferedIO":                  "YES",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"XCTestBundlePath":                testBundlePath,
		"XCTestConfigurationFilePath":     "",
		"XCTestManagerVariant":            "DDI",
		"XCTestSessionIdentifier":         strings.ToUpper(sessionIdentifier),
	}

	for _, entrystring := range testEnv {
		entry := strings.Split(entrystring, "=")
		key := entry[0]
		value := entry[1]
		env[key] = value
		log.Debugf("adding extra env %s=%s", key, value)
	}

	opts := map[string]interface{}{
		"ActivateSuspended": uint64(1),
		"StartSuspendedKey": uint64(0),
	}

	appLaunch, err := appserviceConn.LaunchApp(
		bundleID,
		args,
		env,
		opts,
		true,
	)

	if err != nil {
		return 0, err
	}

	return appLaunch.Pid, nil
}

func setupXcuiTest(device ios.DeviceEntry, bundleID string, testRunnerBundleID string, xctestConfigFileName string) (uuid.UUID, string, nskeyedarchiver.XCTestConfiguration, testInfo, error) {
	testSessionID := uuid.New()
	installationProxy, err := installationproxy.New(device)
	if err != nil {
		return uuid.UUID{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	defer installationProxy.Close()

	apps, err := installationProxy.BrowseUserApps()
	if err != nil {
		return uuid.UUID{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}

	info, err := getAppInfos(bundleID, testRunnerBundleID, apps)
	if err != nil {
		return uuid.UUID{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	log.Debugf("app info found: %+v", info)

	houseArrestService, err := house_arrest.New(device, testRunnerBundleID)
	defer houseArrestService.Close()
	if err != nil {
		return uuid.UUID{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}
	log.Debugf("creating test config")
	testConfigPath, testConfig, err := createTestConfigOnDevice(testSessionID, info, houseArrestService, xctestConfigFileName)
	if err != nil {
		return uuid.UUID{}, "", nskeyedarchiver.XCTestConfiguration{}, testInfo{}, err
	}

	return testSessionID, testConfigPath, testConfig, info, nil
}

func createTestConfigOnDevice(testSessionID uuid.UUID, info testInfo, houseArrestService *house_arrest.Connection, xctestConfigFileName string) (string, nskeyedarchiver.XCTestConfiguration, error) {
	relativeXcTestConfigPath := path.Join("tmp", testSessionID.String()+".xctestconfiguration")
	xctestConfigPath := path.Join(info.testRunnerHomePath, relativeXcTestConfigPath)

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

func createTestConfig(info testInfo, testSessionID uuid.UUID, xctestConfigFileName string) nskeyedarchiver.XCTestConfiguration {
	return nskeyedarchiver.NewXCTestConfiguration(info.targetAppBundleName, testSessionID, info.targetAppBundleID, info.targetAppPath, "PlugIns/"+xctestConfigFileName)
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
