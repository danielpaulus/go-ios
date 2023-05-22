package ios

import log "github.com/sirupsen/logrus"
import "fmt"

const voiceOverTouchKey = "VoiceOverTouchEnabledByiTunes"

// EnableVoiceOver creates a new lockdown session for the device and enables or disables
// VoiceOver (the on-screen software home button), by using the special key VoiceOverTouchEnabledByiTunes.
// Setting to true will enable VoiceOver
// Setting to false will disable VoiceOver, regardless of whether it was previously-enabled through
// a non-iTunes-related method.
func SetVoiceOver(device DeviceEntry, enabled bool) error {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return err
	}
	log.Debugf("Setting %s: %t", voiceOverTouchKey, enabled)
	defer lockDownConn.Close()
	err = lockDownConn.SetValueForDomain(voiceOverTouchKey, accessibilityDomain, enabled)
	return err
}
func GetVoiceOver(device DeviceEntry) (bool, error) {
	lockDownConn, err := ConnectLockdownWithSession(device)

	if err != nil {
		return false, err
	}
	defer lockDownConn.Close()

	enabledIntf, err := lockDownConn.GetValueForDomain(voiceOverTouchKey, accessibilityDomain)

	if err != nil {
		return false, err
	}
	if enabledIntf == nil {
		// In testing, nil was returned in only one case, on an iOS 14.7 device that should already have been paired.
		// Calling SetVoiceOver() directly returned the somewhat more useful error: SetProhibited
		// After re-running "go-ios pair", full functionality returned.
		return false, fmt.Errorf("Received null response when querying %s.%s. Try re-pairing the device.", accessibilityDomain, voiceOverTouchKey)
	}
	enabledUint64, ok := enabledIntf.(uint64)
	if !ok {
		// On iOS 10.x at least, "false" is returned, perhaps for any key at all.  Attempting to manipulate VoiceOverTouchEnabledByiTunes had no effect
		return false, fmt.Errorf("Expected unit64 0 or 1 when querying %s.%s, but received %T:%+v. Is this device running iOS 11+?", accessibilityDomain, voiceOverTouchKey, enabledIntf, enabledIntf)
	} else if enabledUint64 != 0 && enabledUint64 != 1 {
		// So far this has never happened
		return false, fmt.Errorf("Expected a value of 0 or 1 for %s.%s, received %d instead!", accessibilityDomain, voiceOverTouchKey, enabledUint64)
	}

	return enabledUint64 == 1, nil
}
