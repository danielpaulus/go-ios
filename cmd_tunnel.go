package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/danielpaulus/go-ios/ios/tunnel"
	"github.com/docopt/docopt-go"
)

type tunnelCommandContext struct {
	Args           docopt.Opts
	TunnelInfoHost string
	TunnelInfoPort int
}

func isTunnelCommand(args docopt.Opts) bool {
	return boolArg(args, "tunnel")
}

func dispatchTunnelCommand(ctx tunnelCommandContext) bool {
	if !isTunnelCommand(ctx.Args) {
		return false
	}

	startCommand, _ := ctx.Args.Bool("start")
	useUserspaceNetworking, _ := ctx.Args.Bool("--userspace")
	if startCommand && !useUserspaceNetworking {
		err := tunnel.CheckPermissions()
		exitIfError("If --userspace is not set, we need sudo, an admin shell on Windows, or CAP_NET_ADMIN on Linux", err)
	}
	if useUserspaceNetworking {
		slog.Info("Using userspace networking")
	}
	stopagent, _ := ctx.Args.Bool("stopagent")
	listCommand, _ := ctx.Args.Bool("ls")
	if startCommand {
		pairRecordsPath, _ := ctx.Args.String("--pair-record-path")
		if len(pairRecordsPath) == 0 {
			pairRecordsPath = "."
		}
		if strings.ToLower(pairRecordsPath) == "default" {
			pairRecordsPath = "/var/db/lockdown/RemotePairing/user_501"
			slog.Warn("'--pair-record-path=default' reads Apple's own RemotePairing identity, "+
				"which macOS 26 (Tahoe) and newer block for third-party binaries via TCC "+
				"('operation not permitted'). If the tunnel fails to read selfIdentity.plist, "+
				"drop '=default' and pass a stable writable directory instead "+
				"(e.g. --pair-record-path=/Users/Shared/go-ios); go-ios then manages its own "+
				"tunnel identity and pairs it on first use. See https://github.com/danielpaulus/go-ios/issues/710",
				"pairRecordsPath", pairRecordsPath)
		}
		startTunnel(context.TODO(), pairRecordsPath, ctx.TunnelInfoHost, ctx.TunnelInfoPort, useUserspaceNetworking)
	} else if listCommand {
		tunnels, err := tunnel.ListRunningTunnels(ctx.TunnelInfoHost, ctx.TunnelInfoPort)
		exitIfError("failed to get tunnel infos", err)
		if JSONdisabled {
			for index, t := range tunnels {
				if 0 != index {
					fmt.Println()
				}
				fmt.Printf("Udid: %s\n  Address: %s\n  RsdPort: %d\n  UserspaceTUN: %v\n  UserspaceTUNPort: %d\n",
					t.Udid, t.Address, t.RsdPort, t.UserspaceTUN, t.UserspaceTUNPort)
			}
		} else {
			fmt.Println(convertToJSONString(tunnels))
		}
	}
	if stopagent {
		err := tunnel.CloseAgent()
		if err != nil {
			exitIfError("failed to close agent", err)
		}
	}
	return true
}
