package ios

type MemoryInfo struct {
	AmountDataAvailable    uint64
	AmountDataReserved     uint64
	AmountRestoreAvailable uint64
	TotalDataAvailable     uint64
	TotalDataCapacity      uint64
	TotalDiskCapacity      uint64
	TotalSystemAvailable   uint64
	TotalSystemCapacity    uint64
}

const memoryDoMain = "com.apple.disk_usage"

func GetMemoryInfo(device DeviceEntry) (MemoryInfo, error) {
	lock, err := ConnectLockdownWithSession(device)
	if err != nil {
		return MemoryInfo{}, err
	}
	defer lock.Close()

	amountDataAvailable, err := lock.GetValueForDomain("AmountDataAvailable", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}
	amountDataReserved, err := lock.GetValueForDomain("AmountDataReserved", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}
	amountRestoreAvailable, err := lock.GetValueForDomain("AmountRestoreAvailable", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}
	totalDataAvailable, err := lock.GetValueForDomain("TotalDataAvailable", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}
	totalDataCapacity, err := lock.GetValueForDomain("TotalDataCapacity", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}
	totalDiskCapacity, err := lock.GetValueForDomain("TotalDiskCapacity", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}
	totalSystemAvailable, err := lock.GetValueForDomain("TotalSystemAvailable", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}
	totalSystemCapacity, err := lock.GetValueForDomain("TotalSystemCapacity", memoryDoMain)
	if err != nil {
		return MemoryInfo{}, err
	}

	return MemoryInfo{AmountDataAvailable: amountDataAvailable.(uint64), AmountDataReserved: amountDataReserved.(uint64), AmountRestoreAvailable: amountRestoreAvailable.(uint64), TotalDataAvailable: totalDataAvailable.(uint64), TotalDataCapacity: totalDataCapacity.(uint64), TotalDiskCapacity: totalDiskCapacity.(uint64), TotalSystemAvailable: totalSystemAvailable.(uint64), TotalSystemCapacity: totalSystemCapacity.(uint64)}, nil
}
