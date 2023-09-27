package ios

import log "github.com/sirupsen/logrus"
import "fmt"

const uses24HourClockKey = "Uses24HourClock"

// Enable24HourClock creates a new lockdown session for the device and enables or disables
// 24HourClock, by using the special key Uses24HourClock.
// Setting to true will enable 24Hour Clock
// Setting to false will disable 24Hour Clock, regardless of whether it was previously-enabled through
// a non-iTunes-related method.
func SetUses24HourClock(device DeviceEntry, enabled bool) error {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return err
	}
	log.Debugf("Setting %s: %t", uses24HourClockKey, enabled)
	defer lockDownConn.Close()
	err = lockDownConn.SetValueForDomain(uses24HourClockKey, "", enabled)
	return err
}
func GetUses24HourClock(device DeviceEntry) (bool, error) {
	lockDownConn, err := ConnectLockdownWithSession(device)

	if err != nil {
		return false, err
	}
	defer lockDownConn.Close()

	enabledBoolf, err := lockDownConn.GetValueForDomain(uses24HourClockKey, "")

	if err != nil {
		return false, err
	}
	if enabledBoolf == nil {
		// In testing, nil was returned in only one case, on an iOS 14.7 device that should already have been paired.
		// Calling SetUses24HourClock() directly returned the somewhat more useful error: SetProhibited
		// After re-running "go-ios pair", full functionality returned.
		return false, fmt.Errorf("Received null response when querying %s.%s. Try re-pairing the device.", "", uses24HourClockKey)
	}
	enabledBool, ok := enabledBoolf.(bool)
	if !ok {
		return false, fmt.Errorf("Expected bool false or true when querying %s.%s, but received %T:%+v. Is this device running iOS 11+?", "", uses24HourClockKey, enabledBoolf, enabledBoolf)
	}

	return enabledBool, nil
}
