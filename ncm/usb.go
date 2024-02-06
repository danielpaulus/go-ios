package ncm

import (
	"fmt"
	"github.com/google/gousb"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"io"
	"log/slog"
	"os"
	"time"
)

func Start(c chan os.Signal) error {
	ctx := gousb.NewContext()
	defer ctx.Close()
	for {
		select {
		case <-time.After(5 * time.Second):
			checkDevices(ctx)
		case <-c:
			return nil
		}
	}
}

func checkDevices(ctx *gousb.Context) {
	devices, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		//slog.Info("found device", slog.Int64("product", int64(desc.Product)), slog.Int64("vendor", int64(desc.Vendor)))
		return desc.Vendor == 0x05ac && desc.Product == 0x12a8
	})
	if err != nil {
		slog.Error("failed opening devices", "err", err)
	}
	slog.Info("device list", "length", len(devices))
	for _, d := range devices {
		err := handleDevice(d)
		if err != nil {
			slog.Error("failed opening network adapter for device", "device", d.String(), "err", err)
		}
	}
}

func handleDevice(device *gousb.Device) error {
	serial, err := device.SerialNumber()
	if err != nil {
		slog.Info("failed to get serial")
		return
	}
	slog.Info("got device", slog.String("serial", serial))

	activeConfig, err := device.ActiveConfigNum()
	if err != nil {
		return
	}
	slog.Info("active config", slog.Int("active", activeConfig))
	confLen := len(device.Desc.Configs)
	slog.Info("available configs", "configs", device.Desc.Configs, "len", confLen)

	if confLen != 5 {
		_, err = device.Control(0xc0, 69, 0, 0, make([]byte, 4))
		if err != nil {
			slog.Error("failed sending control1", slog.Any("error", err))
			return
		}

		_, err = device.Control(0xc0, 82, 0, 3, make([]byte, 1))
		if err != nil {
			slog.Error("failed sending control2", slog.Any("error", err))
			return
		}
	}

	cfg, err := device.Config(5)
	if err != nil {
		slog.Error("failed activating config", "err", slog.AnyValue(err))
		return
	}
	slog.Info("got config", slog.String("config", cfg.String()))

	for _, iface := range cfg.Desc.Interfaces {
		for _, alt := range iface.AltSettings {
			if alt.Class == 10 && alt.SubClass == 0 && len(alt.Endpoints) == 2 {
				slog.Info("alt setting", slog.String("alt", alt.String()), slog.Int("class", int(alt.Class)), slog.Int("subclass", int(alt.SubClass)), slog.String("protocol", alt.Protocol.String()))
			}
		}
	}
	iface, err := cfg.Interface(5, 1)
	if err != nil {
		slog.Error("failed to open interface", slog.AnyValue(err))
		return
	}
	var inEndpoint int
	var outEndpoint int
	for endpoint, i := range iface.Setting.Endpoints {
		slog.Info(endpoint.String())
		if i.Direction == gousb.EndpointDirectionIn {
			inEndpoint = i.Number
		}
		if i.Direction == gousb.EndpointDirectionOut {
			outEndpoint = i.Number
		}
		slog.Info(i.String())
	}

	in, err := iface.InEndpoint(inEndpoint)
	if err != nil {
		slog.Error("failed to get in-endpoint", slog.AnyValue(err))
	}

	out, err := iface.OutEndpoint(outEndpoint)
	if err != nil {
		slog.Error("failed to get out-endpoint", slog.AnyValue(err))
	}
	slog.Info("claimed interfaces")

	inDesc, _ := getEndpointDescriptions(cfg.Desc.Interfaces[5].AltSettings[1])

	inStream, err := in.NewStream(inDesc.MaxPacketSize*3, 1)
	if err != nil {
		return
	}
	defer inStream.Close()

	slog.Info("created streams")

	ifce, err := createConfig(serial)
	rw(out, inStream, ifce)

}

func getEndpointDescriptions(s gousb.InterfaceSetting) (in gousb.EndpointDesc, out gousb.EndpointDesc) {
	for _, e := range s.Endpoints {
		if e.Direction == gousb.EndpointDirectionIn {
			in = e
		}
		if e.Direction == gousb.EndpointDirectionOut {
			out = e
		}
	}
	return
}

func createConfig(serial string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = "iphone_" + serial
	slog.Info("creating TAP device", "device", config.Name)
	ifce, err := water.New(config)
	if err != nil {
		return &water.Interface{}, fmt.Errorf("createConfig: failed creating ifce %w", err)
	}
	output, err := SetInterfaceUp(config.Name)
	if err != nil {
		return &water.Interface{}, fmt.Errorf("createConfig: err calling interface up %w", err)
	}
	slog.Info("ethernet device is up:", "device", config.Name, "cmd", output)

	time.Sleep(10 * time.Second)

	return ifce, err
}

func rw(w io.Writer, r io.Reader, ifce *water.Interface) error {
	wr := NewWrapper(r, w)
	go func() {
		var frame ethernet.Frame
		for {
			frame.Resize(1500)
			n, err := ifce.Read([]byte(frame))
			if err != nil {
				slog.Fatal(err)
			}
			frame = frame[:n]
			_, err = wr.Write(frame)
			if err != nil {
				slog.Error("failed to copy from iface to usb", slog.Any("error", err))
				continue
			}
			slog.Info("write to USB ok")
		}

	}()

	go func() {
		for {
			frames, err := wr.ReadDatagrams()
			if err != nil {
				slog.Error("failed reading datagrams with err", err)
			}
			for _, frame := range frames {
				_, err := ifce.Write(frame)
				if err != nil {
					slog.Error("failed sending frame to virtual device", "err", err)
				}
			}
		}
	}()

}
