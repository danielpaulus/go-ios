[![](https://dcbadge.vercel.app/api/server/Zr8J3bCdkv)](https://discord.gg/Zr8J3bCdkv)
[![NPM](https://nodei.co/npm/go-ios.png?mini=true)](https://npmjs.org/package/go-ios)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/daniel1paulus.svg?style=social&label=Follow%20%40daniel1paulus)](https://twitter.com/daniel1paulus)
[![NPM](https://img.shields.io/npm/dw/go-ios?style=flat-square)](https://npmjs.org/package/go-ios)

<img src="logo.png" width="256"/>

Welcome 👋

`npm install -g go-ios` can be used to get going. Run `ios --help` after the installation for details. 
For iOS 17+ devices you need to run `sudo ios tunnel start` for go ios to work. This will start a tunnel daemon. 
To make this work on Windows, download the latest wintun.dll from here `https://git.zx2c4.com/wintun` and copy it to `C:/Windows/system32`

The goal of this project is to provide a stable and production ready opensource solution to automate iOS device on Linux, Windows and Mac OS X. I am delighted to announce that a few companies including [headspin.io](https://www.headspin.io/) and [Sauce Labs](https://saucelabs.com/) will use or are using go-iOS. 

Follow my twitter for updates or check out my medium blog: https://daniel-paulus.medium.com/

If you are interested in using go-iOS please get in touch on LinkedIn, Twitter or the Github discussions above, I always love to hear what people are doing with it. 

If you miss something your Mac can do but go-iOS can't, just request a feature in the issues tab.
# New REST-API
Go-iOS is getting an experimental REST-API check it out [https://github.com/danielpaulus/go-ios/tree/main/restapi](https://github.com/danielpaulus/go-ios/tree/main/restapi) 

# Design principles:
1. Using golang to compile static, small and fast binaries for all platforms very easily. 
   
   *Build Manual*: Install golang and run `go build`
2. All output as JSON so you can easily use go-iOS from any other programming language
3. Everything is a module, you can use go-iOS in golang projects as a module dependency easily

# Features:
 Most notable:
 - Install apps zipped as ipa or unzipped from their .app folder `ios install --path=/path/to/app`
 - Run XCTests including WebdriverAgent on Linux, Windows and Mac
 - Start and Stop apps
 - Use a debug proxy to reverse engineer every tool Mac OSX has, so you can contrib to go-ios or build      your own
 - Pair devices without manual tapping on a popup
 - Install developer images automatically by running `ios image auto`
 - Set thermal states and network emulation on the device with the `ios devicestate` command

All features:

```
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

```
