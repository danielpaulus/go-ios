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

	"github.com/danielpaulus/go-ios/usbmux/screenshotr"
	syslog "github.com/danielpaulus/go-ios/usbmux/syslog"
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
  ios -h | --help
  ios --version | version [options]

Options:
  -v --verbose   Enable Debug Logging.
  -t --trace     Enable Trace Logging (dump every message).
  --nojson       Disable JSON output (default).
  -h --help      Show this screen.
  --udid=<udid>  UDID of the device.
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

	b, _ = arguments.Bool("list")
	if b {
		b, _ = arguments.Bool("--details")
		printDeviceList(b)
		return
	}

	b, _ = arguments.Bool("pair")
	if b {
		pairDevice(device)
		return
	}

	b, _ = arguments.Bool("forward")
	if b {
		hostPort, _ := arguments.Int("<hostPort>")
		targetPort, _ := arguments.Int("<targetPort>")
		startForwarding(device, hostPort, targetPort)
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

func startForwarding(device usbmux.DeviceEntry, hostPort int, targetPort int) {
	// forward.Forward(device, uint16(hostPort), uint16(targetPort))
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// <-c
}

func printDiagnostics(device usbmux.DeviceEntry) {
	// log.Debug("print diagnostics")
	// diagnosticsService := diagnostics.New(device.DeviceID, device.Properties.SerialNumber)
	// fmt.Println(formatOutput(diagnosticsService.AllValues()))
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
	imageBytes := screenshotrService.TakeScreenshot()
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

	syslogConnection, err := syslog.New(device.DeviceID, device.Properties.SerialNumber, usbmux.ReadPairRecord(device.Properties.SerialNumber))
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
				print(convertToJSONString(messageContainer))
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

	allValues := lockdownConnection.GetValues()
	lockdownConnection.StopSession()
	return allValues
}

func pairDevice(device usbmux.DeviceEntry) {
	// err := usbmux.Pair(device)
	// if err != nil {
	// 	println(err)
	// } else {
	// 	fmt.Printf("Paired %s", device.Properties.SerialNumber)
	// }

}

func convertToJSONString(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}
