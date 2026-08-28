package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// discoveryFileName is the fixed name of the discovery file inside the go-ios
// home directory. SDKs (TS/Python/Java/C#) read this exact path to auto-find a
// locally running REST daemon, so the name and JSON shape are a contract shared
// across every language and must not drift. See DISCOVERY-CONTRACT.md.
const discoveryFileName = "rest-api.json"

// DiscoveryInfo is the JSON payload written to <home>/rest-api.json after a
// successful bind. baseUrl is authoritative (scheme + host + port); the other
// fields are informational. The struct tags define the wire format the SDKs
// depend on.
type DiscoveryInfo struct {
	// BaseUrl is the full URL a client should use, including scheme (http/https),
	// host and the real (possibly OS-assigned) port. Authoritative.
	BaseUrl string `json:"baseUrl"`
	// Host is the bound host (e.g. 127.0.0.1). Informational.
	Host string `json:"host"`
	// Port is the real bound TCP port. Informational.
	Port int `json:"port"`
	// Pid is the daemon process id, so an SDK can detect a stale file.
	Pid int `json:"pid"`
	// StartedAt is when the daemon bound, in RFC3339 UTC.
	StartedAt string `json:"startedAt"`
	// Tls reports whether the daemon is serving HTTPS.
	Tls bool `json:"tls"`
}

// goIosHome resolves the go-ios home directory: GO_IOS_HOME if set and
// non-empty, otherwise <user home>/.go-ios. It does not create the directory.
func goIosHome() (string, error) {
	if h := os.Getenv("GO_IOS_HOME"); h != "" {
		return h, nil
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home dir for go-ios home: %w", err)
	}
	return filepath.Join(userHome, ".go-ios"), nil
}

// discoveryFilePath returns the absolute path of the discovery file, creating
// the home directory (0700) if it does not yet exist.
func discoveryFilePath() (string, error) {
	home, err := goIosHome()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		return "", fmt.Errorf("creating go-ios home dir %q: %w", home, err)
	}
	return filepath.Join(home, discoveryFileName), nil
}

// writeDiscoveryFile atomically writes info to <home>/rest-api.json with mode
// 0600. It writes to a temp file in the same directory and renames it over the
// destination so readers never observe a partially written file. It returns the
// path that was written so the caller can log and later remove it.
func writeDiscoveryFile(info DiscoveryInfo) (string, error) {
	path, err := discoveryFilePath()
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling discovery info: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, discoveryFileName+".*.tmp")
	if err != nil {
		return "", fmt.Errorf("creating temp discovery file in %q: %w", dir, err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup of the temp file on any failure past this point.
	defer func() { _ = os.Remove(tmpName) }()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("chmod temp discovery file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("writing temp discovery file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("closing temp discovery file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return "", fmt.Errorf("renaming discovery file into place: %w", err)
	}
	return path, nil
}

// removeDiscoveryFile deletes the discovery file. A missing file is not an
// error (the daemon may never have written one, or it was cleaned already).
func removeDiscoveryFile() error {
	home, err := goIosHome()
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(home, discoveryFileName)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// newDiscoveryInfo builds a DiscoveryInfo for a daemon bound at host:port. tls
// selects the scheme (https vs http) reflected in BaseUrl and the Tls field.
func newDiscoveryInfo(host string, port int, tls bool) DiscoveryInfo {
	scheme := "http"
	if tls {
		scheme = "https"
	}
	return DiscoveryInfo{
		BaseUrl:   fmt.Sprintf("%s://%s:%d", scheme, host, port),
		Host:      host,
		Port:      port,
		Pid:       os.Getpid(),
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Tls:       tls,
	}
}
