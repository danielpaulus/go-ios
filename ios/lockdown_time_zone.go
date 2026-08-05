package ios

import "time"

// SetTime sets the device clock and timezone via the lockdown protocol.
// timezone must be an IANA timezone database name (e.g. "America/Chicago").
// unixTimestamp is seconds since the Unix epoch.
// Note: on iOS 14+ the timezone write is only honoured during the
// setup-assistant phase; the clock sync works at any time.
func SetTime(device DeviceEntry, timezone string, unixTimestamp int64) error {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return err
	}
	defer lockDownConn.Close()
	err = lockDownConn.SetValueForDomain("TimeIntervalSince1970", "", unixTimestamp)
	if err != nil {
		return err
	}
	return lockDownConn.SetValueForDomain("TimeZone", "", timezone)
}

// SetSystemTime syncs the device clock and timezone to the host system values.
func SetSystemTime(device DeviceEntry) error {
	return SetTime(device, SystemTimezone(), time.Now().Unix())
}
