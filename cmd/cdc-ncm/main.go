package main

import (
	ncm "go-ios-cdcncm"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/google/gousb"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

func main() {
	ctx := gousb.NewContext()
	devices, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		//slog.Info("found device", slog.Int64("product", int64(desc.Product)), slog.Int64("vendor", int64(desc.Vendor)))
		return desc.Vendor == 0x05ac && desc.Product == 0x12a8
	})
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("device list", "length", len(devices))
	for _, d := range devices {
		handleDevice(d)
	}
}

func handleDevice(device *gousb.Device) {
	defer device.Close()

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
	//for i := 0; i < 100; i++ {
	//	s, err := device.GetStringDescriptor(13)
	//	if err != nil {
	//		return
	//	}
	//	slog.Info("string descriptor", slog.String("value", s), slog.Int64("idx", int64(i)))
	//
	//}

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

	/*	outStream, err := out.NewStream(outDesc.MaxPacketSize*3, 1)
		if err != nil {
			return
		}

		defer outStream.Close()
	*/
	slog.Info("created streams")

	createConfig(out, inStream)
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

func createConfig(w io.Writer, r io.Reader) {
	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = "iphone"

	ifce, err := water.New(config)
	if err != nil {
		log.Fatal(err)
	}
	out, err := exec.Command("/bin/sh", "-c", "sudo ip link set dev iphone up").CombinedOutput()
	if err != nil {
		slog.Error("failed setting up ethernet device", "err", err)
	}
	slog.Info("ethernet device is up:", "cmd", string(out))

	wr := ncm.NewWrapper(loggingReader{r}, w)

	time.Sleep(10 * time.Second)
	slog.Info("start copying")

	go func() {
		var frame ethernet.Frame
		for {
			frame.Resize(1500)
			n, err := ifce.Read([]byte(frame))
			if err != nil {
				log.Fatal(err)
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	//_, err = io.CopyBuffer(io.Discard, wr, outBuf)

	//slog.Error("failed to copy from usb to iface", slog.Any("error", err))
	//
	//for {
	//	n, err := ifce.Read(inBuf)
	//	//frame.Resize(1500)
	//	//n, err := ifce.Read([]byte(frame))
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	frame := inBuf[:n]
	//
	//	log.Printf("%s\n", hex.EncodeToString(frame))
	//	//frame = frame[:n]
	//	//log.Printf("Dst: %s\n", frame.Destination())
	//	//log.Printf("Src: %s\n", frame.Source())
	//	//log.Printf("Ethertype: % x\n", frame.Ethertype())
	//	//log.Printf("Payload: % x\n", frame.Payload())
	//
	//}
}

type loggingReader struct {
	r io.Reader
}

func (l loggingReader) Read(p []byte) (n int, err error) {
	n, err = l.r.Read(p)
	if err != nil {
		return n, err
	}
	//slog.Info("read", slog.String("bytes", hex.EncodeToString(p[:n])))
	return
}
