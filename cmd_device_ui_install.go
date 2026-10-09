package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios/signing"
)

const (
	defaultWDAArtifactURL       = "https://deviceboxhq.com/WebDriverAgentRunner-13.2.0.zip"
	defaultDeviceKitArtifactURL = "https://deviceboxhq.com/devicekit-ios-runner-0.0.18.ipa"
	defaultWDAArtifactSHA256    = "448ffe591cf64abf7ff0751897b6860cd45b8032b39677a5dd79f827c4d781cc"
	defaultDeviceKitSHA256      = "45457d3f11de2b5370b14ba45e6c8502328e28825ee8222e8517245898a4c57f"
	defaultWDABundleID          = "com.deviceboxhq.goios.WebDriverAgentRunner.xctrunner"
	defaultDeviceKitBundleID    = "com.deviceboxhq.goios.devicekit.runner"
)

func runUIInstallCommand(ctx commandContext) {
	switch {
	case boolArg(ctx.Args, "wda"):
		runUIInstallApp(ctx, uiInstallTarget{
			Name:           "wda",
			DefaultURL:     defaultWDAArtifactURL,
			ExpectedSHA256: defaultWDAArtifactSHA256,
			DefaultBundle:  defaultWDABundleID,
			DefaultName:    "go-ios WDA",
			OutputBaseName: "WebDriverAgentRunner",
		})
	case boolArg(ctx.Args, "devicekit"):
		runUIInstallApp(ctx, uiInstallTarget{
			Name:           "devicekit",
			DefaultURL:     defaultDeviceKitArtifactURL,
			ExpectedSHA256: defaultDeviceKitSHA256,
			DefaultName:    "go-ios DeviceKit",
			OutputBaseName: "devicekit-ios-runner",
		})
	default:
		logFatal("unknown ui install target; use 'ios ui install wda' or 'ios ui install devicekit'. Run 'ios ui download' to pre-download artifacts, or pass --path to use a local artifact.")
	}
}

type uiInstallTarget struct {
	Name           string
	DefaultURL     string
	ExpectedSHA256 string
	DefaultBundle  string
	DefaultName    string
	OutputBaseName string
}

func runUIInstallApp(ctx commandContext, target uiInstallTarget) {
	artifactPath, cleanup := uiInstallArtifactPath(ctx, target)
	defer cleanup()

	outputPath, _ := ctx.Args.String("--output")
	if outputPath == "" {
		outputPath = filepath.Join(os.TempDir(), target.OutputBaseName+"-signed-"+time.Now().UTC().Format("20060102150405")+filepath.Ext(artifactPath))
		if strings.EqualFold(filepath.Ext(artifactPath), ".app") {
			outputPath = filepath.Join(os.TempDir(), target.OutputBaseName+"-signed-"+time.Now().UTC().Format("20060102150405")+".app")
		}
	}

	bundleID, _ := ctx.Args.String("--bundleid")
	if bundleID == "" {
		bundleID = target.DefaultBundle
	}
	p12Password, _ := ctx.Args.String("--p12password")
	p12Path, _ := ctx.Args.String("--p12file")
	profilePath, _ := ctx.Args.String("--profile")

	result, err := signing.SignWithFiles(signing.SignWithFilesOptions{
		AppPath:     artifactPath,
		OutputPath:  outputPath,
		BundleID:    bundleID,
		P12Path:     p12Path,
		P12Password: p12Password,
		ProfilePath: profilePath,
	})
	exitIfError("failed signing "+target.Name, err)
	slog.Info("signed UI automation app", "target", target.Name, "appPath", result.OutputPath, "bundleID", result.BundleID, "udid", ctx.Device.Properties.SerialNumber)
	installApp(ctx.Device, result.OutputPath)
}

func uiInstallArtifactPath(ctx commandContext, target uiInstallTarget) (string, func()) {
	pathArg, _ := ctx.Args.String("--path")
	if pathArg != "" {
		slog.Warn("using explicitly trusted local UI automation artifact", "target", target.Name, "path", pathArg)
		return prepareUIInstallAppPath(pathArg)
	}

	tempDir, err := os.MkdirTemp("", "go-ios-ui-install-*")
	exitIfError("failed creating temp dir", err)
	artifactURL := target.DefaultURL
	artifactPath := filepath.Join(tempDir, filepath.Base(artifactURL))
	digest, err := downloadUIArtifact(artifactURL, artifactPath, target.ExpectedSHA256)
	exitIfError("failed downloading "+target.Name, err)
	slog.Info("verified UI automation artifact", "target", target.Name, "sha256", digest)
	appPath, cleanupApp := prepareUIInstallAppPath(artifactPath)
	return appPath, func() {
		cleanupApp()
		_ = os.RemoveAll(tempDir)
	}
}

func prepareUIInstallAppPath(path string) (string, func()) {
	if !strings.EqualFold(filepath.Ext(path), ".zip") {
		return path, func() {}
	}
	tempDir, err := os.MkdirTemp("", "go-ios-ui-install-zip-*")
	exitIfError("failed creating unzip dir", err)
	exitIfError("failed extracting "+path, unzipUIArtifact(path, tempDir))
	appPath, err := findUIInstallApp(tempDir)
	exitIfError("failed finding .app in "+path, err)
	return appPath, func() { _ = os.RemoveAll(tempDir) }
}

func downloadUIArtifact(rawURL string, targetPath string, expectedSHA256 string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("download returned %s", resp.Status)
	}
	return writeVerifiedUIArtifact(resp.Body, targetPath, expectedSHA256)
}

func writeVerifiedUIArtifact(source io.Reader, targetPath string, expectedSHA256 string) (string, error) {
	if len(expectedSHA256) != sha256.Size*2 {
		return "", fmt.Errorf("invalid expected SHA-256 digest")
	}
	if _, err := hex.DecodeString(expectedSHA256); err != nil {
		return "", fmt.Errorf("invalid expected SHA-256 digest: %w", err)
	}

	out, err := os.CreateTemp(filepath.Dir(targetPath), "."+filepath.Base(targetPath)+"-*")
	if err != nil {
		return "", err
	}
	temporaryPath := out.Name()
	verified := false
	defer func() {
		if !verified {
			_ = os.Remove(temporaryPath)
		}
	}()

	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(out, hash), source)
	if err := errors.Join(copyErr, out.Close()); err != nil {
		return "", err
	}

	digest := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(digest, expectedSHA256) {
		return "", fmt.Errorf("artifact SHA-256 mismatch: got %s, want %s", digest, expectedSHA256)
	}
	if err := os.Rename(temporaryPath, targetPath); err != nil {
		return "", fmt.Errorf("store verified artifact: %w", err)
	}
	verified = true
	return digest, nil
}

func unzipUIArtifact(zipPath string, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath := filepath.Join(targetDir, file.Name)
		if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("zip entry escapes target dir: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = src.Close()
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			_ = src.Close()
			_ = dst.Close()
			return err
		}
		_ = src.Close()
		_ = dst.Close()
	}
	return nil
}

func findUIInstallApp(root string) (string, error) {
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
		return "", err
	}
	if appPath == "" {
		return "", fmt.Errorf("no .app found")
	}
	return appPath, nil
}
