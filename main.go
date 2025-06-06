package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/danielpaulus/go-ios/ios/debugproxy"
	"github.com/danielpaulus/go-ios/ios/deviceinfo"
	"github.com/danielpaulus/go-ios/ios/tunnel"

	"github.com/danielpaulus/go-ios/ios/amfi"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"

	"github.com/danielpaulus/go-ios/ios/afc"

	"github.com/danielpaulus/go-ios/ios/crashreport"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"

	"github.com/danielpaulus/go-ios/ios/debugserver"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/zipconduit"

	"github.com/danielpaulus/go-ios/ios/simlocation"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/forward"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/danielpaulus/go-ios/ios/notificationproxy"
	"github.com/danielpaulus/go-ios/ios/pcap"
	syslog "github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"
)

// JSONdisabled enables or disables output in JSON format
var (
	JSONdisabled = false
	prettyJSON   = false
)

func main() {
	Main()
}

const version = "local-build"

// Main Exports main for testing
func Main() {

	usage := fmt.Sprintf(`go-ios %s

Usage:
    ios --version | version [options]
  ios -h | --help
  ios activate [options]
  ios apps [--system] [--all] [--list] [--filesharing] [options]
  ios assistivetouch (enable | disable | toggle | get) [--force] [options]
  ios ax [--font=<fontSize>] [options]
  ios batterycheck [options]
  ios batteryregistry [options]
  ios crash cp <srcpattern> <target> [options]
  ios crash ls [<pattern>] [options]
  ios crash rm <cwd> <pattern> [options]
  ios date [options]
  ios debug [options] [--stop-at-entry] <app_path>
  ios devicename [options]
  ios devicestate enable <profileTypeId> <profileId> [options]
  ios devicestate list [options]
  ios devmode (enable | get) [--enable-post-restart] [options]
  ios diagnostics list [options]
  ios diskspace [options]
  ios dproxy [--binary] [--mode=<all(default)|usbmuxd|utun>] [--iface=<iface>] [options]
  ios erase [--force] [options]
  ios forward [options] <hostPort> <targetPort>
  ios fsync [--app=bundleId] [options] (pull | push) --srcPath=<srcPath> --dstPath=<dstPath>
  ios fsync [--app=bundleId] [options] (rm [--r] | tree | mkdir) --path=<targetPath>
  ios httpproxy <host> <port> [<user>] [<pass>] --p12file=<orgid> --password=<p12password> [options]
  ios httpproxy remove [options]
  ios image auto [--basedir=<where_dev_images_are_stored>] [options]
  ios image list [options]
  ios image mount [--path=<imagepath>] [options]
  ios image unmount [options]
  ios info [display | lockdown] [options]
  ios install --path=<ipaOrAppFolder> [options]
  ios instruments notifications [options]
  ios ip [options]
  ios kill (<bundleID> | --pid=<processID> | --process=<processName>) [options]
  ios lang [--setlocale=<locale>] [--setlang=<newlang>] [options]
  ios launch <bundleID> [--wait] [--kill-existing] [--arg=<a>]... [--env=<e>]... [options]
  ios list [options] [--details]
  ios listen [options]
  ios memlimitoff (--process=<processName>) [options]
  ios mobilegestalt <key>... [--plist] [options]
  ios pair [--p12file=<orgid>] [--password=<p12password>] [options]
  ios pcap [options] [--pid=<processID>] [--process=<processName>]
  ios prepare [--skip-all] [--skip=<option>]... [--certfile=<cert_file_path>] [--orgname=<org_name>] [--locale] [--lang] [options]
  ios prepare create-cert
  ios prepare printskip
  ios profile add <profileFile> [--p12file=<orgid>] [--password=<p12password>] [options]
  ios profile list [options]
  ios profile remove <profileName> [options]
  ios ps [--apps] [options]
  ios readpair [options]
  ios reboot [options]
  ios resetax [options]
  ios resetlocation [options]
  ios rsd ls [options]
  ios runtest [--bundle-id=<bundleid>] [--test-runner-bundle-id=<testrunnerbundleid>] [--xctest-config=<xctestconfig>] [--log-output=<file>] [--xctest] [--test-to-run=<tests>]... [--test-to-skip=<tests>]... [--env=<e>]... [options]
  ios runwda [--bundleid=<bundleid>] [--testrunnerbundleid=<testbundleid>] [--xctestconfig=<xctestconfig>] [--log-output=<file>] [--arg=<a>]... [--env=<e>]... [options]
  ios runxctest [--xctestrun-file-path=<xctestrunFilePath>] [--log-output=<file>] [options]
  ios screenshot [options] [--output=<outfile>] [--stream] [--port=<port>]
  ios setlocation [options] [--lat=<lat>] [--lon=<lon>]
  ios setlocationgpx [options] [--gpxfilepath=<gpxfilepath>]
  ios syslog [--parse] [options]
  ios sysmontap [options]
  ios timeformat (24h | 12h | toggle | get) [--force] [options]
  ios tunnel ls [options]
  ios tunnel start [options] [--pair-record-path=<pairrecordpath>] [--userspace]
  ios tunnel stopagent
  ios uninstall <bundleID> [options]
  ios voiceover (enable | disable | toggle | get) [--force] [options]
  ios zoom (enable | disable | toggle | get) [--force] [options]

Options:
  -v --verbose              Enable Debug Logging.
  -t --trace                Enable Trace Logging (dump every message).
  --nojson                  Disable JSON output
  --pretty                  Pretty-print JSON command output
  -h --help                 Show this screen.
  --udid=<udid>             UDID of the device.
  --tunnel-info-port=<port> When go-ios is used to manage tunnels for iOS 17+ it exposes them on an HTTP-API for localhost (default port: 28100)
  --address=<ipv6addrr>     Address of the device on the interface. This parameter is optional and can be set if a tunnel created by MacOS needs to be used.
  >                         To get this value run "log stream --debug --info --predicate 'eventMessage LIKE "*Tunnel established*" OR eventMessage LIKE "*for server port*"'",
  >                         connect a device and open Xcode
  --rsd-port=<port>         Port of remote service discovery on the device through the tunnel
  >                         This parameter is similar to '--address' and can be obtained by the same log filter
  --proxyurl=<url>          Set this if you want go-ios to use a http proxy for outgoing requests, like for downloading images or contacting Apple during device activation.
  >                         A simple format like: "http://PROXY_LOGIN:PROXY_PASS@proxyIp:proxyPort" works. Otherwise use the HTTP_PROXY system env var.
  --userspace-port=<port>   Optional. Set this if you run a command supplying rsd-port and address and your device is using userspace tunnel

The commands work as following:
	The default output of all commands is JSON. Should you prefer human readable outout, specify the --nojson option with your command.
	By default, the first device found will be used for a command unless you specify a --udid=some_udid switch.
	Specify -v for debug logging and -t for dumping every message.

      ios --version | version [options]                                  Prints the version
   ios -h | --help                                                    Prints this screen.
   ios activate [options]                                             Activate a device
   ios apps [--system] [--all] [--list] [--filesharing]               Retrieves a list of installed applications. --system prints out preinstalled system apps. --all prints all apps, including system, user, and hidden apps. --list only prints bundle ID, bundle name and version number. --filesharing only prints apps which enable documents sharing.
   ios assistivetouch (enable | disable | toggle | get) [--force] [options] Enables, disables, toggles, or returns the state of the "AssistiveTouch" software home-screen button. iOS 11+ only (Use --force to try on older versions).
   ios ax [--font=<fontSize>] [options]                               Access accessibility inspector features.
   ios batterycheck [options]                                         Prints battery info.
   ios batteryregistry [options]                                      Prints battery registry stats like Temperature, Voltage.
   ios crash cp <srcpattern> <target> [options]                       copy "file pattern" to the target dir. Ex.: 'ios crash cp "*" "./crashes"'
   ios crash ls [<pattern>] [options]                                 run "ios crash ls" to get all crashreports in a list,
   >                                                                  or use a pattern like 'ios crash ls "*ips*"' to filter
   ios crash rm <cwd> <pattern> [options]                             remove file pattern from dir. Ex.: 'ios crash rm "." "*"' to delete everything
   ios date [options]                                                 Prints the device date
   ios debug [--stop-at-entry] <app_path>                             Start debug with lldb
   ios devicename [options]                                           Prints the devicename
   ios devicestate enable <profileTypeId> <profileId> [options]       Enables a profile with ids (use the list command to see options). It will only stay active until the process is terminated.
   >                                                                  Ex. "ios devicestate enable SlowNetworkCondition SlowNetwork3GGood"
   ios devicestate list [options]                                     Prints a list of all supported device conditions, like slow network, gpu etc.
   ios devmode (enable | get) [--enable-post-restart] [options]	  Enable developer mode on the device or check if it is enabled. Can also completely finalize developer mode setup after device is restarted.
   ios diagnostics list [options]                                     List diagnostic infos
   ios diskspace [options]											  Prints disk space info.
   ios dproxy [--binary] [--mode=<all(default)|usbmuxd|utun>] [--iface=<iface>] [options] Starts the reverse engineering proxy server.
   >                                                                  It dumps every communication in plain text so it can be implemented easily.
   >                                                                  Use "sudo launchctl unload -w /Library/Apple/System/Library/LaunchDaemons/com.apple.usbmuxd.plist"
   >                                                                  to stop usbmuxd and load to start it again should the proxy mess up things.
   >                                                                  The --binary flag will dump everything in raw binary without any decoding.
   ios erase [--force] [options]                                      Erase the device. It will prompt you to input y+Enter unless --force is specified.
   ios forward [options] <hostPort> <targetPort>                      Similar to iproxy, forward a TCP connection to the device.
   ios fsync [--app=bundleId] [options] (pull | push) --srcPath=<srcPath> --dstPath=<dstPath>    Pull or Push file from srcPath to dstPath.
   ios fsync [--app=bundleId] [options] (rm [--r] | tree | mkdir) --path=<targetPath>            Remove | treeview | mkdir in target path. --r used alongside rm will recursively remove all files and directories from target path.
   ios httpproxy <host> <port> [<user>] [<pass>] --p12file=<orgid> [--password=<p12password>] set global http proxy on supervised device. Use the password argument or set the environment variable 'P12_PASSWORD'
   >                                                                  Specify proxy password either as argument or using the environment var: PROXY_PASSWORD
   >                                                                  Use p12 file and password for silent installation on supervised devices.
   ios httpproxy remove [options]                                     Removes the global http proxy config. Only works with http proxies set by go-ios!
   ios image auto [--basedir=<where_dev_images_are_stored>] [options] Automatically download correct dev image from the internets and mount it.
   >                                                                  You can specify a dir where images should be cached.
   >                                                                  The default is the current dir.
   ios image list [options]                                           List currently mounted developers images' signatures
   ios image mount [--path=<imagepath>] [options]                     Mount a image from <imagepath>
   >                                                                  For iOS 17+ (personalized developer disk images) <imagepath> must point to the "Restore" directory inside the developer disk
   ios image unmount [options]                                        Unmount developer disk image
   ios info [display | lockdown] [options]                            Prints a dump of device information from the given source.
   ios install --path=<ipaOrAppFolder> [options]                      Specify a .app folder or an installable ipa file that will be installed.
   ios instruments notifications [options]                            Listen to application state notifications
   ios ip [options]                                                   Uses the live pcap iOS packet capture to wait until it finds one that contains the IP address of the device.
   >                                                                  It relies on the MAC address of the WiFi adapter to know which is the right IP.
   >                                                                  You have to disable the "automatic wifi address"-privacy feature of the device for this to work.
   >                                                                  If you wanna speed it up, open apple maps or similar to force network traffic.
   >                                                                  f.ex. "ios launch com.apple.Maps"
   ios kill (<bundleID> | --pid=<processID> | --process=<processName>) [options] Kill app with the specified bundleID, process id, or process name on the device.
   ios lang [--setlocale=<locale>] [--setlang=<newlang>] [options]    Sets or gets the Device language. ios lang will print the current language and locale, as well as a list of all supported langs and locales.
   ios launch <bundleID> [--wait] [--kill-existing] [--arg=<a>]... [--env=<e>]... [options] Launch app with the bundleID on the device. Get your bundle ID from the apps command. --wait keeps the connection open if you want logs.
   ios list [options] [--details]                                     Prints a list of all connected device's udids. If --details is specified, it includes version, name and model of each device.
   ios listen [options]                                               Keeps a persistent connection open and notifies about newly connected or disconnected devices.
   ios memlimitoff (--process=<processName>) [options]                Waives memory limit set by iOS (For instance a Broadcast Extension limit is 50 MB).
   ios mobilegestalt <key>... [--plist] [options]                     Lets you query mobilegestalt keys. Standard output is json but if desired you can get
   >                                                                  it in plist format by adding the --plist param.
   >                                                                  Ex.: "ios mobilegestalt MainScreenCanvasSizes ArtworkTraits --plist"
   ios pair [--p12file=<orgid>] [--password=<p12password>] [options]  Pairs the device. If the device is supervised, specify the path to the p12 file
   >                                                                  to pair without a trust dialog. Specify the password either with the argument or
   >                                                                  by setting the environment variable 'P12_PASSWORD'
   ios pcap [options] [--pid=<processID>] [--process=<processName>]   Starts a pcap dump of network traffic, use --pid or --process to filter specific processes.
   ios prepare [--skip-all] [--skip=<option>]... [--certfile=<cert_file_path>] [--orgname=<org_name>] [--locale] [--lang] [options] prepare a device. Use skip-all to skip everything multiple --skip args to skip only a subset.
   >                                                                  You can use 'ios prepare printskip' to get a list of all options to skip. Use certfile and orgname if you want to supervise the device. If you need certificates
   >                                                                  to supervise, run 'ios prepare create-cert' and go-ios will generate one you can use. locale and lang are optional, the default is en_US and en.
   >                                                                  Run 'ios lang' to see a list of all supported locales and languages.
   ios prepare create-cert                                            A nice util to generate a certificate you can use for supervising devices. Make sure you rename and store it in a safe place.
   ios prepare printskip                                              Print all options you can skip.
   ios profile add <profileFile> [--p12file=<orgid>] [--password=<p12password>] Install profile file on the device. If supervised set p12file and password or the environment variable 'P12_PASSWORD'
   ios profile list                                                   List the profiles on the device
   ios profile remove <profileName>                                   Remove the profileName from the device
   ios ps [--apps] [options]                                          Dumps a list of running processes on the device.
   >                                                                  Use --nojson for a human-readable listing including BundleID when available. (not included with JSON output)
   >                                                                  --apps limits output to processes flagged by iOS as "isApplication". This greatly-filtered list
   >                                                                  should at least include user-installed software.  Additional packages will also be displayed depending on the version of iOS.
   ios readpair                                                       Dump detailed information about the pairrecord for a device.
   ios reboot [options]                                               Reboot the given device
   ios resetax [options]                                              Reset accessibility settings to defaults.
   ios resetlocation [options]                                        Resets the location of the device to the actual one
   ios rsd ls [options]											  List RSD services and their port.
   ios runtest [--bundle-id=<bundleid>] [--test-runner-bundle-id=<testbundleid>] [--xctest-config=<xctestconfig>] [--log-output=<file>] [--xctest] [--test-to-run=<tests>]... [--test-to-skip=<tests>]... [--env=<e>]... [options]                    Run a XCUITest. If you provide only bundle-id go-ios will try to dynamically create test-runner-bundle-id and xctest-config.
   >                                                                  If you provide '-' as log output, it prints resuts to stdout.
   >                                                                  To be able to filter for tests to run or skip, use one argument per test selector. Example: runtest --test-to-run=(TestTarget.)TestClass/testMethod --test-to-run=(TestTarget.)TestClass/testMethod (the value for 'TestTarget' is optional)
   >                                                                  The method name can also be omitted and in this case all tests of the specified class are run
   ios runwda [--bundleid=<bundleid>] [--testrunnerbundleid=<testbundleid>] [--xctestconfig=<xctestconfig>] [--log-output=<file>] [--arg=<a>]... [--env=<e>]...[options]  runs WebDriverAgents
   >                                                                  specify runtime args and env vars like --env ENV_1=something --env ENV_2=else  and --arg ARG1 --arg ARG2
   ios runxctest [--xctestrun-file-path=<xctestrunFilePath>]  [--log-output=<file>] [options]                    Run a XCTest. The --xctestrun-file-path specifies the path to the .xctestrun file to configure the test execution.
   >                                                                  If you provide '-' as log output, it prints resuts to stdout.
   ios screenshot [options] [--output=<outfile>] [--stream] [--port=<port>]  Takes a screenshot and writes it to the current dir or to <outfile>  If --stream is supplied it
   >                                                                  starts an mjpeg server at 0.0.0.0:3333. Use --port to set another port.
   ios setlocation [options] [--lat=<lat>] [--lon=<lon>]              Updates the location of the device to the provided by latitude and longitude coordinates. Example: setlocation --lat=40.730610 --lon=-73.935242
   ios setlocationgpx [options] [--gpxfilepath=<gpxfilepath>]         Updates the location of the device based on the data in a GPX file. Example: setlocationgpx --gpxfilepath=/home/username/location.gpx
   ios syslog [--parse] [options]                                     Prints a device's log output, Use --parse to parse the fields from the log
   ios sysmontap                                                      Get system stats like MEM, CPU
   ios timeformat (24h | 12h | toggle | get) [--force] [options] Sets, or returns the state of the "time format". iOS 11+ only (Use --force to try on older versions).
   ios tunnel ls                                                      List currently started tunnels. Use --enabletun to activate using TUN devices rather than user space network. Requires sudo/admin shells. 
   ios tunnel start [options] [--pair-record-path=<pairrecordpath>] [--enabletun]   Creates a tunnel connection to the device. If the device was not paired with the host yet, device pairing will also be executed.
   >           														  On systems with System Integrity Protection enabled the argument '--pair-record-path=default' can be used to point to /var/db/lockdown/RemotePairing/user_501.
   >                                                                  If nothing is specified, the current dir is used for the pair record.
   >                                                                  This command needs to be executed with admin privileges.
   >                                                                  (On MacOS the process 'remoted' must be paused before starting a tunnel is possible 'sudo pkill -SIGSTOP remoted', and 'sudo pkill -SIGCONT remoted' to resume)
   ios voiceover (enable | disable | toggle | get) [--force] [options] Enables, disables, toggles, or returns the state of the "VoiceOver" software home-screen button. iOS 11+ only (Use --force to try on older versions).
   ios zoom (enable | disable | toggle | get) [--force] [options] Enables, disables, toggles, or returns the state of the "ZoomTouch" software home-screen button. iOS 11+ only (Use --force to try on older versions).

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
	if pretty {
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
	// log.SetReportCaller(true)
	log.Debug(arguments)

	skipAgent, _ := os.LookupEnv("ENABLE_GO_IOS_AGENT")
	if skipAgent == "user" || skipAgent == "kernel" {
		tunnel.RunAgent(skipAgent)
	}

	if !tunnel.IsAgentRunning() {
		log.Warn("go-ios agent is not running. You might need to start it with 'ios tunnel start' for ios17+. Use ENABLE_GO_IOS_AGENT=user for userspace tunnel or ENABLE_GO_IOS_AGENT=kernel for kernel tunnel for the experimental daemon mode.")
	}
	shouldPrintVersionNoDashes, _ := arguments.Bool("version")
	shouldPrintVersion, _ := arguments.Bool("--version")
	if shouldPrintVersionNoDashes || shouldPrintVersion {
		printVersion()
		return
	}
	proxyUrl, _ := arguments.String("--proxyurl")
	exitIfError("could not parse proxy url", ios.UseHttpProxy(proxyUrl))

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

	tunnelInfoHost, err := arguments.String("--tunnel-info-host")
	if err != nil {
		tunnelInfoHost = ios.HttpApiHost()
	}

	tunnelInfoPort, err := arguments.Int("--tunnel-info-port")
	if err != nil {
		tunnelInfoPort = ios.HttpApiPort()
	}

	tunnelCommand, _ := arguments.Bool("tunnel")

	udid, _ := arguments.String("--udid")
	address, addressErr := arguments.String("--address")
	rsdPort, rsdErr := arguments.Int("--rsd-port")
	userspaceTunnelHost, userspaceTunnelHostErr := arguments.String("--userspace-host")
	if userspaceTunnelHostErr != nil {
		userspaceTunnelHost = ios.HttpApiHost()
	}

	userspaceTunnelPort, userspaceTunnelErr := arguments.Int("--userspace-port")

	device, err := ios.GetDevice(udid)
	// device address and rsd port are only available after the tunnel started
	if !tunnelCommand {
		exitIfError("Device not found: "+udid, err)
		if addressErr == nil && rsdErr == nil {
			if userspaceTunnelErr == nil {
				device.UserspaceTUN = true
				device.UserspaceTUNHost = userspaceTunnelHost
				device.UserspaceTUNPort = userspaceTunnelPort
			}
			device = deviceWithRsdProvider(device, udid, address, rsdPort)
		} else {
			info, err := tunnel.TunnelInfoForDevice(device.Properties.SerialNumber, tunnelInfoHost, tunnelInfoPort)
			if err == nil {
				device.UserspaceTUNPort = info.UserspaceTUNPort
				device.UserspaceTUNHost = userspaceTunnelHost
				device.UserspaceTUN = info.UserspaceTUN
				device = deviceWithRsdProvider(device, udid, info.Address, info.RsdPort)
			} else {
				log.WithField("udid", device.Properties.SerialNumber).Warn("failed to get tunnel info")
			}
		}
	}

	b, _ = arguments.Bool("erase")
	if b {
		force, _ := arguments.Bool("--force")
		if !force {
			log.Warnf("are you sure you want to erase device %s? (y/n)", device.Properties.SerialNumber)
			reader := bufio.NewReader(os.Stdin)
			// ReadString will block until the delimiter is entered
			input, err := reader.ReadString('\n')
			exitIfError("An error occured while reading input", err)
			if !strings.HasPrefix(input, "y") {
				log.Errorf("abort")
				return
			}
		}

		exitIfError("failed erasing", mcinstall.Erase(device))
		print(convertToJSONString("ok"))
		return
	}

	rsdCommand, _ := arguments.Bool("rsd")
	if rsdCommand {
		listCommand, _ := arguments.Bool("ls")
		if listCommand {
			services := device.Rsd.GetServices()
			if JSONdisabled {
				fmt.Println(services)
			} else {
				b, err := marshalJSON(services)
				exitIfError("failed json conversion", err)
				println(string(b))
			}
			return
		}
	}

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

	b, _ = arguments.Bool("prepare")
	if b {
		b, _ = arguments.Bool("create-cert")
		if b {
			cert, err := ios.CreateDERFormattedSupervisionCert()
			exitIfError("failed creating cert", err)
			err = os.WriteFile("supervision-cert.der", cert.CertDER, 0o777)
			log.Info("supervision-cert.der")
			exitIfError("failed writing cert", err)
			err = os.WriteFile("supervision-cert.pem", cert.CertPEM, 0o777)
			log.Info("supervision-cert.pem")
			exitIfError("failed writing cert", err)
			err = os.WriteFile("supervision-private-key.key", cert.PrivateKeyDER, 0o777)
			log.Info("supervision-private-key.key")
			exitIfError("failed writing cert", err)
			err = os.WriteFile("supervision-private-key.pem", cert.PrivateKeyPEM, 0o777)
			log.Info("supervision-private-key.pem")
			exitIfError("failed writing key", err)
			err = os.WriteFile("supervision-csr.csr", []byte(cert.Csr), 0o777)
			log.Info("supervision-csr.csr")
			exitIfError("failed writing cert", err)
			log.Info("Golang does not have good PKCS12 format sadly. If you need a p12 file run this: " +
				"'openssl pkcs12 -export -inkey supervision-private-key.pem -in supervision-cert.pem -out certificate.p12 -password pass:a'")
			return
		}
		b, _ = arguments.Bool("printskip")
		if b {
			println(convertToJSONString(mcinstall.GetAllSetupSkipOptions()))
			return
		}
		skip := mcinstall.GetAllSetupSkipOptions()
		skip1 := arguments["--skip"].([]string)
		if len(skip1) > 0 {
			skip = skip1
		}

		certfile, _ := arguments.String("--certfile")
		orgname, _ := arguments.String("--orgname")
		locale, _ := arguments.String("--locale")
		lang, _ := arguments.String("--lang")
		var certBytes []byte
		if certfile != "" {
			certBytes, err = os.ReadFile(certfile)
			exitIfError("failed opening cert file", err)
			if orgname == "" {
				log.Fatal("--orgname must be specified if certfile for supervision is provided")
			}
		}
		exitIfError("failed erasing", mcinstall.Prepare(device, skip, certBytes, orgname, locale, lang))
		print(convertToJSONString("ok"))
		return
	}

	b, _ = arguments.Bool("activate")
	if b {
		exitIfError("failed activation", mobileactivation.Activate(device))
		return
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
	if instrumentsCommand(device, arguments) {
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

	b, _ = arguments.Bool("assistivetouch")
	if b {
		force, _ := arguments.Bool("--force")
		b, _ = arguments.Bool("enable")
		if b {
			assistiveTouch(device, "enable", force)
		}
		b, _ = arguments.Bool("disable")
		if b {
			assistiveTouch(device, "disable", force)
		}
		b, _ = arguments.Bool("toggle")
		if b {
			assistiveTouch(device, "toggle", force)
		}
		b, _ = arguments.Bool("get")
		if b {
			assistiveTouch(device, "get", force)
		}
	}

	b, _ = arguments.Bool("voiceover")
	if b {
		force, _ := arguments.Bool("--force")
		b, _ = arguments.Bool("enable")
		if b {
			voiceOver(device, "enable", force)
		}
		b, _ = arguments.Bool("disable")
		if b {
			voiceOver(device, "disable", force)
		}
		b, _ = arguments.Bool("toggle")
		if b {
			voiceOver(device, "toggle", force)
		}
		b, _ = arguments.Bool("get")
		if b {
			voiceOver(device, "get", force)
		}
	}

	b, _ = arguments.Bool("zoom")
	if b {
		force, _ := arguments.Bool("--force")
		b, _ = arguments.Bool("enable")
		if b {
			zoomTouch(device, "enable", force)
		}
		b, _ = arguments.Bool("disable")
		if b {
			zoomTouch(device, "disable", force)
		}
		b, _ = arguments.Bool("toggle")
		if b {
			zoomTouch(device, "toggle", force)
		}
		b, _ = arguments.Bool("get")
		if b {
			zoomTouch(device, "get", force)
		}
	}

	b, _ = arguments.Bool("dproxy")
	if b {
		log.SetFormatter(&log.TextFormatter{})
		// log.SetLevel(log.DebugLevel)
		binaryMode, _ := arguments.Bool("--binary")
		startDebugProxy(device, binaryMode)
		return
	}

	b, _ = arguments.Bool("info")
	if b {
		if display, _ := arguments.Bool("display"); display {
			deviceInfo, err := deviceinfo.NewDeviceInfo(device)
			exitIfError("Can't connect to deviceinfo service", err)
			defer deviceInfo.Close()

			info, err := deviceInfo.GetDisplayInfo()
			exitIfError("Can't fetch dispaly info", err)

			fmt.Println(convertToJSONString(info))
		} else if lockdown, _ := arguments.Bool("lockdown"); lockdown {
			printDeviceInfo(device)
		} else {
			// When subcommand is missing, it defaults to lockdown.
			// Unknown subcommands don't reach this line and quit early.
			printDeviceInfo(device)
		}
		return
	}

	b, _ = arguments.Bool("syslog")
	if b {
		parse, _ := arguments.Bool("--parse")

		runSyslog(device, parse)
		return
	}

	b, _ = arguments.Bool("screenshot")
	if b {
		stream, _ := arguments.Bool("--stream")
		port, _ := arguments.String("--port")
		path, _ := arguments.String("--output")
		if stream {
			if port == "" {
				port = "3333"
			}
			err := instruments.StartMJPEGStreamingServer(device, port)
			exitIfError("failed starting mjpeg", err)
			return
		}
		saveScreenshot(device, path)
		return
	}

	b, _ = arguments.Bool("setlocation")
	if b {
		lat, _ := arguments.String("--lat")
		lon, _ := arguments.String("--lon")

		if device.SupportsRsd() {
			server, err := instruments.NewLocationSimulationService(device)
			exitIfError("failed to create location simulation service:", err)

			startLocationSimulation(server, lat, lon)
			return
		}

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
		list, _ := arguments.Bool("--list")
		system, _ := arguments.Bool("--system")
		all, _ := arguments.Bool("--all")
		filesharing, _ := arguments.Bool("--filesharing")
		printInstalledApps(device, system, all, list, filesharing)
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

	b, _ = arguments.Bool("timeformat")
	if b {
		force, _ := arguments.Bool("--force")
		b, _ = arguments.Bool("24h")
		if b {
			timeFormat(device, "24h", force)
		}
		b, _ = arguments.Bool("12h")
		if b {
			timeFormat(device, "12h", force)
		}
		b, _ = arguments.Bool("toggle")
		if b {
			timeFormat(device, "toggle", force)
		}
		b, _ = arguments.Bool("get")
		if b {
			timeFormat(device, "get", force)
		}
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
		p12bytes, err := os.ReadFile(p12file)
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
		wait, _ := arguments.Bool("--wait")
		bKillExisting, _ := arguments.Bool("--kill-existing")
		bundleID, _ := arguments.String("<bundleID>")
		if bundleID == "" {
			log.Fatal("please provide a bundleID")
		}
		pControl, err := instruments.NewProcessControl(device)
		exitIfError("processcontrol failed", err)
		opts := map[string]any{}
		if bKillExisting {
			opts["KillExisting"] = 1
		} // end if
		args := toArgs(arguments["--arg"].([]string))
		envs := toEnvs(arguments["--env"].([]string))
		pid, err := pControl.LaunchAppWithArgs(bundleID, args, envs, opts)
		exitIfError("launch app command failed", err)
		log.WithFields(log.Fields{"pid": pid}).Info("Process launched")
		if wait {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			<-c
			log.WithFields(log.Fields{"pid": pid}).Info("stop listening to logs")
		}
	}

	b, _ = arguments.Bool("sysmontap")
	if b {
		printSysmontapStats(device)
	}

	b, _ = arguments.Bool("memlimitoff")
	if b {
		processName, _ := arguments.String("--process")

		pControl, err := instruments.NewProcessControl(device)
		exitIfError("processcontrol failed", err)
		defer pControl.Close()

		svc, err := instruments.NewDeviceInfoService(device)
		exitIfError("failed opening deviceInfoService for getting process list", err)
		defer svc.Close()

		processList, _ := svc.ProcessList()
		for _, process := range processList {
			if process.Pid > 1 && process.Name == processName {
				disabled, err := pControl.DisableMemoryLimit(process.Pid)
				exitIfError("DisableMemoryLimit failed", err)
				log.WithFields(log.Fields{"process": process.Name, "pid": process.Pid}).Info("memory limit is off: ", disabled)
			}
		}
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
		if bundleID != "" {
			response, err = svc.BrowseAllApps()
			exitIfError("browsing apps failed", err)

			for _, app := range response {
				if app.CFBundleIdentifier() == bundleID {
					processName = app.CFBundleExecutable()
					break
				}
			}
			if processName == "" {
				log.Errorf("%s not installed", bundleID)
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
		bundleID, _ := arguments.String("--bundle-id")
		testRunnerBundleId, _ := arguments.String("--test-runner-bundle-id")
		xctestConfig, _ := arguments.String("--xctest-config")

		testsToRunArg := arguments["--test-to-run"]
		var testsToRun []string
		if testsToRunArg != nil && len(testsToRunArg.([]string)) > 0 {
			testsToRun = testsToRunArg.([]string)
		}

		testsToSkipArg := arguments["--test-to-skip"]
		var testsToSkip []string
		testsToSkip = nil
		if testsToSkipArg != nil && len(testsToSkipArg.([]string)) > 0 {
			testsToSkip = testsToSkipArg.([]string)
		}

		rawTestlog, rawTestlogErr := arguments.String("--log-output")
		env := splitKeyValuePairs(arguments["--env"].([]string), "=")
		isXCTest, _ := arguments.Bool("--xctest")

		config := testmanagerd.TestConfig{
			BundleId:           bundleID,
			TestRunnerBundleId: testRunnerBundleId,
			XctestConfigName:   xctestConfig,
			Env:                env,
			TestsToRun:         testsToRun,
			TestsToSkip:        testsToSkip,
			XcTest:             isXCTest,
			Device:             device,
		}

		if rawTestlogErr == nil {
			var writer *os.File = os.Stdout
			if rawTestlog != "-" {
				file, err := os.Create(rawTestlog)
				exitIfError("Cannot open file "+rawTestlog, err)
				writer = file
			}
			defer writer.Close()

			config.Listener = testmanagerd.NewTestListener(writer, writer, os.TempDir())

			testResults, err := testmanagerd.RunTestWithConfig(context.TODO(), config)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Info("Failed running Xcuitest")
			}

			log.Info(fmt.Printf("%+v", testResults))
		} else {
			config.Listener = testmanagerd.NewTestListener(io.Discard, io.Discard, os.TempDir())
			_, err := testmanagerd.RunTestWithConfig(context.TODO(), config)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Info("Failed running Xcuitest")
			}
		}
		return
	}

	b, _ = arguments.Bool("runxctest")
	if b {
		xctestrunFilePath, _ := arguments.String("--xctestrun-file-path")

		rawTestlog, rawTestlogErr := arguments.String("--log-output")

		if rawTestlogErr == nil {
			var writer *os.File = os.Stdout
			if rawTestlog != "-" {
				file, err := os.Create(rawTestlog)
				exitIfError("Cannot open file "+rawTestlog, err)
				writer = file
			}
			defer writer.Close()
			var listener = testmanagerd.NewTestListener(writer, writer, os.TempDir())

			testResults, err := testmanagerd.StartXCTestWithConfig(context.TODO(), xctestrunFilePath, device, listener)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Info("Failed running Xctest")
			}

			log.Info(fmt.Printf("%+v", testResults))
		} else {
			var listener = testmanagerd.NewTestListener(io.Discard, io.Discard, os.TempDir())
			_, err := testmanagerd.StartXCTestWithConfig(context.TODO(), xctestrunFilePath, device, listener)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Info("Failed running Xctest")
			}
		}
		return
	}

	if runWdaCommand(device, arguments) {
		return
	}

	b, _ = arguments.Bool("ax")
	if b {
		startAx(device, arguments)
		return
	}

	b, _ = arguments.Bool("resetax")
	if b {
		resetAx(device)
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

	b, _ = arguments.Bool("batteryregistry")
	if b {
		printBatteryRegistry(device)
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
		containerBundleId, _ := arguments.String("--app")
		var afcService *afc.Connection
		if containerBundleId == "" {
			afcService, err = afc.New(device)
		} else {
			afcService, err = afc.NewContainer(device, containerBundleId)
		}
		exitIfError("fsync: connect afc service failed", err)
		b, _ = arguments.Bool("rm")
		if b {
			path, _ := arguments.String("--path")
			isRecursive, _ := arguments.Bool("--r")
			if isRecursive {
				err = afcService.RemoveAll(path)
			} else {
				err = afcService.Remove(path)
			}
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

	b, _ = arguments.Bool("diskspace")
	if b {
		afcService, err := afc.New(device)
		exitIfError("connect afc service failed", err)
		info, err := afcService.GetSpaceInfo()
		if err != nil {
			exitIfError("get device info push failed", err)
		}
		fmt.Printf("      Model: %s\n", info.Model)
		fmt.Printf("  BlockSize: %d\n", info.BlockSize/8)
		fmt.Printf("  FreeSpace: %s\n", ios.ByteCountDecimal(int64(info.FreeBytes)))
		fmt.Printf("  UsedSpace: %s\n", ios.ByteCountDecimal(int64(info.TotalBytes-info.FreeBytes)))
		fmt.Printf(" TotalSpace: %s\n", ios.ByteCountDecimal(int64(info.TotalBytes)))
		return
	}

	b, _ = arguments.Bool("batterycheck")
	if b {
		printBatteryDiagnostics(device)
		return
	}

	if tunnelCommand {
		startCommand, _ := arguments.Bool("start")
		useUserspaceNetworking, _ := arguments.Bool("--userspace")
		if startCommand && !useUserspaceNetworking {
			err := ios.CheckRoot()
			if err != nil {
				exitIfError("If --userspace is not set, we need sudo or an admin shell on Windows", err)
			}
		}
		if useUserspaceNetworking {
			log.Info("Using userspace networking")
		}
		stopagent, _ := arguments.Bool("stopagent")
		listCommand, _ := arguments.Bool("ls")
		if startCommand {
			pairRecordsPath, _ := arguments.String("--pair-record-path")
			if len(pairRecordsPath) == 0 {
				pairRecordsPath = "."
			}
			if strings.ToLower(pairRecordsPath) == "default" {
				pairRecordsPath = "/var/db/lockdown/RemotePairing/user_501"
			}
			startTunnel(context.TODO(), pairRecordsPath, tunnelInfoPort, useUserspaceNetworking)
		} else if listCommand {
			tunnels, err := tunnel.ListRunningTunnels(tunnelInfoHost, tunnelInfoPort)
			if err != nil {
				exitIfError("failed to get tunnel infos", err)
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(tunnels)
		}
		if stopagent {
			err := tunnel.CloseAgent()
			if err != nil {
				exitIfError("failed to close agent", err)
			}
			return
		}
	}

	b, _ = arguments.Bool("devmode")
	if b {
		enable, _ := arguments.Bool("enable")
		get, _ := arguments.Bool("get")
		enablePostRestart, _ := arguments.Bool("--enable-post-restart")
		if enable {
			err := amfi.EnableDeveloperMode(device, enablePostRestart)
			exitIfError("Failed enabling developer mode", err)
		}

		if get {
			devModeEnabled, _ := imagemounter.IsDevModeEnabled(device)
			fmt.Printf("Developer mode enabled: %v\n", devModeEnabled)
		}

		return
	}
}

func printSysmontapStats(device ios.DeviceEntry) {
	const xcodeDefaultSamplingRate = 10
	sysmon, err := instruments.NewSysmontapService(device, xcodeDefaultSamplingRate)
	if err != nil {
		exitIfError("systemMonitor creation error", err)
	}
	defer sysmon.Close()

	cpuUsageChannel := sysmon.ReceiveCPUUsage()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	log.Info("starting to monitor CPU usage... Press CTRL+C to stop.")

	for {
		select {
		case cpuUsageMsg, ok := <-cpuUsageChannel:
			if !ok {
				log.Info("CPU usage channel closed.")
				return
			}
			log.WithFields(log.Fields{
				"cpu_count":      cpuUsageMsg.CPUCount,
				"enabled_cpus":   cpuUsageMsg.EnabledCPUs,
				"end_time":       cpuUsageMsg.EndMachAbsTime,
				"cpu_total_load": cpuUsageMsg.SystemCPUUsage.CPU_TotalLoad,
			}).Info("received CPU usage data")

		case <-c:
			log.Info("shutting down sysmontap")
			return
		}
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

		path, _ := arguments.String("--path")

		auto, _ := arguments.Bool("auto")
		if auto {
			basedir, _ := arguments.String("--basedir")
			if basedir == "" {
				basedir = "./devimages"
			}

			var err error
			path, err = imagemounter.DownloadImageFor(device, basedir)
			if err != nil {
				log.WithFields(log.Fields{"basedir": basedir, "udid": device.Properties.SerialNumber, "err": err}).
					Error("failed downloading image")
				return false
			}

			log.WithFields(log.Fields{"basedir": basedir, "udid": device.Properties.SerialNumber}).Info("success downloaded image")
		}

		mount, _ := arguments.Bool("mount")
		if mount || auto {
			err := imagemounter.MountImage(device, path)
			if err != nil {
				log.WithFields(log.Fields{"image": path, "udid": device.Properties.SerialNumber, "err": err}).
					Error("error mounting image")
				return true
			}
			log.WithFields(log.Fields{"image": path, "udid": device.Properties.SerialNumber}).Info("success mounting image")
		}

		unmount, _ := arguments.Bool("unmount")
		if unmount {
			err := imagemounter.UnmountImage(device)
			if err != nil {
				log.WithFields(log.Fields{"udid": device.Properties.SerialNumber, "err": err}).
					Error("error unmounting image")
				return true
			}
			log.WithFields(log.Fields{"udid": device.Properties.SerialNumber}).Info("success unmounting image")
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
		wdaenv := splitKeyValuePairs(arguments["--env"].([]string), "=")

		if bundleID == "" && testbundleID == "" && xctestconfig == "" {
			log.Info("no bundle ids specified, falling back to defaults")
			bundleID, testbundleID, xctestconfig = "com.facebook.WebDriverAgentRunner.xctrunner", "com.facebook.WebDriverAgentRunner.xctrunner", "WebDriverAgentRunner.xctest"
		}
		if bundleID == "" || testbundleID == "" || xctestconfig == "" {
			log.WithFields(log.Fields{"bundleid": bundleID, "testbundleid": testbundleID, "xctestconfig": xctestconfig}).Error("please specify either NONE of bundleid, testbundleid and xctestconfig or ALL of them. At least one was empty.")
			return true
		}
		log.WithFields(log.Fields{"bundleid": bundleID, "testbundleid": testbundleID, "xctestconfig": xctestconfig}).Info("Running wda")

		rawTestlog, rawTestlogErr := arguments.String("--log-output")

		var writer io.Writer

		if rawTestlogErr == nil {
			writerCloser := os.Stdout
			writer = writerCloser
			if rawTestlog != "-" {
				file, err := os.Create(rawTestlog)
				exitIfError("Cannot open file "+rawTestlog, err)
				writer = file
			}
			defer writerCloser.Close()
		} else {
			writer = io.Discard
		}

		errorChannel := make(chan error)
		defer close(errorChannel)
		ctx, stopWda := context.WithCancel(context.Background())
		go func() {
			_, err := testmanagerd.RunTestWithConfig(ctx, testmanagerd.TestConfig{
				BundleId:           bundleID,
				TestRunnerBundleId: testbundleID,
				XctestConfigName:   xctestconfig,
				Env:                wdaenv,
				Args:               wdaargs,
				Device:             device,
				Listener:           testmanagerd.NewTestListener(writer, writer, os.TempDir()),
			})
			if err != nil {
				errorChannel <- err
			}
			stopWda()
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-errorChannel:
			log.WithError(err).Error("Failed running WDA")
			stopWda()
			os.Exit(1)
		case <-ctx.Done():
			log.Error("WDA process ended unexpectedly")
			os.Exit(1)
		case signal := <-c:
			log.Infof("os signal:%d received, closing...", signal)
			stopWda()
		}
		log.Info("Done Closing")
	}
	return b
}

func instrumentsCommand(device ios.DeviceEntry, arguments docopt.Opts) bool {
	b, _ := arguments.Bool("instruments")
	if b {
		listenerFunc, closeFunc, err := instruments.ListenAppStateNotifications(device)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			for {
				notification, err := listenerFunc()
				if err != nil {
					log.Error(err)
					return
				}
				s, _ := json.Marshal(notification)
				println(string(s))
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		err = closeFunc()
		if err != nil {
			log.Warnf("timeout during close %v", err)
		}
	}
	return b
}

func toArgs(argsIn []string) []interface{} {
	args := []interface{}{}
	for _, arg := range argsIn {
		args = append(args, arg)
	}
	return args
}

func toEnvs(envsIn []string) map[string]interface{} {
	env := map[string]interface{}{}

	for _, entrystring := range envsIn {
		entry := strings.Split(entrystring, "=")
		key := entry[0]
		value := entry[1]
		env[key] = value
	}

	return env
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
	conn, err := imagemounter.NewImageMounter(device)
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

func assistiveTouch(device ios.DeviceEntry, operation string, force bool) {
	var enable bool

	if !force {
		version, err := ios.GetProductVersion(device)
		exitIfError("failed getting device product version", err)

		if version.LessThan(ios.IOS11()) {
			log.Errorf("iOS Version 11.0+ required to manipulate AssistiveTouch.  iOS version: %s detected. Use --force to override.", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetAssistiveTouch(device)
	if err != nil {
		if force && (operation == "enable" || operation == "disable") {
			log.WithFields(log.Fields{"error": err}).Warn("Failed getting current AssistiveTouch status. Continuing anyway.")
		} else {
			exitIfError("failed getting current AssistiveTouch status", err)
		}
	}

	switch {
	case operation == "enable":
		enable = true
	case operation == "disable":
		enable = false
	case operation == "toggle":
		enable = !wasEnabled
	default: // get
		enable = wasEnabled
	}
	if operation != "get" && (force || wasEnabled != enable) {
		err = ios.SetAssistiveTouch(device, enable)
		exitIfError("failed setting AssistiveTouch", err)
	}
	if operation == "get" {
		if JSONdisabled {
			fmt.Printf("%t\n", enable)
		} else {
			fmt.Println(convertToJSONString(map[string]bool{"AssistiveTouchEnabled": enable}))
		}
	}
}

func voiceOver(device ios.DeviceEntry, operation string, force bool) {
	var enable bool

	if !force {
		version, err := ios.GetProductVersion(device)
		exitIfError("failed getting device product version", err)

		if version.LessThan(ios.IOS11()) {
			log.Errorf("iOS Version 11.0+ required to manipulate VoiceOver.  iOS version: %s detected. Use --force to override.", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetVoiceOver(device)

	if err != nil {
		if force && (operation == "enable" || operation == "disable") {
			log.WithFields(log.Fields{"error": err}).Warn("Failed getting current VoiceOver status. Continuing anyway.")
		} else {
			exitIfError("failed getting current VoiceOver status", err)
		}
	}

	switch {
	case operation == "enable":
		enable = true
	case operation == "disable":
		enable = false
	case operation == "toggle":
		enable = !wasEnabled
	default: // get
		enable = wasEnabled
	}
	if operation != "get" && (force || wasEnabled != enable) {
		err = ios.SetVoiceOver(device, enable)
		exitIfError("failed setting VoiceOver", err)
	}
	if operation == "get" {
		if JSONdisabled {
			fmt.Printf("%t\n", enable)
		} else {
			fmt.Println(convertToJSONString(map[string]bool{"VoiceOverTouchEnabled": enable}))
		}
	}
}

func zoomTouch(device ios.DeviceEntry, operation string, force bool) {
	var enable bool

	if !force {
		version, err := ios.GetProductVersion(device)
		exitIfError("failed getting device product version", err)

		if version.LessThan(ios.IOS11()) {
			log.Errorf("iOS Version 11.0+ required to manipulate VoiceOver.  iOS version: %s detected. Use --force to override.", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetZoomTouch(device)

	if err != nil {
		if force && (operation == "enable" || operation == "disable") {
			log.WithFields(log.Fields{"error": err}).Warn("Failed getting current VoiceOver status. Continuing anyway.")
		} else {
			exitIfError("failed getting current VoiceOver status", err)
		}
	}

	switch {
	case operation == "enable":
		enable = true
	case operation == "disable":
		enable = false
	case operation == "toggle":
		enable = !wasEnabled
	default: // get
		enable = wasEnabled
	}
	if operation != "get" && (force || wasEnabled != enable) {
		err = ios.SetZoomTouch(device, enable)
		exitIfError("failed setting VoiceOver", err)
	}
	if operation == "get" {
		if JSONdisabled {
			fmt.Printf("%t\n", enable)
		} else {
			fmt.Println(convertToJSONString(map[string]bool{"ZoomTouchEnabled": enable}))
		}
	}
}

func timeFormat(device ios.DeviceEntry, operation string, force bool) {
	var enable bool

	if !force {
		version, err := ios.GetProductVersion(device)
		exitIfError("failed getting device product version", err)

		if version.LessThan(ios.IOS11()) {
			log.Errorf("iOS Version 11.0+ required to manipulate Time Format.  iOS version: %s detected. Use --force to override.", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetUses24HourClock(device)

	if err != nil {
		if force && (operation == "24h" || operation == "12h") {
			log.WithFields(log.Fields{"error": err}).Warn("Failed getting current TimeFormat value. Continuing anyway.")
		} else {
			exitIfError("failed getting current TimeFormat value", err)
		}
	}

	switch {
	case operation == "24h":
		enable = true
	case operation == "12h":
		enable = false
	case operation == "toggle":
		enable = !wasEnabled
	default: // get
		enable = wasEnabled
	}
	if operation != "get" && (force || wasEnabled != enable) {
		err = ios.SetUses24HourClock(device, enable)
		exitIfError("failed setting Time Format", err)
	}
	if operation == "get" {
		timeFormat := "24h"
		if !enable {
			timeFormat = "12h"
		}
		if JSONdisabled {
			fmt.Printf("%s\n", timeFormat)
		} else {
			fmt.Println(convertToJSONString(map[string]string{"TimeFormat": timeFormat}))
		}
	}
}

func startAx(device ios.DeviceEntry, arguments docopt.Opts) {
	go func() {
		conn, err := accessibility.New(device)
		exitIfError("failed starting ax", err)

		conn.SwitchToDevice()

		conn.EnableSelectionMode()

		size, _ := arguments.Float64("--font")
		if size != 0 {
			conn.UpdateAccessibilitySetting("DYNAMIC_TYPE", size)
		}

		for i := 0; i < 3; i++ {
			conn.GetElement()
			time.Sleep(time.Second)
		}
		/*	conn.GetElement()
			time.Sleep(time.Second)
			conn.TurnOff()*/
		// conn.GetElement()
		// conn.GetElement()

		exitIfError("ax failed", err)
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func resetAx(device ios.DeviceEntry) {
	conn, err := accessibility.NewWithoutEventChangeListeners(device)
	exitIfError("failed creating ax service", err)

	err = conn.ResetToDefaultAccessibilitySettings()
	exitIfError("failed resetting ax", err)
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
	filebytes, err := os.ReadFile(file)
	exitIfError("could not read profile-file", err)
	err = profileService.AddProfile(filebytes)
	exitIfError("failed adding profile", err)
	log.Info("profile installed, you have to accept it in the device settings")
}

func handleProfileAddSupervised(device ios.DeviceEntry, file string, p12file string, p12password string) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	filebytes, err := os.ReadFile(file)
	exitIfError("could not read profile-file", err)
	p12bytes, err := os.ReadFile(p12file)
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
	cl, err := forward.Forward(device, uint16(hostPort), uint16(targetPort))
	exitIfError("failed to forward port", err)
	defer stopForwarding(cl)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func stopForwarding(cl *forward.ConnListener) {
	err := cl.Close()
	if err != nil {
		exitIfError("failed to close forwarded port", err)
	}
}

func printDiagnostics(device ios.DeviceEntry) {
	log.Debug("print diagnostics")
	diagnosticsService, err := diagnostics.New(device)
	exitIfError("Starting diagnostics service failed with", err)

	values, err := diagnosticsService.AllValues()
	exitIfError("getting valued failed", err)

	fmt.Println(convertToJSONString(values))
}

func printBatteryDiagnostics(device ios.DeviceEntry) {
	battery, err := ios.GetBatteryDiagnostics(device)
	exitIfError("failed getting battery diagnostics", err)

	fmt.Println(convertToJSONString(battery))
}

func printBatteryRegistry(device ios.DeviceEntry) {
	conn, err := diagnostics.New(device)
	if err != nil {
		exitIfError("failed diagnostics service", err)
	}
	defer conn.Close()

	stats, err := conn.Battery()
	if err != nil {
		exitIfError("failed to get battery stats", err)
	}

	fmt.Println(convertToJSONString(stats))
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

func printInstalledApps(device ios.DeviceEntry, system bool, all bool, list bool, filesharing bool) {
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
	} else if filesharing {
		response, err = svc.BrowseFileSharingApps()
		appType = "filesharingapps"
	} else {
		response, err = svc.BrowseUserApps()
		appType = "user"
	}
	exitIfError("browsing "+appType+" apps failed", err)

	if list {
		for _, v := range response {
			fmt.Printf("%s %s %s\n", v.CFBundleIdentifier(), v.CFBundleName(), v.CFBundleShortVersionString())
		}
		return
	}
	if filesharing {
		for _, v := range response {
			if v.UIFileSharingEnabled() {
				fmt.Printf("%s %s %s\n", v.CFBundleIdentifier(), v.CFBundleName(), v.CFBundleShortVersionString())
			}
		}
		return
	}
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
	screenshotService, err := instruments.NewScreenshotService(device)
	exitIfError("Starting screenshot service failed", err)
	defer screenshotService.Close()

	imageBytes, err := screenshotService.TakeScreenshot()
	exitIfError("Taking screenshot failed", err)

	if outputPath == "" {
		timestamp := time.Now().Format("20060102150405")
		outputPath, err = filepath.Abs("./screenshot" + timestamp + ".png")
		exitIfError("getting filepath failed", err)
	}

	err = os.WriteFile(outputPath, imageBytes, 0o777)
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

func startLocationSimulation(service *instruments.LocationSimulationService, lat string, lon string) {
	latitude, err := strconv.ParseFloat(lat, 64)
	exitIfError("location simulation failed to parse lat", err)

	longitude, err := strconv.ParseFloat(lon, 64)
	exitIfError("location simulation failed to parse lon", err)

	err = service.StartSimulateLocation(latitude, longitude)
	exitIfError("location simulation failed to start with", err)

	defer stopLocationSimulation(service)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func stopLocationSimulation(service *instruments.LocationSimulationService) {
	err := service.StopSimulateLocation()
	if err != nil {
		exitIfError("location simulation failed to stop with", err)
	}
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
				applicationProcessList = append(applicationProcessList, processInfo)
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
	appInfoByExecutableName := make(map[string]installationproxy.AppInfo)

	if err != nil {
		log.Error("browsing installed apps failed. bundleID will not be included in output")
	} else {
		for _, app := range response {
			appInfoByExecutableName[app.CFBundleExecutable()] = app
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
	maxPidLength := len(fmt.Sprintf("%d", maxPid))

	fmt.Printf("%*s %-*s %s  %s\n", maxPidLength, "PID", maxNameLength, "NAME", "START_DATE         ", "BUNDLE_ID")
	for _, processInfo := range processes {
		bundleID := ""
		appInfo, exists := appInfoByExecutableName[processInfo.Name]
		if exists {
			bundleID = appInfo.CFBundleIdentifier()
		}
		fmt.Printf("%*d %-*s %s  %s\n", maxPidLength, processInfo.Pid, maxNameLength, processInfo.Name, processInfo.StartDate.Format("2006-01-02 15:04:05"), bundleID)
	}
}

func startListening() {
	go func() {
		for {
			deviceConn, err := ios.NewDeviceConnection(ios.GetUsbmuxdSocket())
			defer deviceConn.Close()
			if err != nil {
				log.Errorf("could not connect to %s with err %+v, will retry in 3 seconds...", ios.GetUsbmuxdSocket(), err)
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
	svc, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		log.Debugf("could not open instruments, probably dev image not mounted %v", err)
	}
	if err == nil {
		info, err := svc.NetworkInformation()
		if err != nil {
			log.Debugf("error getting networkinfo from instruments %v", err)
		} else {
			allValues["instruments:networkInformation"] = info
		}
		info, err = svc.HardwareInformation()
		if err != nil {
			log.Debugf("error getting hardwareinfo from instruments %v", err)
		} else {
			allValues["instruments:hardwareInformation"] = info
		}
	}

	fmt.Println(convertToJSONString(allValues))
}

func runSyslog(device ios.DeviceEntry, parse bool) {
	log.Debug("Run Syslog.")

	syslogConnection, err := syslog.New(device)
	exitIfError("Syslog connection failed", err)

	defer syslogConnection.Close()

	var logFormatter func(string) string
	if JSONdisabled {
		logFormatter = rawSyslog
	} else if parse {
		logFormatter = parsedJsonSyslog()
	} else {
		logFormatter = legacyJsonSyslog()
	}

	go func() {
		for {
			logMessage, err := syslogConnection.ReadLogMessage()
			if err != nil {
				exitIfError("failed reading syslog", err)
			}
			logMessage = strings.TrimSuffix(logMessage, "\x00")
			logMessage = strings.TrimSuffix(logMessage, "\x0A")

			fmt.Println(logFormatter(logMessage))
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func rawSyslog(log string) string {
	return log
}

func legacyJsonSyslog() func(log string) string {
	messageContainer := map[string]string{}

	return func(log string) string {
		messageContainer["msg"] = log
		return convertToJSONString(messageContainer)
	}
}

func parsedJsonSyslog() func(log string) string {
	parser := syslog.Parser()

	return func(log string) string {
		log_entry, err := parser(log)
		if err != nil {
			return convertToJSONString(map[string]string{"msg": log, "error": err.Error()})
		}

		return convertToJSONString(log_entry)
	}
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

func startTunnel(ctx context.Context, recordsPath string, tunnelInfoPort int, userspaceTUN bool) {
	pm, err := tunnel.NewPairRecordManager(recordsPath)
	exitIfError("could not creat pair record manager", err)
	tm := tunnel.NewTunnelManager(pm, userspaceTUN)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := tm.UpdateTunnels(ctx)
				if err != nil {
					log.WithError(err).Warn("failed to update tunnels")
				}
			}
		}
	}()

	go func() {
		err := tunnel.ServeTunnelInfo(tm, tunnelInfoPort)
		if err != nil {
			exitIfError("failed to start tunnel server", err)
		}
	}()
	log.Info("Tunnel server started")
	<-ctx.Done()
}

func deviceWithRsdProvider(device ios.DeviceEntry, udid string, address string, rsdPort int) ios.DeviceEntry {
	rsdService, err := ios.NewWithAddrPortDevice(address, rsdPort, device)
	exitIfError("could not connect to RSD", err)
	defer rsdService.Close()
	rsdProvider, err := rsdService.Handshake()
	device1, err := ios.GetDeviceWithAddress(udid, address, rsdProvider)
	device1.UserspaceTUN = device.UserspaceTUN
	device1.UserspaceTUNHost = device.UserspaceTUNHost
	device1.UserspaceTUNPort = device.UserspaceTUNPort
	exitIfError("error getting devicelist", err)

	return device1
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

func marshalJSON(data interface{}) ([]byte, error) {
	if prettyJSON {
		return json.MarshalIndent(data, "", "    ")
	} else {
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

func splitKeyValuePairs(envArgs []string, sep string) map[string]interface{} {
	env := make(map[string]interface{})
	for _, entrystring := range envArgs {
		entry := strings.Split(entrystring, sep)
		key := entry[0]
		value := entry[1]
		env[key] = value
	}
	return env
}
