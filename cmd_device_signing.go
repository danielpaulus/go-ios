package main

import (
	"context"
	"log/slog"

	"github.com/danielpaulus/go-ios/ios/signing"
)

func runSignCommand(ctx commandContext) {
	switch {
	case boolArg(ctx.Args, "provision") && boolArg(ctx.Args, "appstoreconnect"):
		runSignProvisionAppStoreConnectCommand(ctx)
	case boolArg(ctx.Args, "app"):
		runSignAppCommand(ctx)
	default:
		logFatal("unknown sign command")
	}
}

func runSignProvisionAppStoreConnectCommand(ctx commandContext) {
	bundleID, _ := ctx.Args.String("--bundleid")
	bundleName, _ := ctx.Args.String("--bundle-name")
	profileName, _ := ctx.Args.String("--profile-name")
	deviceName, _ := ctx.Args.String("--device-name")
	p12Password, _ := ctx.Args.String("--p12password")
	p12Output, _ := ctx.Args.String("--p12-output")
	profileOutput, _ := ctx.Args.String("--profile-output")

	keyID, _ := ctx.Args.String("--asc-key-id")
	issuerID, _ := ctx.Args.String("--asc-issuer-id")
	privateKeyPath, _ := ctx.Args.String("--asc-private-key")
	creds, err := signing.LoadAppStoreConnectCredentials(keyID, issuerID, privateKeyPath)
	exitIfError("failed loading App Store Connect credentials", err)

	result, err := signing.PrepareSigningAssets(context.Background(), signing.PrepareAssetsOptions{
		BundleID:    bundleID,
		BundleName:  bundleName,
		ProfileName: profileName,
		DeviceName:  deviceName,
		P12Password: p12Password,
		P12Output:   p12Output,
		ProfileOut:  profileOutput,
		Credentials: creds,
		Device:      ctx.Device,
	})
	exitIfError("failed creating signing assets", err)
	slog.Info("created signing assets", "bundleID", result.BundleID, "p12", result.P12Path, "profile", result.ProfilePath, "udid", ctx.Device.Properties.SerialNumber)
}

func runSignAppCommand(ctx commandContext) {
	appPath, _ := ctx.Args.String("--path")
	outputPath, _ := ctx.Args.String("--output")
	bundleID, _ := ctx.Args.String("--bundleid")
	p12Path, _ := ctx.Args.String("--p12file")
	profilePath, _ := ctx.Args.String("--profile")
	p12Password, _ := ctx.Args.String("--p12password")
	install, _ := ctx.Args.Bool("--install")

	result, err := signing.SignWithFiles(signing.SignWithFilesOptions{
		AppPath:     appPath,
		OutputPath:  outputPath,
		BundleID:    bundleID,
		P12Path:     p12Path,
		P12Password: p12Password,
		ProfilePath: profilePath,
	})
	exitIfError("failed signing app", err)
	slog.Info("signed app", "appPath", result.OutputPath, "bundleID", result.BundleID, "udid", ctx.Device.Properties.SerialNumber)
	if install {
		installApp(ctx.Device, result.OutputPath)
	}
}
