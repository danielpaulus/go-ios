//go:build e2e

// Package harness holds the shared plumbing for the go-ios real-device e2e
// suites. It builds the ios binary once per suite process and exposes
// per-device test helpers, so the tunnel-free suite (test/e2e) and the
// tunnel-requiring suite (test/e2e/tunnel) stay thin and consistent.
package harness

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	iosBin  string
	devices []string
)

// Main builds the ios binary, parses the device list from GO_IOS_E2E_DEVICES,
// and runs the suite. Call it from a TestMain in each suite package.
func Main(m *testing.M) {
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

	os.Exit(m.Run())
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
