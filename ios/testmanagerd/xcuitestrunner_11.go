package testmanagerd

import (
	"context"
	"fmt"
	"maps"

	"github.com/Masterminds/semver"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/instruments"
)

func runXCUIWithBundleIdsXcode11Ctx(
	ctx context.Context,
	config TestConfig,
	version *semver.Version,
) ([]TestSuite, error) {
	golog.Debug("set up xcuitest", "module", logModule, "udid", config.Device.Properties.SerialNumber)
	testSessionId, xctestConfigPath, testConfig, testInfo, err := setupXcuiTest(config.Device, config.BundleId, config.TestRunnerBundleId, config.XctestConfigName, config.TestsToRun, config.TestsToSkip, config.XcTest, version)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot create test config: %w", err)
	}
	golog.Debug("test session setup ok", "module", logModule, "udid", config.Device.Properties.SerialNumber)
	conn, err := dtx.NewUsbmuxdConnection(config.Device, testmanagerd)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot create a usbmuxd connection to testmanagerd: %w", err)
	}
	defer conn.Close()

	ideDaemonProxy := newDtxProxyWithConfig(conn, testConfig, config.Listener)

	conn2, err := dtx.NewUsbmuxdConnection(config.Device, testmanagerd)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot create a usbmuxd connection to testmanagerd: %w", err)
	}
	defer conn2.Close()
	golog.Debug("connections ready", "module", logModule, "udid", config.Device.Properties.SerialNumber)
	ideDaemonProxy2 := newDtxProxyWithConfig(conn2, testConfig, config.Listener)
	ideDaemonProxy2.ideInterface.testConfig = testConfig
	// TODO: fixme
	protocolVersion := uint64(25)
	_, err = ideDaemonProxy.daemonConnection.initiateSessionWithIdentifier(testSessionId, protocolVersion)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot initiate a test session: %w", err)
	}

	pControl, err := instruments.NewProcessControl(config.Device)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot connect to process control: %w", err)
	}
	defer pControl.Close()

	pid, err := startTestRunner11(pControl, xctestConfigPath, config.TestRunnerBundleId, testSessionId.String(), testInfo.testApp.path+"/PlugIns/"+config.XctestConfigName, config.Args, config.Env)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot start the test runner: %w", err)
	}
	golog.Debug("Runner started, waiting for testBundleReady", "module", logModule, "udid", config.Device.Properties.SerialNumber, "pid", pid)

	err = ideDaemonProxy2.daemonConnection.initiateControlSession(pid, protocolVersion)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot initiate a control session with capabilities: %w", err)
	}
	golog.Debug("control session initiated", "module", logModule, "udid", config.Device.Properties.SerialNumber, "pid", pid)
	ideInterfaceChannel := ideDaemonProxy.dtxConnection.ForChannelRequest(proxyDispatcher{id: "emty", testListener: noopTestListener()})

	golog.Debug("start executing testplan", "module", logModule, "udid", config.Device.Properties.SerialNumber)
	err = ideDaemonProxy2.daemonConnection.startExecutingTestPlanWithProtocolVersion(ideInterfaceChannel, 25)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot start executing test plan: %w", err)
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

func startTestRunner11(pControl *instruments.ProcessControl, xctestConfigPath string, bundleID string,
	sessionIdentifier string, testBundlePath string, wdaargs []string, wdaenv map[string]interface{},
) (uint64, error) {
	args := []interface{}{}
	for _, arg := range wdaargs {
		args = append(args, arg)
	}
	env := map[string]interface{}{
		"XCTestBundlePath":            testBundlePath,
		"XCTestConfigurationFilePath": xctestConfigPath,
		"XCTestSessionIdentifier":     sessionIdentifier,
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
