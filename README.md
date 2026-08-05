[![](https://dcbadge.vercel.app/api/server/Zr8J3bCdkv)](https://discord.gg/Zr8J3bCdkv)
[![NPM](https://nodei.co/npm/go-ios.png?mini=true)](https://npmjs.org/package/go-ios)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/daniel1paulus.svg?style=social&label=Follow%20%40daniel1paulus)](https://twitter.com/daniel1paulus)
[![NPM](https://img.shields.io/npm/dw/go-ios?style=flat-square)](https://npmjs.org/package/go-ios)

<img src="logo.png" width="256"/>

Welcome 👋

`npm install -g go-ios` can be used to get going. Run `ios --help` after the installation for details.
For iOS 17+ devices you need to run `sudo ios tunnel start` for go ios to work. This will start a tunnel daemon.
To keep the tunnel agent running permanently (surviving reboots), register it as a managed OS service with `sudo ios tunnel service install [--userspace]` (systemd on Linux, launchd on macOS, the service manager on Windows) — installing a system service needs sudo/admin. Pass `--user` for a user-level service instead; on Linux enable lingering (`loginctl enable-linger`) so it runs without a login session. Remove it again with `ios tunnel service uninstall`.
To make this work on Windows, download the latest wintun.dll from here `https://git.zx2c4.com/wintun` and copy it to `C:/Windows/system32`

The goal of this project is to provide a stable and production ready opensource solution to automate iOS device on Linux, Windows and Mac OS X. I am delighted to announce that a few companies including [headspin.io](https://www.headspin.io/) and [Sauce Labs](https://saucelabs.com/) will use or are using go-iOS.

Follow my twitter for updates or check out my medium blog: https://daniel-paulus.medium.com/

If you are interested in using go-iOS please get in touch on LinkedIn, Twitter or the Github discussions above, I always love to hear what people are doing with it.

If you miss something your Mac can do but go-iOS can't, just request a feature in the issues tab.

# New REST-API

Go-iOS is getting an experimental REST-API check it out [https://github.com/danielpaulus/go-ios/tree/main/restapi](https://github.com/danielpaulus/go-ios/tree/main/restapi)

# Design principles:

1. Using golang to compile static, small and fast binaries for all platforms very easily.

   _Build Manual_: Install golang and run `go build`

2. All output as JSON so you can easily use go-iOS from any other programming language
3. Everything is a module, you can use go-iOS in golang projects as a module dependency easily

# Features:

Most notable:

- Install apps zipped as ipa or unzipped from their .app folder `ios install --path=/path/to/app`
- Run XCTests including WebdriverAgent on Linux, Windows and Mac
- Start and Stop apps
- Use a debug proxy to reverse engineer every tool Mac OSX has, so you can contrib to go-ios or build your own
- Pair devices without manual tapping on a popup
- Install developer images automatically by running `ios image auto`
- Set thermal states and network emulation on the device with the `ios devicestate` command

Help:

```bash
ios --help
ios help <command>
ios <command> --help
```

<!-- help begin -->

```text
go-ios local-build

Cross-platform iOS automation CLI.

Usage:
  ios [--help]
  ios help [<command>...]
  ios <command> [<args>...]

Global options:
  -h, --help                 Show help.
  -v, --verbose              Enable debug logging.
  -t, --trace                Enable trace logging.
  --nojson                   Disable JSON output.
  --pretty                   Pretty-print JSON output.
  --udid=<udid>              Target a specific device.
  --tunnel-info-port=<port>  Tunnel info API port (default 28100).
  --address=<ipv6addr>       Device tunnel address.
  --rsd-port=<port>          Device tunnel RSD port.
  --proxyurl=<url>           Outbound HTTP proxy URL.
  --userspace-port=<port>    Userspace tunnel port.

Commands:
  activate                   Activate a device.
  apps                       List installed applications.
  assistivetouch             Manage AssistiveTouch state.
  ax                         Accessibility inspector features.
  batterycheck               Battery information.
  batteryregistry            Battery registry metrics.
  crash cp                   Copy crash reports.
  crash ls                   List crash reports.
  crash rm                   Remove crash reports.
  date                       Print device date.
  debug                      Start LLDB debug session.
  devicename                 Print device name.
  devicestate enable         Enable device condition profile.
  devicestate list           List device condition profiles.
  devmode                    Manage developer mode.
  diagnostics list           List diagnostics.
  diskspace                  Print disk usage.
  dproxy                     Start debug proxy.
  erase                      Erase device.
  file ls                    List files in app/group/temp/crash container.
  file pull                  Pull file from device.
  file push                  Push file to device.
  forward                    Forward host port to device.
  fsync                      App container file sync operations.
  httpproxy                  Install global HTTP proxy profile.
  httpproxy remove           Remove go-ios HTTP proxy profile.
  image auto                 Auto-download and mount developer image.
  image list                 List mounted developer images.
  image mount                Mount developer image.
  image unmount              Unmount developer image.
  info                       Dump device info.
  install                    Install app bundle or IPA.
  instruments notifications  Stream app state notifications.
  ip                         Detect device IP from packet capture.
  kill                       Kill app by bundle ID, PID, or process.
  lang                       Read or set device language and locale.
  launch                     Launch app by bundle ID.
  list                       List connected devices.
  listen                     Listen for device connect/disconnect.
  lockdown get               Query lockdown values.
  memlimitoff                Disable process memory limit.
  mobilegestalt              Query mobilegestalt keys.
  ostrace                    Stream os_trace_relay logs.
  pair                       Pair host with device.
  pcap                       Capture network packets.
  prepare                    Prepare device for automation.
  prepare cloudconfig        Print cloud configuration.
  prepare create-cert        Create supervision certificate.
  prepare printskip          Print prepare skip options.
  profile add                Install profile on device.
  profile list               List installed profiles.
  profile remove             Remove installed profile.
  ps                         List running processes.
  readpair                   Dump pair record.
  reboot                     Reboot device.
  resetax                    Reset accessibility settings.
  resetlocation              Reset simulated location.
  rsd ls                     List RSD services.
  runtest                    Run XCUITest bundles.
  runwda                     Run WebDriverAgent.
  runxctest                  Run XCTest from .xctestrun file.
  screenshot                 Capture screenshot or stream MJPEG.
  setlocation                Set simulated location coordinates.
  setlocationgpx             Set simulated location from GPX.
  syslog                     Stream device syslog.
  sysmontap                  Stream CPU and memory metrics.
  timeformat                 Manage time format setting.
  tunnel ls                  List running tunnels.
  tunnel service             Run the tunnel agent as a managed OS service.
  tunnel start               Start iOS 17+ tunnel and pair if needed.
  tunnel stopagent           Stop tunnel agent.
  uninstall                  Uninstall app by bundle ID.
  version                    Print version.
  voiceover                  Manage VoiceOver state.
  zoom                       Manage Zoom state.

Run 'ios help <command>' or 'ios <command> --help' for command details.
```

<!-- help end -->
