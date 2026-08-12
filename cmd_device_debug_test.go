package main

import (
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/docopt/docopt-go"
)

func TestInstrumentsSubcommand(t *testing.T) {
	testCases := []struct {
		name string
		args docopt.Opts
		want string
	}{
		{name: "fps", args: docopt.Opts{"instruments": true, "fps": true}, want: "fps"},
		{name: "network", args: docopt.Opts{"instruments": true, "network": true}, want: "network"},
		{name: "notifications", args: docopt.Opts{"instruments": true, "notifications": true}, want: "notifications"},
		{name: "no subcommand", args: docopt.Opts{"instruments": true}, want: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := instrumentsSubcommand(testCase.args); got != testCase.want {
				t.Fatalf("instrumentsSubcommand() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestInstrumentsCommandDispatch(t *testing.T) {
	for _, subcommand := range []string{"fps", "network", "notifications"} {
		args := docopt.Opts{"instruments": true, subcommand: true}
		matched := false
		dispatchCommand(commandContext{Args: args}, []command{
			commandByBool("instruments", func(commandContext) { matched = true }),
		})
		if !matched {
			t.Fatalf("instruments %s did not dispatch to the instruments command", subcommand)
		}
	}
}

func TestInstrumentsSampleDuration(t *testing.T) {
	testCases := []struct {
		name    string
		args    docopt.Opts
		want    time.Duration
		wantErr bool
	}{
		{name: "missing", args: docopt.Opts{}, want: 0},
		{name: "empty", args: docopt.Opts{"--duration": ""}, want: 0},
		{name: "nil", args: docopt.Opts{"--duration": nil}, want: 0},
		{name: "seconds", args: docopt.Opts{"--duration": "5"}, want: 5 * time.Second},
		{name: "fractional", args: docopt.Opts{"--duration": "0.5"}, want: 500 * time.Millisecond},
		{name: "not a number", args: docopt.Opts{"--duration": "abc"}, wantErr: true},
		{name: "negative", args: docopt.Opts{"--duration": "-1"}, wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := instrumentsSampleDuration(testCase.args)
			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != testCase.want {
				t.Fatalf("instrumentsSampleDuration() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func withJSONOutput(t *testing.T, disabled bool) {
	t.Helper()
	previousDisabled, previousPretty := JSONdisabled, prettyJSON
	JSONdisabled, prettyJSON = disabled, false
	t.Cleanup(func() {
		JSONdisabled, prettyJSON = previousDisabled, previousPretty
	})
}

func TestFormatFPSSampleJSON(t *testing.T) {
	withJSONOutput(t, false)
	sample := instruments.FramesPerSecondSample{CoreAnimationFramesPerSecond: 59.5}
	want := `{"fps":59.5}`
	if got := formatFPSSample(sample); got != want {
		t.Fatalf("formatFPSSample() = %q, want %q", got, want)
	}
}

func TestFormatFPSSampleHuman(t *testing.T) {
	withJSONOutput(t, true)
	sample := instruments.FramesPerSecondSample{CoreAnimationFramesPerSecond: 59.5}
	want := "fps=59.50"
	if got := formatFPSSample(sample); got != want {
		t.Fatalf("formatFPSSample() = %q, want %q", got, want)
	}
}

func TestFormatNetworkSampleJSON(t *testing.T) {
	withJSONOutput(t, false)
	sample := instruments.NetworkSample{
		Type: 2,
		Data: map[string]interface{}{
			"rxBytes":       uint64(42),
			"interfaceName": "en0",
		},
	}
	want := `{"type":2,"data":{"interfaceName":"en0","rxBytes":42}}`
	if got := formatNetworkSample(sample); got != want {
		t.Fatalf("formatNetworkSample() = %q, want %q", got, want)
	}
}

func TestFormatNetworkSampleHuman(t *testing.T) {
	withJSONOutput(t, true)
	sample := instruments.NetworkSample{
		Type: 2,
		Data: map[string]interface{}{
			"rxBytes":       uint64(42),
			"interfaceName": "en0",
		},
	}
	want := "type=2 interfaceName=en0 rxBytes=42"
	if got := formatNetworkSample(sample); got != want {
		t.Fatalf("formatNetworkSample() = %q, want %q", got, want)
	}
}

func TestStreamInstrumentsSamplesStopsAfterDuration(t *testing.T) {
	samples := make(chan instruments.FramesPerSecondSample)
	done := make(chan struct{})
	go func() {
		defer close(done)
		streamInstrumentsSamples(samples, 10*time.Millisecond, formatFPSSample)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("streamInstrumentsSamples did not stop after the duration elapsed")
	}
}

func TestStreamInstrumentsSamplesStopsOnClosedChannel(t *testing.T) {
	samples := make(chan instruments.FramesPerSecondSample)
	close(samples)
	done := make(chan struct{})
	go func() {
		defer close(done)
		streamInstrumentsSamples(samples, 0, formatFPSSample)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("streamInstrumentsSamples did not stop on a closed channel")
	}
}
