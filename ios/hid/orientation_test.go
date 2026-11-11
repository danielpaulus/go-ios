package hid

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDeviceControl answers every orientation request with the orientation it
// was last changed to.
type fakeDeviceControl struct {
	now      string
	requests []map[string]interface{}
}

func (f *fakeDeviceControl) request(msg map[string]interface{}) (map[string]interface{}, error) {
	f.requests = append(f.requests, msg)
	if change, ok := payloadOf(msg)["changeOrientation"].(map[string]interface{}); ok {
		f.now = change["_0"].(string)
	}
	return map[string]interface{}{
		"currentDeviceOrientation":        f.now,
		"currentDeviceNonFlatOrientation": f.now,
	}, nil
}

func (f *fakeDeviceControl) Close() error { return nil }

// duoDisplays lists a foldable's cover and inner displays, with the inner one
// active when open.
func duoDisplays(open bool) map[string]interface{} {
	display := func(id int, w, h float64, active bool) map[string]interface{} {
		return map[string]interface{}{
			"displayId": id, "external": false, "active": active,
			"bounds": []interface{}{[]interface{}{0.0, 0.0}, []interface{}{w, h}},
		}
	}
	return map[string]interface{}{"displays": []interface{}{
		display(1, 1398, 2034, !open),
		display(2, 2853, 2007, open),
	}}
}

type orientationFakes struct {
	indigo        *fakeMessages
	universal     *fakeMessages
	deviceControl *fakeDeviceControl
}

func openOrientationControls(foldable, open bool) (*Controls, orientationFakes) {
	f := orientationFakes{
		indigo:        &fakeMessages{},
		universal:     &fakeMessages{},
		deviceControl: &fakeDeviceControl{now: "portrait"},
	}
	c := &Controls{
		indigo:            f.indigo,
		foldable:          foldable,
		dialDeviceControl: func() (requestConn, error) { return f.deviceControl, nil },
		dialUniversalHID:  func() (messageConn, error) { return f.universal, nil },
		readDisplays:      func() (map[string]interface{}, error) { return duoDisplays(open), nil },
	}
	return c, f
}

func virtualizationOf(t *testing.T, msg map[string]interface{}) map[string]interface{} {
	t.Helper()
	return decodeVirtualization(t, payloadOf(msg)["data"].([]byte))
}

func TestSetOrientationOnAFoldableUsesItsPicker(t *testing.T) {
	for o, picker := range map[Orientation]string{
		OrientationPortrait:           "portrait",
		OrientationLandscapeLeft:      "landscape-left",
		OrientationPortraitUpsideDown: "pud",
		OrientationLandscapeRight:     "landscape-right",
	} {
		c, f := openOrientationControls(true, true)
		require.NoError(t, c.SetOrientation(o))
		require.Len(t, f.indigo.sent, 2)
		assert.Equal(t, map[string]interface{}{
			"provider": "com.apple.Virtualization.VirtualMachines",
			"source":   "orientation-picker-control",
			"type":     "enum",
			"value":    picker,
		}, virtualizationOf(t, f.indigo.sent[1]), o.String())
		assert.Empty(t, f.deviceControl.requests)
	}
}

// After a restart a foldable ignores its picker until the virtualization
// service got an empty report and the hinge was set.
func TestSetOrientationArmsAFoldablesPickerFirst(t *testing.T) {
	c, f := openOrientationControls(true, true)
	require.NoError(t, c.SetOrientation(OrientationLandscapeLeft))

	require.Len(t, f.universal.sent, 1)
	assert.Equal(t, map[string]interface{}{
		"featureIdentifier": "com.apple.coredevice.feature.remote.universalhidservice",
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"send": map[string]interface{}{"_0": []byte{0x81, 0x61, 0xff, 0x5b, 0x00, 0x00}, "_1": uint64(0x500)},
		},
	}, f.universal.sent[0])

	require.Len(t, f.indigo.sent, 2)
	hinge := virtualizationOf(t, f.indigo.sent[0])
	assert.Equal(t, "hinge-slider-control", hinge["source"], "the hinge comes before the picker")
}

func TestSetOrientationKeepsTheHingeWhereItIs(t *testing.T) {
	for open, want := range map[bool]float64{true: HingeOpenFlat, false: HingeClosed} {
		c, f := openOrientationControls(true, open)
		require.NoError(t, c.SetOrientation(OrientationPortrait))
		assert.Equal(t, want, virtualizationOf(t, f.indigo.sent[0])["value"], "open=%v", open)
	}
}

func TestSetOrientationOnAFoldableFailsWithoutItsDisplays(t *testing.T) {
	c, f := openOrientationControls(true, true)
	c.readDisplays = func() (map[string]interface{}, error) { return nil, errors.New("no deviceinfo") }
	assert.Error(t, c.SetOrientation(OrientationPortrait))
	assert.Empty(t, f.indigo.sent, "nothing is sent")
	assert.Empty(t, f.universal.sent)
}

func TestSetOrientationOnASingleDisplayDevice(t *testing.T) {
	c, f := openOrientationControls(false, false)
	require.NoError(t, c.SetOrientation(OrientationLandscapeRight))

	require.Len(t, f.deviceControl.requests, 1)
	assert.Equal(t, "com.apple.coredevice.feature.remote.devicecontrol.orientation", f.deviceControl.requests[0]["featureIdentifier"])
	assert.Equal(t, "OrientationRequest", f.deviceControl.requests[0]["messageType"])
	assert.Equal(t, map[string]interface{}{"changeOrientation": map[string]interface{}{"_0": "landscapeRight"}}, payloadOf(f.deviceControl.requests[0]))
	assert.Empty(t, f.indigo.sent)
	assert.Empty(t, f.universal.sent)
}

// devicecontrol answers with the orientation the device is in, which differs
// when it could not turn, such as with the rotation lock on.
func TestSetOrientationReportsADeviceThatDidNotTurn(t *testing.T) {
	c, _ := openOrientationControls(false, false)
	c.dialDeviceControl = func() (requestConn, error) { return stuckDeviceControl{}, nil }
	assert.ErrorContains(t, c.SetOrientation(OrientationLandscapeLeft), "the device is in portrait")
}

type stuckDeviceControl struct{}

func (stuckDeviceControl) request(map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"currentDeviceNonFlatOrientation": "portrait"}, nil
}

func (stuckDeviceControl) Close() error { return nil }

func TestSetOrientationRefusesAnUnknownOrientation(t *testing.T) {
	c, f := openOrientationControls(true, true)
	assert.Error(t, c.SetOrientation(Orientation(42)))
	assert.Empty(t, f.indigo.sent)
}

func TestClosedControlsTurnNothing(t *testing.T) {
	for _, foldable := range []bool{true, false} {
		c, f := openOrientationControls(foldable, true)
		require.NoError(t, c.Close())
		assert.Error(t, c.SetOrientation(OrientationLandscapeLeft))
		assert.Empty(t, f.indigo.sent)
		assert.Empty(t, f.universal.sent)
		assert.Empty(t, f.deviceControl.requests)
	}
}

func TestParseOrientation(t *testing.T) {
	for _, o := range Orientations() {
		parsed, err := ParseOrientation(o.String())
		require.NoError(t, err)
		assert.Equal(t, o, parsed)
	}
	o, err := ParseOrientation("LandscapeLeft")
	require.NoError(t, err)
	assert.Equal(t, OrientationLandscapeLeft, o, "ignores case")
	_, err = ParseOrientation("sideways")
	assert.ErrorContains(t, err, "portrait, landscapeleft, portraitupsidedown, landscaperight")
}

func TestInnerDisplayActive(t *testing.T) {
	assert.True(t, innerDisplayActive(duoDisplays(true)))
	assert.False(t, innerDisplayActive(duoDisplays(false)))
	assert.False(t, innerDisplayActive(map[string]interface{}{}))
}
