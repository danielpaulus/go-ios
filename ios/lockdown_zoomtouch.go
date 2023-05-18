package ios

import log "github.com/sirupsen/logrus"
import "fmt"

const zoomTouchKey = "ZoomTouchEnabledByiTunes"

// EnableZoomTouch creates a new lockdown session for the device and enables or disables
// ZoomTouch (the on-screen software home button), by using the special key ZoomTouchEnabledByiTunes.
// Setting to true will enable ZoomTouch
// Setting to false will disable ZoomTouch, regardless of whether it was previously-enabled through
// a non-iTunes-related method.
func SetZoomTouch(device DeviceEntry, enabled bool) error {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return err
	}
	log.Debugf("Setting %s: %t", zoomTouchKey, enabled)
	defer lockDownConn.Close()
	err = lockDownConn.SetValueForDomain(zoomTouchKey, accessibilityDomain, enabled)
	return err
}
func GetZoomTouch(device DeviceEntry) (bool, error) {
	lockDownConn, err := ConnectLockdownWithSession(device)

	if err != nil {
		return false, err
	}
	defer lockDownConn.Close()

	enabledIntf, err := lockDownConn.GetValueForDomain(zoomTouchKey, accessibilityDomain)

	if err != nil {
		return false, err
	}
	if enabledIntf == nil {
		// In testing, nil was returned in only one case, on an iOS 14.7 device that should already have been paired.
		// Calling SetZoomTouch() directly returned the somewhat more useful error: SetProhibited
		// After re-running "go-ios pair", full functionality returned.
		return false, fmt.Errorf("Received null response when querying %s.%s. Try re-pairing the device.", accessibilityDomain, zoomTouchKey)
	}
	enabledUint64, ok := enabledIntf.(uint64)
	if !ok {
		// On iOS 10.x at least, "false" is returned, perhaps for any key at all.  Attempting to manipulate ZoomTouchEnabledByiTunes had no effect
		return false, fmt.Errorf("Expected unit64 0 or 1 when querying %s.%s, but received %T:%+v. Is this device running iOS 11+?", accessibilityDomain, zoomTouchKey, enabledIntf, enabledIntf)
	} else if enabledUint64 != 0 && enabledUint64 != 1 {
		// So far this has never happened
		return false, fmt.Errorf("Expected a value of 0 or 1 for %s.%s, received %d instead!", accessibilityDomain, zoomTouchKey, enabledUint64)
	}

	return enabledUint64 == 1, nil
}
