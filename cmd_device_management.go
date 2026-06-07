package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/amfi"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"
)

func runActivateCommand(ctx commandContext) {
	exitIfError("failed activation", mobileactivation.Activate(ctx.Device))
}

func runLangCommand(ctx commandContext) {
	locale, _ := ctx.Args.String("--setlocale")
	newlang, _ := ctx.Args.String("--setlang")
	slog.Debug("lang", "setlocale", locale, "setlang", newlang)
	language(ctx.Device, locale, newlang)
}

func runEraseCommand(ctx commandContext) {
	force, _ := ctx.Args.Bool("--force")
	if !force {
		slog.Warn("are you sure you want to erase device? (y/n)", "udid", ctx.Device.Properties.SerialNumber)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		exitIfError("An error occured while reading input", err)
		if !strings.HasPrefix(input, "y") {
			slog.Error("abort")
			return
		}
	}

	exitIfError("failed erasing", mcinstall.Erase(ctx.Device))
	fmt.Print(convertToJSONString("ok"))
}

func runWifiCommand(ctx commandContext) {
	ssid, _ := ctx.Args.String("--ssid")
	psw, _ := ctx.Args.String("--password")
	encType, _ := ctx.Args.String("--enc-type")
	remove, _ := ctx.Args.Bool("--remove")

	if encType == "" {
		encType = "WPA"
	}

	if remove {
		exitIfError("failed removing wifi", mcinstall.RemoveWifi(ctx.Device, ssid))
	} else {
		exitIfError("failed preparing wifi", mcinstall.PrepareWifi(ctx.Device, ssid, psw, encType))
	}
	fmt.Print(convertToJSONString("ok"))
}

func runPrepareCommand(ctx commandContext) {
	if createCert, _ := ctx.Args.Bool("create-cert"); createCert {
		cert, err := ios.CreateDERFormattedSupervisionCert()
		exitIfError("failed creating cert", err)
		err = os.WriteFile("supervision-cert.der", cert.CertDER, 0o777)
		slog.Info("supervision-cert.der")
		exitIfError("failed writing cert", err)
		err = os.WriteFile("supervision-cert.pem", cert.CertPEM, 0o777)
		slog.Info("supervision-cert.pem")
		exitIfError("failed writing cert", err)
		err = os.WriteFile("supervision-private-key.key", cert.PrivateKeyDER, 0o777)
		slog.Info("supervision-private-key.key")
		exitIfError("failed writing cert", err)
		err = os.WriteFile("supervision-private-key.pem", cert.PrivateKeyPEM, 0o777)
		slog.Info("supervision-private-key.pem")
		exitIfError("failed writing key", err)
		err = os.WriteFile("supervision-csr.csr", []byte(cert.Csr), 0o777)
		slog.Info("supervision-csr.csr")
		exitIfError("failed writing cert", err)
		slog.Info("Golang does not have good PKCS12 format sadly. If you need a p12 file run this: " +
			"'openssl pkcs12 -export -inkey supervision-private-key.pem -in supervision-cert.pem -out certificate.p12 -password pass:a'")
		return
	}
	if printSkip, _ := ctx.Args.Bool("printskip"); printSkip {
		fmt.Println(convertToJSONString(mcinstall.GetAllSetupSkipOptions()))
		return
	}
	if cloudConfig, _ := ctx.Args.Bool("cloudconfig"); cloudConfig {
		conn, err := mcinstall.New(ctx.Device)
		exitIfError("failed connecting to mcinstall", err)
		defer conn.Close()
		config, err := conn.GetCloudConfiguration()
		exitIfError("failed getting cloud configuration", err)
		fmt.Println(convertToJSONString(config))
		return
	}
	skip := mcinstall.GetAllSetupSkipOptions()
	skipArg := ctx.Args["--skip"].([]string)
	if len(skipArg) > 0 {
		skip = skipArg
	}

	certfile, _ := ctx.Args.String("--certfile")
	orgname, _ := ctx.Args.String("--orgname")
	locale, _ := ctx.Args.String("--locale")
	lang, _ := ctx.Args.String("--lang")
	p12password, _ := ctx.Args.String("--p12password")
	if p12password == "" {
		p12password = os.Getenv("P12_PASSWORD")
	}
	var certBytes []byte
	if certfile != "" {
		rawCertBytes, err := os.ReadFile(certfile)
		exitIfError("failed opening cert file", err)
		if orgname == "" {
			logFatal("--orgname must be specified if certfile for supervision is provided")
		}
		certBytes, err = extractDERCertificate(rawCertBytes, p12password)
		exitIfError("failed to parse supervision certificate", err)
	}
	exitIfError("failed erasing", mcinstall.Prepare(ctx.Device, skip, certBytes, orgname, locale, lang))
	fmt.Print(convertToJSONString("ok"))
}

func runSetWallpaperCommand(ctx commandContext) {
	imagePath, _ := ctx.Args.String("<imagePath>")
	p12file, _ := ctx.Args.String("--p12file")
	p12password, _ := ctx.Args.String("--password")
	if p12password == "" {
		p12password = os.Getenv("P12_PASSWORD")
	}
	screen, _ := ctx.Args.String("--screen")
	if screen == "" {
		screen = "home"
	}
	handleSetWallpaper(ctx.Device, imagePath, screen, p12file, p12password)
}

func runGetWallpaperCommand(ctx commandContext) {
	out, _ := ctx.Args.String("--output")
	if out == "" {
		out = "wallpaper.png"
	}
	handleGetWallpaper(ctx.Device, out)
}

func runGetIconLayoutCommand(ctx commandContext) {
	out, _ := ctx.Args.String("--output")
	handleGetIconLayout(ctx.Device, out)
}

func runSetIconLayoutCommand(ctx commandContext) {
	layoutFile, _ := ctx.Args.String("<layoutFile>")
	handleSetIconLayout(ctx.Device, layoutFile)
}

func runHTTPProxyCommand(ctx commandContext) {
	if removeCommand, _ := ctx.Args.Bool("remove"); removeCommand {
		err := mcinstall.RemoveProxy(ctx.Device)
		exitIfError("failed removing proxy", err)
		slog.Info("success")
		return
	}
	host, _ := ctx.Args.String("<host>")
	port, _ := ctx.Args.String("<port>")
	user, _ := ctx.Args.String("<user>")
	pass, _ := ctx.Args.String("<pass>")
	if pass == "" {
		pass = os.Getenv("PROXY_PASSWORD")
	}
	p12file, _ := ctx.Args.String("--p12file")
	p12password, _ := ctx.Args.String("--password")
	if p12password == "" {
		p12password = os.Getenv("P12_PASSWORD")
	}
	p12bytes, err := os.ReadFile(p12file)
	exitIfError("could not read p12-file", err)

	err = mcinstall.SetHttpProxy(ctx.Device, host, port, user, pass, p12bytes, p12password)
	exitIfError("failed", err)
	slog.Info("success")
}

func runProfileCommand(ctx commandContext) {
	if listCommand, _ := ctx.Args.Bool("list"); listCommand {
		handleProfileList(ctx.Device)
	}
	if add, _ := ctx.Args.Bool("add"); add {
		name, _ := ctx.Args.String("<profileFile>")
		p12file, _ := ctx.Args.String("--p12file")
		p12password, _ := ctx.Args.String("--password")
		if p12password == "" {
			p12password = os.Getenv("P12_PASSWORD")
		}
		if p12file != "" {
			handleProfileAddSupervised(ctx.Device, name, p12file, p12password)
			return
		}
		handleProfileAdd(ctx.Device, name)
	}
	if remove, _ := ctx.Args.Bool("remove"); remove {
		name, _ := ctx.Args.String("<profileName>")
		handleProfileRemove(ctx.Device, name)
	}
}

func runDeviceNameCommand(ctx commandContext) {
	printDeviceName(ctx.Device)
}

func runPairCommand(ctx commandContext) {
	org, _ := ctx.Args.String("--p12file")
	pwd, _ := ctx.Args.String("--password")
	if pwd == "" {
		pwd = os.Getenv("P12_PASSWORD")
	}
	pairDevice(ctx.Device, org, pwd)
}

func runReadPairCommand(ctx commandContext) {
	readPair(ctx.Device)
}

func runRebootCommand(ctx commandContext) {
	err := diagnostics.Reboot(ctx.Device)
	if err != nil {
		slog.Error("reboot failed", "error", err)
	} else {
		slog.Info("ok")
	}
}

func runShutdownCommand(ctx commandContext) {
	err := diagnostics.Shutdown(ctx.Device)
	if err != nil {
		slog.Error("shutdown failed", "error", err)
	} else {
		slog.Info("ok")
	}
}

func runDevModeCommand(ctx commandContext) {
	enable, _ := ctx.Args.Bool("enable")
	get, _ := ctx.Args.Bool("get")
	enablePostRestart, _ := ctx.Args.Bool("--enable-post-restart")
	if enable {
		err := amfi.EnableDeveloperMode(ctx.Device, enablePostRestart)
		exitIfError("Failed enabling developer mode", err)
	}

	if get {
		devModeEnabled, _ := imagemounter.IsDevModeEnabled(ctx.Device)
		if JSONdisabled {
			fmt.Printf("Developer mode enabled: %v\n", devModeEnabled)
		} else {
			result := map[string]interface{}{"DeveloperModeEnabled": devModeEnabled}
			fmt.Println(convertToJSONString(result))
		}
	}

	if reveal, _ := ctx.Args.Bool("reveal"); reveal {
		conn, err := amfi.New(ctx.Device)
		exitIfError("Failed connecting to AMFI service", err)
		defer conn.Close()
		err = conn.RevealDevMode()
		exitIfError("Failed revealing developer mode menu", err)
		slog.Info("Developer Mode menu has been revealed on the device. Go to Settings → Privacy & Security → Developer Mode to enable it.")
	}
}
