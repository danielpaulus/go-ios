package hid

import (
	"errors"
	"math"
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
		backlight := "off"
		if active {
			backlight = "activeOn"
		}
		return map[string]interface{}{
			"displayId": id, "external": false, "active": active, "backlightState": backlight,
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

// openOrientationControls opens controls on a device whose hinge reads open
// flat when open, and closed otherwise.
func openOrientationControls(foldable, open bool) (*Controls, orientationFakes) {
	hinge := HingeClosed
	if open {
		hinge = HingeOpenFlat
	}
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
		readHingeAngle:    func() (float64, error) { return hinge, nil },
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
		require.Len(t, f.indigo.sent, 1, "the picker alone, the hinge stays")
		assert.Equal(t, map[string]interface{}{
			"provider": "com.apple.Virtualization.VirtualMachines",
			"source":   "orientation-picker-control",
			"type":     "enum",
			"value":    picker,
		}, virtualizationOf(t, f.indigo.sent[0]), o.String())
		assert.Empty(t, f.universal.sent)
		assert.Empty(t, f.deviceControl.requests)
	}
}

// After a restart a foldable ignores its picker until the virtualization
// service got an empty report and the hinge was set.
func TestUnstickOrientationSendsTheReportThenTheHinge(t *testing.T) {
	c, f := openOrientationControls(true, true)
	require.NoError(t, c.UnstickOrientation())

	require.Len(t, f.universal.sent, 1)
	assert.Equal(t, map[string]interface{}{
		"featureIdentifier": "com.apple.coredevice.feature.remote.universalhidservice",
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"send": map[string]interface{}{"_0": []byte{0x81, 0x61, 0xff, 0x5b, 0x00, 0x00}, "_1": uint64(0x500)},
		},
	}, f.universal.sent[0])

	require.Len(t, f.indigo.sent, 1)
	assert.Equal(t, "hinge-slider-control", virtualizationOf(t, f.indigo.sent[0])["source"])
}

func TestUnstickOrientationKeepsTheHingeWhereItIs(t *testing.T) {
	c, f := openOrientationControls(true, true)
	c.readHingeAngle = func() (float64, error) { return 120.5, nil }
	require.NoError(t, c.UnstickOrientation())
	require.Len(t, f.indigo.sent, 1)
	assert.Equal(t, 120.5, virtualizationOf(t, f.indigo.sent[0])["value"])
}

// Without the hinge angle, the hinge goes to the end the display in use goes
// with, so that display stays.
func TestUnstickOrientationWithoutTheHingeAngleKeepsTheDisplayInUse(t *testing.T) {
	for open, want := range map[bool]float64{true: HingeOpenFlat, false: HingeClosed} {
		c, f := openOrientationControls(true, open)
		c.readHingeAngle = func() (float64, error) { return 0, errors.New("no hinge updates") }
		require.NoError(t, c.UnstickOrientation())
		assert.Equal(t, want, virtualizationOf(t, f.indigo.sent[0])["value"], "open=%v", open)
	}
}

func TestUnstickOrientationAtKeepsTheHingeAngle(t *testing.T) {
	c, f := openOrientationControls(true, true)
	require.NoError(t, c.UnstickOrientationAt(130))
	require.Len(t, f.universal.sent, 1)
	require.Len(t, f.indigo.sent, 1)
	assert.Equal(t, 130.0, virtualizationOf(t, f.indigo.sent[0])["value"])
}

func TestUnstickOrientationAtRefusesAnAngleOutOfRange(t *testing.T) {
	for _, degrees := range []float64{-1, 181, math.NaN()} {
		c, f := openOrientationControls(true, true)
		assert.Error(t, c.UnstickOrientationAt(degrees), "%v", degrees)
		assert.Empty(t, f.universal.sent)
		assert.Empty(t, f.indigo.sent)
	}
}

func TestUnstickOrientationFailsWithoutTheHingeAngleOrTheDisplays(t *testing.T) {
	c, f := openOrientationControls(true, true)
	noHinge, noDisplays := errors.New("no hinge updates"), errors.New("no displays")
	c.readHingeAngle = func() (float64, error) { return 0, noHinge }
	c.readDisplays = func() (map[string]interface{}, error) { return nil, noDisplays }
	err := c.UnstickOrientation()
	assert.ErrorIs(t, err, noHinge)
	assert.ErrorIs(t, err, noDisplays)
	assert.Empty(t, f.indigo.sent, "nothing is sent")
	assert.Empty(t, f.universal.sent)
}

func TestUnstickOrientationRefusesADeviceThatDoesNotFold(t *testing.T) {
	c, f := openOrientationControls(false, false)
	assert.ErrorIs(t, c.UnstickOrientation(), ErrNotFoldable)
	assert.ErrorIs(t, c.UnstickOrientationAt(90), ErrNotFoldable)
	assert.Empty(t, f.indigo.sent)
	assert.Empty(t, f.universal.sent)
	assert.Empty(t, f.deviceControl.requests)
}

// brokenMessages fails to send or close with the errors given.
type brokenMessages struct {
	sendErr, closeErr error
}

func (b brokenMessages) send(map[string]interface{}) error { return b.sendErr }
func (b brokenMessages) Close() error                      { return b.closeErr }

func TestUnstickOrientationSaysWhichStepFailed(t *testing.T) {
	broken := errors.New("broken")
	for step, breakIt := range map[string]func(c *Controls){
		"connect to the universal HID service": func(c *Controls) {
			c.dialUniversalHID = func() (messageConn, error) { return nil, broken }
		},
		"send the empty report": func(c *Controls) {
			c.dialUniversalHID = func() (messageConn, error) { return brokenMessages{sendErr: broken}, nil }
		},
		"close the universal HID service": func(c *Controls) {
			c.dialUniversalHID = func() (messageConn, error) { return brokenMessages{closeErr: broken}, nil }
		},
		"set the hinge to 180 degrees": func(c *Controls) {
			c.indigo = brokenMessages{sendErr: broken}
		},
	} {
		c, _ := openOrientationControls(true, true)
		breakIt(c)
		err := c.UnstickOrientation()
		assert.ErrorIs(t, err, broken, step)
		assert.ErrorContains(t, err, step)
	}
}

func TestHingeAngle(t *testing.T) {
	c, _ := openOrientationControls(true, true)
	c.readHingeAngle = func() (float64, error) { return 120.5, nil }
	degrees, err := c.HingeAngle()
	require.NoError(t, err)
	assert.Equal(t, 120.5, degrees)

	c.readHingeAngle = func() (float64, error) { return 0.05, nil }
	degrees, err = c.HingeAngle()
	require.NoError(t, err)
	assert.Equal(t, HingeOpenFlat, degrees, "the read lags a fold by hand; the inner display is in use")

	broken := errors.New("no hinge updates")
	c.readHingeAngle = func() (float64, error) { return 0, broken }
	_, err = c.HingeAngle()
	assert.ErrorIs(t, err, broken)
}

func TestHingeAngleIsWithinTheHingesRange(t *testing.T) {
	c, _ := openOrientationControls(true, false)
	c.readHingeAngle = func() (float64, error) { return -0.06, nil }
	degrees, err := c.HingeAngle()
	require.NoError(t, err)
	assert.Equal(t, HingeClosed, degrees)
}

func TestHingeAngleOfADeviceThatDoesNotFold(t *testing.T) {
	c, _ := openOrientationControls(false, false)
	c.readHingeAngle = func() (float64, error) { panic("not read") }
	_, err := c.HingeAngle()
	assert.ErrorIs(t, err, ErrNotFoldable)
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
		c.readHingeAngle = func() (float64, error) { panic("not read") }
		require.NoError(t, c.Close())
		assert.Error(t, c.SetOrientation(OrientationLandscapeLeft))
		if foldable {
			_, err := c.HingeAngle()
			assert.Error(t, err)
			assert.Error(t, c.UnstickOrientation())
			assert.Error(t, c.UnstickOrientationAt(90))
		}
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

func TestInnerDisplayInUse(t *testing.T) {
	assert.True(t, innerDisplayInUse(duoDisplays(true)))
	assert.False(t, innerDisplayInUse(duoDisplays(false)))
	assert.False(t, innerDisplayInUse(map[string]interface{}{}))
	assert.False(t, innerDisplayInUse(openWhileSetClosed()), "the lit cover display, not the dark active inner one")
}

// openWhileSetClosed is a foldable open while its hinge is set closed: the
// inner display stays active but goes dark, and the cover display lights up
// without becoming active.
func openWhileSetClosed() map[string]interface{} {
	info := duoDisplays(true)
	displays := info["displays"].([]interface{})
	displays[0].(map[string]interface{})["backlightState"] = "activeOn"
	displays[1].(map[string]interface{})["backlightState"] = "off"
	return info
}

// The angle read can lag a fold by hand, while the displays follow whichever
// came last, so it is kept only when it fits the display in use.
func TestHingeAngleInUse(t *testing.T) {
	for _, tc := range []struct {
		name     string
		displays map[string]interface{}
		read     float64
		want     float64
		fits     bool
	}{
		{"inner display, half open", duoDisplays(true), 130, 130, true},
		{"inner display, read closed", duoDisplays(true), 0.05, HingeOpenFlat, false},
		{"inner display, at the cover's limit", duoDisplays(true), 90, HingeOpenFlat, false},
		{"cover display, closed", duoDisplays(false), 0.05, 0.05, true},
		{"cover display, below closed", duoDisplays(false), -0.06, HingeClosed, true},
		{"cover display, folded to 45", duoDisplays(false), 45, 45, true},
		{"cover display, read open", duoDisplays(false), 180, HingeClosed, false},
		{"open while set closed", openWhileSetClosed(), 0, 0, true},
	} {
		degrees, fits := HingeAngleInUse(tc.read, tc.displays)
		assert.Equal(t, tc.want, degrees, tc.name)
		assert.Equal(t, tc.fits, fits, tc.name)
	}
}

func TestUnstickOrientationKeepsAnAngleThatFitsTheDisplayInUse(t *testing.T) {
	for _, tc := range []struct {
		name     string
		displays map[string]interface{}
		read     float64
		want     float64
	}{
		{"inner display, half open", duoDisplays(true), 130, 130},
		{"inner display, read closed", duoDisplays(true), 0.05, HingeOpenFlat},
		{"cover display, read open", duoDisplays(false), 180, HingeClosed},
	} {
		c, f := openOrientationControls(true, true)
		displays, read := tc.displays, tc.read
		c.readDisplays = func() (map[string]interface{}, error) { return displays, nil }
		c.readHingeAngle = func() (float64, error) { return read, nil }
		require.NoError(t, c.UnstickOrientation(), tc.name)
		assert.Equal(t, tc.want, virtualizationOf(t, f.indigo.sent[0])["value"], tc.name)
	}
}

func TestUnstickOrientationKeepsTheReadWithoutTheDisplays(t *testing.T) {
	c, f := openOrientationControls(true, true)
	c.readHingeAngle = func() (float64, error) { return 130, nil }
	c.readDisplays = func() (map[string]interface{}, error) { return nil, errors.New("no displays") }
	require.NoError(t, c.UnstickOrientation())
	assert.Equal(t, 130.0, virtualizationOf(t, f.indigo.sent[0])["value"])
}
