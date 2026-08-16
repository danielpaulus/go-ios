// Command go-ncm is a standalone CDC-NCM USB ethernet driver for iOS devices.
// It sets up virtual TAP network devices for connected iOS devices so tools
// like go-ios or pymobiledevice3 can talk to them over the network.
// It is Linux-only and needs root (USB access + TAP device setup).
package main

import (
	"flag"
	"fmt"
	ncm "go-ios-cdcncm"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
)

// version is stamped by the release workflow via
// go build -ldflags "-X main.version=x.y.z"
var version = "local-build"

func checkRoot() {
	u := os.Geteuid()
	if u != 0 {
		slog.Error("go-ncm needs root. run with sudo.")
		os.Exit(1)
	}
}

func checkLinux() {
	if runtime.GOOS != "linux" {
		slog.Error("go-ncm only works on linux", slog.String("os", runtime.GOOS))
		os.Exit(1)
	}
}

// accepts these cmd line arguments:
//
//	--prometheusport=8080  if specified, prometheus metrics will be available at
//	                       http://0.0.0.0:prometheusport/metrics. If not, the
//	                       prometheus endpoint will not be started.
//	--version              print the version and exit.
func main() {
	// Define a string flag with a default value and a short description.
	// This will read the command-line argument for --prometheusport.
	prometheusPort := flag.Int("prometheusport", -1, "The port for Prometheus metrics")
	printVersion := flag.Bool("version", false, "Print the version and exit")
	// Parse the flags from the command-line arguments.
	flag.Parse()
	if *printVersion {
		fmt.Println(version)
		return
	}
	checkLinux()
	checkUsbMux()
	checkRoot()
	slog.Info("starting go-ncm", slog.String("version", version))
	if *prometheusPort != -1 {
		go ncm.StartPrometheus(*prometheusPort)
	} else {
		slog.Info("prometheus metrics not configured. start with '--prometheusport=8080' to expose prometheus metrics.")
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	err := ncm.Start(c)
	if err != nil {
		slog.Error("error looking for devices", slog.Any("error", err))
		os.Exit(1)
	}
}

func checkUsbMux() {
	v, err := ncm.CheckUSBMUXVersion()
	if err != nil {
		slog.Error("error getting usbmuxd version", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("usbmuxd version", slog.Any("version", v))
}
