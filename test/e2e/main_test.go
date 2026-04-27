//go:build e2e

package e2e_test

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

func TestMain(m *testing.M) {
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

// runIOS executes the ios binary with the given args and returns stdout.
// On non-zero exit it fails the test with stderr + stdout for debugging.
func runIOS(t *testing.T, args ...string) []byte {
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

// runIOSForDevice runs ios with --udid=<udid> appended.
func runIOSForDevice(t *testing.T, udid string, args ...string) []byte {
	t.Helper()
	return runIOS(t, append(args, "--udid="+udid)...)
}

// smoke runs ios for the given device and fails the test if stdout is empty.
// It returns the captured stdout for further inspection by the caller.
func smoke(t *testing.T, udid string, args ...string) []byte {
	t.Helper()
	out := runIOSForDevice(t, udid, args...)
	if len(bytes.TrimSpace(out)) == 0 {
		t.Fatalf("ios %v: empty output", args)
	}
	return out
}

// forEachDevice runs fn as a parallel subtest per UDID from GO_IOS_E2E_DEVICES.
// Fails the parent test if the env var is unset.
func forEachDevice(t *testing.T, fn func(t *testing.T, udid string)) {
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
