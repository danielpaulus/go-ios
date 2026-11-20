package instruments

import (
	"fmt"
	"sort"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

// ProcessSample is one process from a sysmontap sample.
type ProcessSample struct {
	Pid  uint64
	Name string
	// Attributes holds every per-process attribute the device reported, keyed
	// by its sysmonProcessAttributes name (e.g. "cpuUsage", "physFootprint",
	// "memResidentSize", "threadCount", "pid", "name"). The set depends on the
	// iOS version. Values the device left unset are nil; in the first sample of
	// a session that includes "cpuUsage", which needs a previous sample to
	// diff against.
	Attributes map[string]interface{}
}

// SysmontapProcessSnapshot holds every process reported in one sysmontap
// sample, sorted by pid.
type SysmontapProcessSnapshot struct {
	StartMachAbsTime uint64
	EndMachAbsTime   uint64
	Processes        []ProcessSample
}

// ReceiveProcessSamples returns a chan with one per-process snapshot per
// sysmontap sample (CPU and memory per pid, like Xcode's Activity Monitor).
// It consumes the same message stream as ReceiveCPUUsage, so use only one of
// the two on a service. The channel is closed when the service is closed.
func (s *sysmontapService) ReceiveProcessSamples() chan SysmontapProcessSnapshot {
	snapshots := make(chan SysmontapProcessSnapshot)
	go func() {
		defer close(snapshots)

		for msg := range s.msgDispatcher.messages {
			snapshot, ok, err := mapToProcessSnapshot(msg, s.processAttributes)
			if err != nil {
				golog.Debug("failed decoding sysmontap process sample", "module", logModule, "error", err)
				continue
			}
			if !ok {
				continue
			}
			snapshots <- snapshot
		}
	}()

	return snapshots
}

// sysmontapRows returns the sample rows carried by a sysmontap message, or nil
// for messages without rows (tap heartbeats and control messages).
//
// System rows arrive as regular archived payloads, but the per-process rows
// arrive in DTX data messages (message type 1). The dtx codec passes the
// payload of those through as the raw NSKeyedArchiver bytes, so they are
// unarchived here.
func sysmontapRows(msg dtx.Message) ([]map[string]interface{}, error) {
	if len(msg.Payload) == 0 {
		return nil, nil
	}
	payload := msg.Payload[0]
	if raw, ok := payload.([]byte); ok {
		decoded, err := nskeyedarchiver.Unarchive(raw)
		if err != nil {
			return nil, fmt.Errorf("unarchiving sysmontap data message: %w", err)
		}
		if len(decoded) == 0 {
			return nil, nil
		}
		payload = decoded[0]
	}

	list, ok := payload.([]interface{})
	if !ok {
		return nil, nil
	}
	rows := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		if row, ok := item.(map[string]interface{}); ok {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

// mapToProcessSnapshot decodes the "Processes" row of a sysmontap message.
// ok is false if the message carries no process row.
func mapToProcessSnapshot(msg dtx.Message, attributes []string) (snapshot SysmontapProcessSnapshot, ok bool, err error) {
	rows, err := sysmontapRows(msg)
	if err != nil {
		return SysmontapProcessSnapshot{}, false, err
	}

	for _, row := range rows {
		processes, found := row["Processes"].(map[string]interface{})
		if !found {
			continue
		}

		snapshot.StartMachAbsTime, _ = toUint64(row["StartMachAbsTime"])
		snapshot.EndMachAbsTime, _ = toUint64(row["EndMachAbsTime"])
		snapshot.Processes = make([]ProcessSample, 0, len(processes))
		for _, raw := range processes {
			values, isList := raw.([]interface{})
			if !isList {
				return SysmontapProcessSnapshot{}, false, fmt.Errorf("expected []interface{} process values, got %T", raw)
			}
			process, err := mapToProcessSample(values, attributes)
			if err != nil {
				return SysmontapProcessSnapshot{}, false, err
			}
			snapshot.Processes = append(snapshot.Processes, process)
		}
		sort.Slice(snapshot.Processes, func(i, j int) bool {
			return snapshot.Processes[i].Pid < snapshot.Processes[j].Pid
		})
		return snapshot, true, nil
	}

	return SysmontapProcessSnapshot{}, false, nil
}

// mapToProcessSample pairs one process's value array with the attribute names
// that were sent as procAttrs; the device reports the values in that order.
func mapToProcessSample(values []interface{}, attributes []string) (ProcessSample, error) {
	if len(values) != len(attributes) {
		return ProcessSample{}, fmt.Errorf("process has %d values for %d requested attributes", len(values), len(attributes))
	}

	sample := ProcessSample{Attributes: make(map[string]interface{}, len(values))}
	for i, name := range attributes {
		value := values[i]
		if _, isNull := value.(nskeyedarchiver.NSNull); isNull {
			value = nil
		}
		sample.Attributes[name] = value
	}

	pid, ok := toUint64(sample.Attributes["pid"])
	if !ok {
		return ProcessSample{}, fmt.Errorf("expected numeric pid in process sample, got %T", sample.Attributes["pid"])
	}
	sample.Pid = pid
	sample.Name, _ = sample.Attributes["name"].(string)
	return sample, nil
}

// attributeNames converts the attribute list returned by the deviceinfo
// service (sysmonProcessAttributes) to strings, keeping positions intact.
func attributeNames(attributes []interface{}) []string {
	names := make([]string, len(attributes))
	for i, attribute := range attributes {
		names[i], _ = attribute.(string)
	}
	return names
}
