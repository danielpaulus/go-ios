package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/kardianos/service"
)

// `ios tunnel service install|uninstall|status` registers the tunnel agent
// (`ios tunnel start …`) as a managed OS service (systemd, launchd, or the
// Windows service manager) via github.com/kardianos/service, the way
// `cloudflared service install` does. The kardianos dependency lives only in
// package main so the ios library packages stay dependency-clean.

const tunnelServiceName = "go-ios-tunnel"

// capturedServiceEnv is the allowlist of environment variables baked into the
// service definition when they are set at install time. A service does not
// inherit the operator's shell environment, so a known-good config has to be
// captured explicitly (systemd `Environment=` lines, launchd
// EnvironmentVariables).
var capturedServiceEnv = []string{
	"ORCHESTRATOR_URL",
	"GO_IOS_AGENT_HOST",
}

// tunnelServiceConfig captures the current invocation's flags and relevant
// environment into a service definition that reproduces this exact tunnel
// agent configuration on every boot. Relative paths are anchored to workingDir
// so the definition stays valid no matter where the service manager starts us.
func tunnelServiceConfig(args docopt.Opts, getenv func(string) string, executable string, workingDir string) *service.Config {
	arguments := []string{"tunnel", "start"}

	pairRecordPath, _ := args.String("--pair-record-path")
	switch {
	case pairRecordPath == "":
		// `ios tunnel start` defaults the pair-record dir to ".", which for a
		// service depends on the init system. Pin it to the install-time cwd.
		pairRecordPath = workingDir
	case strings.ToLower(pairRecordPath) == "default":
		// keep the literal: `ios tunnel start` resolves it to Apple's
		// RemotePairing directory itself.
	case !filepath.IsAbs(pairRecordPath):
		pairRecordPath = filepath.Join(workingDir, pairRecordPath)
	}
	arguments = append(arguments, "--pair-record-path="+pairRecordPath)

	if boolArg(args, "--userspace") {
		arguments = append(arguments, "--userspace")
	}
	udid, _ := args.String("--udid")
	if udid == "" {
		udid = getenv("GO_IOS_UDID")
	}
	if udid != "" {
		arguments = append(arguments, "--udid="+udid)
	}
	if host, err := args.String("--tunnel-info-host"); err == nil && host != "" {
		arguments = append(arguments, "--tunnel-info-host="+host)
	}
	if port, err := args.String("--tunnel-info-port"); err == nil && port != "" {
		arguments = append(arguments, "--tunnel-info-port="+port)
	}

	envVars := map[string]string{}
	for _, name := range capturedServiceEnv {
		if value := getenv(name); value != "" {
			envVars[name] = value
		}
	}

	return &service.Config{
		Name:             tunnelServiceName,
		DisplayName:      "go-ios tunnel agent",
		Description:      "go-ios tunnel agent (ios tunnel start) providing tunnels for iOS 17+ devices",
		Executable:       executable,
		Arguments:        arguments,
		WorkingDirectory: workingDir,
		EnvVars:          envVars,
		Option: service.KeyValue{
			"Restart":     "always",                // systemd restart policy
			"RunAtLoad":   true,                    // launchd: start at boot/login, not only on demand
			"UserService": boolArg(args, "--user"), // systemd --user unit / launchd LaunchAgent
		},
	}
}

// serviceController is the subset of service.Service the install, uninstall
// and status flows use; tests inject a fake implementation.
type serviceController interface {
	Install() error
	Uninstall() error
	Start() error
	Stop() error
	Status() (service.Status, error)
	Platform() string
}

// checkServiceSystem rejects init systems we would write a broken or untested
// definition for. Linux is systemd-first: kardianos/service would happily
// generate upstart/openrc/sysv scripts, but those paths are untested here.
func checkServiceSystem(goos string, platform string) error {
	if goos == "linux" && platform != "linux-systemd" {
		return fmt.Errorf("no systemd detected (found %q): 'ios tunnel service' currently supports systemd on Linux; "+
			"on other init systems create the init script manually and use 'ios tunnel start' as the service command", platform)
	}
	return nil
}

// installTunnelService registers, enables, and starts the service. It is
// idempotent: re-running install replaces any existing definition and
// re-enables it.
func installTunnelService(ctl serviceController, goos string, userService bool) error {
	if err := checkServiceSystem(goos, ctl.Platform()); err != nil {
		return err
	}
	if _, err := ctl.Status(); !isServiceNotInstalled(err) {
		// Ignore stop errors: the previous instance may already be stopped or broken.
		_ = ctl.Stop()
		if err := ctl.Uninstall(); err != nil && !isServiceNotInstalled(err) {
			return decorateServicePrivilegeError(fmt.Errorf("failed removing the existing service before reinstall: %w", err), userService)
		}
	}
	if err := ctl.Install(); err != nil {
		return decorateServicePrivilegeError(fmt.Errorf("failed installing the service: %w", err), userService)
	}
	if err := ctl.Start(); err != nil {
		return decorateServicePrivilegeError(fmt.Errorf("failed starting the service: %w", err), userService)
	}
	return nil
}

// uninstallTunnelService stops, disables, and removes the service. A service
// that is not installed is not an error; removed reports whether anything was
// actually deleted.
func uninstallTunnelService(ctl serviceController, userService bool) (removed bool, err error) {
	if _, err := ctl.Status(); isServiceNotInstalled(err) {
		return false, nil
	}
	_ = ctl.Stop()
	if err := ctl.Uninstall(); err != nil {
		if isServiceNotInstalled(err) {
			return false, nil
		}
		return false, decorateServicePrivilegeError(fmt.Errorf("failed uninstalling the service: %w", err), userService)
	}
	return true, nil
}

func isServiceNotInstalled(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, service.ErrNotInstalled) || strings.Contains(err.Error(), "not installed")
}

// decorateServicePrivilegeError turns bare permission errors from the service
// manager into an actionable message.
func decorateServicePrivilegeError(err error, userService bool) error {
	if !isPermissionError(err) {
		return err
	}
	if userService {
		return fmt.Errorf("%w: insufficient privileges to manage the user service", err)
	}
	return fmt.Errorf("%w: managing a system service needs elevated privileges: rerun with sudo (or an admin shell on Windows), "+
		"or pass --user for a user-level service (enable lingering with 'loginctl enable-linger' to run it without a login session)", err)
}

func isPermissionError(err error) bool {
	if errors.Is(err, os.ErrPermission) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access is denied") ||
		strings.Contains(msg, "operation not permitted")
}

// noopTunnelServiceProgram satisfies service.Interface for install/uninstall/
// status management; the installed service runs `ios tunnel start` directly
// and never enters this program.
type noopTunnelServiceProgram struct{}

func (noopTunnelServiceProgram) Start(service.Service) error { return nil }
func (noopTunnelServiceProgram) Stop(service.Service) error  { return nil }

func isTunnelServiceCommand(args docopt.Opts) bool {
	return boolArg(args, "tunnel") && boolArg(args, "service")
}

func runTunnelServiceCommand(ctx commandContext) {
	executable, err := os.Executable()
	exitIfError("failed determining the ios binary path", err)
	executable, err = filepath.Abs(executable)
	exitIfError("failed determining the ios binary path", err)
	workingDir, err := os.Getwd()
	exitIfError("failed determining the working directory", err)

	cfg := tunnelServiceConfig(ctx.Args, os.Getenv, executable, workingDir)
	ctl, err := service.New(noopTunnelServiceProgram{}, cfg)
	exitIfError("no supported service manager detected (need systemd, launchd, or the Windows service manager)", err)

	userService := boolArg(ctx.Args, "--user")
	switch {
	case boolArg(ctx.Args, "install"):
		if userService && !boolArg(ctx.Args, "--userspace") {
			slog.Warn("installing a user service without --userspace: the kernel tunnel needs root, so the service will likely fail to start; pass --userspace for a root-less agent")
		}
		err := installTunnelService(ctl, runtime.GOOS, userService)
		exitIfError("failed installing the tunnel service", err)
		commandLine := strings.Join(append([]string{cfg.Executable}, cfg.Arguments...), " ")
		slog.Info("tunnel service installed", "name", tunnelServiceName, "platform", ctl.Platform(), "command", commandLine)
		printTunnelServiceState(ctl, "installed")
	case boolArg(ctx.Args, "uninstall"):
		removed, err := uninstallTunnelService(ctl, userService)
		exitIfError("failed uninstalling the tunnel service", err)
		if removed {
			printTunnelServiceState(ctl, "uninstalled")
		} else {
			printTunnelServiceState(ctl, "not installed")
		}
	case boolArg(ctx.Args, "status"):
		status, err := ctl.Status()
		if isServiceNotInstalled(err) {
			printTunnelServiceState(ctl, "not installed")
			return
		}
		exitIfError("failed querying the tunnel service status", err)
		printTunnelServiceState(ctl, tunnelServiceStatusString(status))
	}
}

func printTunnelServiceState(ctl serviceController, state string) {
	if JSONdisabled {
		fmt.Printf("Service: %s\n  Platform: %s\n  Status: %s\n", tunnelServiceName, ctl.Platform(), state)
		return
	}
	fmt.Println(convertToJSONString(map[string]string{
		"name":     tunnelServiceName,
		"platform": ctl.Platform(),
		"status":   state,
	}))
}

func tunnelServiceStatusString(status service.Status) string {
	switch status {
	case service.StatusRunning:
		return "running"
	case service.StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}
