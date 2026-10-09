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
		{name: "processes", args: docopt.Opts{"instruments": true, "processes": true}, want: "processes"},
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
	for _, subcommand := range []string{"fps", "network", "processes", "notifications"} {
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

func TestProcessSampleFilterFromArgs(t *testing.T) {
	testCases := []struct {
		name    string
		args    docopt.Opts
		want    processSampleFilter
		wantErr bool
	}{
		{name: "no filter", args: docopt.Opts{"--pid": nil, "--process": nil}, want: processSampleFilter{}},
		{name: "pid", args: docopt.Opts{"--pid": "42"}, want: processSampleFilter{pid: 42, hasPid: true}},
		{name: "pid zero", args: docopt.Opts{"--pid": "0"}, want: processSampleFilter{pid: 0, hasPid: true}},
		{name: "process", args: docopt.Opts{"--process": "SpringBoard"}, want: processSampleFilter{name: "SpringBoard"}},
		{name: "both", args: docopt.Opts{"--pid": "42", "--process": "SpringBoard"}, want: processSampleFilter{pid: 42, hasPid: true, name: "SpringBoard"}},
		{name: "invalid pid", args: docopt.Opts{"--pid": "abc"}, wantErr: true},
		{name: "negative pid", args: docopt.Opts{"--pid": "-1"}, wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := processSampleFilterFromArgs(testCase.args)
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
				t.Fatalf("processSampleFilterFromArgs() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func testProcessSample(pid uint64, name string) instruments.ProcessSample {
	return instruments.ProcessSample{
		Pid:  pid,
		Name: name,
		Attributes: map[string]interface{}{
			"pid":           pid,
			"name":          name,
			"cpuUsage":      2.5,
			"physFootprint": uint64(81462792),
		},
	}
}

func TestProcessSampleFilterMatches(t *testing.T) {
	springboard := testProcessSample(76, "SpringBoard")
	testCases := []struct {
		name   string
		filter processSampleFilter
		want   bool
	}{
		{name: "no filter", filter: processSampleFilter{}, want: true},
		{name: "pid match", filter: processSampleFilter{pid: 76, hasPid: true}, want: true},
		{name: "pid mismatch", filter: processSampleFilter{pid: 77, hasPid: true}, want: false},
		{name: "pid zero is not a wildcard", filter: processSampleFilter{pid: 0, hasPid: true}, want: false},
		{name: "name match", filter: processSampleFilter{name: "SpringBoard"}, want: true},
		{name: "name is exact", filter: processSampleFilter{name: "Spring"}, want: false},
		{name: "pid and name must both match", filter: processSampleFilter{pid: 77, hasPid: true, name: "SpringBoard"}, want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.filter.matches(springboard); got != testCase.want {
				t.Fatalf("matches() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestFormatProcessSampleJSON(t *testing.T) {
	withJSONOutput(t, false)
	want := `{"cpuUsage":2.5,"name":"SpringBoard","physFootprint":81462792,"pid":76}`
	if got := formatProcessSample(testProcessSample(76, "SpringBoard")); got != want {
		t.Fatalf("formatProcessSample() = %q, want %q", got, want)
	}
}

func TestFormatProcessSampleJSONNullCPU(t *testing.T) {
	withJSONOutput(t, false)
	sample := testProcessSample(1, "launchd")
	sample.Attributes["cpuUsage"] = nil
	want := `{"cpuUsage":null,"name":"launchd","physFootprint":81462792,"pid":1}`
	if got := formatProcessSample(sample); got != want {
		t.Fatalf("formatProcessSample() = %q, want %q", got, want)
	}
}

func TestFormatProcessSampleHuman(t *testing.T) {
	withJSONOutput(t, true)
	want := "pid=76 name=SpringBoard cpuUsage=2.50 physFootprint=81462792"
	if got := formatProcessSample(testProcessSample(76, "SpringBoard")); got != want {
		t.Fatalf("formatProcessSample() = %q, want %q", got, want)
	}
}

func TestFilterProcessSamplesSkipsFirstSnapshotAndFilters(t *testing.T) {
	snapshots := make(chan instruments.SysmontapProcessSnapshot, 2)
	snapshots <- instruments.SysmontapProcessSnapshot{Processes: []instruments.ProcessSample{
		testProcessSample(1, "launchd"), testProcessSample(76, "SpringBoard"),
	}}
	snapshots <- instruments.SysmontapProcessSnapshot{Processes: []instruments.ProcessSample{
		testProcessSample(1, "launchd"), testProcessSample(76, "SpringBoard"),
	}}
	close(snapshots)

	var got []instruments.ProcessSample
	for sample := range filterProcessSamples(snapshots, processSampleFilter{name: "SpringBoard"}) {
		got = append(got, sample)
	}
	if len(got) != 1 || got[0].Pid != 76 {
		t.Fatalf("filterProcessSamples() = %+v, want only SpringBoard from the second snapshot", got)
	}
}
