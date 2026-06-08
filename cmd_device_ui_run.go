package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielpaulus/go-ios/ios/forward"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
)

type uiRunTarget struct {
	name          string
	defaultBundle string
	devicePort    uint16
	healthPath    string
	xctestConfig  string
}

// runUIRunCommand runs a UI automation runner (WebDriverAgent or DeviceKit) and
// forwards its port to the host, so `ios ui ...` and any HTTP client can reach it
// at http://127.0.0.1:<host-port>. It is the run counterpart to `ios ui download`
// and `ios ui install`, so users don't have to hand-roll `ios runtest`. It blocks
// until interrupted, then stops the runner and tears the forward down.
func runUIRunCommand(ctx commandContext) {
	var target uiRunTarget
	switch {
	case boolArg(ctx.Args, "wda"):
		target = uiRunTarget{name: "wda", defaultBundle: defaultWDABundleID, devicePort: 8100, healthPath: "/status", xctestConfig: "WebDriverAgentRunner.xctest"}
	case boolArg(ctx.Args, "devicekit"):
		target = uiRunTarget{name: "devicekit", defaultBundle: defaultDeviceKitBundleID, devicePort: 12004, healthPath: "/health", xctestConfig: "devicekit-iosUITests.xctest"}
	default:
		logFatal("unknown ui run target; use 'ios ui run wda' or 'ios ui run devicekit'")
	}

	bundleID, _ := ctx.Args.String("--bundleid")
	if bundleID == "" {
		bundleID = target.defaultBundle
	}
	// The runner bundle is also the test host. Allow overrides for non-default
	// builds, but default to the runner bundle id and the target's xctest config.
	testRunnerBundleID, _ := ctx.Args.String("--test-runner-bundleid")
	if testRunnerBundleID == "" {
		testRunnerBundleID = bundleID
	}
	xctestConfig, _ := ctx.Args.String("--xctest-config")
	if xctestConfig == "" {
		xctestConfig = target.xctestConfig
	}
	hostPort := target.devicePort
	if hp, err := ctx.Args.Int("--host-port"); err == nil && hp > 0 {
		hostPort = uint16(hp)
	}

	// Run the runner. testmanagerd derives the test-runner bundle id and xctest
	// config from the bundle id when the others are left empty, so callers only
	// need the runner's bundle id.
	runCtx, stopRunner := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		_, err := testmanagerd.RunTestWithConfig(runCtx, testmanagerd.TestConfig{
			BundleId:           bundleID,
			TestRunnerBundleId: testRunnerBundleID,
			XctestConfigName:   xctestConfig,
			Device:             ctx.Device,
			Listener:           testmanagerd.NewTestListener(uiRunLogWriter(ctx), uiRunLogWriter(ctx), os.TempDir()),
		})
		if err != nil {
			runErr <- err
		}
		stopRunner()
	}()

	cl, err := forward.Forward(ctx.Device, hostPort, target.devicePort)
	exitIfError("failed to forward "+target.name+" port", err)
	defer func() { _ = cl.Close() }()

	localURL := fmt.Sprintf("http://127.0.0.1:%d", hostPort)
	go reportUIRunReady(localURL+target.healthPath, target.name, localURL, ctx.Device.Properties.SerialNumber)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-runErr:
		slog.Error("ui run: runner failed", "target", target.name, "error", err)
		stopRunner()
		os.Exit(1)
	case <-runCtx.Done():
		slog.Error("ui run: runner ended unexpectedly", "target", target.name)
		os.Exit(1)
	case sig := <-c:
		slog.Info("signal received, stopping ui run", "signal", sig.String(), "target", target.name)
		stopRunner()
	}
}

func uiRunLogWriter(ctx commandContext) io.Writer {
	rawTestlog, err := ctx.Args.String("--log-output")
	if err != nil {
		return io.Discard
	}
	if rawTestlog == "" || rawTestlog == "-" {
		return os.Stdout
	}
	file, ferr := os.Create(rawTestlog)
	exitIfError("cannot open "+rawTestlog, ferr)
	return file
}

// reportUIRunReady polls the runner's health endpoint over the forward and logs
// once it answers, so users see when `ui run` is usable. Best-effort: the runner
// keeps running regardless.
func reportUIRunReady(healthURL, name, localURL, udid string) {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(120 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				slog.Info("ui run ready", "target", name, "url", localURL, "udid", udid)
				return
			}
		}
		time.Sleep(time.Second)
	}
	slog.Warn("ui run: backend not reachable yet; still running", "target", name, "url", localURL)
}
