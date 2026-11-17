package deviceinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sideChannel = "16dea760-2442-4eb3-becb-390ed4433392"

// hingeUpdate is a hinge update as an iPhone Duo with iOS 27 sends it, folded
// by hand.
func hingeUpdate(sideChannel string, valid bool) map[string]interface{} {
	degrees := func(v float64) map[string]interface{} {
		return map[string]interface{}{
			"unit": map[string]interface{}{
				"converter": map[string]interface{}{"coefficient": int64(1), "constant": int64(0)},
				"symbol":    "°",
			},
			"value": v,
		}
	}
	return map[string]interface{}{
		"CoreDevice.XPCMessageKey.sideChannelStatus": map[string]interface{}{
			"pushing": map[string]interface{}{
				"elements": []interface{}{map[string]interface{}{
					"angle":           degrees(90.97),
					"mechanicalAngle": degrees(89.44),
					"minAngle":        degrees(0),
					"maxAngle":        degrees(180),
					"isAngleValid":    valid,
					"isVelocityValid": true,
					"velocity":        degrees(-15.3),
					"timestamp":       1864.66,
				}},
			},
		},
		"XPCSideChannel.uniqueIdentifier": sideChannel,
	}
}

func TestHingeAngleInTakesTheMechanicalAngle(t *testing.T) {
	degrees, ours, err := hingeAngleIn(hingeUpdate("16DEA760-2442-4EB3-BECB-390ED4433392", true), sideChannel)
	require.NoError(t, err)
	assert.True(t, ours, "the side channel is matched ignoring case")
	assert.Equal(t, 89.44, degrees)
}

func TestHingeAngleInSkipsAnotherSideChannel(t *testing.T) {
	_, ours, err := hingeAngleIn(hingeUpdate("AD348699-6370-4BDD-B889-A31728D7ED84", true), sideChannel)
	require.NoError(t, err)
	assert.False(t, ours)
}

func TestHingeAngleInRefusesAnInvalidAngle(t *testing.T) {
	_, ours, err := hingeAngleIn(hingeUpdate(sideChannel, false), sideChannel)
	assert.True(t, ours)
	assert.ErrorContains(t, err, "no valid hinge angle")
}

func TestHingeAngleInRefusesAnUpdateWithoutAnAngle(t *testing.T) {
	_, ours, err := hingeAngleIn(map[string]interface{}{"XPCSideChannel.uniqueIdentifier": sideChannel}, sideChannel)
	assert.True(t, ours)
	assert.Error(t, err)
}
