package main

import (
	"flag"
	ncm "go-ios-cdcncm"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
)

func checkRoot() {
	u := os.Geteuid()
	if u != 0 {
		slog.Error("go-ncm needs root. run with sudo.")
		os.Exit(1)
	}
}

func checkLinux() {
	if runtime.GOOS != "linux" {
		slog.Error("go-ncm only works on linux")
		os.Exit(1)
	}
}

// accepts one cmd line argument: --prometheusport=8080
// if specified, prometheus metrics will be available at http://0.0.0.0:prometheusport/metrics
// if not specified, the prometheus endpoint will not be started and not be available.
func main() {
	checkLinux()
	checkUsbMux()
	checkRoot()
	// Define a string flag with a default value and a short description.
	// This will read the command-line argument for --prometheusport.
	prometheusPort := flag.Int("prometheusport", -1, "The port for Prometheus metrics")
	// Parse the flags from the command-line arguments.
	flag.Parse()
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
