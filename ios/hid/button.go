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

// Button is a hardware button, as the device hub presses it.
type Button int

const (
	// ButtonHome is the home button. Two presses about 150 ms apart open the app
	// switcher.
	ButtonHome Button = iota
	ButtonVolumeUp
	ButtonVolumeDown
	ButtonLock
	// ButtonCameraControl is the Camera Control. Whether one press opens the
	// camera or two do, about 150 ms apart, is a setting on the device.
	ButtonCameraControl
	// ButtonAction is the Action button, which only acts when held for about a
	// second.
	ButtonAction
)

// buttonUsage is how the device hub presses a button: its HID usage page and
// code, and how long it holds it down for a click.
type buttonUsage struct {
	name      string
	usagePage uint64
	usageCode uint64
	hold      time.Duration
}

func (b Button) usage() (buttonUsage, bool) {
	switch b {
	case ButtonHome:
		return buttonUsage{"home", 0x0c, 0x40, time.Millisecond}, true
	case ButtonVolumeUp:
		return buttonUsage{"volumeup", 0x0c, 0xe9, 250 * time.Millisecond}, true
	case ButtonVolumeDown:
		return buttonUsage{"volumedown", 0x0c, 0xea, 250 * time.Millisecond}, true
	case ButtonLock:
		return buttonUsage{"lock", 0x0c, 0x30, 250 * time.Millisecond}, true
	case ButtonCameraControl:
		return buttonUsage{"cameracontrol", 0xff00, 0x66, 250 * time.Millisecond}, true
	case ButtonAction:
		return buttonUsage{"action", 0x0b, 0x2d, time.Second}, true
	}
	return buttonUsage{}, false
}

// String is the button's name, as ParseButton takes it.
func (b Button) String() string {
	if u, ok := b.usage(); ok {
		return u.name
	}
	return fmt.Sprintf("Button(%d)", int(b))
}

// Buttons lists every button.
func Buttons() []Button {
	return []Button{ButtonHome, ButtonVolumeUp, ButtonVolumeDown, ButtonLock, ButtonCameraControl, ButtonAction}
}

// ParseButton finds a button by name, ignoring case.
func ParseButton(name string) (Button, error) {
	names := make([]string, 0, len(Buttons()))
	for _, b := range Buttons() {
		if strings.EqualFold(b.String(), name) {
			return b, nil
		}
		names = append(names, b.String())
	}
	return 0, fmt.Errorf("unknown button %q, expected one of %s", name, strings.Join(names, ", "))
}

// PressButton presses the button and lets it go after hold, or after the
// button's own hold when hold is zero. It returns once the button is up.
// A device without the button ignores it.
func (c *Controls) PressButton(b Button, hold time.Duration) error {
	u, ok := b.usage()
	if !ok {
		return fmt.Errorf("PressButton: unknown button %d", int(b))
	}
	if hold <= 0 {
		hold = u.hold
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.send(buildButtonPayload(u, buttonDown)); err != nil {
		return fmt.Errorf("PressButton: failed to press %s: %w", u.name, err)
	}
	time.Sleep(hold)
	if err := c.send(buildButtonPayload(u, buttonUp)); err != nil {
		return fmt.Errorf("PressButton: failed to release %s: %w", u.name, err)
	}
	return nil
}

func buildButtonPayload(u buttonUsage, state buttonState) map[string]interface{} {
	return map[string]interface{}{
		"featureIdentifier": buttonFeatureIdentifier,
		"messageType":       buttonMessageType,
		"payload": map[string]interface{}{
			"usagePage": u.usagePage,
			"usageCode": u.usageCode,
			"state":     uint64(state),
		},
	}
}

// The ring/silent switch, on iPhones that have one instead of an Action button.
// The device hub sends one event for each position the switch is moved to,
// rather than a press and a release. The position belongs to the connection:
// once it closes, the device goes by its physical switch again.
const (
	muteSwitchUsagePage uint64 = 0x0b
	muteSwitchUsageCode uint64 = 0x2e

	muteSwitchRing   buttonState = 1
	muteSwitchSilent buttonState = 2
)

// SetMuteSwitch moves the ring/silent switch: to silent when muted is true,
// to ring otherwise. It holds there while these Controls stay open; closing them
// hands the ringer back to the physical switch. A device without the switch,
// such as one with an Action button, ignores it.
func (c *Controls) SetMuteSwitch(muted bool) error {
	state := muteSwitchRing
	if muted {
		state = muteSwitchSilent
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	u := buttonUsage{name: "muteswitch", usagePage: muteSwitchUsagePage, usageCode: muteSwitchUsageCode}
	if err := c.send(buildButtonPayload(u, state)); err != nil {
		return fmt.Errorf("SetMuteSwitch: %w", err)
	}
	return nil
}
