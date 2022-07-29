package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/danielpaulus/go-ios/ios/afc"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"syscall"

	"github.com/danielpaulus/go-ios/ios/crashreport"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"

	"github.com/danielpaulus/go-ios/ios/debugserver"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/zipconduit"

	"os"
	"os/signal"
	"time"

	"github.com/danielpaulus/go-ios/ios/simlocation"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
	"github.com/danielpaulus/go-ios/ios/debugproxy"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/forward"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/danielpaulus/go-ios/ios/notificationproxy"
	"github.com/danielpaulus/go-ios/ios/pcap"
	"github.com/danielpaulus/go-ios/ios/screenshotr"
	syslog "github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"
)

//JSONdisabled enables or disables output in JSON format
var JSONdisabled = false
var prettyJSON = false

func main() {
	Main()
}

const version = "local-build"

// Main Exports main for testing
func Main() {
	usage := fmt.Sprintf(`go-ios %s

Usage:
  ios listen [options]
  ios list [options] [--details]
  ios info [options]
  ios image list [options]
  ios image mount [--path=<imagepath>] [options]
  ios image auto [--basedir=<where_dev_images_are_stored>] [options]
  ios syslog [options]
  ios screenshot [options] [--output=<outfile>]
  ios crash ls [<pattern>] [options]
  ios crash cp <srcpattern> <target> [options]
  ios crash rm <cwd> <pattern> [options]
  ios devicename [options] 
  ios date [options]
  ios devicestate list [options]
  ios devicestate enable <profileTypeId> <profileId> [options]
  ios lang [--setlocale=<locale>] [--setlang=<newlang>] [options]
  ios mobilegestalt <key>... [--plist] [options]
  ios diagnostics list [options]
  ios profile list [options]
  ios profile remove <profileName> [options]
  ios profile add <profileFile> [--p12file=<orgid>] [--password=<p12password>] [options]
  ios httpproxy <host> <port> [<user>] [<pass>] --p12file=<orgid> --password=<p12password> [options]
  ios httpproxy remove [options]
  ios pair [--p12file=<orgid>] [--password=<p12password>] [options]
  ios ps [--apps] [options]
  ios ip [options]
  ios forward [options] <hostPort> <targetPort>
  ios dproxy [--binary]
  ios readpair [options]
  ios pcap [options] [--pid=<processID>] [--process=<processName>]
  ios install --path=<ipaOrAppFolder> [options]
  ios uninstall <bundleID> [options]
  ios apps [--system] [--all] [options]
  ios launch <bundleID> [options]
  ios kill (<bundleID> | --pid=<processID> | --process=<processName>) [options]
  ios runtest <bundleID> [options]
  ios runwda [--bundleid=<bundleid>] [--testrunnerbundleid=<testbundleid>] [--xctestconfig=<xctestconfig>] [--arg=<a>]... [--env=<e>]... [options]
  ios ax [options]
  ios debug [options] [--stop-at-entry] <app_path>
  ios fsync (rm | tree | mkdir) --path=<targetPath>
  ios fsync (pull | push) --srcPath=<srcPath> --dstPath=<dstPath> 
  ios reboot [options]
  ios -h | --help
  ios --version | version [options]
  ios setlocation [options] [--lat=<lat>] [--lon=<lon>]
  ios setlocationgpx [options] [--gpxfilepath=<gpxfilepath>]
  ios resetlocation [options]

Options:
  -v --verbose   Enable Debug Logging.
  -t --trace     Enable Trace Logging (dump every message).
  --nojson       Disable JSON output
  --pretty       Pretty-print JSON command output
  -h --help      Show this screen.
  --udid=<udid>  UDID of the device.

The commands work as following:
	The default output of all commands is JSON. Should you prefer human readable outout, specify the --nojson option with your command. 
	By default, the first device found will be used for a command unless you specify a --udid=some_udid switch.
	Specify -v for debug logging and -t for dumping every message.

   ios listen [options]                                               Keeps a persistent connection open and notifies about newly connected or disconnected devices.
   ios list [options] [--details]                                     Prints a list of all connected device's udids. If --details is specified, it includes version, name and model of each device.
   ios info [options]                                                 Prints a dump of Lockdown getValues.
   ios image list [options]                                           List currently mounted developers images' signatures
   ios image mount [--path=<imagepath>] [options]                     Mount a image from <imagepath>
   ios image auto [--basedir=<where_dev_images_are_stored>] [options] Automatically download correct dev image from the internets and mount it.
   >                                                                  You can specify a dir where images should be cached.
   >                                                                  The default is the current dir. 
   ios syslog [options]                                               Prints a device's log output
   ios screenshot [options] [--output=<outfile>]                      Takes a screenshot and writes it to the current dir or to <outfile>
   ios crash ls [<pattern>] [options]                                 run "ios crash ls" to get all crashreports in a list, 
   >                                                                  or use a pattern like 'ios crash ls "*ips*"' to filter
   ios crash cp <srcpattern> <target> [options]                       copy "file pattern" to the target dir. Ex.: 'ios crash cp "*" "./crashes"'
   ios crash rm <cwd> <pattern> [options]                             remove file pattern from dir. Ex.: 'ios crash rm "." "*"' to delete everything
   ios devicename [options]                                           Prints the devicename
   ios date [options]                                                 Prints the device date
   ios devicestate list [options]                                     Prints a list of all supported device conditions, like slow network, gpu etc.
   ios devicestate enable <profileTypeId> <profileId> [options]       Enables a profile with ids (use the list command to see options). It will only stay active until the process is terminated.
   >                                                                  Ex. "ios devicestate enable SlowNetworkCondition SlowNetwork3GGood"
   ios lang [--setlocale=<locale>] [--setlang=<newlang>] [options]    Sets or gets the Device language
   ios mobilegestalt <key>... [--plist] [options]                     Lets you query mobilegestalt keys. Standard output is json but if desired you can get
   >                                                                  it in plist format by adding the --plist param. 
   >                                                                  Ex.: "ios mobilegestalt MainScreenCanvasSizes ArtworkTraits --plist"
   ios diagnostics list [options]                                     List diagnostic infos
   ios pair [--p12file=<orgid>] [--password=<p12password>] [options]  Pairs the device. If the device is supervised, specify the path to the p12 file 
   >                                                                  to pair without a trust dialog. Specify the password either with the argument or
   >                                                                  by setting the environment variable 'P12_PASSWORD'
   ios profile list                                                   List the profiles on the device
   ios profile remove <profileName>                                   Remove the profileName from the device
   ios profile add <profileFile> [--p12file=<orgid>] [--password=<p12password>] Install profile file on the device. If supervised set p12file and password or the environment variable 'P12_PASSWORD'
   ios httpproxy <host> <port> [<user>] [<pass>] --p12file=<orgid> [--password=<p12password>] set global http proxy on supervised device. Use the password argument or set the environment variable 'P12_PASSWORD'
   >                                                                  Specify proxy password either as argument or using the environment var: PROXY_PASSWORD
   >                                                                  Use p12 file and password for silent installation on supervised devices.
   ios httpproxy remove [options]                                     Removes the global http proxy config. Only works with http proxies set by go-ios!
   ios ps [--apps] [options]                                          Dumps a list of running processes on the device.
   >                                                                  Use --nojson for a human-readable listing including BundleID when available. (not included with JSON output)
   >                                                                  --apps limits output to processes flagged by iOS as "isApplication". This greatly-filtered list
   >                                                                  should at least include user-installed software.  Additional packages will also be displayed depending on the version of iOS.
   ios ip [options]                                                   Uses the live pcap iOS packet capture to wait until it finds one that contains the IP address of the device.
   >                                                                  It relies on the MAC address of the WiFi adapter to know which is the right IP. 
   >                                                                  You have to disable the "automatic wifi address"-privacy feature of the device for this to work.
   >                                                                  If you wanna speed it up, open apple maps or similar to force network traffic.
   >                                                                  f.ex. "ios launch com.apple.Maps"
   ios forward [options] <hostPort> <targetPort>                      Similar to iproxy, forward a TCP connection to the device.
   ios dproxy [--binary]                                              Starts the reverse engineering proxy server. 
   >                                                                  It dumps every communication in plain text so it can be implemented easily. 
   >                                                                  Use "sudo launchctl unload -w /Library/Apple/System/Library/LaunchDaemons/com.apple.usbmuxd.plist"
   >                                                                  to stop usbmuxd and load to start it again should the proxy mess up things.
   >                                                                  The --binary flag will dump everything in raw binary without any decoding. 
   ios readpair                                                       Dump detailed information about the pairrecord for a device.
   ios install --path=<ipaOrAppFolder> [options]                      Specify a .app folder or an installable ipa file that will be installed.  
   ios pcap [options] [--pid=<processID>] [--process=<processName>]   Starts a pcap dump of network traffic, use --pid or --process to filter specific processes.
   ios apps [--system] [--all]                                        Retrieves a list of installed applications. --system prints out preinstalled system apps. --all prints all apps, including system, user, and hidden apps.
   ios launch <bundleID>                                              Launch app with the bundleID on the device. Get your bundle ID from the apps command.
   ios kill (<bundleID> | --pid=<processID> | --process=<processName>) [options] Kill app with the specified bundleID, process id, or process name on the device.
   ios runtest <bundleID>                                             Run a XCUITest. 
   ios runwda [--bundleid=<bundleid>] [--testrunnerbundleid=<testbundleid>] [--xctestconfig=<xctestconfig>] [--arg=<a>]... [--env=<e>]...[options]  runs WebDriverAgents
   >                                                                  specify runtime args and env vars like --env ENV_1=something --env ENV_2=else  and --arg ARG1 --arg ARG2
   ios ax [options]                                                   Access accessibility inspector features. 
   ios debug [--stop-at-entry] <app_path>                             Start debug with lldb
   ios fsync (rm | tree | mkdir) --path=<targetPath>                  Remove | treeview | mkdir in target path.
   ios fsync (pull | push) --srcPath=<srcPath> --dstPath=<dstPath>    Pull or Push file from srcPath to dstPath.
   ios reboot [options]                                               Reboot the given device
   ios -h | --help                                                    Prints this screen.
   ios --version | version [options]                                  Prints the version
   ios setlocation [options] [--lat=<lat>] [--lon=<lon>]			  Updates the location of the device to the provided by latitude and longitude coordinates. Example: setlocation --lat=40.730610 --lon=-73.935242
   ios setlocationgpx [options] [--gpxfilepath=<gpxfilepath>]		  Updates the location of the device based on the data in a GPX file. Example: setlocationgpx --gpxfilepath=/home/username/location.gpx
   ios resetlocation [options]										  Resets the location of the device to the actual one

  `, version)
	arguments, err := docopt.ParseDoc(usage)

	exitIfError("failed parsing args", err)
	disableJSON, _ := arguments.Bool("--nojson")
	if disableJSON {
		JSONdisabled = true
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	pretty, _ := arguments.Bool("--pretty")
	if pretty{
		prettyJSON = true
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

	listCommand, _ := arguments.Bool("list")
	diagnosticsCommand, _ := arguments.Bool("diagnostics")
	imageCommand, _ := arguments.Bool("image")
	deviceStateCommand, _ := arguments.Bool("devicestate")
	profileCommand, _ := arguments.Bool("profile")

	if listCommand && !diagnosticsCommand && !imageCommand && !deviceStateCommand && !profileCommand {
		b, _ = arguments.Bool("--details")
		printDeviceList(b)
		return
	}

	udid, _ := arguments.String("--udid")
	device, err := ios.GetDevice(udid)
	exitIfError("error getting devicelist", err)

	if mobileGestaltCommand(device, arguments) {
		return
	}

	if deviceStateCommand {
		if listCommand {
			deviceState(device, true, false, "", "")
			return
		}
		enable, _ := arguments.Bool("enable")
		profileTypeId, _ := arguments.String("<profileTypeId>")
		profileId, _ := arguments.String("<profileId>")
		deviceState(device, false, enable, profileTypeId, profileId)
	}

	b, _ = arguments.Bool("ip")
	if b {
		ip, err := pcap.FindIp(device)
		exitIfError("failed", err)
		println(convertToJSONString(ip))
		return
	}

	if crashCommand(device, arguments) {
		return
	}

	b, _ = arguments.Bool("pcap")
	if b {
		p, _ := arguments.String("--process")
		i, _ := arguments.Int("--pid")
		pcap.Pid = int32(i)
		pcap.ProcName = p
		err := pcap.Start(device)
		if err != nil {
			exitIfError("pcap failed", err)
		}
		return
	}

	b, _ = arguments.Bool("ps")
	if b {
		applicationsOnly, _ := arguments.Bool("--apps")
		processList(device, applicationsOnly)
		return
	}

	b, _ = arguments.Bool("install")
	if b {
		path, _ := arguments.String("--path")
		installApp(device, path)
		return
	}

	b, _ = arguments.Bool("uninstall")
	if b {
		bundleID, _ := arguments.String("<bundleID>")
		uninstallApp(device, bundleID)
		return
	}

	if imageCommand1(device, arguments) {
		return
	}

	b, _ = arguments.Bool("lang")
	if b {
		locale, _ := arguments.String("--setlocale")
		newlang, _ := arguments.String("--setlang")
		log.Debugf("lang --setlocale:%s --setlang:%s", locale, newlang)
		language(device, locale, newlang)
		return
	}

	b, _ = arguments.Bool("dproxy")
	if b {
		log.SetFormatter(&log.TextFormatter{})
		//log.SetLevel(log.DebugLevel)
		binaryMode, _ := arguments.Bool("--binary")
		startDebugProxy(device, binaryMode)
		return
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

	b, _ = arguments.Bool("setlocation")
	if b {
		lat, _ := arguments.String("--lat")
		lon, _ := arguments.String("--lon")
		setLocation(device, lat, lon)
		return
	}

	b, _ = arguments.Bool("setlocationgpx")
	if b {
		gpxFilePath, _ := arguments.String("--gpxfilepath")
		setLocationGPX(device, gpxFilePath)
		return
	}

	b, _ = arguments.Bool("resetlocation")
	if b {
		resetLocation(device)
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
		all, _ := arguments.Bool("--all")
		printInstalledApps(device, system, all)
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
		org, _ := arguments.String("--p12file")
		pwd, _ := arguments.String("--password")
		if pwd == "" {
			pwd = os.Getenv("P12_PASSWORD")
		}
		pairDevice(device, org, pwd)
		return
	}

	b, _ = arguments.Bool("readpair")
	if b {
		readPair(device)
		return
	}

	b, _ = arguments.Bool("httpproxy")
	if b {
		removeCommand, _ := arguments.Bool("remove")
		if removeCommand {
			mcinstall.RemoveProxy(device)
			exitIfError("failed removing proxy", err)
			log.Info("success")
			return
		}
		host, _ := arguments.String("<host>")
		port, _ := arguments.String("<port>")
		user, _ := arguments.String("<user>")
		pass, _ := arguments.String("<pass>")
		if pass == "" {
			pass = os.Getenv("PROXY_PASSWORD")
		}
		p12file, _ := arguments.String("--p12file")
		p12password, _ := arguments.String("--password")
		if p12password == "" {
			p12password = os.Getenv("P12_PASSWORD")
		}
		p12bytes, err := ioutil.ReadFile(p12file)
		exitIfError("could not read p12-file", err)

		err = mcinstall.SetHttpProxy(device, host, port, user, pass, p12bytes, p12password)
		exitIfError("failed", err)
		log.Info("success")
		return
	}

	b, _ = arguments.Bool("profile")
	if b {
		if listCommand {
			handleProfileList(device)
		}
		b, _ = arguments.Bool("add")
		if b {
			name, _ := arguments.String("<profileFile>")
			p12file, _ := arguments.String("--p12file")
			p12password, _ := arguments.String("--password")
			if p12password == "" {
				p12password = os.Getenv("P12_PASSWORD")
			}
			if p12file != "" {
				handleProfileAddSupervised(device, name, p12file, p12password)
				return
			}
			handleProfileAdd(device, name)
		}
		b, _ = arguments.Bool("remove")
		if b {
			name, _ := arguments.String("<profileName>")
			handleProfileRemove(device, name)
		}

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
		if bundleID == "" {
			log.Fatal("please provide a bundleID")
		}
		pControl, err := instruments.NewProcessControl(device)
		exitIfError("processcontrol failed", err)

		pid, err := pControl.LaunchApp(bundleID)
		exitIfError("launch app command failed", err)

		log.WithFields(log.Fields{"pid": pid}).Info("Process launched")
	}

	b, _ = arguments.Bool("kill")
	if b {
		var response []installationproxy.AppInfo
		bundleID, _ := arguments.String("<bundleID>")
		processIDint, _ := arguments.Int("--pid")
		processName, _ := arguments.String("--process")

		processID := uint64(processIDint)

		// Technically "Mach Kernel" is process 0, I suppose we provide no way to attempt to kill that.
		if bundleID == "" && processID == 0 && processName == "" {
			log.Fatal("please provide a bundleID")
		}
		pControl, err := instruments.NewProcessControl(device)
		exitIfError("processcontrol failed", err)
		svc, _ := installationproxy.New(device)

		// Look for correct process exe name for this bundleID. By default, searches only user-installed apps.
		if bundleID != ""{
			response, err = svc.BrowseAllApps()
			exitIfError("browsing apps failed", err)

			for _, app := range response {
				if app.CFBundleIdentifier == bundleID {
					processName = app.CFBundleExecutable
					break
				}
			}
			if processName == "" {
				log.Errorf(bundleID, " not installed")
				os.Exit(1)
				return
			}
		}

		service, err := instruments.NewDeviceInfoService(device)
		defer service.Close()
		exitIfError("failed opening deviceInfoService for getting process list", err)
		processList, _ := service.ProcessList()
		// ps
		for _, p := range processList {
			if (processID > 0 && p.Pid == processID) || (processName != "" && p.Name == processName) {
				err = pControl.KillProcess(p.Pid)
				exitIfError("kill process failed ", err)
				if bundleID != "" {
					log.Info(bundleID, " killed, Pid: ", p.Pid)
				} else {
					log.Info(p.Name, " killed, Pid: ", p.Pid)
				}
				return
			}
		}
		if bundleID != "" {
			log.Error("process of ", bundleID, " not found")
		} else if processName != "" {
			log.Error("process named ", processName, " not found")
		} else {
			log.Error("process with pid ", processID, " not found")
		}
		os.Exit(1)
		return
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

	if runWdaCommand(device, arguments) {
		return
	}

	b, _ = arguments.Bool("ax")
	if b {
		startAx(device)
		return
	}

	b, _ = arguments.Bool("debug")
	if b {
		appPath, _ := arguments.String("<app_path>")
		if appPath == "" {
			log.Fatal("parameter bundleid and app_path must be specified")
		}
		stopAtEntry, _ := arguments.Bool("--stop-at-entry")
		err = debugserver.Start(device, appPath, stopAtEntry)
		if err != nil {
			log.Error(err.Error())
		}
	}

	b, _ = arguments.Bool("reboot")
	if b {
		err := diagnostics.Reboot(device)
		if err != nil {
			log.Error(err)
		} else {
			log.Info("ok")
		}
		return
	}

	b, _ = arguments.Bool("fsync")
	if b {
		afcService, err := afc.New(device)
		exitIfError("fsync: connect afc service failed", err)
		b, _ = arguments.Bool("rm")
		if b {
			path, _ := arguments.String("--path")
			err = afcService.Remove(path)
			exitIfError("fsync: remove failed", err)
		}

		b, _ = arguments.Bool("tree")
		if b {
			path, _ := arguments.String("--path")
			err = afcService.TreeView(path, "", true)
			exitIfError("fsync: tree view failed", err)
		}

		b, _ = arguments.Bool("mkdir")
		if b {
			path, _ := arguments.String("--path")
			err = afcService.MkDir(path)
			exitIfError("fsync: mkdir failed", err)
		}

		b, _ = arguments.Bool("pull")
		if b {
			sp, _ := arguments.String("--srcPath")
			dp, _ := arguments.String("--dstPath")
			if dp != "" {
				ret, _ := ios.PathExists(dp)
				if !ret {
					err = os.MkdirAll(dp, os.ModePerm)
					exitIfError("mkdir failed", err)
				}
			}
			dp = path.Join(dp, filepath.Base(sp))
			err = afcService.Pull(sp, dp)
			exitIfError("fsync: pull failed", err)
		}
		b, _ = arguments.Bool("push")
		if b {
			sp, _ := arguments.String("--srcPath")
			dp, _ := arguments.String("--dstPath")
			err = afcService.Push(sp, dp)
			exitIfError("fsync: push failed", err)
		}
		afcService.Close()
		return
	}
}

func mobileGestaltCommand(device ios.DeviceEntry, arguments docopt.Opts) bool {
	b, _ := arguments.Bool("mobilegestalt")
	if b {
		conn, _ := diagnostics.New(device)
		keys := arguments["<key>"].([]string)
		plist, _ := arguments.Bool("--plist")
		resp, _ := conn.MobileGestaltQuery(keys)
		if plist {
			fmt.Printf("%s\n", ios.ToPlist(resp))
			return true
		}
		jb, _ := marshalJSON(resp)
		fmt.Printf("%s\n", jb)
		return true
	}
	return b
}

func imageCommand1(device ios.DeviceEntry, arguments docopt.Opts) bool {
	b, _ := arguments.Bool("image")
	if b {
		list, _ := arguments.Bool("list")
		if list {
			listMountedImages(device)
		}
		mount, _ := arguments.Bool("mount")
		if mount {
			path, _ := arguments.String("--path")
			err := imagemounter.MountImage(device, path)
			if err != nil {
				log.WithFields(log.Fields{"image": path, "udid": device.Properties.SerialNumber, "err": err}).
					Error("error mounting image")
				return true
			}
			log.WithFields(log.Fields{"image": path, "udid": device.Properties.SerialNumber}).Info("success mounting image")
		}
		auto, _ := arguments.Bool("auto")
		if auto {
			basedir, _ := arguments.String("--basedir")
			if basedir == "" {
				basedir = "."
			}
			err := imagemounter.FixDevImage(device, basedir)
			if err != nil {
				log.WithFields(log.Fields{"basedir": basedir, "udid": device.Properties.SerialNumber, "err": err}).
					Error("error mounting image")
				return true
			}
			log.WithFields(log.Fields{"basedir": basedir, "udid": device.Properties.SerialNumber}).Info("success mounting image")
		}
	}
	return b
}

func runWdaCommand(device ios.DeviceEntry, arguments docopt.Opts) bool {
	b, _ := arguments.Bool("runwda")
	if b {
		bundleID, _ := arguments.String("--bundleid")
		testbundleID, _ := arguments.String("--testrunnerbundleid")
		xctestconfig, _ := arguments.String("--xctestconfig")
		wdaargs := arguments["--arg"].([]string)
		wdaenv := arguments["--env"].([]string)

		if bundleID == "" && testbundleID == "" && xctestconfig == "" {
			log.Info("no bundle ids specified, falling back to defaults")
			bundleID, testbundleID, xctestconfig = "com.facebook.WebDriverAgentRunner.xctrunner", "com.facebook.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest"
		}
		if bundleID == "" || testbundleID == "" || xctestconfig == "" {
			log.WithFields(log.Fields{"bundleid": bundleID, "testbundleid": testbundleID, "xctestconfig": xctestconfig}).Error("please specify either NONE of bundleid, testbundleid and xctestconfig or ALL of them. At least one was empty.")
			return true
		}
		log.WithFields(log.Fields{"bundleid": bundleID, "testbundleid": testbundleID, "xctestconfig": xctestconfig}).Info("Running wda")
		go func() {
			err := testmanagerd.RunXCUIWithBundleIdsCtx(context.Background(), bundleID, testbundleID, xctestconfig, device, wdaargs, wdaenv)

			if err != nil {
				log.WithFields(log.Fields{"error": err}).Fatal("Failed running WDA")
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		signal := <-c
		log.Infof("os signal:%d received, closing..", signal)

		err := testmanagerd.CloseXCUITestRunner()
		if err != nil {
			log.Error("Failed closing wda-testrunner")
			os.Exit(1)
		}
		log.Info("Done Closing")
	}
	return b
}

func crashCommand(device ios.DeviceEntry, arguments docopt.Opts) bool {
	b, _ := arguments.Bool("crash")
	if b {
		ls, _ := arguments.Bool("ls")
		if ls {
			pattern, err := arguments.String("<pattern>")
			if err != nil || pattern == "" {
				pattern = "*"
			}
			files, err := crashreport.ListReports(device, pattern)
			exitIfError("failed listing crashreports", err)
			println(
				convertToJSONString(
					map[string]interface{}{"files": files, "length": len(files)},
				),
			)
		}
		cp, _ := arguments.Bool("cp")
		if cp {
			pattern, _ := arguments.String("<srcpattern>")
			target, _ := arguments.String("<target>")
			log.Debugf("cp %s %s", pattern, target)
			err := crashreport.DownloadReports(device, pattern, target)
			exitIfError("failed downloading crashreports", err)
		}

		rm, _ := arguments.Bool("rm")
		if rm {
			cwd, _ := arguments.String("<cwd>")
			pattern, _ := arguments.String("<pattern>")
			log.Debugf("rm %s %s", cwd, pattern)
			err := crashreport.RemoveReports(device, cwd, pattern)
			exitIfError("failed deleting crashreports", err)
		}
	}
	return b
}

func deviceState(device ios.DeviceEntry, list bool, enable bool, profileTypeId string, profileId string) {
	control, err := instruments.NewDeviceStateControl(device)
	exitIfError("failed to connect to deviceStateControl", err)
	profileTypes, err := control.List()
	if list {
		if JSONdisabled {
			outputPrettyStateList(profileTypes)
		} else {
			b, err := marshalJSON(profileTypes)
			exitIfError("failed json conversion", err)
			println(string(b))
		}
		return
	}
	exitIfError("failed listing device states", err)
	if enable {
		pType, profile, err := instruments.VerifyProfileAndType(profileTypes, profileTypeId, profileId)
		exitIfError("invalid arguments", err)
		log.Info("Enabling profile.. (this can take a while for ThermalConditions)")
		err = control.Enable(pType, profile)
		exitIfError("could not enable profile", err)
		log.Infof("Profile %s - %s is active! waiting for SIGTERM..", profileTypeId, profileId)
		c := make(chan os.Signal, syscall.SIGTERM)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Infof("Disabling profiletype %s", profileTypeId)
		err = control.Disable(pType)
		exitIfError("could not disable profile", err)
		log.Info("ok")
	}
}

func outputPrettyStateList(types []instruments.ProfileType) {
	var buffer bytes.Buffer
	for i, ptype := range types {
		buffer.WriteString(
			fmt.Sprintf("ProfileType %d\nName:%s\nisActive:%v\nIdentifier:%s\n\n",
				i, ptype.Name, ptype.IsActive, ptype.Identifier,
			),
		)
		for i, profile := range ptype.Profiles {
			buffer.WriteString(fmt.Sprintf("\tProfile %d:%s\n\tIdentifier:%s\n\t%s",
				i, profile.Name, profile.Identifier, profile.Description),
			)
			buffer.WriteString("\n\t------\n")
		}
		buffer.WriteString("\n\n")
	}
	println(buffer.String())
}

func listMountedImages(device ios.DeviceEntry) {
	conn, err := imagemounter.New(device)
	exitIfError("failed connecting to image mounter", err)
	signatures, err := conn.ListImages()
	exitIfError("failed getting image list", err)
	if len(signatures) == 0 {
		log.Infof("none")
		return
	}
	for _, sig := range signatures {
		log.Infof("%x", sig)
	}
}

func installApp(device ios.DeviceEntry, path string) {
	log.WithFields(
		log.Fields{"appPath": path, "device": device.Properties.SerialNumber}).Info("installing")
	conn, err := zipconduit.New(device)
	exitIfError("failed connecting to zipconduit, dev image installed?", err)
	err = conn.SendFile(path)
	exitIfError("failed writing", err)
}

func uninstallApp(device ios.DeviceEntry, bundleId string) {
	log.WithFields(
		log.Fields{"appPath": bundleId, "device": device.Properties.SerialNumber}).Info("uninstalling")
	svc, err := installationproxy.New(device)
	exitIfError("failed connecting to installationproxy", err)
	err = svc.Uninstall(bundleId)
	exitIfError("failed uninstalling", err)
}

func language(device ios.DeviceEntry, locale string, language string) {
	lang, err := ios.GetLanguage(device)
	exitIfError("failed getting language", err)

	err = ios.SetLanguage(device, ios.LanguageConfiguration{Language: language, Locale: locale})
	exitIfError("failed setting language", err)
	if lang.Language != language && language != "" {
		log.Debugf("Language should be changed from %s to %s waiting for Springboard to reboot", lang.Language, language)
		notificationproxy.WaitUntilSpringboardStarted(device)
	}
	lang, err = ios.GetLanguage(device)
	exitIfError("failed getting language", err)

	fmt.Println(convertToJSONString(lang))
}

func startAx(device ios.DeviceEntry) {
	go func() {
		deviceList, err := ios.ListDevices()
		exitIfError("failed converting to json", err)

		device := deviceList.DeviceList[0]

		conn, err := accessibility.New(device)
		exitIfError("failed starting ax", err)

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

		exitIfError("ax failed", err)
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
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

func startDebugProxy(device ios.DeviceEntry, binaryMode bool) {
	proxy := debugproxy.NewDebugProxy()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered a panic: %v", r)
				proxy.Close()
				debug.PrintStack()
				os.Exit(1)
				return
			}

		}()
		err := proxy.Launch(device, binaryMode)
		log.WithFields(log.Fields{"error": err}).Infof("DebugProxy Terminated abnormally")
		os.Exit(0)
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("Shutting down debugproxy")
	proxy.Close()
}

func handleProfileRemove(device ios.DeviceEntry, identifier string) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	err = profileService.RemoveProfile(identifier)
	exitIfError("failed adding profile", err)
	log.Infof("profile '%s' removed", identifier)
}

func handleProfileAdd(device ios.DeviceEntry, file string) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	filebytes, err := ioutil.ReadFile(file)
	exitIfError("could not read profile-file", err)
	err = profileService.AddProfile(filebytes)
	exitIfError("failed adding profile", err)
	log.Info("profile installed, you have to accept it in the device settings")
}

func handleProfileAddSupervised(device ios.DeviceEntry, file string, p12file string, p12password string) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	filebytes, err := ioutil.ReadFile(file)
	exitIfError("could not read profile-file", err)
	p12bytes, err := ioutil.ReadFile(p12file)
	exitIfError("could not read p12-file", err)
	err = profileService.AddProfileSupervised(filebytes, p12bytes, p12password)
	exitIfError("failed adding profile", err)
	log.Info("profile installed")
}

func handleProfileList(device ios.DeviceEntry) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	list, err := profileService.HandleList()
	exitIfError("failed getting profile list", err)
	fmt.Println(convertToJSONString(list))
}

func startForwarding(device ios.DeviceEntry, hostPort int, targetPort int) {
	forward.Forward(device, uint16(hostPort), uint16(targetPort))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func printDiagnostics(device ios.DeviceEntry) {
	log.Debug("print diagnostics")
	diagnosticsService, err := diagnostics.New(device)
	exitIfError("Starting diagnostics service failed with", err)

	values, err := diagnosticsService.AllValues()
	exitIfError("getting valued failed", err)

	fmt.Println(convertToJSONString(values))
}

func printDeviceDate(device ios.DeviceEntry) {
	allValues, err := ios.GetValues(device)
	exitIfError("failed getting values", err)

	formatedDate := time.Unix(int64(allValues.Value.TimeIntervalSince1970), 0).Format(time.RFC850)
	if JSONdisabled {
		fmt.Println(formatedDate)
	} else {
		fmt.Println(convertToJSONString(map[string]interface{}{"formatedDate": formatedDate, "TimeIntervalSince1970": allValues.Value.TimeIntervalSince1970}))
	}

}
func printInstalledApps(device ios.DeviceEntry, system bool, all bool) {
	svc, _ := installationproxy.New(device)
	var err error
	var response []installationproxy.AppInfo
	appType := ""
	if all {
		response, err = svc.BrowseAllApps()
		appType = "all"
	} else if system {
		response, err = svc.BrowseSystemApps()
		appType = "system"
	} else {
		response, err = svc.BrowseUserApps()
		appType = "user"
	}
	exitIfError("browsing " + appType + " apps failed", err)

	if JSONdisabled {
		log.Info(response)
	} else {
		fmt.Println(convertToJSONString(response))
	}
}

func printDeviceName(device ios.DeviceEntry) {
	allValues, err := ios.GetValues(device)
	exitIfError("failed getting values", err)
	if JSONdisabled {
		fmt.Println(allValues.Value.DeviceName)
	} else {
		fmt.Println(convertToJSONString(map[string]string{
			"devicename": allValues.Value.DeviceName,
		}))
	}
}

func saveScreenshot(device ios.DeviceEntry, outputPath string) {
	log.Debug("take screenshot")
	screenshotrService, err := screenshotr.New(device)
	exitIfError("Starting Screenshotr failed with", err)

	imageBytes, err := screenshotrService.TakeScreenshot()
	exitIfError("screenshotr failed", err)

	if outputPath == "" {
		time := time.Now().Format("20060102150405")
		path, _ := filepath.Abs("./screenshot" + time + ".png")
		outputPath = path
	}
	err = ioutil.WriteFile(outputPath, imageBytes, 0777)
	exitIfError("write file failed", err)

	if JSONdisabled {
		fmt.Println(outputPath)
	} else {
		log.WithFields(log.Fields{"outputPath": outputPath}).Info("File saved successfully")
	}
}

func setLocation(device ios.DeviceEntry, lat string, lon string) {
	err := simlocation.SetLocation(device, lat, lon)
	exitIfError("Setting location failed with", err)
}

func setLocationGPX(device ios.DeviceEntry, gpxFilePath string) {
	err := simlocation.SetLocationGPX(device, gpxFilePath)
	exitIfError("Setting location failed with", err)
}

func resetLocation(device ios.DeviceEntry) {
	err := simlocation.ResetLocation(device)
	exitIfError("Resetting location failed with", err)
}

func processList(device ios.DeviceEntry, applicationsOnly bool) {
	service, err := instruments.NewDeviceInfoService(device)
	defer service.Close()
	if err != nil {
		exitIfError("failed opening deviceInfoService for getting process list", err)
	}
	processList, err := service.ProcessList()
	if applicationsOnly {
		var applicationProcessList []instruments.ProcessInfo
		for _, processInfo := range processList {
			if processInfo.IsApplication {
				applicationProcessList = append(applicationProcessList,processInfo)
			}
		}
		processList = applicationProcessList
	}

	if JSONdisabled {
		outputProcessListNoJSON(device, processList)
	} else {
		fmt.Println(convertToJSONString(processList))
	}
}

func printDeviceList(details bool) {
	deviceList, err := ios.ListDevices()
	if err != nil {
		exitIfError("failed getting device list", err)
	}

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

func outputDetailedList(deviceList ios.DeviceList) {
	result := make([]detailsEntry, len(deviceList.DeviceList))
	for i, device := range deviceList.DeviceList {
		udid := device.Properties.SerialNumber
		allValues, err := ios.GetValues(device)
		exitIfError("failed getting values", err)
		result[i] = detailsEntry{udid, allValues.Value.ProductName, allValues.Value.ProductType, allValues.Value.ProductVersion}
	}
	fmt.Println(convertToJSONString(map[string][]detailsEntry{
		"deviceList": result,
	}))
}

func outputDetailedListNoJSON(deviceList ios.DeviceList) {
	for _, device := range deviceList.DeviceList {
		udid := device.Properties.SerialNumber
		allValues, err := ios.GetValues(device)
		exitIfError("failed getting values", err)
		fmt.Printf("%s  %s  %s %s\n", udid, allValues.Value.ProductName, allValues.Value.ProductType, allValues.Value.ProductVersion)
	}
}

func outputProcessListNoJSON(device ios.DeviceEntry, processes []instruments.ProcessInfo) {
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].Pid < processes[j].Pid
	})
	svc, _ := installationproxy.New(device)
	response, err := svc.BrowseAllApps()
	appInfoByExecutableName := make(map[string] installationproxy.AppInfo)

	if err != nil {
		log.Error("browsing installed apps failed. bundleID will not be included in output")
	} else {
		for _, app := range response {
			appInfoByExecutableName[app.CFBundleExecutable] = app
		}
	}

	var maxPid uint64
	maxNameLength := 15

	for _, processInfo := range processes {
		if processInfo.Pid > maxPid {
			maxPid = processInfo.Pid
		}
		if len(processInfo.Name) > maxNameLength {
			maxNameLength = len(processInfo.Name)
		}
	}
	maxPidLength := len(fmt.Sprintf("%d",maxPid))

	fmt.Printf("%*s %-*s %s  %s\n", maxPidLength, "PID", maxNameLength, "NAME", "START_DATE         ", "BUNDLE_ID")
	for _, processInfo := range processes {
		bundleID := ""
		appInfo, exists := appInfoByExecutableName[processInfo.Name]
		if exists{
			bundleID = appInfo.CFBundleIdentifier
		}
		fmt.Printf("%*d %-*s %s  %s\n", maxPidLength, processInfo.Pid, maxNameLength, processInfo.Name, processInfo.StartDate.Format("2006-01-02 15:04:05"), bundleID)
	}
}

func startListening() {
	go func() {
		for {
			deviceConn, err := ios.NewDeviceConnection(ios.DefaultUsbmuxdSocket)
			defer deviceConn.Close()
			if err != nil {
				log.Errorf("could not connect to %s with err %+v, will retry in 3 seconds...", ios.DefaultUsbmuxdSocket, err)
				time.Sleep(time.Second * 3)
				continue
			}
			muxConnection := ios.NewUsbMuxConnection(deviceConn)

			attachedReceiver, err := muxConnection.Listen()
			if err != nil {
				log.Error("Failed issuing Listen command, will retry in 3 seconds", err)
				deviceConn.Close()
				time.Sleep(time.Second * 3)
				continue
			}
			for {
				msg, err := attachedReceiver()
				if err != nil {
					log.Error("Stopped listening because of error")
					break
				}
				fmt.Println(convertToJSONString((msg)))
			}
		}
	}()
	c := make(chan os.Signal, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt)
	<-c
}

func printDeviceInfo(device ios.DeviceEntry) {
	allValues, err := ios.GetValuesPlist(device)
	if err != nil {
		exitIfError("failed getting info", err)
	}
	fmt.Println(convertToJSONString(allValues))
}

func runSyslog(device ios.DeviceEntry) {
	log.Debug("Run Syslog.")

	syslogConnection, err := syslog.New(device)
	exitIfError("Syslog connection failed", err)

	defer syslogConnection.Close()

	go func() {
		messageContainer := map[string]string{}
		for {
			logMessage, err := syslogConnection.ReadLogMessage()
			if err != nil {
				exitIfError("failed reading syslog", err)
			}
			logMessage = strings.TrimSuffix(logMessage, "\x00")
			logMessage = strings.TrimSuffix(logMessage, "\x0A")
			if JSONdisabled {
				fmt.Println(logMessage)
			} else {
				messageContainer["msg"] = logMessage
				fmt.Println(convertToJSONString(messageContainer))
			}
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func pairDevice(device ios.DeviceEntry, orgIdentityP12File string, p12Password string) {
	if orgIdentityP12File == "" {
		err := ios.Pair(device)
		exitIfError("Pairing failed", err)
		log.Infof("Successfully paired %s", device.Properties.SerialNumber)
		return
	}
	p12, err := os.ReadFile(orgIdentityP12File)
	exitIfError("Invalid file:"+orgIdentityP12File, err)
	err = ios.PairSupervised(device, p12, p12Password)
	exitIfError("Pairing failed", err)
	log.Infof("Successfully paired %s", device.Properties.SerialNumber)
}

func readPair(device ios.DeviceEntry) {
	record, err := ios.ReadPairRecord(device.Properties.SerialNumber)
	if err != nil {
		exitIfError("failed reading pairrecord", err)
	}
	json, err := marshalJSON(record)
	if err != nil {
		exitIfError("failed converting to json", err)
	}
	fmt.Printf("%s\n", json)
}

func marshalJSON(data interface{}) ([]byte, error){
	if prettyJSON{
		return json.MarshalIndent(data,"","    ")
	}else{
		return json.Marshal(data)
	}
}

func convertToJSONString(data interface{}) string {
	b, err := marshalJSON(data)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

func exitIfError(msg string, err error) {
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatalf(msg)
	}
}
