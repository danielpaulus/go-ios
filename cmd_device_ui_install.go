package main

import (
	"archive/zip"
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
	defaultWDABundleID          = "com.deviceboxhq.goios.WebDriverAgentRunner.xctrunner"
	defaultDeviceKitBundleID    = "com.deviceboxhq.goios.devicekit.runner"
)

func runUIInstallCommand(ctx commandContext) {
	switch {
	case boolArg(ctx.Args, "wda"):
		runUIInstallApp(ctx, uiInstallTarget{
			Name:           "wda",
			DefaultURL:     defaultWDAArtifactURL,
			DefaultBundle:  defaultWDABundleID,
			DefaultName:    "go-ios WDA",
			OutputBaseName: "WebDriverAgentRunner",
		})
	case boolArg(ctx.Args, "devicekit"):
		runUIInstallApp(ctx, uiInstallTarget{
			Name:           "devicekit",
			DefaultURL:     defaultDeviceKitArtifactURL,
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
		return prepareUIInstallAppPath(pathArg)
	}

	tempDir, err := os.MkdirTemp("", "go-ios-ui-install-*")
	exitIfError("failed creating temp dir", err)
	artifactURL := target.DefaultURL
	artifactPath := filepath.Join(tempDir, filepath.Base(artifactURL))
	exitIfError("failed downloading "+target.Name, downloadUIArtifact(artifactURL, artifactPath))
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

func downloadUIArtifact(rawURL string, targetPath string) error {
	resp, err := http.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("download returned %s", resp.Status)
	}
	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
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
