package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"usbmuxd/usbmux"

	docopt "github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"
)

func main() {
	usage := `iOS client v 0.01

Usage:
  ios list [--details]
  ios info [options]
  ios syslog [options]
  ios screenshot [options]
  ios devicename [options] 
  ios date [options]
  ios diagnostics list [options]
  ios pair [options]
  ios forward [options] <hostPort> <targetPort>
  ios -h | --help
  ios --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  -u=<udid>, --udid     UDID of the device.
  -o=<filepath>, --output
  `
	arguments, _ := docopt.ParseDoc(usage)

	fmt.Println(arguments)
	udid, _ := arguments.String("--udid")
	device, err := getDeviceOrQuit(udid)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := arguments.Bool("info")
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
	// allValues := getValues(device)

	// fmt.Println(time.Unix(int64(allValues.Value.TimeIntervalSince1970), 0).Format(time.RFC850))

}

func printDeviceName(device usbmux.DeviceEntry) {
	// allValues := getValues(device)
	// println(allValues.Value.DeviceName)
}

func saveScreenshot(device usbmux.DeviceEntry, outputPath string) {
	// log.Debug("take screenshot")
	// screenshotrService := screenshotr.New(device.DeviceID, device.Properties.SerialNumber)
	// imageBytes := screenshotrService.TakeScreenshot()
	// if outputPath == "" {
	// 	time := time.Now().Format("20060102150405")
	// 	path, _ := filepath.Abs("./screenshot" + time + ".png")
	// 	outputPath = path
	// }
	// err := ioutil.WriteFile(outputPath, imageBytes, 0777)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Info(outputPath)
}

func printDeviceList(details bool) {
	deviceList := usbmux.ListDevices()
	println(deviceList.String())
	// if details {
	// 	for _, device := range deviceList.DeviceList {
	// 		udid := device.Properties.SerialNumber
	// 		allValues := getValues(device)
	// 		fmt.Printf("%s  %s  %s %s", udid, allValues.Value.ProductName, allValues.Value.ProductType, allValues.Value.ProductVersion)
	// 	}
	// } else {
	// 	deviceList.Print()
	// }
}

func getDeviceOrQuit(udid string) (usbmux.DeviceEntry, error) {
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

func printDeviceInfo(device usbmux.DeviceEntry) {

	// allValues := getValues(device)

	// fmt.Println(formatOutput(allValues.Value))
}

func runSyslog(device usbmux.DeviceEntry) {
	// log.Debug("Run Syslog.")
	// syslogConnection := syslog.New(device.DeviceID, device.Properties.SerialNumber)
	// defer syslogConnection.Close()

	// go func() {
	// 	for {
	// 		print(<-syslogConnection.LogReader)
	// 	}
	// }()
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// <-c
}

// func getValues(device usbmux.DeviceEntry) usbmux.GetAllValuesResponse {
// 	muxConnection := usbmux.NewUsbMuxConnection()
// 	defer muxConnection.Close()

// 	pairRecord := muxConnection.ReadPair(device.Properties.SerialNumber)

// 	lockdownConnection, err := muxConnection.ConnectLockdown(device.DeviceID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	lockdownConnection.StartSession(pairRecord)

// 	allValues := lockdownConnection.GetValues()
// 	lockdownConnection.StopSession()
// 	return allValues
// }

func pairDevice(device usbmux.DeviceEntry) {
	// err := usbmux.Pair(device)
	// if err != nil {
	// 	println(err)
	// } else {
	// 	fmt.Printf("Paired %s", device.Properties.SerialNumber)
	// }

}

func formatOutput(data interface{}) string {
	b, err2 := json.Marshal(data)
	if err2 != nil {
		fmt.Println(err2)
		return ""
	}
	return string(b)
}
