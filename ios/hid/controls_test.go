package hid

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMessages struct {
	sent []map[string]interface{}
}

func (f *fakeMessages) send(msg map[string]interface{}) error {
	f.sent = append(f.sent, msg)
	return nil
}

func (f *fakeMessages) Close() error { return nil }

func openControls(foldable bool) (*Controls, *fakeMessages) {
	m := &fakeMessages{}
	return &Controls{indigo: m, foldable: foldable}, m
}

func payloadOf(msg map[string]interface{}) map[string]interface{} {
	return msg["payload"].(map[string]interface{})
}

func TestSetHingeAngleSendsTheHingeControl(t *testing.T) {
	c, m := openControls(true)
	require.NoError(t, c.SetHingeAngle(HingeHalfOpen))

	require.Len(t, m.sent, 1)
	assert.Equal(t, "com.apple.coredevice.feature.remote.hid.vendordefined", m.sent[0]["featureIdentifier"])
	assert.Equal(t, "IndigoVendorDefinedEvent", m.sent[0]["messageType"])
	p := payloadOf(m.sent[0])
	assert.Equal(t, uint64(0xff61), p["usagePage"])
	assert.Equal(t, uint64(0x5b), p["usage"])
	assert.Equal(t, uint64(0), p["version"])
	assert.Equal(t, map[string]interface{}{
		"provider": "com.apple.Virtualization.VirtualMachines",
		"source":   "hinge-slider-control",
		"type":     "range",
		"value":    130.0,
	}, decodeVirtualization(t, p["data"].([]byte)))
}

func TestSetHingeAngleRefusesWhatCannotFold(t *testing.T) {
	c, m := openControls(false)
	assert.ErrorIs(t, c.SetHingeAngle(HingeClosed), ErrNotFoldable)

	c, m = openControls(true)
	assert.Error(t, c.SetHingeAngle(-1))
	assert.Error(t, c.SetHingeAngle(181))
	assert.Error(t, c.SetHingeAngle(math.NaN()))
	assert.Empty(t, m.sent)
}

func TestClosedControlsSendNothing(t *testing.T) {
	c, m := openControls(true)
	require.NoError(t, c.Close())
	require.NoError(t, c.Close(), "idempotent")
	assert.Error(t, c.SetHingeAngle(HingeOpenFlat))
	assert.Empty(t, m.sent)
}

func TestCloseWaitsForTheLastEventToLand(t *testing.T) {
	c, _ := openControls(true)
	c.grace = 100 * time.Millisecond
	require.NoError(t, c.SetHingeAngle(HingeOpenFlat))
	start := time.Now()
	require.NoError(t, c.Close())
	assert.GreaterOrEqual(t, time.Since(start), 90*time.Millisecond)

	c, _ = openControls(true)
	c.grace = time.Second
	start = time.Now()
	require.NoError(t, c.Close())
	assert.Less(t, time.Since(start), 100*time.Millisecond, "nothing was sent, so nothing to wait for")
}

func TestCountBuiltInDisplays(t *testing.T) {
	duo := map[string]interface{}{"displays": []interface{}{
		map[string]interface{}{"displayId": uint64(1), "active": true},
		map[string]interface{}{"displayId": uint64(2), "active": false},
		map[string]interface{}{"displayId": uint64(3), "external": true},
	}}
	phone := map[string]interface{}{"displays": []interface{}{
		map[string]interface{}{"displayId": uint64(1), "active": true},
	}}
	assert.Equal(t, 2, countBuiltInDisplays(duo))
	assert.Equal(t, 1, countBuiltInDisplays(phone))
	assert.Equal(t, 0, countBuiltInDisplays(map[string]interface{}{}))
}

// Data the device hub sent to an iPhone Duo. Its keys come in no fixed order.
const capturedHinge = "d3000000040000810900000870726f76696465720000000028000009636f6d2e6170706c652e5669727475616c697a6174696f6e2e5669727475616c4d616368696e657307000008736f7572636500001400000968696e67652d736c696465722d636f6e74726f6c0600000876616c75650000003f00000445aef7ffff7f66400500000874797065000000000500008972616e6765000000"

func TestEncodingMatchesTheDeviceHub(t *testing.T) {
	hinge, err := hex.DecodeString(capturedHinge)
	require.NoError(t, err)
	captured := decodeVirtualization(t, hinge)
	ours := encodeVirtualizationControl(hingeControl, "range", captured["value"].(float64))
	assert.Equal(t, captured, decodeVirtualization(t, ours))
	assert.Len(t, ours, len(hinge))
	assert.InDelta(t, 180, captured["value"], 0.01)
}

// decodeVirtualization reads the dictionary encodeVirtualizationControl
// writes, the way the device would, so a test can compare it with captured
// data whatever order the keys come in.
func decodeVirtualization(t *testing.T, b []byte) map[string]interface{} {
	t.Helper()
	require.Equal(t, virtualizationHeader, b[:8])
	out := map[string]interface{}{}
	off := 8
	read := func() (byte, interface{}) {
		require.LessOrEqual(t, off+4, len(b), "truncated at %d", off)
		n := int(b[off]) | int(b[off+1])<<8 | int(b[off+2])<<16
		typ := b[off+3]
		off += 4
		var v interface{}
		switch typ &^ virtualizationLast {
		case virtualizationDouble:
			v = math.Float64frombits(binary.LittleEndian.Uint64(b[off:]))
			off += 8
		case virtualizationKey:
			v = string(b[off : off+n-1])
			off += n
		case virtualizationString:
			v = string(b[off : off+n])
			off += n
		default:
			require.Fail(t, fmt.Sprintf("unknown type 0x%x at %d", typ, off-4))
		}
		for off%4 != 0 {
			off++
		}
		return typ, v
	}
	for {
		_, key := read()
		typ, value := read()
		out[key.(string)] = value
		if typ&virtualizationLast != 0 {
			break
		}
	}
	assert.Equal(t, len(b), off, "trailing bytes")
	return out
}
