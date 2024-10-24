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

	resultArray := payload[0].([]interface{})
	resultMap := resultArray[0].(map[string]interface{})
	cpuCount := resultMap["CPUCount"].(uint64)
	enabledCPUs := resultMap["EnabledCPUs"].(uint64)
	endMachAbsTime := resultMap["EndMachAbsTime"].(uint64)
	typ := resultMap["Type"].(uint64)
	sysmontapMessageMap := resultMap["SystemCPUUsage"].(map[string]interface{})
	cpuTotalLoad := sysmontapMessageMap["CPU_TotalLoad"].(float64)
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
