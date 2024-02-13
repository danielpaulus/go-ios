package ncm

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/gousb"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

var allocatedDevices = sync.Map{}

func Start(c chan os.Signal) error {
	ctx := gousb.NewContext()
	defer ctx.Close()
	for {
		select {
		case <-time.After(5 * time.Second):
			checkDevices(ctx)
			printStatus()
		case <-c:
			return nil
		}
	}
}

func printStatus() {
	var connectedDevices []string
	allocatedDevices.Range(func(key, value interface{}) bool {
		connectedDevices = append(connectedDevices, key.(string))
		return true
	})
	slog.Info("connected devices", "devices", connectedDevices)
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
		go func() {
			err := handleDevice(d)
			if err != nil {
				slog.Error("failed opening network adapter for device", "device", d.String(), "err", err)
			}
		}()
	}
}

func handleDevice(device *gousb.Device) error {
	defer closeWithLog("device "+device.String(), device.Close)
	serial, err := device.SerialNumber()
	if err != nil {
		slog.Info("failed to get serial")
		return fmt.Errorf("handleDevice: failed to get serial for device %s with err %w", device.String(), err)
	}
	serial = strings.Trim(serial, "\x00")
	_, loaded := allocatedDevices.LoadOrStore(serial, true)
	if loaded {
		slog.Info("device already handled", "serial", serial)
		return nil
	}
	defer allocatedDevices.Delete(serial)
	slog.Info("got device", slog.String("serial", serial))

	activeConfig, err := device.ActiveConfigNum()
	if err != nil {
		return fmt.Errorf("handleDevice: failed to get active config for device %s with err %w", serial, err)
	}
	slog.Info("active config", slog.Int("active", activeConfig), "serial", serial)
	confLen := len(device.Desc.Configs)
	slog.Info("available configs", "configs", device.Desc.Configs, "len", confLen, "serial", serial)

	if confLen != 5 {
		_, err = device.Control(0xc0, 69, 0, 0, make([]byte, 4))
		if err != nil {
			slog.Error("failed sending control1", slog.Any("error", err))
			return fmt.Errorf("handleDevice: failed sending control1 for device %s with err %w", serial, err)
		}

		_, err = device.Control(0xc0, 82, 0, 3, make([]byte, 1))
		if err != nil {
			slog.Error("failed sending control2", slog.Any("error", err))
			return fmt.Errorf("handleDevice: failed sending control2 for device %s with err %w", serial, err)
		}
	}

	cfg, err := device.Config(5)
	if err != nil {
		return fmt.Errorf("handleDevice: failed activating config for device %s with err %w", serial, err)
	}
	defer closeWithLog("config "+serial, cfg.Close)
	slog.Info("got config", slog.String("config", cfg.String()), "serial", serial)

	for _, iface := range cfg.Desc.Interfaces {
		for _, alt := range iface.AltSettings {
			if alt.Class == 10 && alt.SubClass == 0 && len(alt.Endpoints) == 2 {
				slog.Info("alt setting", slog.String("alt", alt.String()), slog.Int("class", int(alt.Class)), slog.Int("subclass", int(alt.SubClass)), slog.String("protocol", alt.Protocol.String()), slog.String("serial", serial))
			}
		}
	}
	iface, err := cfg.Interface(5, 1)
	if err != nil {
		return fmt.Errorf("handleDevice: failed to claim interface for device %s. this can happen if some other process already claimed it. err %w", serial, err)
	}
	defer func() {
		slog.Info("closing interface", "serial", serial)
		iface.Close()
	}()
	var inEndpoint = -1
	var outEndpoint = -1
	for endpoint, i := range iface.Setting.Endpoints {
		slog.Info(endpoint.String())
		if i.Direction == gousb.EndpointDirectionIn {
			inEndpoint = i.Number
		}
		if i.Direction == gousb.EndpointDirectionOut {
			outEndpoint = i.Number
		}
	}
	if inEndpoint == -1 {
		return fmt.Errorf("handleDevice: failed to find in-endpoint for device %s", serial)
	}
	if outEndpoint == -1 {
		return fmt.Errorf("handleDevice: failed to find out-endpoint for device %s", serial)
	}

	in, err := iface.InEndpoint(inEndpoint)
	if err != nil {
		return fmt.Errorf("handleDevice: failed to open in-endpoint for device %s with err %w", serial, err)
	}

	out, err := iface.OutEndpoint(outEndpoint)
	if err != nil {
		return fmt.Errorf("handleDevice: failed to open out-endpoint for device %s with err %w", serial, err)
	}
	slog.Info("claimed interfaces", "serial", serial)

	inDesc, _ := getEndpointDescriptions(cfg.Desc.Interfaces[5].AltSettings[1])

	inStream, err := in.NewStream(inDesc.MaxPacketSize*3, 1)
	if err != nil {
		return fmt.Errorf("handleDevice: failed to open in-stream for device %s with err %w", serial, err)
	}
	defer closeWithLog("in stream", inStream.Close)

	slog.Info("created streams", "serial", serial)

	ifce, err := createConfig(serial)
	if err != nil {
		return fmt.Errorf("handleDevice: failed to create config for device %s with err %w", serial, err)
	}
	defer closeWithLog("virtual TUN interface", ifce.Close)
	err = ncmIOCopy(out, inStream, ifce, serial)
	slog.Info("stopping interface for device", "serial", serial)
	if err != nil {
		slog.Error("failed to copy data", "err", err, "serial", serial)
	}
	return err
}

func closeWithLog(msg string, closeable func() error) {
	slog.Info("closing", "what", msg)
	err := closeable()
	if err != nil {
		slog.Error("failed closing", "err", err)
	}
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
	config.Name = "iphone_1"
	slog.Info("creating TAP device", "device", config.Name, "serial", serial)
	ifce, err := water.New(config)
	if err != nil {
		return &water.Interface{}, fmt.Errorf("createConfig: failed creating ifce %w", err)
	}
	hasIp, output := InterfaceHasIP(config.Name)
	if !hasIp {
		slog.Info("add IP address to device", "device", config.Name, "serial", serial)
		output, err := AddInterface(config.Name)
		if err != nil {
			return &water.Interface{}, fmt.Errorf("createConfig: err adding ip to interface. cmd output: '%s' err: %w", output, err)
		}
	} else {
		slog.Info("ip found", "interface", config.Name, "ip", output)
	}

	output, err = SetInterfaceUp(config.Name)
	if err != nil {
		return &water.Interface{}, fmt.Errorf("createConfig: err calling interface up. cmd output: '%s' err: %w", output, err)
	}
	slog.Info("ethernet device is up:", "device", config.Name, "serial", serial)

	return ifce, err
}

// ncmIOCopy copies data between a USB device and a virtual network interface. It is blocking until the connection fails
// for some reason. This happens if the device is disconnected or if the virtual network interface is removed.
// Closing interfaces is the responsibility of the caller.
func ncmIOCopy(w io.Writer, r io.Reader, ifce *water.Interface, serial string) error {
	wr := NewWrapper(r, w)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		var frame ethernet.Frame
		for {
			select {
			case <-ctx.Done():
				return
			default:
				frame.Resize(1500)
				n, err := ifce.Read([]byte(frame))
				if err != nil {
					slog.Error("ncmIOCopy: failed to read from iface", slog.Any("error", err), slog.String("serial", serial))
					cancel()
					continue
				}
				frame = frame[:n]
				_, err = wr.Write(frame)
				if err != nil {
					slog.Error("ncmIOCopy: failed to copy from iface to usb", slog.Any("error", err), slog.String("serial", serial))
					cancel()
				}
			}
		}

	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				frames, err := wr.ReadDatagrams()
				if err != nil {
					slog.Error("ncmIOCopy: failed to read from usb", slog.Any("error", err), slog.String("serial", serial))
					cancel()
					continue
				}
				for _, frame := range frames {
					_, err := ifce.Write(frame)
					if err != nil {
						slog.Error("failed sending frame to virtual device", "err", err)
						cancel()
					}
				}
			}
		}
	}()
	wg.Wait()

	return nil
}
