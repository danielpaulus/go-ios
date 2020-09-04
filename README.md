[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CircleCI](https://circleci.com/gh/danielpaulus/go-ios.svg?style=svg)](https://circleci.com/gh/danielpaulus/go-ios)
[![codecov](https://codecov.io/gh/danielpaulus/go-ios/branch/master/graph/badge.svg)](https://codecov.io/gh/danielpaulus/go-ios)
[![Go Report](https://goreportcard.com/badge/github.com/danielpaulus/go-ios)](https://goreportcard.com/report/github.com/danielpaulus/go-ios)
# go-ios
Is a port of the wonderful https://www.libimobiledevice.org to golang, so you can control iOS devices using go :-)

To run integration tests: `go test ./... --tags=integration`
To setup vscode to run integration tests: `touch ./vscode/settings.json`
and add 
```
{
"go.testTags" : "integration"
}
``` 
to it. 

```
iOS client v 0.01

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
   ios -h | --help                                    Prints this screen.
   ios --version | version [options]                  Prints the version
   ```
