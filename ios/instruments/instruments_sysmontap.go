package instruments

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

const sysmontapName = "com.apple.instruments.server.services.sysmontap"

type SysmontapService struct {
	channel    *dtx.Channel
	conn       *dtx.Connection
	plistCodec ios.PlistCodec
}

// Creates a new SysmontapService
func NewSysmontapService(device ios.DeviceEntry) (*SysmontapService, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(sysmontapName, loggingDispatcher{dtxConn})

	return &SysmontapService{channel: processControlChannel, conn: dtxConn, plistCodec: ios.NewPlistCodec()}, nil
}

// Close closes up the DTX connection
func (s *SysmontapService) Close() {
	s.conn.Close()
}

// Start sends a start method call async and waits until the cpu info & stats come back
func (s *SysmontapService) Start() (SysmontapMessage, error) {
	err := s.channel.MethodCallAsync("start")
	if err != nil {
		return SysmontapMessage{}, err
	}

	globalChannel := s.conn.GlobalChannel()

	// Receive() will block until the message with the CPU usage is delivered
	msg, err := globalChannel.Receive()
	if err != nil {
		return SysmontapMessage{}, err
	}

	sysmontapMessage, err := mapToCPUUsage(msg)
	if err != nil {
		return SysmontapMessage{}, err
	}

	return sysmontapMessage, nil
}

// SetConfig sets configuration to allow the sysmontap service getting desired data points
func (s *SysmontapService) SetConfig(procAttrs, sysAttrs []interface{}) error {
	config := map[string]interface{}{
		"ur":             500,
		"bm":             0,
		"procAttrs":      procAttrs,
		"sysAttrs":       sysAttrs,
		"cpuUsage":       true,
		"physFootprint":  true,
		"sampleInterval": 500000000,
	}

	_, err := s.channel.MethodCall("setConfig:", config)

	if err != nil {
		return err
	}

	return nil
}

type SysmontapMessage struct {
	CPUCount       uint64
	EnabledCPUs    uint64
	EndMachAbsTime uint64
	Type           uint64
	SystemCPUUsage CPUUsage
}

type CPUUsage struct {
	CPU_TotalLoad float64
}

func mapToCPUUsage(msg dtx.Message) (SysmontapMessage, error) {
	payload := msg.Payload
	if len(payload) != 1 {
		return SysmontapMessage{}, fmt.Errorf("payload of message should have only one element: %+v", msg)
	}

	resultArray, ok := payload[0].([]interface{})
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected resultArray of type []interface{}: %+v", payload[0])
	}
	resultMap, ok := resultArray[0].(map[string]interface{})
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected resultMap of type map[string]interface{} as a single element of resultArray: %+v", resultArray[0])
	}
	cpuCount, ok := resultMap["CPUCount"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected CPUCount of type uint64 of resultMap: %+v", resultMap)
	}
	enabledCPUs, ok := resultMap["EnabledCPUs"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected EnabledCPUs of type uint64 of resultMap: %+v", resultMap)
	}
	endMachAbsTime, ok := resultMap["EndMachAbsTime"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected EndMachAbsTime of type uint64 of resultMap: %+v", resultMap)
	}
	typ, ok := resultMap["Type"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected Type of type uint64 of resultMap: %+v", resultMap)
	}
	sysmontapMessageMap, ok := resultMap["SystemCPUUsage"].(map[string]interface{})
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected SystemCPUUsage of type map[string]interface{} of resultMap: %+v", resultMap)
	}
	cpuTotalLoad, ok := sysmontapMessageMap["CPU_TotalLoad"].(float64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected CPU_TotalLoad of type uint64 of sysmontapMessageMap: %+v", sysmontapMessageMap)
	}
	cpuUsage := CPUUsage{CPU_TotalLoad: cpuTotalLoad}

	sysmontapMessage := SysmontapMessage{
		cpuCount,
		enabledCPUs,
		endMachAbsTime,
		typ,
		cpuUsage,
	}
	return sysmontapMessage, nil
}
