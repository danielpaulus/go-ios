package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"os"
	"os/signal"
	"time"

	"github.com/danielpaulus/go-ios/usbmux"

	"github.com/danielpaulus/go-ios/usbmux/accessibility"
	"github.com/danielpaulus/go-ios/usbmux/debugproxy"
	"github.com/danielpaulus/go-ios/usbmux/diagnostics"
	"github.com/danielpaulus/go-ios/usbmux/forward"
	"github.com/danielpaulus/go-ios/usbmux/installationproxy"
	"github.com/danielpaulus/go-ios/usbmux/instruments"
	"github.com/danielpaulus/go-ios/usbmux/screenshotr"
	syslog "github.com/danielpaulus/go-ios/usbmux/syslog"
	"github.com/danielpaulus/go-ios/usbmux/testmanagerd"
	"github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"
)

//JSONdisabled enables or disables output in JSON format
var JSONdisabled = false

func main() {
	Main()
}

const version = "v 0.01"

// Main Exports main for testing
func Main() {
	usage := `iOS client v 0.01

Usage:
  ios listen [options]
  ios list [options] [--details]
  ios info [options]
  ios syslog [options]
  ios screenshot [options] [--output=<outfile>]
  ios devicename [options] 
  ios date [options]
  ios diagnostics list [options]
  ios pair [options]
  ios forward [options] <hostPort> <targetPort>
  ios dproxy
  ios readpair
  ios apps [--system]
  ios launch <bundleID>
  ios runtest <bundleID>  
  ios runwda [options]
  ios ax [options]
  ios -h | --help
  ios --version | version [options]

Options:
  -v --verbose   Enable Debug Logging.
  -t --trace     Enable Trace Logging (dump every message).
  --nojson       Disable JSON output (default).
  -h --help      Show this screen.
  --udid=<udid>  UDID of the device.

The commands work as following:
	The default output of all commands is JSON. Should you prefer human readable outout, specify the --nojson option with your command. 
	By default, the first device found will be used for a command unless you specify a --udid=some_udid switch.
	Specify -v for debug logging and -t for dumping every message.

   ios listen [options]                               Keeps a persistent connection open and notifies about newly connected or disconnected devices.
   ios list [options] [--details]                     Prints a list of all connected device's udids. If --details is specified, it includes version, name and model of each device.
   ios info [options]                                 Prints a dump of Lockdown getValues.
   ios syslog [options]                               Prints a device's log output
   ios screenshot [options] [--output=<outfile>]      Takes a screenshot and writes it to the current dir or to <outfile>
   ios devicename [options]                           Prints the devicename
   ios date [options]                                 Prints the device date
   ios diagnostics list [options]                     List diagnostic infos
   ios pair [options]                                 Pairs the device and potentially triggers the pairing dialog
   ios forward [options] <hostPort> <targetPort>      Similar to iproxy, forward a TCP connection to the device.
   ios dproxy                                         Starts the reverse engineering proxy server. Use "sudo launchctl unload -w /Library/Apple/System/Library/LaunchDaemons/com.apple.usbmuxd.plist" to stop usbmuxd and load to start it again should the proxy mess up things.
   ios readpair                                       Dump detailed information about the pairrecord for a device.
   ios -h | --help                                    Prints this screen.
   ios --version | version [options]                  Prints the version

  `
	arguments, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatal(err)
	}
	disableJSON, _ := arguments.Bool("--nojson")
	if disableJSON {
		JSONdisabled = true
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	traceLevelEnabled, _ := arguments.Bool("--trace")
	if traceLevelEnabled {
		log.Info("Set Trace mode")
		log.SetLevel(log.TraceLevel)
	} else {

		verboseLoggingEnabledLong, _ := arguments.Bool("--verbose")

		if verboseLoggingEnabledLong {
			log.Info("Set Debug mode")
			log.SetLevel(log.DebugLevel)
		}
	}
	//log.SetReportCaller(true)
	log.Debug(arguments)

	shouldPrintVersionNoDashes, _ := arguments.Bool("version")
	shouldPrintVersion, _ := arguments.Bool("--version")
	if shouldPrintVersionNoDashes || shouldPrintVersion {
		printVersion()
		return
	}

	b, _ := arguments.Bool("listen")
	if b {
		startListening()
		return
	}

	b, _ = arguments.Bool("dproxy")
	if b {
		startDebugProxy()
		return
	}

	b, _ = arguments.Bool("list")
	diagnostics, _ := arguments.Bool("diagnostics")
	if b && !diagnostics {
		b, _ = arguments.Bool("--details")
		printDeviceList(b)
		return
	}

	udid, _ := arguments.String("--udid")
	device, err := getDeviceOrQuit(udid)
	if err != nil {
		log.Fatal(err)
	}

	b, _ = arguments.Bool("info")
	if b {
		printDeviceInfo(device)
		return
	}

	b, _ = arguments.Bool("syslog")
	if b {
		runSyslog(device)
		return
	}

	b, _ = arguments.Bool("screenshot")
	if b {
		path, _ := arguments.String("--output")
		saveScreenshot(device, path)
		return
	}
	b, _ = arguments.Bool("devicename")
	if b {
		printDeviceName(device)
		return
	}

	b, _ = arguments.Bool("apps")

	if b {
		system, _ := arguments.Bool("--system")
		printInstalledApps(device, system)
		return
	}

	b, _ = arguments.Bool("date")
	if b {
		printDeviceDate(device)
		return
	}
	b, _ = arguments.Bool("diagnostics")
	if b {
		printDiagnostics(device)
		return
	}

	b, _ = arguments.Bool("pair")
	if b {
		pairDevice(device)
		return
	}

	b, _ = arguments.Bool("readpair")
	if b {
		readPair(device)
		return
	}

	b, _ = arguments.Bool("forward")
	if b {
		hostPort, _ := arguments.Int("<hostPort>")
		targetPort, _ := arguments.Int("<targetPort>")
		startForwarding(device, hostPort, targetPort)
		return
	}

	b, _ = arguments.Bool("launch")
	if b {
		bundleID, _ := arguments.String("<bundleID>")
		pid, err := instruments.LaunchApp(bundleID, device)
		if err != nil {
			log.WithFields(log.Fields{"pid": pid}).Info("Process launched")
		}
		log.Error(err)
	}

	b, _ = arguments.Bool("runtest")
	if b {
		bundleID, _ := arguments.String("<bundleID>")
		err := testmanagerd.RunXCUITest(bundleID, device)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Info("Failed running Xcuitest")
		}
		return
	}

	b, _ = arguments.Bool("runwda")
	if b {
		go func() {
			err := testmanagerd.RunWDA(device)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Fatal("Failed running WDA")
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Info("Closing..")
		testmanagerd.CloseXCUITestRunner()
		log.Info("Done Closing")
		return
	}

	b, _ = arguments.Bool("ax")
	if b {
		go func() {
			device := usbmux.ListDevices().DeviceList[0]

			conn, err := accessibility.New(device)
			if err != nil {
				log.Fatal(err)
			}

			conn.SwitchToDevice()

			conn.EnableSelectionMode()

			for i := 0; i < 3; i++ {
				conn.GetElement()
				time.Sleep(time.Second)
			}
			/*	conn.GetElement()
				time.Sleep(time.Second)
				conn.TurnOff()*/
			//conn.GetElement()
			//conn.GetElement()

			if err != nil {
				log.Fatal(err)
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		return
	}

}

func printVersion() {
	versionMap := map[string]interface{}{
		"version": version,
	}
	if JSONdisabled {
		fmt.Println(version)
	} else {
		fmt.Println(convertToJSONString(versionMap))
	}
}

func startDebugProxy() {
	proxy := debugproxy.NewDebugProxy()
	go func() {
		err := proxy.Launch()
		log.WithFields(log.Fields{"error": err}).Infof("DebugProxy Terminated abnormally")
		os.Exit(0)
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("Shutting down debugproxy")
	proxy.Close()
}

func startForwarding(device usbmux.DeviceEntry, hostPort int, targetPort int) {
	forward.Forward(device, uint16(hostPort), uint16(targetPort))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func printDiagnostics(device usbmux.DeviceEntry) {
	log.Debug("print diagnostics")
	diagnosticsService, err := diagnostics.New(device.DeviceID, device.Properties.SerialNumber, usbmux.ReadPairRecord(device.Properties.SerialNumber))
	if err != nil {
		log.Fatalf("Starting diagnostics service failed with: %s", err)
	}
	values, err := diagnosticsService.AllValues()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(convertToJSONString(values))
}

func printDeviceDate(device usbmux.DeviceEntry) {
	allValues := getValues(device)

	formatedDate := time.Unix(int64(allValues.Value.TimeIntervalSince1970), 0).Format(time.RFC850)
	if JSONdisabled {
		fmt.Println(formatedDate)
	} else {
		fmt.Println(convertToJSONString(map[string]interface{}{"formatedDate": formatedDate, "TimeIntervalSince1970": allValues.Value.TimeIntervalSince1970}))
	}

}
func printInstalledApps(device usbmux.DeviceEntry, system bool) {
	svc, _ := installationproxy.New(device)
	if !system {
		response, err := svc.BrowseUserApps()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(response)
		return
	}
	/*response, err := svc.BrowseSystemApps()
	if err != nil {
		log.Fatal(err)
	}
	log.Info(response)*/
}

func printDeviceName(device usbmux.DeviceEntry) {
	allValues := getValues(device)
	if JSONdisabled {
		println(allValues.Value.DeviceName)
	} else {
		println(convertToJSONString(map[string]string{
			"devicename": allValues.Value.DeviceName,
		}))
	}
}

func saveScreenshot(device usbmux.DeviceEntry, outputPath string) {
	log.Debug("take screenshot")
	screenshotrService, err := screenshotr.New(device.DeviceID, device.Properties.SerialNumber, usbmux.ReadPairRecord(device.Properties.SerialNumber))
	if err != nil {
		log.Fatalf("Starting Screenshotr failed with: %s", err)
	}
	imageBytes, err := screenshotrService.TakeScreenshot()
	if err != nil {
		log.Fatal(err)
	}
	if outputPath == "" {
		time := time.Now().Format("20060102150405")
		path, _ := filepath.Abs("./screenshot" + time + ".png")
		outputPath = path
	}
	err = ioutil.WriteFile(outputPath, imageBytes, 0777)
	if err != nil {
		log.Fatal(err)
	}
	if JSONdisabled {
		println(outputPath)
	} else {
		log.WithFields(log.Fields{"outputPath": outputPath}).Info("File saved successfully")
	}
}

func printDeviceList(details bool) {
	deviceList := usbmux.ListDevices()

	if details {
		if JSONdisabled {
			outputDetailedListNoJSON(deviceList)
		} else {
			outputDetailedList(deviceList)
		}
	} else {
		if JSONdisabled {
			fmt.Print(deviceList.String())
		} else {
			fmt.Println(convertToJSONString(deviceList.CreateMapForJSONConverter()))
		}
	}
}

type detailsEntry struct {
	Udid           string
	ProductName    string
	ProductType    string
	ProductVersion string
}

func outputDetailedList(deviceList usbmux.DeviceList) {
	result := make([]detailsEntry, len(deviceList.DeviceList))
	for i, device := range deviceList.DeviceList {
		udid := device.Properties.SerialNumber
		allValues := getValues(device)
		result[i] = detailsEntry{udid, allValues.Value.ProductName, allValues.Value.ProductType, allValues.Value.ProductVersion}
	}
	fmt.Println(convertToJSONString(map[string][]detailsEntry{
		"deviceList": result,
	}))
}

func outputDetailedListNoJSON(deviceList usbmux.DeviceList) {
	for _, device := range deviceList.DeviceList {
		udid := device.Properties.SerialNumber
		allValues := getValues(device)
		fmt.Printf("%s  %s  %s %s\n", udid, allValues.Value.ProductName, allValues.Value.ProductType, allValues.Value.ProductVersion)
	}
}

func getDeviceOrQuit(udid string) (usbmux.DeviceEntry, error) {
	log.Debugf("Looking for device '%s'", udid)
	deviceList := usbmux.ListDevices()
	if udid == "" {
		if len(deviceList.DeviceList) == 0 {
			return usbmux.DeviceEntry{}, errors.New("no iOS devices are attached to this host")
		}
		return deviceList.DeviceList[0], nil
	}
	for _, device := range deviceList.DeviceList {
		if device.Properties.SerialNumber == udid {
			return device, nil
		}
	}
	return usbmux.DeviceEntry{}, fmt.Errorf("Device '%s' not found. Is it attached to the machine?", udid)
}

func startListening() {
	muxConnection := usbmux.NewUsbMuxConnection()
	defer muxConnection.Close()
	attachedReceiver, err := muxConnection.Listen()
	if err != nil {
		log.Fatal("Failed issuing Listen command", err)
	}
	for {
		msg, err := attachedReceiver()
		if err != nil {
			log.Error("Stopped listening because of error")
			return
		}
		println(convertToJSONString((msg)))
	}

}

func printDeviceInfo(device usbmux.DeviceEntry) {
	allValues := getValues(device)
	fmt.Println(convertToJSONString(allValues.Value))
}

func runSyslog(device usbmux.DeviceEntry) {
	log.Debug("Run Syslog.")

	syslogConnection, err := syslog.New(device.DeviceID, device.Properties.SerialNumber)
	if err != nil {
		log.Fatalf("Syslog connection failed, %s", err)
	}
	defer syslogConnection.Close()

	go func() {
		messageContainer := map[string]string{}
		for {
			logMessage := syslogConnection.ReadLogMessage()
			if JSONdisabled {
				print(logMessage)
			} else {
				messageContainer["msg"] = logMessage
				println(convertToJSONString(messageContainer))
			}
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func getValues(device usbmux.DeviceEntry) usbmux.GetAllValuesResponse {
	muxConnection := usbmux.NewUsbMuxConnection()
	defer muxConnection.Close()

	pairRecord := muxConnection.ReadPair(device.Properties.SerialNumber)

	lockdownConnection, err := muxConnection.ConnectLockdown(device.DeviceID)
	if err != nil {
		log.Fatal(err)
	}
	lockdownConnection.StartSession(pairRecord)

	allValues, err := lockdownConnection.GetValues()
	if err != nil {
		log.Fatal(err)
	}
	lockdownConnection.StopSession()
	return allValues
}

func pairDevice(device usbmux.DeviceEntry) {
	println("not yet copied from branch go-ios-old")
	// err := usbmux.Pair(device)
	// if err != nil {
	// 	println(err)
	// } else {
	// 	fmt.Printf("Paired %s", device.Properties.SerialNumber)
	// }

}

func readPair(device usbmux.DeviceEntry) {
	record := usbmux.ReadPairRecord(device.Properties.SerialNumber)
	log.Info(record.String())
}

func convertToJSONString(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}
