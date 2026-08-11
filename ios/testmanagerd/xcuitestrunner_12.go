package testmanagerd

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/Masterminds/semver"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/google/uuid"
)

func runXUITestWithBundleIdsXcode12Ctx(ctx context.Context, config TestConfig, version *semver.Version,
) ([]TestSuite, error) {
	conn, err := dtx.NewUsbmuxdConnection(config.Device, testmanagerdiOS14)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot create a usbmuxd connection to testmanagerd: %w", err)
	}
	defer conn.Close()

	// iOS 14-16 accept the XCTestConfiguration over the DTX session, the same way
	// the iOS 17+ path does, so we build it in memory instead of writing an
	// .xctestconfiguration file to the device and passing its path. Serializing
	// that file on-device is what broke test runs on these versions.
	testSessionId := uuid.New()
	testInfo, err := getTestInfo(config.Device, config.BundleId, config.TestRunnerBundleId)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot get test info: %w", err)
	}
	testConfig := createTestConfig(testInfo, testSessionId, config.XctestConfigName, config.TestsToRun, config.TestsToSkip, config.XcTest, version)

	ideDaemonProxy := newDtxProxyWithConfig(conn, testConfig, config.Listener)

	conn2, err := dtx.NewUsbmuxdConnection(config.Device, testmanagerdiOS14)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot create a usbmuxd connection to testmanagerd: %w", err)
	}
	defer conn2.Close()
	golog.Debug("connections ready", "module", logModule, "udid", config.Device.Properties.SerialNumber)
	ideDaemonProxy2 := newDtxProxyWithConfig(conn2, testConfig, config.Listener)
	ideDaemonProxy2.ideInterface.testConfig = testConfig
	caps, err := ideDaemonProxy.daemonConnection.initiateControlSessionWithCapabilities(nskeyedarchiver.XCTCapabilities{})
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot initiate a control session with capabilities: %w", err)
	}
	golog.Debug("control session capabilities", "module", logModule, "udid", config.Device.Properties.SerialNumber, "caps", caps)
	localCaps := nskeyedarchiver.XCTCapabilities{CapabilitiesDictionary: map[string]interface{}{
		"XCTIssue capability":     uint64(1),
		"skipped test capability": uint64(1),
		"test timeout capability": uint64(1),
	}}

	caps2, err := ideDaemonProxy2.daemonConnection.initiateSessionWithIdentifierAndCaps(testSessionId, localCaps)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot initiate a session with identifier and capabilities: %w", err)
	}
	golog.Debug("session capabilities", "module", logModule, "udid", config.Device.Properties.SerialNumber, "caps", caps2)
	pControl, err := instruments.NewProcessControl(config.Device)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot connect to process control: %w", err)
	}
	defer pControl.Close()

	pid, err := startTestRunner12(pControl, config.TestRunnerBundleId, testSessionId.String(), testInfo.testApp.path+"/PlugIns/"+config.XctestConfigName, config.Args, config.Env)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot start test runner: %w", err)
	}
	golog.Debug("Runner started, waiting for testBundleReady", "module", logModule, "udid", config.Device.Properties.SerialNumber, "pid", pid)

	ideInterfaceChannel := ideDaemonProxy2.dtxConnection.ForChannelRequest(proxyDispatcher{id: "emty", testListener: noopTestListener()})

	time.Sleep(time.Second)

	success, _ := ideDaemonProxy.daemonConnection.authorizeTestSessionWithProcessID(pid)
	golog.Debug("authorizing test session", "module", logModule, "udid", config.Device.Properties.SerialNumber, "pid", pid, "success", success)
	err = ideDaemonProxy2.daemonConnection.startExecutingTestPlanWithProtocolVersion(ideInterfaceChannel, 36)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("runXUITestWithBundleIdsXcode12Ctx: cannot start executing test plan: %w", err)
	}

	select {
	case <-conn.Closed():
		golog.Debug("conn closed", "module", logModule, "udid", config.Device.Properties.SerialNumber)
		if conn.Err() != dtx.ErrConnectionClosed {
			golog.Error("conn closed unexpectedly", "module", logModule, "udid", config.Device.Properties.SerialNumber, "error", conn.Err())
		}
		break
	case <-conn2.Closed():
		golog.Debug("conn2 closed", "module", logModule, "udid", config.Device.Properties.SerialNumber)
		if conn2.Err() != dtx.ErrConnectionClosed {
			golog.Error("conn2 closed unexpectedly", "module", logModule, "udid", config.Device.Properties.SerialNumber, "error", conn2.Err())
		}
		break
	case <-config.Listener.Done():
		break
	case <-ctx.Done():
		break
	}
	golog.Info("Killing test runner", "module", logModule, "udid", config.Device.Properties.SerialNumber, "pid", pid)
	err = pControl.KillProcess(pid)
	if err != nil {
		golog.Info("Nothing to kill, process is already dead", "module", logModule, "udid", config.Device.Properties.SerialNumber, "pid", pid)
	} else {
		golog.Info("Test runner killed with success", "module", logModule, "udid", config.Device.Properties.SerialNumber)
	}

	golog.Debug("Done running test", "module", logModule, "udid", config.Device.Properties.SerialNumber)

	return config.Listener.TestSuites, config.Listener.err
}

func startTestRunner12(pControl *instruments.ProcessControl, bundleID string,
	sessionIdentifier string, testBundlePath string, wdaargs []string, wdaenv map[string]interface{},
) (uint64, error) {
	args := []interface{}{
		"-NSTreatUnknownArgumentsAsOpen", "NO", "-ApplePersistenceIgnoreState", "YES",
	}
	for _, arg := range wdaargs {
		args = append(args, arg)
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

	if len(wdaenv) > 0 {
		maps.Copy(env, wdaenv)

		for key, value := range wdaenv {
			golog.Debug("adding extra env", "module", logModule, "bundleID", bundleID, "key", key, "value", value)
		}
	}

	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"ActivateSuspended": uint64(1),
	}

	return pControl.StartProcess(bundleID, env, args, opts)
}
