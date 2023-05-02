package ios

import "time"

func SetTime(device DeviceEntry, timeZone string, time int64) error {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return err
	}
	defer lockDownConn.Close()
	err = lockDownConn.SetValueForDomain("TimeIntervalSince1970", "", time)
	if err != nil {
		return err
	}
	return lockDownConn.SetValueForDomain("TimeZone", "", timeZone)
}

func SetSystemTime(device DeviceEntry) error {
	t := time.Now().Unix()
	return SetTime(device, "Europe/Berlin", t)
}
