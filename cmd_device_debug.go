package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/crashreport"
	"github.com/danielpaulus/go-ios/ios/debugserver"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/ostrace"
	"github.com/danielpaulus/go-ios/ios/pcap"
	"github.com/docopt/docopt-go"
)

func runPCAPCommand(ctx commandContext) {
	p, _ := ctx.Args.String("--process")
	i, _ := ctx.Args.Int("--pid")
	pcap.Pid = int32(i)
	pcap.ProcName = p
	err := pcap.Start(ctx.Device)
	if err != nil {
		exitIfError("pcap failed", err)
	}
}

func runDproxyCommand(ctx commandContext) {
	binaryMode, _ := ctx.Args.Bool("--binary")
	startDebugProxy(ctx.Device, binaryMode)
}

func runSyslogCommand(ctx commandContext) {
	parse, _ := ctx.Args.Bool("--parse")
	runSyslog(ctx.Device, parse)
}

func runOSTraceCommand(ctx commandContext) {
	pidStr, _ := ctx.Args.String("--pid")
	processName, _ := ctx.Args.String("--process")
	levelStr, _ := ctx.Args.String("--level")
	subsystem, _ := ctx.Args.String("--subsystem")
	match, _ := ctx.Args.String("--match")
	exclude, _ := ctx.Args.String("--exclude")
	pid := -1
	if pidStr != "" {
		var err error
		pid, err = strconv.Atoi(pidStr)
		exitIfError("invalid --pid value", err)
	}
	levelFilter, err := ostrace.ParseLevelFilter(levelStr)
	exitIfError("invalid --level value", err)
	clientFilter := ostrace.ClientFilter{
		Levels:    levelFilter.ClientLevels,
		Subsystem: subsystem,
		Match:     match,
		Exclude:   exclude,
	}
	follow, _ := ctx.Args.Bool("--follow")
	runOsTrace(ctx.Device, pid, processName, levelFilter.MessageFilter, levelFilter.StreamFlags, clientFilter, follow)
}

func runCrashCommand(ctx commandContext) {
	if ls, _ := ctx.Args.Bool("ls"); ls {
		pattern, err := ctx.Args.String("<pattern>")
		if err != nil || pattern == "" {
			pattern = "*"
		}
		files, err := crashreport.ListReports(ctx.Device, pattern)
		exitIfError("failed listing crashreports", err)
		fmt.Println(convertToJSONString(map[string]interface{}{"files": files, "length": len(files)}))
	}
	if cp, _ := ctx.Args.Bool("cp"); cp {
		pattern, _ := ctx.Args.String("<srcpattern>")
		target, _ := ctx.Args.String("<target>")
		slog.Debug("cp", "srcpattern", pattern, "target", target)
		err := crashreport.DownloadReports(ctx.Device, pattern, target)
		exitIfError("failed downloading crashreports", err)
	}
	if rm, _ := ctx.Args.Bool("rm"); rm {
		cwd, _ := ctx.Args.String("<cwd>")
		pattern, _ := ctx.Args.String("<pattern>")
		slog.Debug("rm", "cwd", cwd, "pattern", pattern)
		err := crashreport.RemoveReports(ctx.Device, cwd, pattern)
		exitIfError("failed deleting crashreports", err)
	}
}

func runInstrumentsCommand(ctx commandContext) {
	duration, err := instrumentsSampleDuration(ctx.Args)
	exitIfError("failed parsing --duration", err)

	switch instrumentsSubcommand(ctx.Args) {
	case "fps":
		streamInstrumentsFPS(ctx.Device, duration)
	case "network":
		streamInstrumentsNetwork(ctx.Device, duration)
	default:
		listenAppStateNotifications(ctx.Device)
	}
}

// instrumentsSubcommand returns which `ios instruments <subcommand>` was
// requested, or "" if none matched.
func instrumentsSubcommand(args docopt.Opts) string {
	for _, name := range []string{"fps", "network", "notifications"} {
		if boolArg(args, name) {
			return name
		}
	}
	return ""
}

// instrumentsSampleDuration parses the optional --duration=<seconds> flag.
// A zero duration means "stream until interrupted".
func instrumentsSampleDuration(args docopt.Opts) (time.Duration, error) {
	value, err := args.String("--duration")
	if err != nil || value == "" {
		return 0, nil
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid --duration %q: %w", value, err)
	}
	if seconds < 0 {
		return 0, fmt.Errorf("--duration must not be negative, got %q", value)
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

func streamInstrumentsFPS(device ios.DeviceEntry, duration time.Duration) {
	service, err := instruments.NewGraphicsOpenGLService(device)
	exitIfError("failed starting graphics service", err)
	defer service.Close()

	streamInstrumentsSamples(service.ReceiveFramesPerSecondSamples(), duration, formatFPSSample)
}

func streamInstrumentsNetwork(device ios.DeviceEntry, duration time.Duration) {
	service, err := instruments.NewNetworkService(device)
	exitIfError("failed starting network monitoring service", err)
	defer service.Close()

	streamInstrumentsSamples(service.ReceiveNetworkSamples(), duration, formatNetworkSample)
}

// streamInstrumentsSamples prints one formatted line per sample until the
// channel closes, the optional duration elapses, or the process is interrupted.
func streamInstrumentsSamples[T any](samples chan T, duration time.Duration, format func(T) string) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	var timeout <-chan time.Time
	if duration > 0 {
		timer := time.NewTimer(duration)
		defer timer.Stop()
		timeout = timer.C
	}

	for {
		select {
		case sample, ok := <-samples:
			if !ok {
				return
			}
			fmt.Println(format(sample))
		case <-timeout:
			return
		case <-stop:
			return
		}
	}
}

type fpsSampleOutput struct {
	FPS float64 `json:"fps"`
}

func formatFPSSample(sample instruments.FramesPerSecondSample) string {
	if JSONdisabled {
		return fmt.Sprintf("fps=%.2f", sample.CoreAnimationFramesPerSecond)
	}
	return convertToJSONString(fpsSampleOutput{FPS: sample.CoreAnimationFramesPerSecond})
}

type networkSampleOutput struct {
	Type uint64                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func formatNetworkSample(sample instruments.NetworkSample) string {
	if JSONdisabled {
		keys := make([]string, 0, len(sample.Data))
		for key := range sample.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var builder strings.Builder
		fmt.Fprintf(&builder, "type=%d", sample.Type)
		for _, key := range keys {
			fmt.Fprintf(&builder, " %s=%v", key, sample.Data[key])
		}
		return builder.String()
	}
	return convertToJSONString(networkSampleOutput{Type: sample.Type, Data: sample.Data})
}

func listenAppStateNotifications(device ios.DeviceEntry) {
	listenerFunc, closeFunc, err := instruments.ListenAppStateNotifications(device)
	if err != nil {
		logFatal("failed listening to app state notifications", "error", err)
	}
	go func() {
		for {
			notification, err := listenerFunc()
			if err != nil {
				slog.Error("listener error", "error", err)
				return
			}
			s, _ := json.Marshal(notification)
			fmt.Println(string(s))
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	err = closeFunc()
	if err != nil {
		slog.Warn("timeout during close", "error", err)
	}
}

func runImageCommand(ctx commandContext) {
	if list, _ := ctx.Args.Bool("list"); list {
		listMountedImages(ctx.Device)
	}

	imagePath, _ := ctx.Args.String("--path")
	auto, _ := ctx.Args.Bool("auto")
	if auto {
		basedir, _ := ctx.Args.String("--basedir")
		if basedir == "" {
			basedir = "./devimages"
		}

		var err error
		imagePath, err = imagemounter.DownloadImageFor(ctx.Device, basedir)
		if err != nil {
			slog.Error("failed downloading image", "basedir", basedir, "udid", ctx.Device.Properties.SerialNumber, "err", err)
			return
		}

		slog.Info("success downloaded image", "basedir", basedir, "udid", ctx.Device.Properties.SerialNumber)
	}

	mount, _ := ctx.Args.Bool("mount")
	if mount || auto {
		err := imagemounter.MountImage(ctx.Device, imagePath)
		if err != nil {
			slog.Error("error mounting image", "image", imagePath, "udid", ctx.Device.Properties.SerialNumber, "err", err)
			os.Exit(1)
		}
		slog.Info("success mounting image", "image", imagePath, "udid", ctx.Device.Properties.SerialNumber)
	}

	if unmount, _ := ctx.Args.Bool("unmount"); unmount {
		err := imagemounter.UnmountImage(ctx.Device)
		if err != nil {
			slog.Error("error unmounting image", "udid", ctx.Device.Properties.SerialNumber, "err", err)
			os.Exit(1)
		}
		slog.Info("success unmounting image", "udid", ctx.Device.Properties.SerialNumber)
	}
}

func runDebugCommand(ctx commandContext) {
	appPath, _ := ctx.Args.String("<app_path>")
	if appPath == "" {
		logFatal("parameter bundleid and app_path must be specified")
	}
	stopAtEntry, _ := ctx.Args.Bool("--stop-at-entry")
	exitIfError("debug server failed", debugserver.Start(ctx.Device, appPath, stopAtEntry))
}
