//go:build e2e

// Package harness holds the shared plumbing for the go-ios real-device e2e
// suites. It builds the ios binary once per suite process and exposes
// per-device test helpers, so the tunnel-free suite (test/e2e) and the
// tunnel-requiring suite (test/e2e/tunnel) stay thin and consistent.
package harness

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	iosBin  string
	devices []string
)

// Main builds the ios binary, parses the device list from GO_IOS_E2E_DEVICES,
// runs any setup hooks, and runs the suite. Call it from a TestMain in each
// suite package; the tunnel suite passes MountDeveloperImage as a setup hook.
func Main(m *testing.M, setup ...func()) {
	root, err := repoRoot()
	if err != nil {
		panic(err)
	}

	dir, err := os.MkdirTemp("", "ios-e2e-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	iosBin = filepath.Join(dir, "ios")
	cmd := exec.Command("go", "build", "-o", iosBin, ".")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + err.Error() + "\n" + string(out))
	}

	if raw := strings.TrimSpace(os.Getenv("GO_IOS_E2E_DEVICES")); raw != "" {
		for _, u := range strings.Split(raw, ",") {
			if u = strings.TrimSpace(u); u != "" {
				devices = append(devices, u)
			}
		}
	}

	for _, s := range setup {
		s()
	}

	os.Exit(m.Run())
}

// MountDeveloperImage downloads and mounts the developer disk image on every
// configured device. The tunnel suite uses it as a setup hook: CoreDevice
// services such as "info display" require a mounted DDI, and a device reboot
// (e.g. after enabling Developer Mode) unmounts it. Best-effort: failures are
// logged, and the individual tests that need the DDI will report clear errors.
func MountDeveloperImage() {
	imgDir, err := os.MkdirTemp("", "ddi-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness: could not create DDI temp dir: %v\n", err)
		return
	}
	for _, udid := range devices {
		out, err := exec.Command(iosBin, "image", "auto", "--basedir="+imgDir, "--udid="+udid).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "harness: image auto failed for %s: %v\n%s\n", udid, err, out)
		}
	}
}

func repoRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSpace(string(out))), nil
}

// RunIOS executes the ios binary with the given args and returns stdout.
// On non-zero exit it fails the test with stderr + stdout for debugging.
func RunIOS(t *testing.T, args ...string) []byte {
	t.Helper()
	var stderr bytes.Buffer
	cmd := exec.Command(iosBin, args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ios %v: %v\nstderr: %s\nstdout: %s", args, err, stderr.String(), out)
	}
	return out
}

// RunForDevice runs ios with --udid=<udid> appended.
func RunForDevice(t *testing.T, udid string, args ...string) []byte {
	t.Helper()
	return RunIOS(t, append(args, "--udid="+udid)...)
}

// Smoke runs ios for the given device and fails the test if stdout is empty.
// It returns the captured stdout for further inspection by the caller.
func Smoke(t *testing.T, udid string, args ...string) []byte {
	t.Helper()
	out := RunForDevice(t, udid, args...)
	if len(bytes.TrimSpace(out)) == 0 {
		t.Fatalf("ios %v: empty output", args)
	}
	return out
}

// StreamSmoke runs a streaming ios command (e.g. syslog) for the given device,
// lets it stream for window, then kills its process group and fails the test if
// nothing was written to stdout. Use this for commands that run until killed.
func StreamSmoke(t *testing.T, udid string, window time.Duration, args ...string) {
	t.Helper()
	var out bytes.Buffer
	cmd := exec.Command(iosBin, append(args, "--udid="+udid)...)
	cmd.Stdout = &out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // own group so we can kill children too
	if err := cmd.Start(); err != nil {
		t.Fatalf("ios %v: start: %v", args, err)
	}

	time.Sleep(window)
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	_ = cmd.Wait() // returns the kill signal error; ignored

	if len(bytes.TrimSpace(out.Bytes())) == 0 {
		t.Fatalf("ios %v: no streamed output within %s", args, window)
	}
}

// StreamInTempDir runs a streaming ios command in a fresh temp directory for
// window, then kills its process group, and returns the directory so the caller
// can inspect files the command wrote there (e.g. pcap's dump-*.pcap).
func StreamInTempDir(t *testing.T, udid string, window time.Duration, args ...string) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command(iosBin, append(args, "--udid="+udid)...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // own group so we can kill children too
	if err := cmd.Start(); err != nil {
		t.Fatalf("ios %v: start: %v", args, err)
	}

	time.Sleep(window)
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	_ = cmd.Wait() // returns the kill signal error; ignored

	return dir
}

// ForEachDevice runs fn as a parallel subtest per UDID from GO_IOS_E2E_DEVICES.
// Fails the parent test if the env var is unset.
func ForEachDevice(t *testing.T, fn func(t *testing.T, udid string)) {
	t.Helper()
	if len(devices) == 0 {
		t.Fatal("GO_IOS_E2E_DEVICES not set: at least one UDID is required")
	}
	for _, udid := range devices {
		udid := udid
		t.Run(udid, func(t *testing.T) {
			t.Parallel()
			fn(t, udid)
		})
	}
}
