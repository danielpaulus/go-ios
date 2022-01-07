[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/daniel1paulus.svg?style=social&label=Follow%20%40daniel1paulus)](https://twitter.com/daniel1paulus)
# Go-iOS
Welcome 👋

The goal of this project is to provide a stable and production ready opensource solution to automate iOS device on Linux, Windows and Mac OS X. I am delighted to announce that a few companies including [headspin.io](https://www.headspin.io/) will use or are using go-iOS. 
If you are interested in controlling iOS devices, please also check out https://github.com/nanoscopic/controlfloor written by my good friend @nanoscopic 

Follow my twitter for updates or check out my medium blog: https://daniel-paulus.medium.com/

If you are interested in using go-iOS please get in touch on LinkedIn, Twitter or the Github discussions above, I always love to hear what people are doing with it. 

If you miss something your Mac can do but go-iOS can't, just request a feature in the issues tab.

# Design principles:
1. Using golang to compile static, small and fast binaries for all platforms very easily. 
   
   *Build Manual*: Install golang and run `go build`
2. All output as JSON so you can easily use go-iOS form any other programming language
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
   ios ps [options]                                                   Dumps a list of running processes on the device
   ios forward [options] <hostPort> <targetPort>                      Similar to iproxy, forward a TCP connection to the device.
   ios dproxy [--binary]                                              Starts the reverse engineering proxy server. 
   >                                                                  It dumps every communication in plain text so it can be implemented easily. 
   >                                                                  Use "sudo launchctl unload -w /Library/Apple/System/Library/LaunchDaemons/com.apple.usbmuxd.plist"
   >                                                                  to stop usbmuxd and load to start it again should the proxy mess up things.
   >                                                                  The --binary flag will dump everything in raw binary without any decoding. 
   ios readpair                                                       Dump detailed information about the pairrecord for a device.                                              Starts a pcap dump of network traffic
   ios install --path=<ipaOrAppFolder> [options]                      Specify a .app folder or an installable ipa file that will be installed.  
   ios pcap [options] [--pid=<processID>] [--process=<processName>]   Starts a pcap dump of network traffic, use --pid or --process to filter specific processes.
   ios apps [--system]                                                Retrieves a list of installed applications. --system prints out preinstalled system apps.
   ios launch <bundleID>                                              Launch app with the bundleID on the device. Get your bundle ID from the apps command.
   ios kill <bundleID> [options]                                      Kill app with the bundleID on the device.
   ios runtest <bundleID>                                             Run a XCUITest. 
   ios runwda [--bundleid=<bundleid>] [--testrunnerbundleid=<testbundleid>] [--xctestconfig=<xctestconfig>] [--arg=<a>]... [--env=<e>]...[options]  runs WebDriverAgents
   >                                                                  specify runtime args and env vars like --env ENV_1=something --env ENV_2=else  and --arg ARG1 --arg ARG2
   ios ax [options]                                                   Access accessibility inspector features. 
   ios debug [--stop-at-entry] <app_path>                             Start debug with lldb
   ios reboot [options]                                               Reboot the given device
   ios -h | --help                                                    Prints this screen.
   ios --version | version [options]                                  Prints the version

```
