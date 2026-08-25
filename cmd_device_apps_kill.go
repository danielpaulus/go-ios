package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
)

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
// application-state notification channel (instruments.ListenAndKill) rather
// than polling — the device notifies us on launch instead of us catching it
// in a poll interval, so the launch->kill window is wire latency, not a poll
// period. Runs until interrupted (CTRL+C), printing one JSON line per kill.
//
// ListenAndKill itself is policy-free (just a shared connection exposing a
// notification stream and a kill func); the blocklist decision
// (shouldWatchKill) lives here, since which processes to watch for and what
// to do about them is specific to this command, not a general-purpose
// instruments capability.
func runKillWatch(ctx commandContext, processNames map[string]string) {
	for name, bundleID := range processNames {
		if bundleID != "" {
			slog.Info("watching", "bundleID", bundleID, "process", name)
		} else {
			slog.Info("watching", "process", name)
		}
	}

	receive, kill, closeFunc, err := instruments.ListenAndKill(ctx.Device)
	exitIfError("failed starting watch", err)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		if err := closeFunc(); err != nil {
			slog.Debug("closing watch", "error", err)
		}
	}()

	for {
		event, err := receive()
		if err != nil {
			return
		}
		if _, ok := shouldWatchKill(event, processNames); !ok {
			continue
		}
		if err := kill(event.Pid); err != nil {
			slog.Error("failed killing blocked app", "process", event.ProcessName, "pid", event.Pid, "error", err)
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

// shouldWatchKill reports whether event is one of processNames entering the
// running state — the pure decision at the core of --watch, split out so it
// can be unit-tested without a device connection.
func shouldWatchKill(event instruments.AppStateEvent, processNames map[string]string) (bundleID string, ok bool) {
	if event.State != "Running" {
		return "", false
	}
	bundleID, ok = processNames[event.ProcessName]
	return bundleID, ok
}
