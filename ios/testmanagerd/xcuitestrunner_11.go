package testmanagerd

import (
	"context"
	"fmt"
	"maps"

	"github.com/Masterminds/semver"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/instruments"
	log "github.com/sirupsen/logrus"
)

func runXCUIWithBundleIdsXcode11Ctx(
	ctx context.Context,
	config TestConfig,
	version *semver.Version,
) ([]TestSuite, error) {
	log.Debugf("set up xcuitest")
	testSessionId, xctestConfigPath, testConfig, testInfo, err := setupXcuiTest(config.Device, config.BundleId, config.TestRunnerBundleId, config.XctestConfigName, config.TestsToRun, config.TestsToSkip, config.XcTest, version)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot create test config: %w", err)
	}
	log.Debugf("test session setup ok")
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
	log.Debug("connections ready")
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
	log.Debugf("Runner started with pid:%d, waiting for testBundleReady", pid)

	err = ideDaemonProxy2.daemonConnection.initiateControlSession(pid, protocolVersion)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot initiate a control session with capabilities: %w", err)
	}
	log.Debugf("control session initiated")
	ideInterfaceChannel := ideDaemonProxy.dtxConnection.ForChannelRequest(proxyDispatcher{id: "emty"})

	log.Debug("start executing testplan")
	err = ideDaemonProxy2.daemonConnection.startExecutingTestPlanWithProtocolVersion(ideInterfaceChannel, 25)
	if err != nil {
		return make([]TestSuite, 0), fmt.Errorf("RunXCUIWithBundleIdsXcode11Ctx: cannot start executing test plan: %w", err)
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
			log.Debugf("adding extra env %s=%s", key, value)
		}
	}

	opts := map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"ActivateSuspended": uint64(1),
	}

	return pControl.StartProcess(bundleID, env, args, opts)
}
