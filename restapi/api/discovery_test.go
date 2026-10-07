package api

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

// TestParseServerConfigDefaultAddrIsEphemeralLoopback pins the security-relevant
// default: with no --addr the daemon binds an ephemeral loopback port
// (127.0.0.1:0) rather than squatting on :8080 / all interfaces.
func TestParseServerConfigDefaultAddrIsEphemeralLoopback(t *testing.T) {
	cfg := parseServerConfig(nil)
	if cfg.addr != "127.0.0.1:0" {
		t.Fatalf("default addr = %q, want 127.0.0.1:0", cfg.addr)
	}
}

// TestParseServerConfigAddrOverride confirms --addr still pins/exposes.
func TestParseServerConfigAddrOverride(t *testing.T) {
	cfg := parseServerConfig([]string{"--addr", "0.0.0.0:9000"})
	if cfg.addr != "0.0.0.0:9000" {
		t.Fatalf("override addr = %q, want 0.0.0.0:9000", cfg.addr)
	}
}

// TestGoIosHomeHonorsEnv checks GO_IOS_HOME wins over the default ~/.go-ios.
func TestGoIosHomeHonorsEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GO_IOS_HOME", dir)
	got, err := goIosHome()
	if err != nil {
		t.Fatalf("goIosHome: %v", err)
	}
	if got != dir {
		t.Fatalf("goIosHome = %q, want %q", got, dir)
	}
}

func TestGoIosHomeDefault(t *testing.T) {
	t.Setenv("GO_IOS_HOME", "")
	got, err := goIosHome()
	if err != nil {
		t.Fatalf("goIosHome: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no user home dir: %v", err)
	}
	if want := filepath.Join(home, ".go-ios"); got != want {
		t.Fatalf("goIosHome = %q, want %q", got, want)
	}
}

// TestWriteDiscoveryFileShapeAndPerms binds a real ephemeral listener, writes
// the discovery file with the actual bound port and asserts the JSON shape,
// values, path, and 0600 permissions.
func TestWriteDiscoveryFileShapeAndPerms(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GO_IOS_HOME", home)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	tcpAddr := ln.Addr().(*net.TCPAddr)
	host := tcpAddr.IP.String()
	port := tcpAddr.Port
	if port == 0 {
		t.Fatal("expected a non-zero OS-assigned port")
	}

	info := newDiscoveryInfo(host, port, false)
	path, err := writeDiscoveryFile(info)
	if err != nil {
		t.Fatalf("writeDiscoveryFile: %v", err)
	}

	wantPath := filepath.Join(home, "rest-api.json")
	if path != wantPath {
		t.Fatalf("discovery path = %q, want %q", path, wantPath)
	}

	// 0600 perms (skip the numeric check on Windows, which has no POSIX bits).
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if perm := fi.Mode().Perm(); perm != 0o600 {
			t.Fatalf("discovery file perm = %o, want 600", perm)
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got DiscoveryInfo
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	wantBase := "http://" + host + ":" + strconv.Itoa(port)
	if got.BaseUrl != wantBase {
		t.Fatalf("baseUrl = %q, want %q", got.BaseUrl, wantBase)
	}
	if got.Host != host {
		t.Fatalf("host = %q, want %q", got.Host, host)
	}
	if got.Port != port {
		t.Fatalf("port = %d, want %d", got.Port, port)
	}
	if got.Pid != os.Getpid() {
		t.Fatalf("pid = %d, want %d", got.Pid, os.Getpid())
	}
	if got.Tls {
		t.Fatal("tls = true, want false")
	}
	if got.StartedAt == "" {
		t.Fatal("startedAt is empty")
	}

	// The raw JSON must use the exact contract keys the SDKs read.
	var keyed map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keyed); err != nil {
		t.Fatalf("unmarshal keyed: %v", err)
	}
	for _, k := range []string{"baseUrl", "host", "port", "pid", "startedAt", "tls"} {
		if _, ok := keyed[k]; !ok {
			t.Fatalf("discovery JSON missing key %q", k)
		}
	}
}

// TestWriteDiscoveryFileTLS asserts the scheme flips to https and tls=true.
func TestWriteDiscoveryFileTLS(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GO_IOS_HOME", home)

	info := newDiscoveryInfo("127.0.0.1", 8443, true)
	if _, err := writeDiscoveryFile(info); err != nil {
		t.Fatalf("writeDiscoveryFile: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(home, "rest-api.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got DiscoveryInfo
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.BaseUrl != "https://127.0.0.1:8443" {
		t.Fatalf("baseUrl = %q, want https://127.0.0.1:8443", got.BaseUrl)
	}
	if !got.Tls {
		t.Fatal("tls = false, want true")
	}
}

// TestWriteDiscoveryFileCreatesHomeDir verifies the home dir is created (0700)
// when it does not already exist.
func TestWriteDiscoveryFileCreatesHomeDir(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "nested", "go-ios-home")
	t.Setenv("GO_IOS_HOME", home)

	if _, err := writeDiscoveryFile(newDiscoveryInfo("127.0.0.1", 1234, false)); err != nil {
		t.Fatalf("writeDiscoveryFile: %v", err)
	}
	fi, err := os.Stat(home)
	if err != nil {
		t.Fatalf("home dir not created: %v", err)
	}
	if !fi.IsDir() {
		t.Fatal("home path is not a directory")
	}
}

// TestWriteDiscoveryFileAtomicOverwrite confirms a second write replaces the
// first cleanly (temp+rename) and leaves no stray temp files behind.
func TestWriteDiscoveryFileAtomicOverwrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GO_IOS_HOME", home)

	if _, err := writeDiscoveryFile(newDiscoveryInfo("127.0.0.1", 1111, false)); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if _, err := writeDiscoveryFile(newDiscoveryInfo("127.0.0.1", 2222, false)); err != nil {
		t.Fatalf("second write: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(home, "rest-api.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got DiscoveryInfo
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Port != 2222 {
		t.Fatalf("port = %d, want 2222 (overwrite failed)", got.Port)
	}

	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "rest-api.json" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("home dir should contain only rest-api.json, got %v", names)
	}
}

// TestRemoveDiscoveryFile verifies removal (the shutdown path) deletes the file
// and is idempotent when the file is already gone.
func TestRemoveDiscoveryFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GO_IOS_HOME", home)

	if _, err := writeDiscoveryFile(newDiscoveryInfo("127.0.0.1", 3333, false)); err != nil {
		t.Fatalf("write: %v", err)
	}
	path := filepath.Join(home, "rest-api.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist before removal: %v", err)
	}

	if err := removeDiscoveryFile(); err != nil {
		t.Fatalf("removeDiscoveryFile: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file should be gone after removal, stat err = %v", err)
	}

	// Idempotent: removing again is not an error.
	if err := removeDiscoveryFile(); err != nil {
		t.Fatalf("second removeDiscoveryFile: %v", err)
	}
}
