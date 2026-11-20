package instruments

import (
	"context"
	"errors"
	"testing"

	"github.com/danielpaulus/go-ios/ios/screencapture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func displayEntry(id string, active, external, primary bool) map[string]interface{} {
	return map[string]interface{}{
		"uniqueId": id, "active": active, "external": external, "primary": primary,
	}
}

func TestActiveDisplayID(t *testing.T) {
	for _, tt := range []struct {
		name     string
		displays []interface{}
		want     string
	}{
		{"unfolded", []interface{}{displayEntry("outer", false, false, true), displayEntry("inner", true, false, false)}, "inner"},
		{"folded", []interface{}{displayEntry("outer", true, false, true), displayEntry("inner", false, false, false)}, "outer"},
		{"external ignored", []interface{}{displayEntry("external", true, true, false), displayEntry("inner", true, false, false)}, "inner"},
		{"primary preferred", []interface{}{displayEntry("inner", true, false, false), displayEntry("outer", true, false, true)}, "outer"},
		{"locked", []interface{}{displayEntry("outer", false, false, true), displayEntry("inner", false, false, false)}, ""},
		{"missing id", []interface{}{displayEntry("", true, false, true), displayEntry("inner", true, false, false)}, "inner"},
		{"malformed entry", []interface{}{nil, "invalid", map[string]interface{}{}}, ""},
		{"no displays", nil, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, activeDisplayID(map[string]interface{}{"displays": tt.displays}))
		})
	}
}

type fakeDisplayInfo struct {
	info   map[string]interface{}
	err    error
	closed bool
}

func (f *fakeDisplayInfo) GetDisplayInfo() (map[string]interface{}, error) { return f.info, f.err }

func (f *fakeDisplayInfo) Close() error {
	f.closed = true
	return nil
}

type fakeScreenshotCapturer struct {
	ids    []string
	closed bool
}

func (f *fakeScreenshotCapturer) TakeScreenshot(_ context.Context, opts screencapture.ScreenshotOptions) ([]byte, error) {
	f.ids = append(f.ids, opts.DisplayUniqueID)
	return []byte("png"), nil
}
func (f *fakeScreenshotCapturer) Close() error {
	f.closed = true
	return nil
}

func TestScreenshotFollowsActiveDisplay(t *testing.T) {
	outer := displayEntry("outer", false, false, true)
	inner := displayEntry("inner", true, false, false)
	info := &fakeDisplayInfo{info: map[string]interface{}{"displays": []interface{}{outer, inner}}}
	capture := &fakeScreenshotCapturer{}
	s := &ScreenshotService{capture: capture, displayInfo: info}

	image, err := s.TakeScreenshot()
	require.NoError(t, err)
	assert.Equal(t, []byte("png"), image)

	outer["active"], inner["active"] = true, false
	_, err = s.TakeScreenshot()
	require.NoError(t, err)
	assert.Equal(t, []string{"inner", "outer"}, capture.ids)

	s.Close()
	assert.True(t, capture.closed)
	assert.True(t, info.closed)
}

func TestScreenshotDisplayInfoError(t *testing.T) {
	cause := errors.New("display info unavailable")
	capture := &fakeScreenshotCapturer{}
	s := &ScreenshotService{capture: capture, displayInfo: &fakeDisplayInfo{err: cause}}
	_, err := s.TakeScreenshot()
	assert.ErrorIs(t, err, cause)
	assert.Empty(t, capture.ids, "do not silently capture the wrong display")
}
