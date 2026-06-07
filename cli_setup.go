package main

import (
	"log/slog"
	"os"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	"github.com/docopt/docopt-go"
)

func configureCLI(arguments docopt.Opts) {
	disableJSON, _ := arguments.Bool("--nojson")
	if disableJSON {
		JSONdisabled = true
	}

	pretty, _ := arguments.Bool("--pretty")
	if pretty {
		prettyJSON = true
	}

	traceLevelEnabled, _ := arguments.Bool("--trace")
	verboseLoggingEnabledLong, _ := arguments.Bool("--verbose")

	level := slog.LevelInfo
	if verboseLoggingEnabledLong {
		level = slog.LevelDebug
	}
	if traceLevelEnabled {
		level = ios.LevelTrace
	}

	handlerOpts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				if lvl, ok := a.Value.Any().(slog.Level); ok && lvl == ios.LevelTrace {
					a.Value = slog.StringValue("TRACE")
				}
			}
			return a
		},
	}

	var handler slog.Handler
	if disableJSON {
		handler = slog.NewTextHandler(os.Stderr, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(os.Stderr, handlerOpts)
	}
	logger := slog.New(handler)

	slog.SetDefault(logger)
	ios.SetLogger(logger)

	if traceLevelEnabled {
		slog.Info("Set Trace mode")
	} else if verboseLoggingEnabledLong {
		slog.Info("Set Debug mode")
	}
	slog.Debug("parsed arguments", "args", redactArgs(arguments))

	startAgentFromEnvironment()
	warnIfAgentIsNotRunning()
}

func redactArgs(arguments docopt.Opts) map[string]interface{} {
	redacted := make(map[string]interface{}, len(arguments))
	for key, value := range arguments {
		switch key {
		case "--password", "--p12password", "--proxyurl":
			if value != nil && value != "" {
				redacted[key] = "<redacted>"
				continue
			}
		}
		redacted[key] = value
	}
	return redacted
}

func startAgentFromEnvironment() {
	skipAgent, _ := os.LookupEnv("ENABLE_GO_IOS_AGENT")
	if skipAgent == "user" || skipAgent == "kernel" {
		tunnel.RunAgent(skipAgent)
	}
}

func warnIfAgentIsNotRunning() {
	if !tunnel.IsAgentRunning() {
		slog.Warn("go-ios agent is not running. You might need to start it with 'ios tunnel start' for ios17+. Use ENABLE_GO_IOS_AGENT=user for userspace tunnel or ENABLE_GO_IOS_AGENT=kernel for kernel tunnel for the experimental daemon mode.")
	}
}
