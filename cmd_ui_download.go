package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type uiDownloadResult struct {
	Target       string `json:"target"`
	URL          string `json:"url"`
	ArtifactPath string `json:"artifactPath"`
	ArtifactType string `json:"artifactType"`
	SizeBytes    int64  `json:"sizeBytes"`
	AppPath      string `json:"appPath,omitempty"`
}

func runUIDownloadCommand(ctx commandContext) {
	targets := uiDownloadTargets(ctx)
	outputDir, _ := ctx.Args.String("--output")
	if outputDir == "" {
		outputDir = "."
	}
	absOutputDir, err := filepath.Abs(outputDir)
	exitIfError("failed resolving output dir", err)
	exitIfError("failed creating output dir", os.MkdirAll(absOutputDir, 0755))

	results := make([]uiDownloadResult, 0, len(targets))
	for _, target := range targets {
		results = append(results, downloadUIInstallTarget(target, absOutputDir))
	}
	fmt.Println(convertToJSONString(map[string]interface{}{
		"outputDir":  absOutputDir,
		"downloaded": results,
	}))
}

func uiDownloadTargets(ctx commandContext) []uiInstallTarget {
	targets := []uiInstallTarget{
		{
			Name:           "wda",
			DefaultURL:     defaultWDAArtifactURL,
			DefaultBundle:  defaultWDABundleID,
			DefaultName:    "go-ios WDA",
			OutputBaseName: "WebDriverAgentRunner",
		},
		{
			Name:           "devicekit",
			DefaultURL:     defaultDeviceKitArtifactURL,
			DefaultName:    "go-ios DeviceKit",
			OutputBaseName: "devicekit-ios-runner",
		},
	}
	switch {
	case boolArg(ctx.Args, "wda"):
		return targets[:1]
	case boolArg(ctx.Args, "devicekit"):
		return targets[1:]
	default:
		return targets
	}
}

func downloadUIInstallTarget(target uiInstallTarget, outputDir string) uiDownloadResult {
	artifactPath := filepath.Join(outputDir, filepath.Base(target.DefaultURL))
	exitIfError("failed downloading "+target.Name, downloadUIArtifact(target.DefaultURL, artifactPath))
	info, err := os.Stat(artifactPath)
	exitIfError("failed stat "+artifactPath, err)

	result := uiDownloadResult{
		Target:       target.Name,
		URL:          target.DefaultURL,
		ArtifactPath: artifactPath,
		ArtifactType: strings.TrimPrefix(strings.ToLower(filepath.Ext(artifactPath)), "."),
		SizeBytes:    info.Size(),
	}
	if strings.EqualFold(filepath.Ext(artifactPath), ".zip") {
		extractDir := filepath.Join(outputDir, target.OutputBaseName+"-"+time.Now().UTC().Format("20060102150405"))
		exitIfError("failed creating extract dir", os.MkdirAll(extractDir, 0755))
		exitIfError("failed extracting "+artifactPath, unzipUIArtifact(artifactPath, extractDir))
		appPath, err := findUIInstallApp(extractDir)
		exitIfError("failed finding .app in "+artifactPath, err)
		result.AppPath = appPath
	}
	return result
}
