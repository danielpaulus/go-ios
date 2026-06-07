package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
)

func runPSCommand(ctx commandContext) {
	applicationsOnly, _ := ctx.Args.Bool("--apps")
	processList(ctx.Device, applicationsOnly)
}

func runInstallCommand(ctx commandContext) {
	path, _ := ctx.Args.String("--path")
	installApp(ctx.Device, path)
}

func runUninstallCommand(ctx commandContext) {
	bundleID, _ := ctx.Args.String("<bundleID>")
	uninstallApp(ctx.Device, bundleID)
}

func runAppsCommand(ctx commandContext) {
	list, _ := ctx.Args.Bool("--list")
	system, _ := ctx.Args.Bool("--system")
	all, _ := ctx.Args.Bool("--all")
	filesharing, _ := ctx.Args.Bool("--filesharing")
	printInstalledApps(ctx.Device, system, all, list, filesharing)
}

func runLaunchCommand(ctx commandContext) {
	wait, _ := ctx.Args.Bool("--wait")
	bKillExisting, _ := ctx.Args.Bool("--kill-existing")
	bundleID, _ := ctx.Args.String("<bundleID>")
	if bundleID == "" {
		logFatal("please provide a bundleID")
	}
	pControl, err := instruments.NewProcessControl(ctx.Device)
	exitIfError("processcontrol failed", err)
	opts := map[string]any{}
	if bKillExisting {
		opts["KillExisting"] = 1
	}
	args := toArgs(ctx.Args["--arg"].([]string))
	envs := toEnvs(ctx.Args["--env"].([]string))
	pid, err := pControl.LaunchAppWithArgs(bundleID, args, envs, opts)
	exitIfError("launch app command failed", err)
	slog.Info("Process launched", "pid", pid)
	if wait {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		slog.Info("stop listening to logs", "pid", pid)
	}
}

func runMemlimitOffCommand(ctx commandContext) {
	processName, _ := ctx.Args.String("--process")

	pControl, err := instruments.NewProcessControl(ctx.Device)
	exitIfError("processcontrol failed", err)
	defer pControl.Close()

	svc, err := instruments.NewDeviceInfoService(ctx.Device)
	exitIfError("failed opening deviceInfoService for getting process list", err)
	defer svc.Close()

	process, err := svc.ProcessByName(processName)
	exitIfError("process not found", err)
	if process.Pid > 1 {
		disabled, err := pControl.DisableMemoryLimit(process.Pid)
		exitIfError("DisableMemoryLimit failed", err)
		slog.Info("memory limit is off", "process", process.Name, "pid", process.Pid, "disabled", disabled)
	}
}

func runKillCommand(ctx commandContext) {
	var response []installationproxy.AppInfo
	bundleID, _ := ctx.Args.String("<bundleID>")
	processIDint, _ := ctx.Args.Int("--pid")
	processName, _ := ctx.Args.String("--process")

	processID := uint64(processIDint)

	if bundleID == "" && processID == 0 && processName == "" {
		logFatal("please provide a bundleID")
	}
	pControl, err := instruments.NewProcessControl(ctx.Device)
	exitIfError("processcontrol failed", err)
	svc, _ := installationproxy.New(ctx.Device)

	if bundleID != "" {
		response, err = svc.BrowseAllApps()
		exitIfError("browsing apps failed", err)

		for _, app := range response {
			if app.CFBundleIdentifier() == bundleID {
				processName = app.CFBundleExecutable()
				break
			}
		}
		if processName == "" {
			slog.Error("not installed", "bundleID", bundleID)
			os.Exit(1)
			return
		}
	}

	service, err := instruments.NewDeviceInfoService(ctx.Device)
	defer service.Close()
	exitIfError("failed opening deviceInfoService for getting process list", err)
	processList, _ := service.ProcessList()
	for _, p := range processList {
		if (processID > 0 && p.Pid == processID) || (processName != "" && p.Name == processName) {
			err = pControl.KillProcess(p.Pid)
			exitIfError("kill process failed ", err)
			if bundleID != "" {
				slog.Info("killed", "bundleID", bundleID, "pid", p.Pid)
			} else {
				slog.Info("killed", "process", p.Name, "pid", p.Pid)
			}
			return
		}
	}
	if bundleID != "" {
		slog.Error("process not found", "bundleID", bundleID)
	} else if processName != "" {
		slog.Error("process not found", "process", processName)
	} else {
		slog.Error("process not found", "pid", processID)
	}
	os.Exit(1)
}
