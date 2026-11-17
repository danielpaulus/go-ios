package main

import (
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios/hid"
	"github.com/docopt/docopt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHingeAngle(t *testing.T) {
	for name, tc := range map[string]struct {
		args docopt.Opts
		want float64
	}{
		"close":    {docopt.Opts{"close": true}, hid.HingeClosed},
		"halfopen": {docopt.Opts{"halfopen": true}, hid.HingeHalfOpen},
		"open":     {docopt.Opts{"open": true}, hid.HingeOpenFlat},
		"degrees":  {docopt.Opts{"<degrees>": "95.5"}, 95.5},
	} {
		got, err := hingeAngle(tc.args)
		require.NoError(t, err, name)
		assert.Equal(t, tc.want, got, name)
	}
	_, err := hingeAngle(docopt.Opts{"<degrees>": "ajar"})
	assert.Error(t, err)
}

func TestPressedButton(t *testing.T) {
	for _, b := range hid.Buttons() {
		assert.Equal(t, b, pressedButton(docopt.Opts{"button": true, b.String(): true}), b.String())
	}
}

func TestChosenOrientation(t *testing.T) {
	for _, o := range hid.Orientations() {
		assert.Equal(t, o, chosenOrientation(docopt.Opts{"orientation": true, o.String(): true}), o.String())
	}
}

func TestButtonHold(t *testing.T) {
	hold, err := buttonHold(docopt.Opts{})
	require.NoError(t, err)
	assert.Zero(t, hold, "no --hold uses the button's own")

	hold, err = buttonHold(docopt.Opts{"--hold": "1500"})
	require.NoError(t, err)
	assert.Equal(t, 1500*time.Millisecond, hold)

	for _, bad := range []string{"0", "-5", "soon"} {
		_, err = buttonHold(docopt.Opts{"--hold": bad})
		assert.Error(t, err, bad)
	}
}
