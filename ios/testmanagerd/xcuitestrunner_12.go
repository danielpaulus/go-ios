package testmanagerd

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/Masterminds/semver"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func runXUITestWithBundleIdsXcode12Ctx(ctx context.Context, config TestConfig, version *semver.Version,
) ([]TestSuite, error) {
	conn, err := dtx.NewUsbmuxdConnection(config.Device, testmanagerdiOS14)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot create a usbmuxd connection to testmanagerd: %w", err)
	}
	defer conn.Close()

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
	log.Debug("connections ready")
	ideDaemonProxy2 := newDtxProxyWithConfig(conn2, testConfig, config.Listener)
	ideDaemonProxy2.ideInterface.testConfig = testConfig
	caps, err := ideDaemonProxy.daemonConnection.initiateControlSessionWithCapabilities(nskeyedarchiver.XCTCapabilities{})
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot initiate a control session with capabilities: %w", err)
	}
	log.Debug(caps)
	localCaps := nskeyedarchiver.XCTCapabilities{CapabilitiesDictionary: map[string]interface{}{
		"XCTIssue capability":     uint64(1),
		"skipped test capability": uint64(1),
		"test timeout capability": uint64(1),
	}}

	caps2, err := ideDaemonProxy2.daemonConnection.initiateSessionWithIdentifierAndCaps(testSessionId, localCaps)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot initiate a session with identifier and capabilities: %w", err)
	}
	log.Debug(caps2)
	pControl, err := instruments.NewProcessControl(config.Device)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot connect to process control: %w", err)
	}
	defer pControl.Close()

	pid, err := startTestRunner12(pControl, config.TestRunnerBundleId, testSessionId.String(), testInfo.testApp.path+"/PlugIns/"+config.XctestConfigName, config.Args, config.Env)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXUITestWithBundleIdsXcode12Ctx: cannot start test runner: %w", err)
	}
	log.Debugf("Runner started with pid:%d, waiting for testBundleReady", pid)

	ideInterfaceChannel := ideDaemonProxy2.dtxConnection.ForChannelRequest(proxyDispatcher{id: "emty"})

	time.Sleep(time.Second)

	success, _ := ideDaemonProxy.daemonConnection.authorizeTestSessionWithProcessID(pid)
	log.Debugf("authorizing test session for pid %d successful %t", pid, success)
	err = ideDaemonProxy2.daemonConnection.startExecutingTestPlanWithProtocolVersion(ideInterfaceChannel, 36)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("runXUITestWithBundleIdsXcode12Ctx: cannot start executing test plan: %w", err)
	}

	select {
	case <-conn.Closed():
		log.Debug("conn closed")
		if conn.Err() != dtx.ErrConnectionClosed {
			log.WithError(conn.Err()).Error("conn closed unexpectedly")
		}
		break
	case <-conn2.Closed():
		log.Debug("conn2 closed")
		if conn2.Err() != dtx.ErrConnectionClosed {
			log.WithError(conn2.Err()).Error("conn2 closed unexpectedly")
		}
		break
	case <-config.Listener.Done():
		break
	case <-ctx.Done():
		break
	}
	log.Infof("Killing test runner with pid %d ...", pid)
	err = pControl.KillProcess(pid)
	if err != nil {
		log.Infof("Nothing to kill, process with pid %d is already dead", pid)
	} else {
		log.Info("Test runner killed with success")
	}

	log.Debugf("Done running test")

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
			log.Debugf("adding extra env %s=%s", key, value)
		}
	}

	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"ActivateSuspended": uint64(1),
	}

	return pControl.StartProcess(bundleID, env, args, opts)
}
