package main

import (
	"context"
	"encoding/json"
	"fmt"
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

// runKillCommand handles `ios kill`. Given one or more bundleIDs (or a single
// --pid/--process, unchanged from before) it kills whatever's running right
// now. Given --watch, it instead stays resident and kills each target the
// instant it's next launched — bundleIDs/--process only, since --pid can't
// name a process that hasn't started yet.
func runKillCommand(ctx commandContext) {
	bundleIDs := ctx.Args["<bundleIDs>"].([]string)
	processIDint, _ := ctx.Args.Int("--pid")
	processID := uint64(processIDint)
	processName, _ := ctx.Args.String("--process")
	watch, _ := ctx.Args.Bool("--watch")

	if len(bundleIDs) == 0 && processID == 0 && processName == "" {
		logFatal("please provide at least one bundleID, or --pid, or --process")
	}
	if watch && processID != 0 {
		logFatal("--watch can't be combined with --pid: a future launch's pid isn't known yet")
	}

	processNames := resolveBundleIDsToProcessNames(ctx, bundleIDs)
	if processName != "" {
		processNames[processName] = ""
	}

	if watch {
		runKillWatch(ctx, processNames)
		return
	}
	runKillOnce(ctx, processNames, processID)
}

// resolveBundleIDsToProcessNames looks up the process/executable name the
// device reports for each bundleID (same lookup the original single-bundleID
// runKillCommand used), returning a processName->bundleID map for logging.
// Exits the process if any bundleID isn't installed.
func resolveBundleIDsToProcessNames(ctx commandContext, bundleIDs []string) map[string]string {
	processNames := make(map[string]string, len(bundleIDs))
	if len(bundleIDs) == 0 {
		return processNames
	}

	svc, err := installationproxy.New(ctx.Device)
	exitIfError("failed creating installationproxy", err)
	apps, err := svc.BrowseAllApps()
	exitIfError("browsing apps failed", err)

	byBundleID := make(map[string]string, len(apps))
	for _, app := range apps {
		byBundleID[app.CFBundleIdentifier()] = app.CFBundleExecutable()
	}
	for _, bundleID := range bundleIDs {
		name, ok := byBundleID[bundleID]
		if !ok {
			slog.Error("not installed", "bundleID", bundleID)
			os.Exit(1)
		}
		processNames[name] = bundleID
	}
	return processNames
}

// runKillOnce kills whichever of processNames/pid are currently running.
// Unlike the original single-target version, a not-found target is logged
// and skipped rather than aborting immediately, so one bad bundleID in a
// list doesn't stop the rest from being killed; the process still exits
// non-zero if anything was missed.
func runKillOnce(ctx commandContext, processNames map[string]string, pid uint64) {
	pControl, err := instruments.NewProcessControl(ctx.Device)
	exitIfError("processcontrol failed", err)
	defer pControl.Close()

	service, err := instruments.NewDeviceInfoService(ctx.Device)
	exitIfError("failed opening deviceInfoService for getting process list", err)
	defer service.Close()
	processList, _ := service.ProcessList()

	found := make(map[string]bool, len(processNames))
	pidFound := false
	for _, p := range processList {
		bundleID, wantedByName, wantedByPid := matchesKillTarget(p, processNames, pid)
		if !wantedByName && !wantedByPid {
			continue
		}
		err = pControl.KillProcess(p.Pid)
		exitIfError("kill process failed", err)
		if bundleID != "" {
			slog.Info("killed", "bundleID", bundleID, "pid", p.Pid)
		} else {
			slog.Info("killed", "process", p.Name, "pid", p.Pid)
		}
		if wantedByName {
			found[p.Name] = true
		}
		if wantedByPid {
			pidFound = true
		}
	}

	missed := false
	for name, bundleID := range processNames {
		if found[name] {
			continue
		}
		missed = true
		if bundleID != "" {
			slog.Error("process not found", "bundleID", bundleID)
		} else {
			slog.Error("process not found", "process", name)
		}
	}
	if pid > 0 && !pidFound {
		missed = true
		slog.Error("process not found", "pid", pid)
	}
	if missed {
		os.Exit(1)
	}
}

// matchesKillTarget reports whether p is one of the requested kill targets:
// by name (via processNames, a processName->bundleID map — bundleID is ""
// for a plain --process target) or by pid. pid==0 means "no --pid given"
// (docopt's zero value for an absent int option) — wantedByPid must
// therefore gate on pid>0, not just p.Pid==pid, since pid 0 is also the Mach
// kernel's real, legitimate entry in the process list; matching it on the
// zero value would try to kill the kernel whenever --pid was left unset.
func matchesKillTarget(p instruments.ProcessInfo, processNames map[string]string, pid uint64) (bundleID string, wantedByName, wantedByPid bool) {
	bundleID, wantedByName = processNames[p.Name]
	wantedByPid = pid > 0 && p.Pid == pid
	return bundleID, wantedByName, wantedByPid
}

// runKillWatch stays resident, killing any of processNames the instant it's
// observed launching (state transitions to "Running"), via the push-based
// application-state notification channel (instruments.Guard) rather than
// polling — the device notifies us on launch instead of us catching it in a
// poll interval, so the launch->kill window is wire latency, not a poll
// period. Runs until interrupted (CTRL+C), printing one JSON line per kill.
func runKillWatch(ctx commandContext, processNames map[string]string) {
	names := make([]string, 0, len(processNames))
	for name, bundleID := range processNames {
		names = append(names, name)
		if bundleID != "" {
			slog.Info("watching", "bundleID", bundleID, "process", name)
		} else {
			slog.Info("watching", "process", name)
		}
	}

	watchCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := instruments.Guard(watchCtx, ctx.Device, names)
	exitIfError("failed starting watch", err)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	for event := range events {
		if event.Err != nil {
			slog.Error("failed killing blocked app", "process", event.ProcessName, "pid", event.Pid, "error", event.Err)
			continue
		}
		s, _ := json.Marshal(map[string]interface{}{
			"process": event.ProcessName,
			"pid":     event.Pid,
			"killed":  true,
		})
		fmt.Println(string(s))
	}
}
