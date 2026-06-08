//go:build e2e

// Package harness holds the shared plumbing for the go-ios real-device e2e
// suites. It builds the ios binary once per suite process and exposes
// per-device test helpers, so the tunnel-free suite (test/e2e) and the
// tunnel-requiring suite (test/e2e/tunnel) stay thin and consistent.
package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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

// Devices returns the parsed GO_IOS_E2E_DEVICES list.
func Devices() []string { return append([]string(nil), devices...) }

// TryRun runs ios with exactly args (nothing appended) and returns stdout,
// stderr and the run error (nil on exit 0) without failing the test. Use it for
// negative tests that expect the command to fail.
func TryRun(t *testing.T, args ...string) (stdout, stderr []byte, err error) {
	t.Helper()
	var o, e bytes.Buffer
	cmd := exec.Command(iosBin, args...)
	cmd.Stdout = &o
	cmd.Stderr = &e
	err = cmd.Run()
	return o.Bytes(), e.Bytes(), err
}

// AuditAfterLaunch launches bundleID and runs the accessibility audit against
// it, asserting it reports at least one issue and that each issue carries an
// issueType. The audit targets the frontmost app, and `launch` is asynchronous,
// so the first audits can come back empty before the app settles into the
// foreground — poll until the audit reports issues (or time out) rather than
// auditing once after a fixed sleep, which flaked when the app was slow to show.
func AuditAfterLaunch(t *testing.T, udid, bundleID string) {
	t.Helper()
	RunForDevice(t, udid, "launch", bundleID)

	var issues []any
	deadline := time.Now().Add(30 * time.Second)
	for {
		out, _, err := TryRun(t, "ax", "audit", "--udid="+udid)
		if err == nil && json.Unmarshal(out, &issues) == nil && len(issues) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("ax audit reported no issues within 30s for %s after launching %s", udid, bundleID)
		}
		time.Sleep(2 * time.Second)
	}

	for i, raw := range issues {
		m, _ := raw.(map[string]any)
		if _, ok := m["issueType"]; !ok {
			t.Fatalf("ax audit issue %d missing issueType: %v", i, m)
		}
	}
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
// nothing was written to stdout. It returns the captured stdout for the caller
// to inspect. Use this for commands that run until killed.
func StreamSmoke(t *testing.T, udid string, window time.Duration, args ...string) []byte {
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

	b := out.Bytes()
	if len(bytes.TrimSpace(b)) == 0 {
		t.Fatalf("ios %v: no streamed output within %s", args, window)
	}
	return b
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

// SmokeJSON runs ios for the device and fails the test unless stdout is
// non-empty AND valid JSON. Most go-ios commands emit JSON by default, so this
// catches error text, log noise, or truncated output that a bare non-empty
// check (Smoke) would let through. Returns the raw stdout for further checks.
func SmokeJSON(t *testing.T, udid string, args ...string) []byte {
	t.Helper()
	out := Smoke(t, udid, args...)
	if !json.Valid(bytes.TrimSpace(out)) {
		t.Fatalf("ios %v: output is not valid JSON:\n%s", args, snippet(out))
	}
	return out
}

// SmokeJSONObject runs ios, asserts the output is a JSON object containing every
// key in requiredKeys, and returns the decoded map for further value checks.
func SmokeJSONObject(t *testing.T, udid string, requiredKeys []string, args ...string) map[string]any {
	t.Helper()
	out := Smoke(t, udid, args...)
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("ios %v: not a JSON object: %v\n%s", args, err, snippet(out))
	}
	for _, k := range requiredKeys {
		if _, ok := m[k]; !ok {
			t.Fatalf("ios %v: missing expected key %q (got %v)", args, k, keysOf(m))
		}
	}
	return m
}

// SmokeJSONArray runs ios, asserts the output is a non-empty JSON array, and
// (when elements are objects) that the first element contains every key in
// elemKeys. Returns the decoded slice.
func SmokeJSONArray(t *testing.T, udid string, elemKeys []string, args ...string) []any {
	t.Helper()
	out := Smoke(t, udid, args...)
	var a []any
	if err := json.Unmarshal(out, &a); err != nil {
		t.Fatalf("ios %v: not a JSON array: %v\n%s", args, err, snippet(out))
	}
	if len(a) == 0 {
		t.Fatalf("ios %v: empty array", args)
	}
	if first, ok := a[0].(map[string]any); ok {
		for _, k := range elemKeys {
			if _, ok := first[k]; !ok {
				t.Fatalf("ios %v: array element missing key %q (got %v)", args, k, keysOf(first))
			}
		}
	}
	return a
}

func keysOf(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// SmokeContains runs ios for the device and fails the test unless stdout
// contains want. Use for commands whose output is not JSON (e.g. text listings).
func SmokeContains(t *testing.T, udid, want string, args ...string) []byte {
	t.Helper()
	out := Smoke(t, udid, args...)
	if !bytes.Contains(out, []byte(want)) {
		t.Fatalf("ios %v: output does not contain %q:\n%s", args, want, snippet(out))
	}
	return out
}

var (
	snapshotOnce sync.Once
	snapshots    map[string]map[string]string
)

// ExpectedDevice returns the recorded identity snapshot for a device UDID (from
// test/e2e/testdata/devices.json) and whether one exists. The test devices are
// not updated, so identity fields (ProductType, ProductVersion, ...) are stable
// and can be asserted exactly. Unknown UDIDs return ok=false so tests can skip
// the exact-match check rather than fail for a newly added device.
func ExpectedDevice(udid string) (map[string]string, bool) {
	snapshotOnce.Do(func() {
		root, err := repoRoot()
		if err != nil {
			return
		}
		b, err := os.ReadFile(filepath.Join(root, "test", "e2e", "testdata", "devices.json"))
		if err != nil {
			return
		}
		_ = json.Unmarshal(b, &snapshots)
	})
	d, ok := snapshots[udid]
	return d, ok
}

func snippet(b []byte) string {
	const max = 400
	if len(b) > max {
		return string(b[:max]) + "..."
	}
	return string(b)
}

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// StartBackground starts an ios command that runs until signalled (e.g.
// "devicestate enable" or "setlocation", which hold device state only while
// running), capturing its combined output. It returns output() to read what the
// command has printed so far, and stop() which sends stopSig to the process
// group so the command can clean up (reverting device state), escalating to
// SIGKILL if it does not exit promptly. The caller must defer stop().
//
// stopSig must match the signal the command waits for: devicestate enable uses
// SIGTERM, setlocation uses SIGINT (os.Interrupt).
func StartBackground(t *testing.T, udid string, stopSig syscall.Signal, args ...string) (output func() string, stop func()) {
	t.Helper()
	return StartBackgroundWithEnv(t, udid, nil, stopSig, args...)
}

// StartBackgroundWithEnv is StartBackground with extra environment variables.
func StartBackgroundWithEnv(t *testing.T, udid string, extraEnv []string, stopSig syscall.Signal, args ...string) (output func() string, stop func()) {
	t.Helper()
	buf := &syncBuf{}
	cmd := exec.Command(iosBin, append(args, "--udid="+udid)...)
	cmd.Stdout = buf
	cmd.Stderr = buf
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("ios %v: start: %v", args, err)
	}
	return buf.String, func() {
		pgid := cmd.Process.Pid
		_ = syscall.Kill(-pgid, stopSig)
		done := make(chan struct{})
		go func() { _ = cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			<-done
		}
	}
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
