//go:build e2e

package e2e_test

// These tests are opt-in inside the normal real-device suite: they run when the
// shared signing identity is available (GO_IOS_E2E_SIGNING_P12_B64 +
// GO_IOS_E2E_SIGNING_CERT_ID, refreshed by the refresh-signing-identity
// workflow) plus the App Store Connect credentials to mint per-bundle profiles
// (GO_IOS_E2E_ASC_KEY_ID, GO_IOS_E2E_ASC_ISSUER_ID, GO_IOS_E2E_ASC_PRIVATE_KEY).
// They provision, install, and run WDA / DeviceKit themselves via `ios ui run`,
// so no external server URL is needed. GO_IOS_E2E_WDA_PATH /
// GO_IOS_E2E_DEVICEKIT_PATH can point at a local .ipa/.zip/.app; otherwise
// GO_IOS_E2E_WDA_ARTIFACT_URL / GO_IOS_E2E_DEVICEKIT_ARTIFACT_URL are downloaded.

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

const (
	defaultE2EWDAArtifactURL       = "https://deviceboxhq.com/WebDriverAgentRunner-13.2.0.zip"
	defaultE2EDeviceKitArtifactURL = "https://deviceboxhq.com/devicekit-ios-runner-0.0.18.ipa"
	defaultE2EWDABundleID          = "com.deviceboxhq.goios.WebDriverAgentRunner.xctrunner"
	defaultE2EDeviceKitBundleID    = "com.deviceboxhq.goios.devicekit.runner"
	defaultE2EWDAConfig            = "WebDriverAgentRunner.xctest"
)

func TestUIInstallWDARunAndUI(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		wdaMu.Lock()
		defer wdaMu.Unlock()
		wdaURL, stop := runWDA(t, udid)
		defer stop()
		smoke(t, udid, "ui", "status", "--driver=wda", "--wda-url="+wdaURL)
		smoke(t, udid, "ui", "api", "--driver=wda", "--method=GET", "--http-path=/status", "--wda-url="+wdaURL)
		// Screen-capture (screenshot/stream) is disabled for now — the CI devices'
		// screens are often off. WDA `tap` is also skipped: go-ios drives it via the
		// /wda/tap/{x}/{y} endpoint, which this WDA build rejects ("unknown
		// command") — tracked separately. DeviceKit's suite covers tap interaction.
	})
}

// WebDriverAgent and DeviceKit each allow only one running instance per device
// (they bind a fixed device port), so the tests that bring one up must not run
// concurrently. These mutexes serialize per backend; WDA and DeviceKit still run
// in parallel with each other.
var (
	wdaMu       sync.Mutex
	deviceKitMu sync.Mutex
)

// runWDA provisions, installs, and runs WebDriverAgent via `ios ui run`,
// returning a base URL the host can reach and a stop func. Hold wdaMu.
func runWDA(t *testing.T, udid string) (string, func()) {
	return runUIBackend(t, udid, "wda", "WDA E2E", defaultE2EWDABundleID, "GO_IOS_E2E_WDA_BUNDLE_ID")
}

// runDeviceKit provisions, installs, and runs DeviceKit via `ios ui run`,
// returning a base URL the host can reach and a stop func. Hold deviceKitMu.
func runDeviceKit(t *testing.T, udid string) (string, func()) {
	return runUIBackend(t, udid, "devicekit", "DeviceKit E2E", defaultE2EDeviceKitBundleID, "GO_IOS_E2E_DEVICEKIT_BUNDLE_ID")
}

// runUIBackend provisions + installs the backend, runs it with `ios ui run`
// (which forwards a local port to the runner on the device), and polls until the
// backend answers. It returns the base URL and a stop func the caller must defer
// (before releasing the backend mutex, so the runner is torn down first).
func runUIBackend(t *testing.T, udid, target, label, defaultBundle, bundleEnv string) (string, func()) {
	t.Helper()
	bundleID := e2eEnv(bundleEnv)
	if bundleID == "" {
		bundleID = defaultBundle
	}
	p12Path, profilePath := provisionSigningAssets(t, udid, bundleID, label)

	runIOSForDevice(t, udid,
		"ui", "install", target,
		"--bundleid="+bundleID,
		"--p12file="+p12Path,
		"--profile="+profilePath,
		"--p12password=go-ios-e2e",
	)

	hostPort := freeLocalPort(t)
	_, stop := harness.StartBackground(t, udid, syscall.SIGINT,
		"ui", "run", target, "--bundleid="+bundleID, "--host-port="+strconv.Itoa(hostPort))

	localURL := fmt.Sprintf("http://127.0.0.1:%d", hostPort)
	driverArg, urlArg := uiDriverArgs(target, localURL)
	deadline := time.Now().Add(120 * time.Second)
	for {
		if _, _, err := harness.TryRun(t, "ui", "status", driverArg, urlArg, "--udid="+udid); err == nil {
			return localURL, stop
		}
		if time.Now().After(deadline) {
			stop()
			t.Fatalf("%s did not become reachable at %s within 120s", target, localURL)
		}
		time.Sleep(2 * time.Second)
	}
}

// uiDriverArgs returns the --driver and --*-url flags for a backend + URL.
func uiDriverArgs(target, baseURL string) (driverArg, urlArg string) {
	if target == "devicekit" {
		return "--driver=devicekit", "--devicekit-url=" + baseURL
	}
	return "--driver=wda", "--wda-url=" + baseURL
}

func freeLocalPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate local port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

func TestUIInstallDeviceKit(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		deviceKitMu.Lock()
		defer deviceKitMu.Unlock()
		bundleID := e2eEnv("GO_IOS_E2E_DEVICEKIT_BUNDLE_ID")
		if bundleID == "" {
			bundleID = defaultE2EDeviceKitBundleID
		}
		p12Path, profilePath := provisionSigningAssets(t, udid, bundleID, "DeviceKit E2E")

		runIOSForDevice(t, udid,
			"ui", "install", "devicekit",
			"--bundleid="+bundleID,
			"--p12file="+p12Path,
			"--profile="+profilePath,
			"--p12password=go-ios-e2e",
		)
	})
}

func TestSignCommandDeviceKit(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		deviceKitMu.Lock()
		defer deviceKitMu.Unlock()
		bundleID := e2eEnv("GO_IOS_E2E_DEVICEKIT_BUNDLE_ID")
		if bundleID == "" {
			bundleID = defaultE2EDeviceKitBundleID
		}
		p12Path, profilePath := provisionSigningAssets(t, udid, bundleID, "DeviceKit Sign E2E")
		appPath := prepareDeviceKitArtifact(t)
		signedPath := filepath.Join(t.TempDir(), "devicekit-ios-runner-signed"+filepath.Ext(appPath))
		if strings.EqualFold(filepath.Ext(appPath), ".app") {
			signedPath = filepath.Join(t.TempDir(), "devicekit-ios-runner-signed.app")
		}

		runIOSForDevice(t, udid,
			"sign", "app",
			"--path="+appPath,
			"--output="+signedPath,
			"--bundleid="+bundleID,
			"--p12file="+p12Path,
			"--profile="+profilePath,
			"--p12password=go-ios-e2e",
			"--install",
		)
	})
}

func TestDeviceKitUI(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		deviceKitMu.Lock()
		defer deviceKitMu.Unlock()
		deviceKitURL, stop := runDeviceKit(t, udid)
		defer stop()
		urlArg := "--devicekit-url=" + deviceKitURL

		smoke(t, udid, "ui", "status", urlArg)
		smoke(t, udid, "ui", "api", "--rpc-method=device.info", "--params={}", urlArg)
		smoke(t, udid, "ui", "raw", "--rpc-method=device.info", "--params={}", urlArg)
		smoke(t, udid, "ui", "size", urlArg)
		smoke(t, udid, "ui", "source", urlArg)
		smoke(t, udid, "ui", "orientation", "get", urlArg)
		smoke(t, udid, "ui", "orientation", "set", "PORTRAIT", urlArg)
		smoke(t, udid, "ui", "app", "foreground", urlArg)

		smoke(t, udid, "ui", "app", "launch", "com.apple.Preferences", urlArg)
		smoke(t, udid, "ui", "tap", "--x=10", "--y=10", urlArg)
		smoke(t, udid, "ui", "swipe", "--from-x=40", "--from-y=400", "--to-x=40", "--to-y=200", "--duration=0.1", urlArg)
		smoke(t, udid, "ui", "longpress", "--x=10", "--y=10", "--duration=0.1", urlArg)
		smoke(t, udid, "ui", "type", "--text=go-ios", urlArg)
		smoke(t, udid, "ui", "button", "home", urlArg)
		smoke(t, udid, "ui", "app", "terminate", "com.apple.Preferences", urlArg)
		// Screen-capture (screenshot / stream mjpeg / stream h264) is disabled for
		// now — it needs an active display, which the CI devices often lack.
	})
}

func TestWDAUICommands(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		wdaMu.Lock()
		defer wdaMu.Unlock()
		wdaURL, stop := runWDA(t, udid)
		defer stop()
		urlArg := "--wda-url=" + wdaURL
		driverArg := "--driver=wda"

		smoke(t, udid, "ui", "status", driverArg, urlArg)
		smoke(t, udid, "ui", "api", driverArg, "--method=GET", "--http-path=/status", urlArg)
		smoke(t, udid, "ui", "raw", driverArg, "--method=GET", "--http-path=/status", urlArg)
		smoke(t, udid, "ui", "size", driverArg, urlArg)
		smoke(t, udid, "ui", "source", driverArg, urlArg)
		smoke(t, udid, "ui", "orientation", "get", driverArg, urlArg)
		// Screen-capture (screenshot / stream mjpeg) and `tap` are skipped for WDA:
		// capture needs the display on (often off on CI), and go-ios's WDA tap hits
		// an endpoint this WDA build rejects. The read-only commands above prove the
		// WDA round-trip; DeviceKit's suite covers capture + interaction.
	})
}

func prepareDeviceKitArtifact(t *testing.T) string {
	t.Helper()
	if path := e2eEnv("GO_IOS_E2E_DEVICEKIT_PATH"); path != "" {
		return prepareAppPath(t, path)
	}

	artifactURL := e2eEnv("GO_IOS_E2E_DEVICEKIT_ARTIFACT_URL")
	if artifactURL == "" {
		artifactURL = defaultE2EDeviceKitArtifactURL
	}
	artifactPath := filepath.Join(t.TempDir(), filepath.Base(artifactURL))
	downloadFile(t, artifactURL, artifactPath)
	return prepareAppPath(t, artifactPath)
}

func appStoreConnectCredentials(t *testing.T) (string, string, string) {
	t.Helper()
	ascKeyID := e2eEnv("GO_IOS_E2E_ASC_KEY_ID", "GO_IOS_ASC_KEY_ID")
	ascIssuerID := e2eEnv("GO_IOS_E2E_ASC_ISSUER_ID", "GO_IOS_ASC_ISSUER_ID")
	ascPrivateKey := e2eEnv("GO_IOS_E2E_ASC_PRIVATE_KEY", "GO_IOS_ASC_PRIVATE_KEY")
	if ascKeyID == "" || ascIssuerID == "" || ascPrivateKey == "" {
		t.Skip("set GO_IOS_E2E_ASC_KEY_ID, GO_IOS_E2E_ASC_ISSUER_ID, and GO_IOS_E2E_ASC_PRIVATE_KEY to run signing e2e")
	}
	return ascKeyID, ascIssuerID, ascPrivateKey
}

// provisionSigningAssets returns a (p12, profile) pair for signing/installing
// the given bundle. It reuses the shared signing identity refreshed by the
// refresh-signing-identity workflow and published as secrets
// (GO_IOS_E2E_SIGNING_P12_B64 + GO_IOS_E2E_SIGNING_CERT_ID): the P12 comes
// straight from the secret and only a per-bundle provisioning profile is minted
// against that existing certificate. No certificate is created here, so the
// tests never hit Apple's "one development certificate" limit and run in
// parallel.
func provisionSigningAssets(t *testing.T, udid string, bundleID string, label string) (string, string) {
	t.Helper()
	ascKeyID, ascIssuerID, ascPrivateKey := appStoreConnectCredentials(t)

	p12B64 := e2eEnv("GO_IOS_E2E_SIGNING_P12_B64")
	certID := e2eEnv("GO_IOS_E2E_SIGNING_CERT_ID")
	if p12B64 == "" || certID == "" {
		t.Skip("set GO_IOS_E2E_SIGNING_P12_B64 and GO_IOS_E2E_SIGNING_CERT_ID (refreshed by the refresh-signing-identity workflow) to run signing e2e")
	}
	p12, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p12B64))
	if err != nil {
		t.Fatalf("decode GO_IOS_E2E_SIGNING_P12_B64: %v", err)
	}

	tempDir := t.TempDir()
	slug := strings.ToLower(strings.ReplaceAll(label, " ", "-"))
	p12Path := filepath.Join(tempDir, slug+".p12")
	if err := os.WriteFile(p12Path, p12, 0600); err != nil {
		t.Fatalf("write p12: %v", err)
	}
	profilePath := filepath.Join(tempDir, slug+".mobileprovision")

	runIOSForDevice(t, udid,
		"sign", "provision", "appstoreconnect",
		"--bundleid="+bundleID,
		"--bundle-name=go-ios "+label,
		"--profile-name=go-ios "+label+" "+time.Now().UTC().Format("20060102150405"),
		"--device-name=go-ios-e2e-"+udid,
		"--profile-output="+profilePath,
		"--certificate-id="+certID,
		"--asc-key-id="+ascKeyID,
		"--asc-issuer-id="+ascIssuerID,
		"--asc-private-key="+ascPrivateKey,
	)
	return p12Path, profilePath
}

func prepareAppPath(t *testing.T, path string) string {
	t.Helper()
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		dir := filepath.Join(t.TempDir(), "wda")
		unzip(t, path, dir)
		appPath := findFirstApp(t, dir)
		if appPath == "" {
			t.Fatalf("no .app found in %s", path)
		}
		return appPath
	}
	return path
}

func downloadFile(t *testing.T, url string, target string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("download %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("download %s: status %s", url, resp.Status)
	}
	out, err := os.Create(target)
	if err != nil {
		t.Fatalf("create %s: %v", target, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		t.Fatalf("write %s: %v", target, err)
	}
}

func unzip(t *testing.T, zipPath string, targetDir string) {
	t.Helper()
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip %s: %v", zipPath, err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath := filepath.Join(targetDir, file.Name)
		if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			t.Fatalf("zip entry escapes target dir: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				t.Fatalf("mkdir %s: %v", targetPath, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(targetPath), err)
		}
		src, err := file.Open()
		if err != nil {
			t.Fatalf("open zip entry %s: %v", file.Name, err)
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = src.Close()
			t.Fatalf("create %s: %v", targetPath, err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			_ = src.Close()
			_ = dst.Close()
			t.Fatalf("extract %s: %v", file.Name, err)
		}
		_ = src.Close()
		_ = dst.Close()
	}
}

func findFirstApp(t *testing.T, root string) string {
	t.Helper()
	var appPath string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.EqualFold(filepath.Ext(path), ".app") {
			appPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return appPath
}

func e2eEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
