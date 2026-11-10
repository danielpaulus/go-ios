package hid

import (
	"fmt"
	"strings"
	"time"
)

const (
	buttonFeatureIdentifier = "com.apple.coredevice.feature.remote.hid.button"
	buttonMessageType       = "IndigoButtonEvent"
)

type buttonState uint64

const (
	buttonDown buttonState = 1
	buttonUp   buttonState = 2
)

// Button is a hardware button, as HID usage page and code.
type Button struct {
	Name      string
	usagePage uint64
	usageCode uint64
	// How long the device hub holds it down for a click.
	hold time.Duration
}

var (
	// ButtonHome is the home button. Two presses about 160 ms apart open the app
	// switcher.
	ButtonHome       = Button{Name: "home", usagePage: 0x0c, usageCode: 0x40, hold: time.Millisecond}
	ButtonVolumeUp   = Button{Name: "volumeup", usagePage: 0x0c, usageCode: 0xe9, hold: 250 * time.Millisecond}
	ButtonVolumeDown = Button{Name: "volumedown", usagePage: 0x0c, usageCode: 0xea, hold: 250 * time.Millisecond}
	ButtonLock       = Button{Name: "lock", usagePage: 0x0c, usageCode: 0x30, hold: 250 * time.Millisecond}
	// ButtonCameraControl is the Camera Control. Whether one press opens the
	// camera or two do, about 150 ms apart, is a setting on the device.
	ButtonCameraControl = Button{Name: "cameracontrol", usagePage: 0xff00, usageCode: 0x66, hold: 250 * time.Millisecond}
	// ButtonAction is the Action button, which only acts when held for about a
	// second.
	ButtonAction = Button{Name: "action", usagePage: 0x0b, usageCode: 0x2d, hold: time.Second}
)

// Buttons lists every button, for looking one up by name.
var Buttons = []Button{ButtonHome, ButtonVolumeUp, ButtonVolumeDown, ButtonLock, ButtonCameraControl, ButtonAction}

// ParseButton finds a button by name, ignoring case.
func ParseButton(name string) (Button, error) {
	for _, b := range Buttons {
		if strings.EqualFold(b.Name, name) {
			return b, nil
		}
	}
	names := make([]string, len(Buttons))
	for i, b := range Buttons {
		names[i] = b.Name
	}
	return Button{}, fmt.Errorf("unknown button %q, expected one of %s", name, strings.Join(names, ", "))
}

// PressButton presses the button and lets it go after hold, or after the
// button's own hold when hold is zero. It returns once the button is up.
// A device without the button ignores it.
func (c *Controls) PressButton(b Button, hold time.Duration) error {
	if hold <= 0 {
		hold = b.hold
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.send(buildButtonPayload(b, buttonDown)); err != nil {
		return fmt.Errorf("PressButton: failed to press %s: %w", b.Name, err)
	}
	time.Sleep(hold)
	if err := c.send(buildButtonPayload(b, buttonUp)); err != nil {
		return fmt.Errorf("PressButton: failed to release %s: %w", b.Name, err)
	}
	return nil
}

func buildButtonPayload(b Button, state buttonState) map[string]interface{} {
	return map[string]interface{}{
		"featureIdentifier": buttonFeatureIdentifier,
		"messageType":       buttonMessageType,
		"payload": map[string]interface{}{
			"usagePage": b.usagePage,
			"usageCode": b.usageCode,
			"state":     uint64(state),
		},
	}
}
