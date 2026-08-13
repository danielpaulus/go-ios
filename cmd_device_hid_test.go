package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHIDGesturesAcceptsEveryGesture(t *testing.T) {
	script := `
# a comment on its own line
tap 100 200
drag 0 0 65535 65535
drag 0 0 100 100 16 0.25
move -50 50
type hello world
button 12 64
sleep 0.5
tap 1 1   # trailing comment
`

	gestures, err := parseHIDGestures(strings.NewReader(script))
	require.NoError(t, err)
	require.Len(t, gestures, 8)

	ops := make([]string, 0, len(gestures))
	for _, g := range gestures {
		ops = append(ops, g.op)
		assert.NotNil(t, g.run, "every parsed gesture must be runnable")
	}
	assert.Equal(t, []string{"tap", "drag", "drag", "move", "type", "button", "sleep", "tap"}, ops)

	// Line numbers are reported for error messages, so they must survive
	// comments and blank lines.
	assert.Equal(t, 3, gestures[0].line)
	assert.Equal(t, 10, gestures[7].line)
}

func TestParseHIDGesturesRejectsBadInput(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		wantMessage string
	}{
		{name: "unknown gesture", script: "wiggle 1 2", wantMessage: `unknown gesture "wiggle"`},
		{name: "tap missing coordinate", script: "tap 100", wantMessage: "tap wants X Y"},
		{name: "tap out of range", script: "tap 100 70000", wantMessage: "out of the 0..65535 range"},
		{name: "tap non numeric", script: "tap a b", wantMessage: "invalid coordinate"},
		{name: "drag too few args", script: "drag 1 2 3", wantMessage: "drag wants X Y TOX TOY"},
		{name: "drag bad steps", script: "drag 1 2 3 4 many", wantMessage: "invalid STEPS"},
		{name: "drag bad duration", script: "drag 1 2 3 4 8 soon", wantMessage: "invalid DURATION"},
		{name: "move missing arg", script: "move 1", wantMessage: "move wants X Y"},
		{name: "type without text", script: "type", wantMessage: "type wants TEXT"},
		{name: "button missing code", script: "button 12", wantMessage: "button wants USAGEPAGE USAGECODE"},
		{name: "button bad page", script: "button page 64", wantMessage: "invalid USAGEPAGE"},
		{name: "sleep bad seconds", script: "sleep soon", wantMessage: "invalid SECONDS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseHIDGestures(strings.NewReader(tt.script))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantMessage)
			assert.Contains(t, err.Error(), "line 1", "errors should point at the offending line")
		})
	}
}

func TestParseHIDGesturesIgnoresBlankAndCommentOnlyScripts(t *testing.T) {
	gestures, err := parseHIDGestures(strings.NewReader("\n  \n# only a comment\n"))
	require.NoError(t, err)
	assert.Empty(t, gestures)
}

func TestParseHIDGesturesKeepsWholeTypedLine(t *testing.T) {
	gestures, err := parseHIDGestures(strings.NewReader("type  several   spaced words"))
	require.NoError(t, err)
	require.Len(t, gestures, 1)
	assert.Equal(t, "type", gestures[0].op)
}

func TestParseNormalisedCoordinate(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    uint16
		wantErr bool
	}{
		{name: "min", in: "0", want: 0},
		{name: "max", in: "65535", want: 65535},
		{name: "above max", in: "65536", wantErr: true},
		{name: "negative", in: "-1", wantErr: true},
		{name: "not a number", in: "x", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNormalisedCoordinate(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
