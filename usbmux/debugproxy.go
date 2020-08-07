package usbmux

type DebugProxy struct{}

func NewDebugProxy() *DebugProxy {
	return &DebugProxy{}
}

func (d *DebugProxy) Launch() error {
	NewUsbMuxServerConnection()
	return nil
}

func (d *DebugProxy) Close() {}
