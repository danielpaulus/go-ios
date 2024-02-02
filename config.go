package main

import (
	"github.com/google/gousb"
	"log"
	"log/slog"
)

func main() {
	ctx := gousb.NewContext()
	devices, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		slog.Info("found device", slog.Int64("product", int64(desc.Product)), slog.Int64("vendor", int64(desc.Vendor)))
		return desc.Vendor == 0x05ac && desc.Product == 0x12a8
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range devices {
		defer d.Close()

		serial, err := d.SerialNumber()
		if err != nil {
			slog.Info("failed to get serial")
			continue
		}
		slog.Info("got device", slog.String("serial", serial))

		_, err = d.Control(0x0, 0x9, 0x5, 0, nil)
		if err != nil {
			slog.Warn("could not send control message", slog.Any("err", err))
		}
	}
}
