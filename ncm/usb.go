package ncm

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/gousb"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

const VID_APPLE = 0x5ac
const PID_RANGE_LOW = 0x1290
const PID_RANGE_MAX = 0x12af
const PID_APPLE_T2_COPROCESSOR = 0x8600
const PID_APPLE_SILICON_RESTORE_LOW = 0x1901
const PID_APPLE_SILICON_RESTORE_MAX = 0x1905

var allocatedDevices = sync.Map{}
var serialToInterface = map[string]string{}
var deviceLock = sync.Mutex{}
var deviceCounter = 0

func Start(c chan os.Signal) error {
	ctx := gousb.NewContext()
	defer ctx.Close()
	for {
		select {
		case <-time.After(5 * time.Second):
			checkDevices(ctx)
			printStatus()
		case <-c:
			slog.Info("shut down complete")
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
	deviceCount.Set(float64(len(connectedDevices)))
	slices.Sort[[]string](connectedDevices)
	slog.Debug("connected devices", "devices", connectedDevices)
}

func checkDevices(ctx *gousb.Context) {
	devices, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if desc.Vendor != VID_APPLE {
			return false
		}
		if desc.Product == PID_APPLE_T2_COPROCESSOR {
			return false
		}
		if desc.Product < PID_RANGE_LOW || desc.Product > PID_RANGE_MAX {
			return false
		}
		return true
	})
	if err != nil {
		slog.Error("failed opening devices", "err", err)
	}
	slog.Debug("device list", "length", len(devices))
	for _, d := range devices {
		go func(d *gousb.Device) {
			err := handleDevice(d)
			if err != nil {
				slog.Error("failed opening network adapter for device", "device", d.String(), "err", err)
			}
		}(d)
	}
}

func updateInterface(serial string) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	_, ok := serialToInterface[serial]
	if ok {
		return
	}
	ifaceName := fmt.Sprintf("iphone%d", deviceCounter)
	slog.Info("assigning interface", "iface", ifaceName, "serial", serial)
	serialToInterface[serial] = ifaceName
	deviceCounter++
}

func interfaceName(serial string) string {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	return serialToInterface[serial]
}

// handleDevice checks if the device has 5 usb configurations.
// USBMUXD should make sure they are already enabled. Without doing anything devices usually have 4.
// The 5th one should be the ncm driver. Then finds endpoints for the ncm network device, starts a virtual
// TAP network device and starts a read/write loop to send data from virtual network to USB and back
func handleDevice(device *gousb.Device) error {
	defer closeWithLog("device "+device.String(), device.Close)
	serial, err := device.SerialNumber()
	if err != nil {
		slog.Info("failed to get serial" + device.String())
		return fmt.Errorf("handleDevice: failed to get serial for device %s with err %w", device.String(), err)
	}
	serial = strings.Trim(serial, "\x00")
	updateInterface(serial)
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

	var ifNumber = -1
	var altS = -1
	for _, iface := range cfg.Desc.Interfaces {
		for _, alt := range iface.AltSettings {
			if alt.Class == 10 && alt.SubClass == 0 && len(alt.Endpoints) == 2 {
				slog.Info("alt setting", slog.String("alt", alt.String()), slog.Int("class", int(alt.Class)), slog.Int("subclass", int(alt.SubClass)), slog.String("protocol", alt.Protocol.String()), slog.String("serial", serial))
				ifNumber = iface.Number
				altS = alt.Alternate
			}
		}
	}
	if ifNumber == -1 || altS == -1 {
		return fmt.Errorf("handleDevice: could not find interface or altsetting")
	}
	iface, err := cfg.Interface(ifNumber, altS)
	if err != nil {
		return fmt.Errorf("handleDevice: failed to claim interface for device %s. this can happen if some other process already claimed it. err %w", serial, err)
	}
	defer func() {
		slog.Info("closing interface", "serial", serial)
		iface.Close()
	}()
	var inEndpoint = -1
	var outEndpoint = -1
	var inDesc gousb.EndpointDesc
	for endpoint, i := range iface.Setting.Endpoints {
		slog.Info(endpoint.String())
		if i.Direction == gousb.EndpointDirectionIn {
			inEndpoint = i.Number
			inDesc = i
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
	defer closeWithLog("virtual TAP interface", ifce.Close)

	//blocks until the device disconnects or the adapter fails for some reason
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

func createConfig(serial string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = interfaceName(serial)
	slog.Info("creating TAP device", "device", config.Name, "serial", serial)
	ifce, err := water.New(config)
	if err != nil {
		return &water.Interface{}, fmt.Errorf("createConfig: failed creating ifce %w", err)
	}
	hasIp, output := InterfaceHasIP(config.Name)
	if !hasIp {
		ip := "FC00:0000:0000:0000:0000:0000:0000:00FB/64"
		slog.Info("add IP address to device", "device", config.Name, "serial", serial, "ip", ip)
		output, err := AddInterface(config.Name, ip)
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
	wr := NewWrapper(r, w, serial)
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
				networkReceiveBytes.WithLabelValues(ifce.Name(), serial).Add(float64(n))
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
					networkSendBytes.WithLabelValues(ifce.Name(), serial).Add(float64(len(frame)))
				}
			}
		}
	}()
	wg.Wait()

	return nil
}
