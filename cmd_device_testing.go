package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielpaulus/go-ios/ios/testmanagerd"
)

func runTestCommand(ctx commandContext) {
	bundleID, _ := ctx.Args.String("--bundle-id")
	testRunnerBundleId, _ := ctx.Args.String("--test-runner-bundle-id")
	xctestConfig, _ := ctx.Args.String("--xctest-config")

	testsToRunArg := ctx.Args["--test-to-run"]
	var testsToRun []string
	if testsToRunArg != nil && len(testsToRunArg.([]string)) > 0 {
		testsToRun = testsToRunArg.([]string)
	}

	testsToSkipArg := ctx.Args["--test-to-skip"]
	var testsToSkip []string
	if testsToSkipArg != nil && len(testsToSkipArg.([]string)) > 0 {
		testsToSkip = testsToSkipArg.([]string)
	}

	rawTestlog, rawTestlogErr := ctx.Args.String("--log-output")
	env := splitKeyValuePairs(ctx.Args["--env"].([]string), "=")
	isXCTest, _ := ctx.Args.Bool("--xctest")

	config := testmanagerd.TestConfig{
		BundleId:           bundleID,
		TestRunnerBundleId: testRunnerBundleId,
		XctestConfigName:   xctestConfig,
		Env:                env,
		TestsToRun:         testsToRun,
		TestsToSkip:        testsToSkip,
		XcTest:             isXCTest,
		Device:             ctx.Device,
	}

	if rawTestlogErr == nil {
		var writer *os.File = os.Stdout
		if rawTestlog != "-" {
			file, err := os.Create(rawTestlog)
			exitIfError("Cannot open file "+rawTestlog, err)
			writer = file
		}
		defer writer.Close()

		config.Listener = testmanagerd.NewTestListener(writer, writer, os.TempDir())

		testResults, err := testmanagerd.RunTestWithConfig(context.TODO(), config)
		if err != nil {
			slog.Info("Failed running Xcuitest", "error", err)
		}

		slog.Info("test results", "results", testResults)
	} else {
		config.Listener = testmanagerd.NewTestListener(io.Discard, io.Discard, os.TempDir())
		_, err := testmanagerd.RunTestWithConfig(context.TODO(), config)
		if err != nil {
			slog.Info("Failed running Xcuitest", "error", err)
		}
	}
}

func runXCTestCommand(ctx commandContext) {
	xctestrunFilePath, _ := ctx.Args.String("--xctestrun-file-path")

	rawTestlog, rawTestlogErr := ctx.Args.String("--log-output")

	if rawTestlogErr == nil {
		var writer *os.File = os.Stdout
		if rawTestlog != "-" {
			file, err := os.Create(rawTestlog)
			exitIfError("Cannot open file "+rawTestlog, err)
			writer = file
		}
		defer writer.Close()
		listener := testmanagerd.NewTestListener(writer, writer, os.TempDir())

		testResults, err := testmanagerd.StartXCTestWithConfig(context.TODO(), xctestrunFilePath, ctx.Device, listener)
		if err != nil {
			slog.Info("Failed running Xctest", "error", err)
		}

		slog.Info("test results", "results", testResults)
	} else {
		listener := testmanagerd.NewTestListener(io.Discard, io.Discard, os.TempDir())
		_, err := testmanagerd.StartXCTestWithConfig(context.TODO(), xctestrunFilePath, ctx.Device, listener)
		if err != nil {
			slog.Info("Failed running Xctest", "error", err)
		}
	}
}

func runWDACommand(ctx commandContext) {
	bundleID, _ := ctx.Args.String("--bundleid")
	testbundleID, _ := ctx.Args.String("--testrunnerbundleid")
	xctestconfig, _ := ctx.Args.String("--xctestconfig")
	wdaargs := ctx.Args["--arg"].([]string)
	wdaenv := splitKeyValuePairs(ctx.Args["--env"].([]string), "=")

	if bundleID == "" && testbundleID == "" && xctestconfig == "" {
		slog.Info("no bundle ids specified, falling back to defaults")
		bundleID, testbundleID, xctestconfig = "com.facebook.WebDriverAgentRunner.xctrunner", "com.facebook.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest"
	}
	if bundleID == "" || testbundleID == "" || xctestconfig == "" {
		slog.Error("please specify either NONE of bundleid, testbundleid and xctestconfig or ALL of them. At least one was empty.", "bundleid", bundleID, "testbundleid", testbundleID, "xctestconfig", xctestconfig)
		return
	}
	slog.Info("Running wda", "bundleid", bundleID, "testbundleid", testbundleID, "xctestconfig", xctestconfig)

	rawTestlog, rawTestlogErr := ctx.Args.String("--log-output")

	var writer io.Writer

	if rawTestlogErr == nil {
		writerCloser := os.Stdout
		writer = writerCloser
		if rawTestlog != "-" {
			file, err := os.Create(rawTestlog)
			exitIfError("Cannot open file "+rawTestlog, err)
			writer = file
		}
		defer writerCloser.Close()
	} else {
		writer = io.Discard
	}

	errorChannel := make(chan error)
	defer close(errorChannel)
	ctxWDA, stopWda := context.WithCancel(context.Background())
	go func() {
		_, err := testmanagerd.RunTestWithConfig(ctxWDA, testmanagerd.TestConfig{
			BundleId:           bundleID,
			TestRunnerBundleId: testbundleID,
			XctestConfigName:   xctestconfig,
			Env:                wdaenv,
			Args:               wdaargs,
			Device:             ctx.Device,
			Listener:           testmanagerd.NewTestListener(writer, writer, os.TempDir()),
		})
		if err != nil {
			errorChannel <- err
		}
		stopWda()
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errorChannel:
		slog.Error("Failed running WDA", "error", err)
		stopWda()
		os.Exit(1)
	case <-ctxWDA.Done():
		slog.Error("WDA process ended unexpectedly")
		os.Exit(1)
	case signal := <-c:
		slog.Info(fmt.Sprintf("os signal %d received, closing...", signal))
		stopWda()
	}
	slog.Info("Done Closing")
}
