package main

import (
	"testing"

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
