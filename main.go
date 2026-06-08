package main

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/pkcs12"
	"golang.org/x/term"

	"github.com/danielpaulus/go-ios/internal/clihelp"
	"github.com/danielpaulus/go-ios/ios/debugproxy"
	"github.com/danielpaulus/go-ios/ios/tunnel"

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
	"github.com/danielpaulus/go-ios/ios/ostrace"
	"github.com/danielpaulus/go-ios/ios/springboard"
	syslog "github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/docopt/docopt-go"
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
	helpCatalog, err := clihelp.Load()
	exitIfError("failed loading help definitions", err)
	if handled, exitCode := helpCatalog.WriteHelp(os.Args[1:], version, os.Stdout, os.Stderr); handled {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return
	}

	usage := fmt.Sprintf(`go-ios %s

Usage:
  ios --version | version [options]
  ios -h | --help
  ios activate [options]
  ios apps [--system] [--all] [--list] [--filesharing] [options]
  ios assistivetouch (enable | disable | toggle | get) [--force] [options]
  ios ax [--font=<fontSize>] [options]
  ios ax audit [options]
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
  ios devmode (enable | get | reveal) [--enable-post-restart] [options]
  ios diagnostics list [options]
  ios diskspace [options]
  ios dproxy [--binary] [--mode=<all(default)|usbmuxd|utun>] [--iface=<iface>] [options]
  ios erase [--force] [options]
  ios file ls [--app=<bundleID> | --app-group=<groupID> | --crash | --temp] [--path=<path>] [options]
  ios file pull [--app=<bundleID> | --app-group=<groupID> | --crash | --temp] --remote=<remotePath> --local=<localPath> [options]
  ios file push [--app=<bundleID> | --app-group=<groupID> | --crash | --temp] --local=<localPath> --remote=<remotePath> [options]
  ios forward [options] [<hostPort> <targetPort>] [--port=<mapping>]...
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
  ios lockdown get [<key>] [--domain=<domain>] [options]
  ios memlimitoff (--process=<processName>) [options]
  ios mobilegestalt <key>... [--plist] [options]
  ios pair [--p12file=<orgid>] [--password=<p12password>] [options]
  ios pcap [options] [--pid=<processID>] [--process=<processName>]
  ios prepare [--skip-all] [--skip=<option>]... [--certfile=<cert_file_path>] [--orgname=<org_name>] [--p12password=<p12password>] [--locale=<locale>] [--lang=<lang>] [options]
  ios prepare cloudconfig [options]
  ios prepare create-cert
  ios prepare printskip
  ios wifi [--ssid=<ssid>] [--password=<password>] [--enc-type=<encType>] [--remove] [options]
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
  ios sign certificate appstoreconnect --asc-key-id=<keyid> --asc-issuer-id=<issuerid> --asc-private-key=<p8file> [--p12-output=<p12file>] [--p12password=<password>] [--revoke-existing] [options]
  ios sign provision appstoreconnect --bundleid=<bundleid> --asc-key-id=<keyid> --asc-issuer-id=<issuerid> --asc-private-key=<p8file> --profile-output=<mobileprovision> [--p12-output=<p12file>] [--certificate-id=<id>] [--revoke-existing] [--p12password=<password>] [--bundle-name=<name>] [--profile-name=<name>] [--device-name=<name>] [options]
  ios sign app --path=<ipaOrAppFolder> --p12file=<p12file> --profile=<mobileprovision> [--p12password=<password>] [--output=<signedPath>] [--bundleid=<bundleid>] [--install] [options]
  ios setlocation [options] [--lat=<lat>] [--lon=<lon>]
  ios setlocationgpx [options] [--gpxfilepath=<gpxfilepath>]
  ios shutdown [options]
  ios set-wallpaper <imagePath> [--screen=<screen>] [--p12file=<orgid>] [--password=<p12password>] [options]
  ios get-wallpaper [--output=<outfile>] [options]
  ios get-icon-layout [--output=<outfile>] [options]
  ios set-icon-layout <layoutFile> [options]
  ios syslog [--parse] [options]
  ios ostrace [--pid=<processID>] [--process=<processName>] [--follow] [--level=<levels>] [--subsystem=<sub>] [--match=<str>] [--exclude=<str>] [options]
  ios sysmontap [options]
  ios timeformat (24h | 12h | toggle | get) [--force] [options]
  ios tunnel ls [options]
  ios tunnel stop [options]
  ios tunnel refresh [options]
  ios tunnel start [options] [--pair-record-path=<pairrecordpath>] [--userspace]
  ios tunnel stopagent
  ios ui install (wda | devicekit) --p12file=<p12file> --profile=<mobileprovision> [--p12password=<password>] [--path=<ipaOrZipOrApp>] [--output=<signedPath>] [--bundleid=<bundleid>] [options]
  ios ui download [(wda | devicekit | all)] [--output=<dir>] [options]
  ios ui status [--driver=<driver>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui api [--driver=<driver>] [--method=<method>] [--http-path=<path>] [--body=<json>] [--body-file=<file>] [--rpc-method=<method>] [--params=<json>] [--params-file=<file>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui raw [--driver=<driver>] [--method=<method>] [--http-path=<path>] [--body=<json>] [--body-file=<file>] [--rpc-method=<method>] [--params=<json>] [--params-file=<file>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui tap --x=<x> --y=<y> [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui swipe --from-x=<x> --from-y=<y> --to-x=<x> --to-y=<y> [--duration=<seconds>] [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui longpress --x=<x> --y=<y> [--duration=<seconds>] [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui type --text=<text> [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui button <button> [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui screenshot [--output=<outfile>] [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui source [--output=<outfile>] [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui size [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui orientation get [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui orientation set <orientation> [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui app launch <bundleID> [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui app terminate <bundleID> [--driver=<driver>] [--session-id=<sessionid>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios ui app foreground [--driver=<driver>] [--devicekit-url=<url>] [options]
  ios ui stream (mjpeg | h264) [--fps=<fps>] [--quality=<quality>] [--scale=<scale>] [--bitrate=<bitrate>] [--driver=<driver>] [--wda-url=<url>] [--devicekit-url=<url>] [options]
  ios uninstall <bundleID> [options]
  ios webinspector list [--timeout=<seconds>] [options]
  ios webinspector launch <url> [--bundle-id=<bundleID>] [--timeout=<seconds>] [options]
  ios webinspector eval <pageID> <expression> [--timeout=<seconds>] [--console-enable] [options]
  ios webinspector js-shell [<url>] [--bundle-id=<bundleID>] [--open-safari] [--timeout=<seconds>] [--console-enable] [options]
  ios webinspector cdp [--host=<host>] [--port=<port>] [options]
  ios voiceover (enable | disable | toggle | get) [--force] [options]
  ios zoom (enable | disable | toggle | get) [--force] [options]

Options:
  -v --verbose              Enable Debug Logging.
  -t --trace                Enable Trace Logging (dump every message).
  --nojson                  Disable JSON output
  --pretty                  Pretty-print JSON command output
  -h --help                 Show this screen.
  --udid=<udid>             UDID of the device. Can also be set via GO_IOS_UDID environment variable.
  --tunnel-info-port=<port> When go-ios is used to manage tunnels for iOS 17+,
                            it exposes them on an HTTP-API (default port: 28100)
  --tunnel-info-host=<host> Host the tunnel-info HTTP-API binds to and is queried on.
                            Defaults to 127.0.0.1, or the GO_IOS_AGENT_HOST environment variable
                            if set. Use 0.0.0.0 to reach the API from another host or container.
  --address=<ipv6addrr>     Address of the device on the interface.
                            This parameter is optional and can be set if a tunnel created by MacOS needs to be used.
                            To get this value run "log stream --debug --info --predicate 'eventMessage LIKE "*Tunnel established*" OR eventMessage LIKE "*for server port*"'",
                            connect a device and open Xcode
  --rsd-port=<port>         Port of remote service discovery on the device through the tunnel
                            This parameter is similar to '--address' and can be obtained by the same log filter
  --proxyurl=<url>          Set this if you want go-ios to use a http proxy for outgoing requests,
                            like for downloading images or contacting Apple during device activation.
                            A simple format like: "http://PROXY_LOGIN:PROXY_PASS@proxyIp:proxyPort" works.
                            Otherwise use the HTTP_PROXY system env var.
  --userspace-port=<port>   Optional. Set this if you run a command supplying rsd-port and address and your device is using userspace tunnel
  --asc-key-id=<keyid>      App Store Connect API key id. Can also be set via GO_IOS_ASC_KEY_ID.
  --asc-issuer-id=<issuerid>
                            App Store Connect API issuer id. Can also be set via GO_IOS_ASC_ISSUER_ID.
  --asc-private-key=<p8file>
                            App Store Connect API .p8 private key path. Can also be set via GO_IOS_ASC_PRIVATE_KEY.
  --revoke-existing         Revoke every existing iOS Development certificate before creating a new one. Apple allows only one current development certificate, so set this to keep provisioning repeatable (use only on a dedicated signing account, e.g. CI).
  --certificate-id=<id>     Provision a profile against an existing App Store Connect certificate (by resource id) instead of creating one. No certificate is created/revoked and no P12 is written; reuse a pre-provisioned signing identity.
  --p12file=<p12file>       P12 identity file path.
  --profile=<mobileprovision>
                            Provisioning profile path for app signing.
  --driver=<driver>         UI automation backend: devicekit, wda, or auto. Defaults to devicekit.
  --wda-url=<url>           WebDriverAgent base URL. Defaults to http://127.0.0.1:8100 or GO_IOS_WDA_URL.
  --devicekit-url=<url>     DeviceKit base URL. Defaults to http://127.0.0.1:12004 or GO_IOS_DEVICEKIT_URL.

The commands work as following:
  The default output of all commands is JSON. Should you prefer human readable outout, specify the --nojson option with your command.
  By default, the first device found will be used for a command unless you specify a --udid=some_udid switch.
  Specify -v for debug logging and -t for dumping every message.

    ios --version | version [options]                     Prints the version
    ios -h | --help                                       Prints this screen.
    ios activate [options]                                Activate a device

    ios apps [--system] [--all] [--list] [--filesharing]  Retrieves a list of installed applications.
                                                          --system prints out preinstalled system apps.
                                                          --all prints all apps, including system, user, and hidden apps.
                                                          --list only prints bundle ID, bundle name and version number.
                                                          --filesharing only prints apps which enable documents sharing.

    ios assistivetouch (enable | disable | toggle | get) [--force] [options]
                                                          Enables, disables, toggles, or returns the state of the "AssistiveTouch" software home-screen button.
                                                          iOS 11+ only (Use --force to try on older versions).

    ios ax [--font=<fontSize>] [options]          Access accessibility inspector features.
    ios ax audit [options]                        Run the accessibility audit on the focused app and print the issues as JSON.
                                                  Each issue includes its type, the element's label and on-screen rect.
    ios batterycheck [options]                    Prints battery info.
    ios batteryregistry [options]                 Prints battery registry stats like Temperature, Voltage.
    ios crash cp <srcpattern> <target> [options]  Copy "file pattern" to the target dir. Ex.: 'ios crash cp "*" "./crashes"'

    ios crash ls [<pattern>] [options]            Run "ios crash ls" to get all crashreports in a list,
                                                  or use a pattern like 'ios crash ls "*ips*"' to filter

    ios crash rm <cwd> <pattern> [options]        Remove file pattern from dir. Ex.: 'ios crash rm "." "*"' to delete everything
    ios date [options]                            Prints the device date
    ios debug [--stop-at-entry] <app_path>        Start debug with lldb
    ios devicename [options]                      Prints the devicename

    ios devicestate enable <profileTypeId> <profileId> [options]  Enables a profile with ids (use the list command to see options).
                                                                  It will only stay active until the process is terminated.
                                                                  Ex. "ios devicestate enable SlowNetworkCondition SlowNetwork3GGood"

    ios devicestate list [options]                                Prints a list of all supported device conditions, like slow network, gpu etc.

    ios devmode (enable | get | reveal) [--enable-post-restart] [options]
                                                                  Enable developer mode on the device, check if it is enabled,
                                                                  or reveal the Developer Mode toggle in Settings.
                                                                  Can also completely finalize developer mode setup after device is restarted.

    ios diagnostics list [options]                                List diagnostic infos
    ios diskspace [options]                                       Prints disk space info.

    ios dproxy [--binary] [--mode=<all(default)|usbmuxd|utun>] [--iface=<iface>] [options]
                                                                  Starts the reverse engineering proxy server.
                                                                  It dumps every communication in plain text so it can be implemented easily.
                                                                  Use "sudo launchctl unload -w /Library/Apple/System/Library/LaunchDaemons/com.apple.usbmuxd.plist"
                                                                  to stop usbmuxd and load to start it again should the proxy mess up things.
                                                                  The --binary flag will dump everything in raw binary without any decoding.

    ios erase [--force] [options]                                 Erase the device. It will prompt you to input y+Enter unless --force is specified.

    ios file ls [--app=<bundleID> | --app-group=<groupID> | --crash | --temp] [--path=<path>] [options]
                                                                  List files using RemoteXPC (iOS 17+). Requires tunnel.
                                                                  Use --app for app container, --app-group for app group,
                                                                  --crash for crash logs, or --temp for temporary files.

    ios file pull [--app=<bundleID> | --app-group=<groupID> | --crash | --temp] --remote=<remotePath> --local=<localPath> [options]
                                                                  Download file using RemoteXPC (iOS 17+). Requires tunnel.

    ios file push [--app=<bundleID> | --app-group=<groupID> | --crash | --temp] --local=<localPath> --remote=<remotePath> [options]
                                                                  Upload file using RemoteXPC (iOS 17+).
                                                                  Requires tunnel. Preserves source file permissions.

    ios forward [options] [<hostPort> <targetPort>] [--port=<mapping>]...
                                                                  Forward TCP connections to device.
                                                                  Use --port for multiple ports: --port=8100:8100 --port=9191:9191

    ios fsync [--app=bundleId] [options] (pull | push) --srcPath=<srcPath> --dstPath=<dstPath>
                                                                  Pull or Push file from srcPath to dstPath.

    ios fsync [--app=bundleId] [options] (rm [--r] | tree | mkdir) --path=<targetPath>
                                                                  Remove | treeview | mkdir in target path.
                                                                  --r used alongside rm will recursively remove all files and directories from target path.

    ios httpproxy <host> <port> [<user>] [<pass>] --p12file=<orgid> [--password=<p12password>]
                                                                  Set global http proxy on supervised device.
                                                                  Use the password argument or set the environment variable 'P12_PASSWORD'
                                                                  Specify proxy password either as argument or using the environment var: PROXY_PASSWORD
                                                                  Use p12 file and password for silent installation on supervised devices.

    ios httpproxy remove [options]                                Removes the global http proxy config. Only works with http proxies set by go-ios!

    ios image auto [--basedir=<where_dev_images_are_stored>] [options]  Automatically download correct dev image from the internets and mount it.
                                                                        You can specify a dir where images should be cached.
                                                                        The default is the current dir.

    ios image list [options]                        List currently mounted developers images' signatures
    
    ios image mount [--path=<imagepath>] [options]  Mount a image from <imagepath>
                                                    For iOS 17+ (personalized developer disk images),
                                                    <imagepath> must point to the "Restore" directory inside the developer disk

    ios image unmount [options]                     Unmount developer disk image
    ios info [display | lockdown] [options]         Prints a dump of device information from the given source.
    ios install --path=<ipaOrAppFolder> [options]   Specify a .app folder or an installable ipa file that will be installed.
    ios instruments notifications [options]         Listen to application state notifications

    ios ip [options]                                Uses the live pcap iOS packet capture to wait until it finds one that contains the IP address of the device.
                                                    It relies on the MAC address of the WiFi adapter to know which is the right IP.
                                                    You have to disable the "automatic wifi address"-privacy feature of the device for this to work.
                                                    If you wanna speed it up, open apple maps or similar to force network traffic.
                                                    Ex.: "ios launch com.apple.Maps"

    ios kill (<bundleID> | --pid=<processID> | --process=<processName>) [options]
                                                                       Kill app with the specified bundleID, process id, or process name on the device.

    ios lang [--setlocale=<locale>] [--setlang=<newlang>] [options]    Sets or gets the Device language. ios lang will print the current language and locale,
                                                                       as well as a list of all supported langs and locales.

    ios launch <bundleID> [--wait] [--kill-existing] [--arg=<a>]... [--env=<e>]... [options]
                                                                       Launch app with the bundleID on the device. Get your bundle ID from the apps command.
                                                                       --wait keeps the connection open if you want logs.

    ios list [options] [--details]                                     Prints a list of all connected device's udids.
                                                                       If --details is specified, it includes version, name and model of each device.

    ios listen [options]                                               Keeps a persistent connection open and notifies about newly connected or disconnected devices.

    ios lockdown get [<key>] [--domain=<domain>] [options]             Query lockdown values. Without arguments returns all values. Specify a key to get a specific value.
                                                                       Use --domain to query from a specific domain (e.g., com.apple.disk_usage, com.apple.PurpleBuddy).
                                                                       Ex.: "ios lockdown get DeviceName", "ios lockdown get --domain=com.apple.PurpleBuddy"

    ios memlimitoff (--process=<processName>) [options]                Waives memory limit set by iOS (For instance a Broadcast Extension limit is 50 MB).

    ios mobilegestalt <key>... [--plist] [options]                     Lets you query mobilegestalt keys.
                                                                       Standard output is json but if desired you can get it in plist format by adding the --plist param.
                                                                       Ex.: "ios mobilegestalt MainScreenCanvasSizes ArtworkTraits --plist"

    ios pair [--p12file=<orgid>] [--password=<p12password>] [options]  Pairs the device. If the device is supervised, specify the path to the p12 file
                                                                       to pair without a trust dialog. Specify the password either with the argument or
                                                                       by setting the environment variable 'P12_PASSWORD'

    ios pcap [options] [--pid=<processID>] [--process=<processName>]   Starts a pcap dump of network traffic, use --pid or --process to filter specific processes.

    ios prepare [--skip-all] [--skip=<option>]... [--certfile=<cert_file_path>] [--orgname=<org_name>] [--p12password=<p12password>] [--locale] [--lang] [options]
                                                                       Prepare a device. Use skip-all to skip everything multiple --skip args to skip only a subset.
                                                                       You can use 'ios prepare printskip' to get a list of all options to skip.
                                                                       Use certfile and orgname if you want to supervise the device.
                                                                       The certfile can be a DER file, PEM file, or P12 file. For P12 files,
                                                                       specify the password with --p12password or P12_PASSWORD env var.
                                                                       If you need certificates to supervise,
                                                                       run 'ios prepare create-cert' and go-ios will generate one you can use.
                                                                       --locale and --lang are optional, the default is en_US and en.
                                                                       Run 'ios lang' to see a list of all supported locales and languages.

    ios prepare cloudconfig                                            Print the cloud configuration of the device as JSON.

    ios prepare create-cert                                            A nice util to generate a certificate you can use for supervising devices.
                                                                       Make sure you rename and store it in a safe place.

    ios prepare printskip                                              Print all options you can skip.

	ios wifi [--ssid=<ssid>] [--password=<password>] [--enc-type=<encType>] [--remove]
																		Installs a wifi profile on the device forcing a connection to the provided WiFi network
																		If --remove is specified, the wifi profile of the provided ssid will be removed.

    ios profile add <profileFile> [--p12file=<orgid>] [--password=<p12password>]
                                                                       Install profile file on the device.
                                                                       If supervised set p12file and password or the environment variable 'P12_PASSWORD'

    ios profile list                  List the profiles on the device
    ios profile remove <profileName>  Remove the profileName from the device

    ios ps [--apps] [options]         Dumps a list of running processes on the device.
                                      Use --nojson for a human-readable listing including BundleID when available (not included with JSON output).
                                      --apps limits output to processes flagged by iOS as "isApplication".
                                      This greatly-filtered list should at least include user-installed software.
                                      Additional packages will also be displayed depending on the version of iOS.

    ios readpair                      Dump detailed information about the pairrecord for a device.
    ios reboot [options]              Reboot the given device
    ios resetax [options]             Reset accessibility settings to defaults.
    ios resetlocation [options]       Resets the location of the device to the actual one
    ios rsd ls [options]              List RSD services and their port.

    ios runtest [--bundle-id=<bundleid>] [--test-runner-bundle-id=<testbundleid>] [--xctest-config=<xctestconfig>] [--log-output=<file>] [--xctest] [--test-to-run=<tests>]... [--test-to-skip=<tests>]... [--env=<e>]... [options]
                                                                    Run a XCUITest.
                                                                    If you provide only bundle-id go-ios will try to dynamically create test-runner-bundle-id and xctest-config.
                                                                    If you provide '-' as log output, it prints resuts to stdout.
                                                                    To be able to filter for tests to run or skip, use one argument per test selector.
                                                                    Ex.: runtest --test-to-run=(TestTarget.)TestClass/testMethod (the value for 'TestTarget' is optional)
                                                                    The method name can also be omitted and in this case all tests of the specified class are run

    ios runwda [--bundleid=<bundleid>] [--testrunnerbundleid=<testbundleid>] [--xctestconfig=<xctestconfig>] [--log-output=<file>] [--arg=<a>]... [--env=<e>]...[options]
                                                                    Runs WebDriverAgents
                                                                    Specify runtime args and env vars like --env ENV_1=something --env ENV_2=else  and --arg ARG1 --arg ARG2

    ios runxctest [--xctestrun-file-path=<xctestrunFilePath>]  [--log-output=<file>] [options]
                                                                    Run a XCTest.
                                                                    The --xctestrun-file-path specifies the path to the .xctestrun file to configure the test execution.
                                                                    If you provide '-' as log output, it prints resuts to stdout.

    ios screenshot [options] [--output=<outfile>] [--stream] [--port=<port>]
                                                                    Takes a screenshot and writes it to the current dir or to <outfile>
                                                                    If --stream is supplied it starts an mjpeg server at 0.0.0.0:3333.
                                                                    Use --port to set another port.

    ios sign provision appstoreconnect --bundleid=<bundleid> --asc-key-id=<keyid> --asc-issuer-id=<issuerid> --asc-private-key=<p8file> --p12-output=<p12file> --profile-output=<mobileprovision>
                                                                    Creates an iOS development signing certificate, P12, and provisioning profile through App Store Connect.
                                                                    This command does not sign an app. Pass --revoke-existing to revoke a leftover go-ios certificate first (repeatable provisioning).

    ios sign app --path=<ipaOrAppFolder> --p12file=<p12file> --profile=<mobileprovision>
                                                                    Resigns the IPA or .app with go-codesign using local signing files,
                                                                    and optionally installs the signed result with --install.
                                                                    For WDA or DeviceKit artifacts, run "ios ui download" first and pass the downloaded path with --path.

    ios ui install (wda | devicekit) --p12file=<p12file> --profile=<mobileprovision>
                                                                    Downloads the default DeviceKit or WDA artifact from deviceboxhq.com unless --path is provided,
                                                                    signs it with local signing files, and installs it on the selected device.
                                                                    Run "ios ui download" to pre-download artifacts, or pass --path to use your own local build.

    ios ui download [(wda | devicekit | all)] [--output=<dir>]       Downloads default WDA and/or DeviceKit artifacts from deviceboxhq.com,
                                                                    extracts zip artifacts, and prints JSON describing the files.
                                                                    Use the printed artifactPath or appPath with "ios ui install --path" or "ios sign app --path".

    ios ui status [--driver=<driver>]                                Checks the configured UI automation backend. Defaults to DeviceKit.

    ios ui api [--driver=<driver>]                                    Calls a backend-specific API directly.
                                                                    For WDA, pass --http-path=<path>, optionally --method=<method> and --body=<json>.
                                                                    For DeviceKit, pass --rpc-method=<method>, optionally --params=<json>.
                                                                    Use --driver=auto to probe DeviceKit first, then WDA.
                                                                    "raw" is accepted as an alias for "api".

    ios ui tap --x=<x> --y=<y>                                       Taps at screen coordinates.
    ios ui swipe --from-x=<x> --from-y=<y> --to-x=<x> --to-y=<y>     Swipes between screen coordinates.
    ios ui longpress --x=<x> --y=<y> [--duration=<seconds>]          Long-presses at screen coordinates.
    ios ui type --text=<text>                                        Types text.
    ios ui button <button>                                           Presses a button. DeviceKit supports more buttons; WDA supports home.
    ios ui screenshot [--output=<outfile>]                           Saves a screenshot, or writes PNG bytes to stdout.
    ios ui source [--output=<outfile>]                               Dumps the UI hierarchy.
    ios ui size                                                      Prints screen or window size information.
    ios ui orientation (get | set <orientation>)                     Gets or sets orientation.
    ios ui app (launch | terminate) <bundleID>                       Launches or terminates an app.
    ios ui app foreground                                            Prints the foreground app through DeviceKit.
    ios ui stream (mjpeg | h264)                                     Streams video to stdout. H264 requires DeviceKit; WDA supports MJPEG.

    ios setlocation [options] [--lat=<lat>] [--lon=<lon>]           Updates the location of the device to the provided by latitude and longitude coordinates.
                                                                    Ex.: setlocation --lat=40.730610 --lon=-73.935242

    ios setlocationgpx [options] [--gpxfilepath=<gpxfilepath>]      Updates the location of the device based on the data in a GPX file.
                                                                    Ex.: setlocationgpx --gpxfilepath=/home/username/location.gpx

    ios shutdown [options]                                          Shuts down the device

    ios set-wallpaper <imagePath> [--screen=<screen>] [--p12file=<orgid>] [--password=<p12password>] [options]
                                                                    Set the device wallpaper from a JPEG/PNG file. --screen is lock|home|both (default home).
                                                                    Requires supervision: pass the supervisor identity .p12 with --p12file and the password
                                                                    via --password or P12_PASSWORD env var. Note: on iOS 16+ both lock and home screens are
                                                                    linked as a pair, so the device sets both regardless of --screen. The flag is preserved
                                                                    for older-iOS / forward-compat. Apple's own cfgutil exhibits the same behavior.

    ios get-wallpaper [--output=<outfile>] [options]                Save the home screen wallpaper as PNG. Default output is wallpaper.png.
                                                                    Does not require supervision. Lock screen wallpaper is not exposed by iOS.
                                                                    Note: this RPC may EOF on iOS 18 (see pymobiledevice3 #1450).

    ios get-icon-layout [--output=<outfile>] [options]              Save the home screen icon layout as JSON (default stdout). Round-trippable: feed the
                                                                    file back to set-icon-layout to restore. Note: the iOS 14+ "Edit Pages" per-page
                                                                    hidden bit is not exposed by springboardservices, so a fetched layout will not include
                                                                    hidden pages.

    ios set-icon-layout <layoutFile> [options]                      Push a previously-saved icon layout JSON file back to the device.
                                                                    iOS requires every installed app to occupy a slot. Per cfgutil docs: "unexpected
                                                                    behavior may occur if the given layout does not contain every icon on the device".
                                                                    Missing apps are re-paginated, not hidden.

    ios syslog [--parse] [options]                                  Prints a device's log output, Use --parse to parse the fields from the log
    ios ostrace [--pid=<processID>] [--process=<processName>] [--follow] [--level=<levels>] [--subsystem=<sub>] [--match=<str>] [--exclude=<str>]
                                                                     Stream structured syslog via os_trace_relay. Note: streaming logs
                                                                     places significant CPU load on the device.
                                                                       --follow             Keep running and reconnect when the process exits or restarts.
                                                                                            When used with --process, polls until the process appears.
                                                                     Device-side filters (reduce USB traffic):
                                                                       --pid=<pid>           Only stream logs from this process ID
                                                                       --process=<name>      Resolve process name to PID, then filter device-side
                                                                       --level=<levels>      Filter by OS log type (comma-separated): default,info,debug,error,fault
                                                                     Client-side filters (applied after receiving, does not reduce USB traffic):
                                                                       --subsystem=<sub>     Only show entries matching this subsystem (substring match)
                                                                       --match=<str>         Only show entries where the message contains this string
                                                                       --exclude=<str>       Hide entries where the message contains this string
    ios sysmontap                                                   Get system stats like MEM, CPU

    ios timeformat (24h | 12h | toggle | get) [--force] [options]   Sets, or returns the state of the "time format".
                                                                    iOS 11+ only (Use --force to try on older versions).

    ios tunnel ls                                                   List currently started tunnels.
                                                                    Use --enabletun to activate using TUN devices rather than user space network.
                                                                    Requires sudo/admin shells.

    ios tunnel stop --udid=<udid>                                   Stop the tunnel for one device without stopping the tunnel agent.

    ios tunnel refresh --udid=<udid>                                Stop the tunnel for one device and wait until the agent recreates it.

    ios tunnel start [options] [--pair-record-path=<pairrecordpath>] [--enabletun]
                                                                    Creates a tunnel connection to the device.
                                                                    If the device was not paired with the host yet, device pairing will also be executed.
                                                                    On systems with System Integrity Protection enabled the argument '--pair-record-path=default'
                                                                    can be used to point to /var/db/lockdown/RemotePairing/user_501.
                                                                    WARNING: macOS 26 (Tahoe) and newer block that path for third-party binaries via TCC
                                                                    ('operation not permitted'). On those systems do NOT use '=default'; pass a stable
                                                                    writable directory instead (e.g. --pair-record-path=/Users/Shared/go-ios) and go-ios
                                                                    will manage its own tunnel identity. See https://github.com/danielpaulus/go-ios/issues/710
                                                                    If nothing is specified, the current dir is used for the pair record.
                                                                    Pass --udid=<udid> to restrict the agent to a single device (isolated
                                                                    per-device tunnel agent); run one per device on its own --tunnel-info-port.
                                                                    This command needs to be executed with admin privileges.
                                                                    (On MacOS the process 'remoted' must be paused before starting a tunnel,
                                                                    is possible 'sudo pkill -SIGSTOP remoted', and 'sudo pkill -SIGCONT remoted' to resume)

    ios webinspector list [--timeout=<seconds>] [options]           List inspectable Safari/WebView pages.

    ios webinspector launch <url> [--bundle-id=<bundleID>] [--timeout=<seconds>] [options]
                                                                    Launch Safari or another app and navigate by Remote Automation.

    ios webinspector eval <pageID> <expression> [--timeout=<seconds>] [--console-enable] [options]
                                                                    Evaluate JavaScript in an inspectable page.

    ios webinspector js-shell [<url>] [--bundle-id=<bundleID>] [--open-safari] [--timeout=<seconds>] [--console-enable] [options]
                                                                    Start an interactive JavaScript shell for an inspectable page.

    ios webinspector cdp [--host=<host>] [--port=<port>] [options]  Start a Chrome DevTools Protocol bridge.

    ios voiceover (enable | disable | toggle | get) [--force] [options] Enables, disables, toggles, or returns the state of the "VoiceOver" software home-screen button.
                                                                    iOS 11+ only (Use --force to try on older versions).

    ios zoom (enable | disable | toggle | get) [--force] [options]  Enables, disables, toggles, or returns the state of the "ZoomTouch" software home-screen button.
                                                                    iOS 11+ only (Use --force to try on older versions).

  `, version)
	arguments, err := docopt.ParseDoc(usage)
	exitIfError("failed parsing args", err)
	configureCLI(arguments)
	if dispatchCommand(commandContext{Args: arguments}, preProxyCommands) {
		return
	}

	proxyUrl, _ := arguments.String("--proxyurl")
	exitIfError("could not parse proxy url", ios.UseHttpProxy(proxyUrl))

	if dispatchCommand(commandContext{Args: arguments}, globalCommands) {
		return
	}

	tunnelInfo := tunnelInfoConfigFromArgs(arguments)
	device := resolveDevice(arguments, tunnelInfo)

	if dispatchCommand(commandContext{Args: arguments, Device: device}, deviceCommands) {
		return
	}

	if dispatchTunnelCommand(tunnelCommandContext{
		Args:           arguments,
		TunnelInfoHost: tunnelInfo.Host,
		TunnelInfoPort: tunnelInfo.Port,
	}) {
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

	slog.Info("starting to monitor CPU usage... Press CTRL+C to stop.")

	for {
		select {
		case cpuUsageMsg, ok := <-cpuUsageChannel:
			if !ok {
				slog.Info("CPU usage channel closed.")
				return
			}
			slog.Info("received CPU usage data",
				"cpu_count", cpuUsageMsg.CPUCount,
				"enabled_cpus", cpuUsageMsg.EnabledCPUs,
				"end_time", cpuUsageMsg.EndMachAbsTime,
				"cpu_total_load", cpuUsageMsg.SystemCPUUsage.CPU_TotalLoad,
			)

		case <-c:
			slog.Info("shutting down sysmontap")
			return
		}
	}
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
			fmt.Println(string(b))
		}
		return
	}
	exitIfError("failed listing device states", err)
	if enable {
		pType, profile, err := instruments.VerifyProfileAndType(profileTypes, profileTypeId, profileId)
		exitIfError("invalid arguments", err)
		slog.Info("Enabling profile.. (this can take a while for ThermalConditions)")
		err = control.Enable(pType, profile)
		exitIfError("could not enable profile", err)
		slog.Info(fmt.Sprintf("Profile %s - %s is active! waiting for SIGTERM..", profileTypeId, profileId))
		c := make(chan os.Signal, syscall.SIGTERM)
		signal.Notify(c, os.Interrupt)
		<-c
		slog.Info(fmt.Sprintf("Disabling profiletype %s", profileTypeId))
		err = control.Disable(pType)
		exitIfError("could not disable profile", err)
		slog.Info("ok")
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
	fmt.Println(buffer.String())
}

func listMountedImages(device ios.DeviceEntry) {
	conn, err := imagemounter.NewImageMounter(device)
	exitIfError("failed connecting to image mounter", err)
	signatures, err := conn.ListImages()
	exitIfError("failed getting image list", err)
	if len(signatures) == 0 {
		slog.Info("none")
		return
	}
	for _, sig := range signatures {
		slog.Info("image signature", "signature", fmt.Sprintf("%x", sig))
	}
}

func installApp(device ios.DeviceEntry, path string) {
	slog.Info("installing", "appPath", path, "device", device.Properties.SerialNumber)
	conn, err := zipconduit.New(device)
	exitIfError("failed connecting to zipconduit, dev image installed?", err)
	err = conn.SendFile(path)
	exitIfError("failed writing", err)
}

func uninstallApp(device ios.DeviceEntry, bundleId string) {
	slog.Info("uninstalling", "appPath", bundleId, "device", device.Properties.SerialNumber)
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
		slog.Debug("Language should be changed waiting for Springboard to reboot", "from", lang.Language, "to", language)
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
			slog.Error("iOS Version 11.0+ required to manipulate AssistiveTouch. Use --force to override.", "version", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetAssistiveTouch(device)
	if err != nil {
		if force && (operation == "enable" || operation == "disable") {
			slog.Warn("Failed getting current AssistiveTouch status. Continuing anyway.", "error", err)
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
			slog.Error("iOS Version 11.0+ required to manipulate VoiceOver. Use --force to override.", "version", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetVoiceOver(device)

	if err != nil {
		if force && (operation == "enable" || operation == "disable") {
			slog.Warn("Failed getting current VoiceOver status. Continuing anyway.", "error", err)
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
			slog.Error("iOS Version 11.0+ required to manipulate ZoomTouch. Use --force to override.", "version", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetZoomTouch(device)

	if err != nil {
		if force && (operation == "enable" || operation == "disable") {
			slog.Warn("Failed getting current ZoomTouch status. Continuing anyway.", "error", err)
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
			slog.Error("iOS Version 11.0+ required to manipulate Time Format. Use --force to override.", "version", version)
			os.Exit(1)
		}
	}

	wasEnabled, err := ios.GetUses24HourClock(device)

	if err != nil {
		if force && (operation == "24h" || operation == "12h") {
			slog.Warn("Failed getting current TimeFormat value. Continuing anyway.", "error", err)
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
		conn, err := accessibility.NewWithoutEventChangeListeners(device)
		exitIfError("failed starting ax", err)

		conn.SwitchToDevice()

		conn.EnableSelectionMode()

		size, _ := arguments.Float64("--font")
		if size != 0 {
			conn.UpdateAccessibilitySetting("DYNAMIC_TYPE", size)
		}

		for i := 0; i < 3; i++ {
			conn.GetElement(context.Background())
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

// axSilentNotifier is a no-op AccessibilityInspectorNotifier for one-shot AX
// commands that don't stream device events.
type axSilentNotifier struct{}

func (axSilentNotifier) HostAppStateChanged(accessibility.Notification)               {}
func (axSilentNotifier) HostInspectorNotificationReceived(accessibility.Notification) {}

func runAxAudit(device ios.DeviceEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	conn, err := accessibility.New(ctx, device, axSilentNotifier{})
	exitIfError("failed starting ax", err)
	defer conn.Close()

	issues, err := conn.RunAudit(ctx)
	exitIfError("ax audit failed", err)
	fmt.Println(convertToJSONString(issues))
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
				slog.Error("Recovered a panic", "panic", r)
				proxy.Close()
				debug.PrintStack()
				os.Exit(1)
				return
			}
		}()
		err := proxy.Launch(device, binaryMode)
		slog.Info("DebugProxy Terminated abnormally", "error", err)
		os.Exit(0)
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	slog.Info("Shutting down debugproxy")
	proxy.Close()
}

func handleProfileRemove(device ios.DeviceEntry, identifier string) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	err = profileService.RemoveProfile(identifier)
	exitIfError("failed adding profile", err)
	slog.Info(fmt.Sprintf("profile '%s' removed", identifier))
}

func handleProfileAdd(device ios.DeviceEntry, file string) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	filebytes, err := os.ReadFile(file)
	exitIfError("could not read profile-file", err)
	err = profileService.AddProfile(filebytes)
	exitIfError("failed adding profile", err)
	slog.Info("profile installed, you have to accept it in the device settings")
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
	slog.Info("profile installed")
}

func handleProfileList(device ios.DeviceEntry) {
	profileService, err := mcinstall.New(device)
	exitIfError("Starting mcInstall failed with", err)
	list, err := profileService.HandleList()
	exitIfError("failed getting profile list", err)
	fmt.Println(convertToJSONString(list))
}

func handleSetWallpaper(device ios.DeviceEntry, imagePath, screen, p12file, p12password string) {
	if p12file == "" {
		logFatal("--p12file is required (set-wallpaper needs a supervisor identity)")
	}
	screenValue, err := mcinstall.ParseWallpaperScreen(screen)
	exitIfError("invalid --screen", err)
	imageBytes, err := os.ReadFile(imagePath)
	exitIfError("could not read image file", err)
	p12bytes, err := os.ReadFile(p12file)
	exitIfError("could not read p12 file", err)

	conn, err := mcinstall.New(device)
	exitIfError("starting mcinstall failed", err)
	defer conn.Close()

	err = conn.SetWallpaperSupervised(imageBytes, screenValue, p12bytes, p12password)
	exitIfError("failed setting wallpaper", err)
	fmt.Println(convertToJSONString("ok"))
}

func handleGetWallpaper(device ios.DeviceEntry, output string) {
	client, err := springboard.NewClient(device)
	exitIfError("could not connect to springboardservices", err)
	defer client.Close()
	pngBytes, err := client.GetHomeScreenWallpaperPNG()
	exitIfError("could not fetch wallpaper", err)
	err = os.WriteFile(output, pngBytes, 0o644)
	exitIfError("could not write wallpaper file", err)
	fmt.Println(convertToJSONString(map[string]any{"path": output, "bytes": len(pngBytes)}))
}

func handleGetIconLayout(device ios.DeviceEntry, output string) {
	client, err := springboard.NewClient(device)
	exitIfError("could not connect to springboardservices", err)
	defer client.Close()
	state, err := client.GetIconLayout("2")
	exitIfError("could not fetch icon layout", err)
	jsonBytes, err := json.MarshalIndent(state, "", "  ")
	exitIfError("could not marshal layout", err)
	if output == "" {
		fmt.Println(string(jsonBytes))
		return
	}
	err = os.WriteFile(output, jsonBytes, 0o644)
	exitIfError("could not write layout file", err)
	fmt.Println(convertToJSONString(map[string]any{"path": output, "bytes": len(jsonBytes)}))
}

func handleSetIconLayout(device ios.DeviceEntry, layoutFile string) {
	raw, err := os.ReadFile(layoutFile)
	exitIfError("could not read layout file", err)
	var state any
	err = json.Unmarshal(raw, &state)
	exitIfError("layout file is not valid JSON", err)
	client, err := springboard.NewClient(device)
	exitIfError("could not connect to springboardservices", err)
	defer client.Close()
	err = client.SetIconLayout(state)
	exitIfError("could not set icon layout", err)
	fmt.Println(convertToJSONString("ok"))
}

func startForwarding(device ios.DeviceEntry, hostPort uint16, targetPort uint16) {
	cl, err := forward.Forward(device, hostPort, targetPort)
	exitIfError("failed to forward port", err)
	defer stopForwarding(cl)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

func stopForwarding(cl *forward.ConnListener) {
	err := cl.Close()
	if err != nil {
		exitIfError("failed to close forwarded port", err)
	}
}

func startMultiForwarding(device ios.DeviceEntry, mappings []string) {
	var listeners []*forward.ConnListener

	closeAllListeners := func() {
		for _, l := range listeners {
			l.Close()
		}
	}

	for _, mapping := range mappings {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			closeAllListeners()
			exitIfError("invalid mapping format", fmt.Errorf("expected hostPort:targetPort, got %s", mapping))
		}
		hostPort, err := strconv.ParseUint(parts[0], 10, 16)
		if err != nil {
			closeAllListeners()
			exitIfError("invalid host port", err)
		}
		targetPort, err := strconv.ParseUint(parts[1], 10, 16)
		if err != nil {
			closeAllListeners()
			exitIfError("invalid target port", err)
		}

		cl, err := forward.Forward(device, uint16(hostPort), uint16(targetPort))
		if err != nil {
			closeAllListeners()
			exitIfError(fmt.Sprintf("failed to forward %d:%d", hostPort, targetPort), err)
		}
		listeners = append(listeners, cl)
		slog.Info(fmt.Sprintf("Forwarding %d -> %d", hostPort, targetPort))
	}

	slog.Info(fmt.Sprintf("Started %d port forwards", len(listeners)))

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	// Close all listeners
	closeAllListeners()
}

func printDiagnostics(device ios.DeviceEntry) {
	slog.Debug("print diagnostics")
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
		slog.Info("apps", "apps", response)
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
		slog.Info("File saved successfully", "outputPath", outputPath)
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
	ConnectionType string
}

func outputDetailedList(deviceList ios.DeviceList) {
	result := make([]detailsEntry, len(deviceList.DeviceList))
	for i, device := range deviceList.DeviceList {
		udid := device.Properties.SerialNumber
		allValues, err := ios.GetValues(device)
		exitIfError("failed getting values", err)
		result[i] = detailsEntry{
			Udid:           udid,
			ProductName:    allValues.Value.ProductName,
			ProductType:    allValues.Value.ProductType,
			ProductVersion: allValues.Value.ProductVersion,
			ConnectionType: device.ConnectionTypeLabel(),
		}
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
		fmt.Printf("%s  %s  %s  %s  %s\n", udid, allValues.Value.ProductName, allValues.Value.ProductType, allValues.Value.ProductVersion, device.ConnectionTypeLabel())
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
		slog.Error("browsing installed apps failed. bundleID will not be included in output")
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
				slog.Error("could not connect, will retry in 3 seconds...", "socket", ios.GetUsbmuxdSocket(), "error", err)
				time.Sleep(time.Second * 3)
				continue
			}
			muxConnection := ios.NewUsbMuxConnection(deviceConn)

			attachedReceiver, err := muxConnection.Listen()
			if err != nil {
				slog.Error("Failed issuing Listen command, will retry in 3 seconds", "error", err)
				deviceConn.Close()
				time.Sleep(time.Second * 3)
				continue
			}
			for {
				msg, err := attachedReceiver()
				if err != nil {
					slog.Error("Stopped listening because of error")
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
		slog.Debug("could not open instruments, probably dev image not mounted", "error", err)
	}
	if err == nil {
		info, err := svc.NetworkInformation()
		if err != nil {
			slog.Debug("error getting networkinfo from instruments", "error", err)
		} else {
			allValues["instruments:networkInformation"] = info
		}
		info, err = svc.HardwareInformation()
		if err != nil {
			slog.Debug("error getting hardwareinfo from instruments", "error", err)
		} else {
			allValues["instruments:hardwareInformation"] = info
		}
	}

	fmt.Println(convertToJSONString(allValues))
}

func runSyslog(device ios.DeviceEntry, parse bool) {
	slog.Debug("Run Syslog.")

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

func runOsTrace(device ios.DeviceEntry, pid int, processName string, messageFilter uint16, streamFlags uint32, clientFilter ostrace.ClientFilter, follow bool) {
	slog.Debug("Run OsTrace.")
	// Note: streaming log messages places significant CPU load on the device.

	formatEntry := func(e ostrace.LogEntry) string {
		return convertToJSONString(e)
	}
	if JSONdisabled {
		formatEntry = formatEntryPlain
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-sigCh
		cancel()
	}()

	sleepOrCancel := func(d time.Duration) bool {
		select {
		case <-ctx.Done():
			return true
		case <-time.After(d):
			return false
		}
	}

	for {
		if processName != "" && pid == -1 {
			resolved := false
			waitingLogged := false
			for !resolved {
				service, err := instruments.NewDeviceInfoService(device)
				if err != nil {
					if follow {
						slog.Warn("Failed to open deviceInfoService, retrying...", "error", err)
						if sleepOrCancel(2 * time.Second) {
							return
						}
						continue
					}
					exitIfError("failed opening deviceInfoService for resolving process name", err)
				}
				proc, err := service.ProcessByName(processName)
				service.Close()
				if err != nil {
					if follow {
						if !waitingLogged {
							slog.Info(fmt.Sprintf("Waiting for process %q to appear...", processName))
							waitingLogged = true
						}
						if sleepOrCancel(2 * time.Second) {
							return
						}
						continue
					}
					exitIfError("process not found", err)
				}
				pid = int(proc.Pid)
				slog.Info(fmt.Sprintf("Resolved process %q to PID %d", processName, pid))
				resolved = true
			}
		}

		conn, err := ostrace.New(device, pid, messageFilter, streamFlags)
		if err != nil {
			if follow {
				slog.Warn("os_trace connection failed, retrying...", "error", err)
				if processName != "" {
					pid = -1
				}
				if sleepOrCancel(2 * time.Second) {
					return
				}
				continue
			}
			exitIfError("os_trace connection failed", err)
		}

		done := make(chan error, 1)
		reconnect := make(chan struct{}, 1)
		monitorCtx, monitorCancel := context.WithCancel(ctx)

		if follow && pid > 0 {
			go func(targetPid int) {
				ticker := time.NewTicker(3 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-monitorCtx.Done():
						return
					case <-ticker.C:
						if !isProcessAlive(device, uint64(targetPid)) {
							slog.Info(fmt.Sprintf("Process PID %d no longer running", targetPid))
							select {
							case reconnect <- struct{}{}:
							default:
							}
							return
						}
					}
				}
			}(pid)
		}

		go func() {
			for {
				entry, err := conn.ReadFilteredEntry(clientFilter)
				if err != nil {
					done <- err
					return
				}
				fmt.Println(formatEntry(entry))
			}
		}()

		select {
		case <-ctx.Done():
			monitorCancel()
			conn.Close()
			return
		case <-reconnect:
			monitorCancel()
			conn.Close()
			if processName != "" {
				slog.Info("os_trace stream ended, reconnecting...")
				pid = -1
			} else {
				slog.Info(fmt.Sprintf("Process PID %d terminated; stopping follow.", pid))
				return
			}
		case err := <-done:
			monitorCancel()
			conn.Close()
			if follow {
				slog.Warn("os_trace stream ended, reconnecting...", "error", err)
				if processName != "" {
					pid = -1
				}
				if sleepOrCancel(2 * time.Second) {
					return
				}
				continue
			}
			exitIfError("failed reading os_trace entry", err)
		}
	}
}

func colorForLevel(level ostrace.LogLevel) string {
	switch level {
	case ostrace.LogLevelInfo:
		return "\033[36m" // cyan
	case ostrace.LogLevelDebug:
		return "\033[90m" // bright black (gray)
	case ostrace.LogLevelError:
		return "\033[31m" // red
	case ostrace.LogLevelFault:
		return "\033[1;31m" // bold red
	default:
		return "" // no color
	}
}

func isTerminal(fd int) bool {
	// golang.org/x/term is cross-platform (Linux/macOS/BSD/Windows), unlike a
	// raw unix.TCGETS ioctl which only builds on Linux.
	return term.IsTerminal(fd)
}

func formatEntryPlain(entry ostrace.LogEntry) string {
	ts := entry.Timestamp.Format("2006-01-02T15:04:05.000Z07:00")
	useColor := isTerminal(int(os.Stdout.Fd()))
	dim, reset, color := "", "", ""
	if useColor {
		dim = "\033[90m"
		reset = "\033[0m"
		color = colorForLevel(entry.Level)
	}
	if entry.Label != nil {
		return fmt.Sprintf("%s%s%s  %sPID:%-5d%s  %s<%-7s>%s  %s[%s:%s]%s  %s%s%s",
			dim, ts, reset,
			dim, entry.PID, reset,
			color, entry.LevelName, reset,
			dim, entry.Label.Subsystem, entry.Label.Category, reset,
			color, entry.Message, reset)
	}
	return fmt.Sprintf("%s%s%s  %sPID:%-5d%s  %s<%-7s>%s  %s%s%s",
		dim, ts, reset,
		dim, entry.PID, reset,
		color, entry.LevelName, reset,
		color, entry.Message, reset)
}

func isProcessAlive(device ios.DeviceEntry, pid uint64) bool {
	service, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		slog.Warn("isProcessAlive: failed to connect to device", "error", err)
		return false
	}
	defer service.Close()
	procs, err := service.ProcessList()
	if err != nil {
		slog.Warn("isProcessAlive: failed to list processes", "error", err)
		return false
	}
	for _, p := range procs {
		if p.Pid == pid {
			return true
		}
	}
	return false
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
		slog.Info(fmt.Sprintf("Successfully paired %s", device.Properties.SerialNumber))
		return
	}
	p12, err := os.ReadFile(orgIdentityP12File)
	exitIfError("Invalid file:"+orgIdentityP12File, err)
	err = ios.PairSupervised(device, p12, p12Password)
	exitIfError("Pairing failed", err)
	slog.Info(fmt.Sprintf("Successfully paired %s", device.Properties.SerialNumber))
}

func startTunnel(ctx context.Context, recordsPath string, tunnelInfoHost string, tunnelInfoPort int, userspaceTUN bool, udid string) {
	// Optional profiling endpoint: set GO_IOS_PPROF=host:port (e.g. 127.0.0.1:6060)
	// to expose net/http/pprof (CPU, heap, block and mutex profiles) on the agent.
	if addr := os.Getenv("GO_IOS_PPROF"); addr != "" {
		runtime.SetBlockProfileRate(1)     // record goroutine blocking events
		runtime.SetMutexProfileFraction(1) // record mutex contention
		go func() {
			slog.Info("pprof listening", "addr", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				slog.Warn("pprof server stopped", "error", err)
			}
		}()
	}
	pm, err := tunnel.NewPairRecordManager(recordsPath)
	exitIfError("could not creat pair record manager", err)
	if udid != "" {
		slog.Info("restricting tunnel agent to a single device", "udid", udid, "tunnelInfoPort", tunnelInfoPort)
	}
	// Always derive userspace listener ports from THIS agent's tunnel-info port
	// (not the global default), so several agents on different ports — e.g. a
	// general agent plus per-device agents — never collide. An empty udid means
	// "manage all devices".
	tm := tunnel.NewTunnelManagerForDevice(pm, userspaceTUN, udid, tunnelInfoPort)

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
					slog.Warn("failed to update tunnels", "error", err)
				}
			}
		}
	}()

	go func() {
		err := tunnel.ServeTunnelInfo(tm, tunnelInfoHost, tunnelInfoPort)
		if err != nil {
			exitIfError("failed to start tunnel server", err)
		}
	}()
	slog.Info("Tunnel server started")
	<-ctx.Done()
}

func deviceWithRsdProvider(device ios.DeviceEntry, udid string, address string, rsdPort int) ios.DeviceEntry {
	rsdService, err := ios.NewWithAddrPortDevice(address, rsdPort, device)
	exitIfError(fmt.Sprintf("could not connect to RSD, host %s, port %d", address, rsdPort), err)
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
		logFatal(msg, "err", err)
	}
}

// logFatal logs at error level and exits with status 1 (slog has no Fatal).
func logFatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
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

// extractDERCertificate extracts a raw DER certificate from various input formats:
// - Raw DER bytes (passed through as-is)
// - PEM encoded certificate
// - PEM with metadata (e.g., from OpenSSL "Bag Attributes" output)
// - PKCS12/P12 file (if password is provided)
func extractDERCertificate(certBytes []byte, p12Password string) ([]byte, error) {
	if der, ok := tryParseDER(certBytes); ok {
		return der, nil
	}
	if der, ok := tryParsePEM(certBytes); ok {
		return der, nil
	}
	if der, ok := tryParsePEMWithMetadata(certBytes); ok {
		return der, nil
	}
	if der, ok := tryParsePKCS12(certBytes, p12Password); ok {
		return der, nil
	}
	return nil, fmt.Errorf("unable to parse certificate from file: not a valid DER, PEM, or PKCS12 format")
}

// tryParseDER attempts to parse raw DER encoded certificate bytes
func tryParseDER(certBytes []byte) ([]byte, bool) {
	if _, err := x509.ParseCertificate(certBytes); err == nil {
		return certBytes, true
	}
	return nil, false
}

// tryParsePEM attempts to decode a PEM encoded certificate
func tryParsePEM(certBytes []byte) ([]byte, bool) {
	block, _ := pem.Decode(certBytes)
	if block != nil && block.Type == "CERTIFICATE" {
		if _, err := x509.ParseCertificate(block.Bytes); err == nil {
			return block.Bytes, true
		}
	}
	return nil, false
}

// tryParsePEMWithMetadata handles PEM files with metadata (e.g., OpenSSL "Bag Attributes" output)
func tryParsePEMWithMetadata(certBytes []byte) ([]byte, bool) {
	pemStart := bytes.Index(certBytes, []byte("-----BEGIN CERTIFICATE-----"))
	if pemStart != -1 {
		block, _ := pem.Decode(certBytes[pemStart:])
		if block != nil && block.Type == "CERTIFICATE" {
			if _, err := x509.ParseCertificate(block.Bytes); err == nil {
				return block.Bytes, true
			}
		}
	}
	return nil, false
}

// tryParsePKCS12 attempts to decode a PKCS12/P12 file if password is provided
func tryParsePKCS12(certBytes []byte, p12Password string) ([]byte, bool) {
	if p12Password == "" {
		return nil, false
	}
	_, cert, err := pkcs12.Decode(certBytes, p12Password)
	if err != nil {
		slog.Debug("P12 decode failed", "error", err)
		return nil, false
	}
	if cert != nil {
		return cert.Raw, true
	}
	return nil, false
}
