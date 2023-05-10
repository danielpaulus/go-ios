package ios

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	accessibilityDomain = "com.apple.Accessibility"
	assistiveTouchKey   = "AssistiveTouchEnabledByiTunes"
)

// EnableAssistiveTouch creates a new lockdown session for the device and enables or disables
// AssistiveTouch (the on-screen software home button), by using the special key AssistiveTouchEnabledByiTunes.
// Setting to true will enable AssistiveTouch
// Setting to false will disable AssistiveTouch, regardless of whether it was previously-enabled through
// a non-iTunes-related method.
func SetAssistiveTouch(device DeviceEntry, enabled bool) error {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return err
	}
	log.Debugf("Setting %s: %t", assistiveTouchKey, enabled)
	defer lockDownConn.Close()
	err = lockDownConn.SetValueForDomain(assistiveTouchKey, accessibilityDomain, enabled)
	return err
}

func GetAssistiveTouch(device DeviceEntry) (bool, error) {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return false, err
	}
	defer lockDownConn.Close()

	enabledIntf, err := lockDownConn.GetValueForDomain(assistiveTouchKey, accessibilityDomain)
	if err != nil {
		return false, err
	}
	if enabledIntf == nil {
		// In testing, nil was returned in only one case, on an iOS 14.7 device that should already have been paired.
		// Calling SetAssistiveTouch() directly returned the somewhat more useful error: SetProhibited
		// After re-running "go-ios pair", full functionality returned.
		return false, fmt.Errorf("Received null response when querying %s.%s. Try re-pairing the device.", accessibilityDomain, assistiveTouchKey)
	}
	enabledUint64, ok := enabledIntf.(uint64)
	if !ok {
		// On iOS 10.x at least, "false" is returned, perhaps for any key at all.  Attempting to manipulate AssistiveTouchEnabledByiTunes had no effect
		return false, fmt.Errorf("Expected unit64 0 or 1 when querying %s.%s, but received %T:%+v. Is this device running iOS 11+?", accessibilityDomain, assistiveTouchKey, enabledIntf, enabledIntf)
	} else if enabledUint64 != 0 && enabledUint64 != 1 {
		// So far this has never happened
		return false, fmt.Errorf("Expected a value of 0 or 1 for %s.%s, received %d instead!", accessibilityDomain, assistiveTouchKey, enabledUint64)
	}

	return enabledUint64 == 1, nil
}
