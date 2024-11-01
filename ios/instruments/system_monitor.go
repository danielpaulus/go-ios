package instruments

import "github.com/danielpaulus/go-ios/ios"

type systemMonitor struct {
	deviceInfoService *DeviceInfoService
	sysmontapService  *sysmontapService
}

// NewSystemMonitor creates a new instance of systemMonitor
func NewSystemMonitor(device ios.DeviceEntry) (*systemMonitor, error) {
	deviceInfoService, err := NewDeviceInfoService(device)
	if err != nil {
		return nil, err
	}
	sysmontapService, err := newSysmontapService(device)
	if err != nil {
		return nil, err
	}

	return &systemMonitor{deviceInfoService, sysmontapService}, nil
}

func (s *systemMonitor) Close() error {
	s.deviceInfoService.Close()
	return s.sysmontapService.Close()
}

// GetCPUUsage send a request to get CPU usage data and waits until the data back
func (s *systemMonitor) GetCPUUsage() (SysmontapMessage, error) {
	sysAttrs, err := s.deviceInfoService.systemAttributes()
	if err != nil {
		return SysmontapMessage{}, err
	}

	procAttrs, err := s.deviceInfoService.processAttributes()
	if err != nil {
		return SysmontapMessage{}, err
	}

	err = s.sysmontapService.setConfig(procAttrs, sysAttrs)
	if err != nil {
		return SysmontapMessage{}, err
	}

	sysmontapMsg, err := s.sysmontapService.start()
	if err != nil {
		return SysmontapMessage{}, err
	}

	return sysmontapMsg, nil
}
