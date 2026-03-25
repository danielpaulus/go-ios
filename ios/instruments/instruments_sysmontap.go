package instruments

import (
	"fmt"
	"reflect"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

type sysmontapMsgDispatcher struct {
	messages chan dtx.Message
}

func newSysmontapMsgDispatcher() *sysmontapMsgDispatcher {
	return &sysmontapMsgDispatcher{make(chan dtx.Message)}
}

func (p *sysmontapMsgDispatcher) Dispatch(m dtx.Message) {
	p.messages <- m
}

const sysmontapName = "com.apple.instruments.server.services.sysmontap"

type sysmontapService struct {
	channel *dtx.Channel
	conn    *dtx.Connection

	deviceInfoService *DeviceInfoService
	msgDispatcher     *sysmontapMsgDispatcher
}

// NewSysmontapService creates a new sysmontapService
// - samplingInterval is the rate how often to get samples, i.e Xcode's default is 10, which results in sampling output
// each 1 second, with 500 the samples are retrieved every 15 seconds. It doesn't make any correlation between
// the expected rate and the actual rate of samples delivery. We can only conclude, that the lower the rate in digits,
// the faster the samples are delivered
func NewSysmontapService(device ios.DeviceEntry, samplingInterval int) (*sysmontapService, error) {
	deviceInfoService, err := NewDeviceInfoService(device)
	if err != nil {
		return nil, err
	}

	msgDispatcher := newSysmontapMsgDispatcher()
	dtxConn, err := connectInstrumentsWithMsgDispatcher(device, msgDispatcher)
	if err != nil {
		return nil, err
	}

	processControlChannel := dtxConn.RequestChannelIdentifier(sysmontapName, loggingDispatcher{dtxConn})

	sysAttrs, err := deviceInfoService.systemAttributes()
	if err != nil {
		return nil, err
	}

	procAttrs, err := deviceInfoService.processAttributes()
	if err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"ur":             samplingInterval,
		"bm":             0,
		"procAttrs":      procAttrs,
		"sysAttrs":       sysAttrs,
		"cpuUsage":       true,
		"physFootprint":  true,
		"sampleInterval": 500000000,
	}
	_, err = processControlChannel.MethodCall("setConfig:", config)
	if err != nil {
		return nil, err
	}

	err = processControlChannel.MethodCallAsync("start")
	if err != nil {
		return nil, err
	}

	return &sysmontapService{processControlChannel, dtxConn, deviceInfoService, msgDispatcher}, nil
}

// Close closes up the DTX connection, message dispatcher and dtx.Message channel
func (s *sysmontapService) Close() error {
	close(s.msgDispatcher.messages)

	s.deviceInfoService.Close()
	return s.conn.Close()
}

// ReceiveCPUUsage returns a chan of SysmontapMessage with CPU Usage info
// The method will close the result channel automatically as soon as sysmontapMsgDispatcher's
// dtx.Message channel is closed.
func (s *sysmontapService) ReceiveCPUUsage() chan SysmontapMessage {
	messages := make(chan SysmontapMessage)
	go func() {
		defer close(messages)

		for msg := range s.msgDispatcher.messages {
			sysmontapMessage, err := mapToCPUUsage(msg)
			if err != nil {
				log.Debugf("expected `sysmontapMessage` from global channel, but received %v", msg)
				continue
			}

			messages <- sysmontapMessage
		}

		log.Infof("sysmontap message dispatcher channel closed")
	}()

	return messages
}

// SysmontapMessage is a wrapper struct for incoming CPU samples
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

// toUint64 converts any numeric type to uint64. Returns false for
// non-numeric types or negative values.
func toUint64(v interface{}) (uint64, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := rv.Int()
		if i < 0 {
			return 0, false
		}
		return uint64(i), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint(), true
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		if f < 0 {
			return 0, false
		}
		return uint64(f), true
	default:
		return 0, false
	}
}

// toFloat64 converts any numeric type to float64. Returns false for
// non-numeric types.
func toFloat64(v interface{}) (float64, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	default:
		return 0, false
	}
}

// requireUint64 extracts a uint64 from a map field, accepting any numeric type.
func requireUint64(m map[string]interface{}, key string) (uint64, error) {
	v, ok := toUint64(m[key])
	if !ok {
		return 0, fmt.Errorf("expected numeric %s, got %T: %+v", key, m[key], m[key])
	}
	return v, nil
}

// requireFloat64 extracts a float64 from a map field, accepting any numeric type.
func requireFloat64(m map[string]interface{}, key string) (float64, error) {
	v, ok := toFloat64(m[key])
	if !ok {
		return 0, fmt.Errorf("expected numeric %s, got %T: %+v", key, m[key], m[key])
	}
	return v, nil
}

// requireMap extracts a map[string]interface{} from a map field.
func requireMap(m map[string]interface{}, key string) (map[string]interface{}, error) {
	raw, exists := m[key]
	if !exists {
		return nil, fmt.Errorf("%s missing in result map", key)
	}
	sub, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{} for %s, got %T", key, raw)
	}
	return sub, nil
}

// extractResultMap unwraps the DTX payload into the first result map.
// Payload structure: []interface{ []interface{ map[string]interface{}, ... }, ... }
func extractResultMap(payload []interface{}) (map[string]interface{}, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty payload in sysmontap message")
	}

	resultArray, ok := payload[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected []interface{} as payload[0], got %T: %+v", payload[0], payload[0])
	}
	if len(resultArray) == 0 {
		return nil, fmt.Errorf("result array is empty in sysmontap payload: %+v", payload)
	}

	resultMap, ok := resultArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{} for result, got %T: %+v", resultArray[0], resultArray[0])
	}
	return resultMap, nil
}

// mapToCPUUsage parses a DTX sysmontap message into a SysmontapMessage.
// It tolerates numeric type variations (int, int64, uint32, float64, etc.)
// across different iOS versions and device types.
func mapToCPUUsage(msg dtx.Message) (SysmontapMessage, error) {
	resultMap, err := extractResultMap(msg.Payload)
	if err != nil {
		return SysmontapMessage{}, err
	}

	cpuCount, err := requireUint64(resultMap, "CPUCount")
	if err != nil {
		return SysmontapMessage{}, err
	}
	enabledCPUs, err := requireUint64(resultMap, "EnabledCPUs")
	if err != nil {
		return SysmontapMessage{}, err
	}
	endMachAbsTime, err := requireUint64(resultMap, "EndMachAbsTime")
	if err != nil {
		return SysmontapMessage{}, err
	}
	typ, err := requireUint64(resultMap, "Type")
	if err != nil {
		return SysmontapMessage{}, err
	}

	sysCPUMap, err := requireMap(resultMap, "SystemCPUUsage")
	if err != nil {
		return SysmontapMessage{}, err
	}
	cpuTotalLoad, err := requireFloat64(sysCPUMap, "CPU_TotalLoad")
	if err != nil {
		return SysmontapMessage{}, err
	}

	return SysmontapMessage{
		CPUCount:       cpuCount,
		EnabledCPUs:    enabledCPUs,
		EndMachAbsTime: endMachAbsTime,
		Type:           typ,
		SystemCPUUsage: CPUUsage{CPU_TotalLoad: cpuTotalLoad},
	}, nil
}
