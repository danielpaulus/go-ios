package main

import (
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

func main() {
	checkLinux()
	checkUsbMux()
	checkRoot()
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
