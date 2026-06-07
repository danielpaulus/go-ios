package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/fileservice"
	"github.com/danielpaulus/go-ios/ios/house_arrest"
)

func runForwardCommand(ctx commandContext) {
	mappings, _ := ctx.Args["--port"].([]string)
	if len(mappings) > 0 {
		startMultiForwarding(ctx.Device, mappings)
		return
	}
	hostPort, _ := ctx.Args.Int("<hostPort>")
	targetPort, _ := ctx.Args.Int("<targetPort>")
	startForwarding(ctx.Device, uint16(hostPort), uint16(targetPort))
}

func runFileCommand(ctx commandContext) {
	if !ctx.Device.SupportsRsd() {
		exitIfError("file command requires iOS 17+ with tunnel", fmt.Errorf("tunnel not running. Start with: ios tunnel start"))
	}

	bundleID, _ := ctx.Args.String("--app")
	groupID, _ := ctx.Args.String("--app-group")
	useCrash, _ := ctx.Args.Bool("--crash")
	useTemp, _ := ctx.Args.Bool("--temp")

	flagCount := 0
	if bundleID != "" {
		flagCount++
	}
	if groupID != "" {
		flagCount++
	}
	if useCrash {
		flagCount++
	}
	if useTemp {
		flagCount++
	}

	if flagCount > 1 {
		exitIfError("file command", fmt.Errorf("can only specify one of: --app, --app-group, --crash, or --temp"))
	}
	if flagCount == 0 {
		exitIfError("file command", fmt.Errorf("must specify one of: --app=<bundleID>, --app-group=<groupID>, --crash, or --temp"))
	}

	var domain fileservice.Domain
	var identifier string

	if bundleID != "" {
		domain = fileservice.DomainAppDataContainer
		identifier = bundleID
	} else if groupID != "" {
		domain = fileservice.DomainAppGroupDataContainer
		identifier = groupID
	} else if useCrash {
		domain = fileservice.DomainSystemCrashLogs
	} else if useTemp {
		domain = fileservice.DomainTemporary
	}

	conn, err := fileservice.New(ctx.Device, domain, identifier)
	exitIfError("file: failed to connect to file service", err)
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("Failed to close file service connection", "error", closeErr)
		}
	}()

	if ls, _ := ctx.Args.Bool("ls"); ls {
		path, _ := ctx.Args.String("--path")
		if path == "" {
			path = "."
		}

		files, err := conn.ListDirectory(path)
		exitIfError("file ls: failed to list directory", err)

		if !JSONdisabled {
			result := map[string]interface{}{
				"path":  path,
				"files": files,
				"count": len(files),
			}
			fmt.Println(convertToJSONString(result))
		} else {
			fmt.Printf("Files in %s:\n", path)
			for _, file := range files {
				fmt.Printf("  %s\n", file)
			}
			fmt.Printf("\nTotal: %d files\n", len(files))
		}
	}

	if pull, _ := ctx.Args.Bool("pull"); pull {
		remotePath, _ := ctx.Args.String("--remote")
		localPath, _ := ctx.Args.String("--local")

		if remotePath == "" {
			exitIfError("file pull", fmt.Errorf("--remote=<path> is required"))
		}
		if localPath == "" {
			exitIfError("file pull", fmt.Errorf("--local=<path> is required"))
		}

		outputFile, err := os.Create(localPath)
		exitIfError("file pull: failed to create output file", err)
		defer outputFile.Close()

		slog.Info(fmt.Sprintf("Downloading %s to %s...", remotePath, localPath))
		err = conn.PullFile(remotePath, outputFile)
		exitIfError("file pull: failed to download file", err)

		fileInfo, err := outputFile.Stat()
		exitIfError("file pull: failed to get file info", err)
		fileSize := fileInfo.Size()

		if !JSONdisabled {
			result := map[string]interface{}{
				"remote": remotePath,
				"local":  localPath,
				"size":   fileSize,
			}
			fmt.Println(convertToJSONString(result))
		} else {
			slog.Info(fmt.Sprintf("Downloaded %d bytes to %s", fileSize, localPath))
		}
	}

	if push, _ := ctx.Args.Bool("push"); push {
		localPath, _ := ctx.Args.String("--local")
		remotePath, _ := ctx.Args.String("--remote")

		if localPath == "" || remotePath == "" {
			exitIfError("push requires --local and --remote paths", fmt.Errorf("missing required arguments"))
		}

		fileInfo, err := os.Stat(localPath)
		exitIfError("push: failed to stat local file", err)

		permissions := int64(fileInfo.Mode().Perm())
		uid := int64(501)
		gid := int64(501)
		fileSize := fileInfo.Size()

		file, err := os.Open(localPath)
		exitIfError("push: failed to open local file", err)
		defer file.Close()

		slog.Info(fmt.Sprintf("Uploading %s to %s...", localPath, remotePath))
		err = conn.PushFile(remotePath, file, fileSize, permissions, uid, gid)
		exitIfError("push: failed to upload file", err)

		if !JSONdisabled {
			result := map[string]interface{}{
				"remote": remotePath,
				"local":  localPath,
				"size":   fileSize,
			}
			fmt.Println(convertToJSONString(result))
		} else {
			slog.Info(fmt.Sprintf("Uploaded %d bytes to %s", fileSize, remotePath))
		}
	}
}

func runFsyncCommand(ctx commandContext) {
	containerBundleId, _ := ctx.Args.String("--app")
	var afcService *afc.Client
	var err error
	if containerBundleId == "" {
		afcService, err = afc.New(ctx.Device)
	} else {
		afcService, err = house_arrest.New(ctx.Device, containerBundleId)
	}
	exitIfError("fsync: connect afc service failed", err)
	defer afcService.Close()

	if rm, _ := ctx.Args.Bool("rm"); rm {
		path, _ := ctx.Args.String("--path")
		isRecursive, _ := ctx.Args.Bool("--r")
		if isRecursive {
			err = afcService.RemoveAll(path)
		} else {
			err = afcService.Remove(path)
		}
		exitIfError("fsync: remove failed", err)
	}

	if tree, _ := ctx.Args.Bool("tree"); tree {
		path, _ := ctx.Args.String("--path")
		err := afcService.WalkDir(path, func(path string, info afc.FileInfo, err error) error {
			s := strings.Split(path, string(os.PathSeparator))
			_, f := filepath.Split(path)
			prefix := strings.Repeat("|  ", len(s)-1)

			suffix := ""
			if info.Type == afc.S_IFDIR {
				suffix = "/"
			}

			fmt.Printf("%s|-%s%s\n", prefix, f, suffix)
			return nil
		})
		exitIfError("fsync: tree view failed", err)
	}

	if mkdir, _ := ctx.Args.Bool("mkdir"); mkdir {
		path, _ := ctx.Args.String("--path")
		err = afcService.MkDir(path)
		exitIfError("fsync: mkdir failed", err)
	}

	if pull, _ := ctx.Args.Bool("pull"); pull {
		sp, _ := ctx.Args.String("--srcPath")
		dp, _ := ctx.Args.String("--dstPath")
		if dp != "" {
			ret, _ := ios.PathExists(dp)
			if !ret {
				err = os.MkdirAll(dp, os.ModePerm)
				exitIfError("mkdir failed", err)
			}
		}

		dp = path.Join(dp, filepath.Base(sp))
		err = afcService.Pull(sp, dp)
		exitIfError("fsync: pull failed", err)
	}
	if push, _ := ctx.Args.Bool("push"); push {
		sp, _ := ctx.Args.String("--srcPath")
		dp, _ := ctx.Args.String("--dstPath")

		err = afcService.Push(sp, dp)
		exitIfError("fsync: push failed", err)
	}
}
