# go-ios

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

Go-iOS was inspired by the wonderful https://www.libimobiledevice.org
It can do all of what libimobiledevice can do and in addition also run XCUITest, launch and kill apps and much more.
It is written in golang, so you can control iOS devices using fast CLI tools :-)

You can use it for free in any open source project. If you want use it in a closed source project, get in touch :-)

Highlights:

- run XCTests including WebdriverAgent on Linux, Windows and Mac
- start and stop apps
- Use a debug proxy to reverse engineer every tool Mac OSX has, so you can contrib to go-ios or build your own
- use Accessibility Inspector APIs

```
The commands work as following:
	The default output of all commands is JSON. Should you prefer human readable outout, specify the --nojson option with your command.
	By default, the first device found will be used for a command unless you specify a --udid=some_udid switch.
	Specify -v for debug logging and -t for dumping every message.

   ios listen [options]                                               Keeps a persistent connection open and notifies about newly connected or disconnected devices.
   ios list [options] [--details]                                     Prints a list of all connected device's udids. If --details is specified, it includes version, name and model of each device.
   ios info [options]                                                 Prints a dump of Lockdown getValues.
   ios image list [options]                                           List currently mounted developers images' signatures
   ios image mount [--path=<imagepath>] [options]                     Mount a image from <imagepath>
   ios image auto [--basedir=<where_dev_images_are_stored>] [options] Automatically download correct dev image from the internets and mount it. You can specify a dir where images should be cached. The default is the current dir. 
   ios syslog [options]                                               Prints a device's log output
   ios screenshot [options] [--output=<outfile>]                      Takes a screenshot and writes it to the current dir or to <outfile>
   ios devicename [options]                                           Prints the devicename
   ios date [options]                                                 Prints the device date
   ios lang [--setlocale=<locale>] [--setlang=<newlang>] [options]    Sets or gets the Device language
   ios diagnostics list [options]                                     List diagnostic infos
   ios pair [options]                                                 Pairs the device without a dialog for supervised devices
   ios ps [options]                                                   Dumps a list of running processes on the device
   ios forward [options] <hostPort> <targetPort>                      Similar to iproxy, forward a TCP connection to the device.
   ios dproxy [--binary]                                              Starts the reverse engineering proxy server. 
   >                                                                  It dumps every communication in plain text so it can be implemented easily. 
   >                                                                  Use "sudo launchctl unload -w /Library/Apple/System/Library/LaunchDaemons/com.apple.usbmuxd.plist"
   >                                                                  to stop usbmuxd and load to start it again should the proxy mess up things.
   >                                                                  The --binary flag will dump everything in raw binary without any decoding. 
   ios readpair                                                       Dump detailed information about the pairrecord for a device.
   ios install --path=<ipaOrAppFolder> [options]                      Specify a .app folder or an installable ipa file that will be installed.  
   ios pcap [options] [--pid=<processID>] [--process=<processName>]   Starts a pcap dump of network traffic, use --pid or --process to filter specific processes.
   ios apps [--system]                                                Retrieves a list of installed applications. --system prints out preinstalled system apps.
   ios launch <bundleID>                                              Launch app with the bundleID on the device. Get your bundle ID from the apps command.
   ios runtest <bundleID>                                             Run a XCUITest. 
   ios runwda [options]                                               Start WebDriverAgent.
   ios ax [options]                                                   Access accessibility inspector features. 
   ios debug [--stop-at-entry] <app_path>                             Start debug with lldb.
   ios reboot [options]                                               Reboot the given device
   ios -h | --help                                                    Prints this screen.
   ios --version | version [options]                                  Prints the version.
```
