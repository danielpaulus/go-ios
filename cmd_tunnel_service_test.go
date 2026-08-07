package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/docopt/docopt-go"
	"github.com/kardianos/service"
)

var updateTunnelServiceGolden = flag.Bool("update-tunnel-service-golden", false, "update tunnel service golden files")

// renderServiceDefinition renders the captured service definition into a
// stable, human-reviewable form for golden-file comparison.
func renderServiceDefinition(cfg *service.Config) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Name: %s\n", cfg.Name)
	fmt.Fprintf(&b, "DisplayName: %s\n", cfg.DisplayName)
	fmt.Fprintf(&b, "Description: %s\n", cfg.Description)
	fmt.Fprintf(&b, "Executable: %s\n", cfg.Executable)
	fmt.Fprintf(&b, "Arguments: %s\n", strings.Join(cfg.Arguments, " "))
	fmt.Fprintf(&b, "WorkingDirectory: %s\n", cfg.WorkingDirectory)
	optionKeys := make([]string, 0, len(cfg.Option))
	for key := range cfg.Option {
		optionKeys = append(optionKeys, key)
	}
	sort.Strings(optionKeys)
	for _, key := range optionKeys {
		fmt.Fprintf(&b, "Option.%s: %v\n", key, cfg.Option[key])
	}
	envKeys := make([]string, 0, len(cfg.EnvVars))
	for key := range cfg.EnvVars {
		envKeys = append(envKeys, key)
	}
	sort.Strings(envKeys)
	for _, key := range envKeys {
		fmt.Fprintf(&b, "Env.%s: %s\n", key, cfg.EnvVars[key])
	}
	return b.String()
}

func assertTunnelServiceGolden(t *testing.T, goldenName string, got string) {
	t.Helper()
	path := filepath.Join("testdata", "tunnelservice", goldenName)
	if *updateTunnelServiceGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run 'go test -run TestTunnelServiceConfig -update-tunnel-service-golden' to create it)", goldenName, err)
	}
	if got != string(wantBytes) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", goldenName, got, string(wantBytes))
	}
}

func envFromMap(env map[string]string) func(string) string {
	return func(name string) string { return env[name] }
}

func TestTunnelServiceConfigCapture(t *testing.T) {
	tests := []struct {
		name   string
		args   docopt.Opts
		env    map[string]string
		golden string
	}{
		{
			name:   "defaults",
			args:   docopt.Opts{"tunnel": true, "service": true, "install": true},
			env:    map[string]string{},
			golden: "defaults.golden",
		},
		{
			name: "cloud_userspace_user",
			args: docopt.Opts{
				"tunnel": true, "service": true, "install": true,
				"--userspace":        true,
				"--user":             true,
				"--pair-record-path": "records",
			},
			env: map[string]string{
				"ORCHESTRATOR_URL":       "https://orchestrator.example.com",
				"GO_IOS_AGENT_HOST":      "0.0.0.0",
				"GO_IOS_AGENT_PORT":      "60106",
				"USBMUXD_SOCKET_ADDRESS": "UNIX:///var/run/usbmuxd",
				"UNRELATED_VAR":          "must-not-be-captured",
			},
			golden: "cloud_userspace_user.golden",
		},
		{
			name: "per_device_agent",
			args: docopt.Opts{
				"tunnel": true, "service": true, "install": true,
				"--userspace":        true,
				"--pair-record-path": "/opt/go-ios/records",
				"--udid":             "00008110-000A2DE21E10801E",
				"--tunnel-info-host": "127.0.0.1",
				"--tunnel-info-port": "28200",
			},
			env:    map[string]string{},
			golden: "per_device_agent.golden",
		},
		{
			name: "udid_from_environment",
			args: docopt.Opts{"tunnel": true, "service": true, "install": true, "--userspace": true},
			env: map[string]string{
				"GO_IOS_UDID": "00008110-000A2DE21E10801E",
			},
			golden: "udid_from_environment.golden",
		},
		{
			name: "pair_record_path_default_literal",
			args: docopt.Opts{
				"tunnel": true, "service": true, "install": true,
				"--pair-record-path": "default",
			},
			env:    map[string]string{},
			golden: "pair_record_path_default_literal.golden",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tunnelServiceConfig(tt.args, envFromMap(tt.env), "/usr/local/bin/ios", "/home/operator")
			assertTunnelServiceGolden(t, tt.golden, renderServiceDefinition(cfg))
		})
	}
}

// fakeServiceController records the actions the install/uninstall flows take.
type fakeServiceController struct {
	platform  string
	installed bool
	running   bool

	installErr   error
	uninstallErr error
	startErr     error
	stopErr      error
	statusErr    error

	actions []string
}

func (f *fakeServiceController) Install() error {
	f.actions = append(f.actions, "install")
	if f.installErr != nil {
		return f.installErr
	}
	f.installed = true
	return nil
}

func (f *fakeServiceController) Uninstall() error {
	f.actions = append(f.actions, "uninstall")
	if f.uninstallErr != nil {
		return f.uninstallErr
	}
	if !f.installed {
		return service.ErrNotInstalled
	}
	f.installed = false
	return nil
}

func (f *fakeServiceController) Start() error {
	f.actions = append(f.actions, "start")
	if f.startErr != nil {
		return f.startErr
	}
	f.running = true
	return nil
}

func (f *fakeServiceController) Stop() error {
	f.actions = append(f.actions, "stop")
	if f.stopErr != nil {
		return f.stopErr
	}
	f.running = false
	return nil
}

func (f *fakeServiceController) Status() (service.Status, error) {
	if f.statusErr != nil {
		return service.StatusUnknown, f.statusErr
	}
	if !f.installed {
		return service.StatusUnknown, service.ErrNotInstalled
	}
	if f.running {
		return service.StatusRunning, nil
	}
	return service.StatusStopped, nil
}

func (f *fakeServiceController) Platform() string { return f.platform }

func TestInstallTunnelServiceNoSystemd(t *testing.T) {
	for _, platform := range []string{"unix-systemv", "linux-upstart", "linux-openrc"} {
		fake := &fakeServiceController{platform: platform}
		err := installTunnelService(fake, "linux", false)
		if err == nil {
			t.Fatalf("platform %s: expected an error, got none", platform)
		}
		if !strings.Contains(err.Error(), "no systemd detected") {
			t.Fatalf("platform %s: error should explain that systemd is required, got: %v", platform, err)
		}
		if len(fake.actions) != 0 {
			t.Fatalf("platform %s: nothing must be installed on unsupported init systems, got actions %v", platform, fake.actions)
		}
	}
}

func TestInstallTunnelServiceSupportedPlatformsAllowed(t *testing.T) {
	for goos, platform := range map[string]string{"darwin": "darwin-launchd", "linux": "linux-systemd"} {
		fake := &fakeServiceController{platform: platform}
		if err := installTunnelService(fake, goos, false); err != nil {
			t.Fatalf("%s/%s: unexpected error: %v", goos, platform, err)
		}
		if !fake.installed || !fake.running {
			t.Fatalf("%s/%s: service should be installed and started", goos, platform)
		}
	}
}

// installTunnelService must refuse to register a Windows service: the tunnel
// agent does not implement the Windows service-control dispatcher, so the SCM
// would kill it with error 1053 on start. Rejecting it up front (leaving
// nothing installed) beats registering a service that can never run.
func TestInstallTunnelServiceWindowsRejected(t *testing.T) {
	fake := &fakeServiceController{platform: "windows-service"}
	err := installTunnelService(fake, "windows", false)
	if err == nil {
		t.Fatal("expected Windows to be rejected, got no error")
	}
	if !strings.Contains(err.Error(), "not supported on Windows") {
		t.Fatalf("error should explain Windows is unsupported, got: %v", err)
	}
	if len(fake.actions) != 0 {
		t.Fatalf("nothing must be installed on Windows, got actions %v", fake.actions)
	}
}

func TestInstallTunnelServiceInsufficientPrivileges(t *testing.T) {
	fake := &fakeServiceController{
		platform:   "linux-systemd",
		installErr: fmt.Errorf("open /etc/systemd/system/go-ios-tunnel.service: %w", os.ErrPermission),
	}
	err := installTunnelService(fake, "linux", false)
	if err == nil {
		t.Fatal("expected an error, got none")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Fatalf("original permission error must stay wrapped, got: %v", err)
	}
	for _, hint := range []string{"sudo", "--user"} {
		if !strings.Contains(err.Error(), hint) {
			t.Fatalf("error should hint at %q, got: %v", hint, err)
		}
	}
}

func TestInstallTunnelServiceUserServicePrivilegeError(t *testing.T) {
	fake := &fakeServiceController{
		platform:   "linux-systemd",
		installErr: errors.New("permission denied"),
	}
	err := installTunnelService(fake, "linux", true)
	if err == nil {
		t.Fatal("expected an error, got none")
	}
	if strings.Contains(err.Error(), "sudo") {
		t.Fatalf("user-service error must not recommend sudo, got: %v", err)
	}
	if !strings.Contains(err.Error(), "insufficient privileges") {
		t.Fatalf("error should mention insufficient privileges, got: %v", err)
	}
}

func TestInstallTunnelServiceIdempotent(t *testing.T) {
	fake := &fakeServiceController{platform: "linux-systemd", installed: true, running: true}
	if err := installTunnelService(fake, "linux", false); err != nil {
		t.Fatalf("reinstall over an existing service failed: %v", err)
	}
	want := []string{"stop", "uninstall", "install", "start"}
	if strings.Join(fake.actions, ",") != strings.Join(want, ",") {
		t.Fatalf("actions = %v, want %v", fake.actions, want)
	}
	if !fake.installed || !fake.running {
		t.Fatal("service should be installed and running after reinstall")
	}
}

func TestUninstallTunnelService(t *testing.T) {
	fake := &fakeServiceController{platform: "linux-systemd", installed: true, running: true}
	removed, err := uninstallTunnelService(fake, false)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	if !removed {
		t.Fatal("expected the service to be removed")
	}
	if fake.installed || fake.running {
		t.Fatal("service should be stopped and removed")
	}
}

func TestUninstallTunnelServiceNotInstalled(t *testing.T) {
	fake := &fakeServiceController{platform: "linux-systemd"}
	removed, err := uninstallTunnelService(fake, false)
	if err != nil {
		t.Fatalf("uninstalling a missing service must not fail: %v", err)
	}
	if removed {
		t.Fatal("nothing should be reported as removed")
	}
	if len(fake.actions) != 0 {
		t.Fatalf("no service manager actions expected, got %v", fake.actions)
	}
}

func TestIsTunnelServiceCommand(t *testing.T) {
	if !isTunnelServiceCommand(docopt.Opts{"tunnel": true, "service": true, "install": true}) {
		t.Fatal("tunnel service install must match")
	}
	if isTunnelServiceCommand(docopt.Opts{"tunnel": true, "start": true}) {
		t.Fatal("tunnel start must not match")
	}
	if isTunnelServiceCommand(docopt.Opts{"install": true}) {
		t.Fatal("app install must not match")
	}
}
