package instruments

import (
	"testing"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to build a dtx.Message with the standard sysmontap payload structure.
func buildSysmontapMsg(resultMap map[string]interface{}) dtx.Message {
	return dtx.Message{
		Payload: []interface{}{
			[]interface{}{resultMap},
		},
	}
}

func validResultMap() map[string]interface{} {
	return map[string]interface{}{
		"CPUCount":       uint64(8),
		"EnabledCPUs":    uint64(8),
		"EndMachAbsTime": uint64(123456789),
		"Type":           uint64(33),
		"SystemCPUUsage": map[string]interface{}{
			"CPU_TotalLoad": float64(25.5),
		},
	}
}

func TestToUint64_NativeTypes(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want uint64
		ok   bool
	}{
		{"uint64", uint64(42), 42, true},
		{"uint32", uint32(42), 42, true},
		{"uint", uint(42), 42, true},
		{"int positive", int(42), 42, true},
		{"int zero", int(0), 0, true},
		{"int negative", int(-1), 0, false},
		{"int32 positive", int32(42), 42, true},
		{"int32 negative", int32(-1), 0, false},
		{"int64 positive", int64(42), 42, true},
		{"int64 negative", int64(-1), 0, false},
		{"float64 positive", float64(42.0), 42, true},
		{"float64 negative", float64(-1.0), 0, false},
		{"float32 positive", float32(42.0), 42, true},
		{"float32 negative", float32(-1.0), 0, false},
		{"string unsupported", "hello", 0, false},
		{"nil unsupported", nil, 0, false},
		{"bool unsupported", true, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toUint64(tt.in)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToFloat64_NativeTypes(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want float64
		ok   bool
	}{
		{"float64", float64(3.14), 3.14, true},
		{"float32", float32(3.14), float64(float32(3.14)), true},
		{"int", int(7), 7.0, true},
		{"int32", int32(7), 7.0, true},
		{"int64", int64(7), 7.0, true},
		{"uint", uint(7), 7.0, true},
		{"uint64", uint64(7), 7.0, true},
		{"string unsupported", "hello", 0, false},
		{"nil unsupported", nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64(tt.in)
			assert.Equal(t, tt.ok, ok)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestMapToCPUUsage_ValidUint64(t *testing.T) {
	msg := buildSysmontapMsg(validResultMap())

	result, err := mapToCPUUsage(msg)

	require.NoError(t, err)
	assert.Equal(t, uint64(8), result.CPUCount)
	assert.Equal(t, uint64(8), result.EnabledCPUs)
	assert.Equal(t, uint64(123456789), result.EndMachAbsTime)
	assert.Equal(t, uint64(33), result.Type)
	assert.InDelta(t, 25.5, result.SystemCPUUsage.CPU_TotalLoad, 0.001)
}

func TestMapToCPUUsage_NumericTypeVariations(t *testing.T) {
	// Simulates devices returning different numeric types (the core bug the fix addresses).
	tests := []struct {
		name           string
		cpuCount       interface{}
		enabledCPUs    interface{}
		endMachAbsTime interface{}
		typ            interface{}
		cpuTotalLoad   interface{}
	}{
		{"all int", int(4), int(4), int(999), int(33), int(50)},
		{"all int64", int64(4), int64(4), int64(999), int64(33), int64(50)},
		{"all uint32", uint32(4), uint32(4), uint32(999), uint32(33), float64(50.0)},
		{"all float64", float64(4), float64(4), float64(999), float64(33), float64(50.0)},
		{"mixed types", int(4), uint32(4), int64(999), uint64(33), float32(50.0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := map[string]interface{}{
				"CPUCount":       tt.cpuCount,
				"EnabledCPUs":    tt.enabledCPUs,
				"EndMachAbsTime": tt.endMachAbsTime,
				"Type":           tt.typ,
				"SystemCPUUsage": map[string]interface{}{
					"CPU_TotalLoad": tt.cpuTotalLoad,
				},
			}
			result, err := mapToCPUUsage(buildSysmontapMsg(m))

			require.NoError(t, err)
			assert.Equal(t, uint64(4), result.CPUCount)
			assert.Equal(t, uint64(4), result.EnabledCPUs)
			assert.Equal(t, uint64(999), result.EndMachAbsTime)
			assert.Equal(t, uint64(33), result.Type)
			assert.InDelta(t, 50.0, result.SystemCPUUsage.CPU_TotalLoad, 0.001)
		})
	}
}

func TestMapToCPUUsage_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		msg       dtx.Message
		errSubstr string
	}{
		{"empty payload", dtx.Message{Payload: []interface{}{}}, "empty payload"},
		{"nil payload", dtx.Message{Payload: nil}, "empty payload"},
		{"payload not slice", dtx.Message{Payload: []interface{}{"not a slice"}}, "expected []interface{}"},
		{"empty result array", dtx.Message{Payload: []interface{}{[]interface{}{}}}, "result array is empty"},
		{"result not map", dtx.Message{Payload: []interface{}{[]interface{}{"not a map"}}}, "expected map[string]interface{}"},
		{"missing CPUCount", buildSysmontapMsg(withDeletion(validResultMap(), "CPUCount")), "CPUCount"},
		{"missing SystemCPUUsage", buildSysmontapMsg(withDeletion(validResultMap(), "SystemCPUUsage")), "SystemCPUUsage"},
		{"SystemCPUUsage not map", buildSysmontapMsg(withOverride(validResultMap(), "SystemCPUUsage", "not a map")), "SystemCPUUsage"},
		{"missing CPU_TotalLoad", buildSysmontapMsg(withOverride(validResultMap(), "SystemCPUUsage", map[string]interface{}{})), "CPU_TotalLoad"},
		{"negative CPUCount", buildSysmontapMsg(withOverride(validResultMap(), "CPUCount", int(-5))), "CPUCount"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mapToCPUUsage(tt.msg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errSubstr)
		})
	}
}

func withDeletion(m map[string]interface{}, key string) map[string]interface{} {
	delete(m, key)
	return m
}

func withOverride(m map[string]interface{}, key string, val interface{}) map[string]interface{} {
	m[key] = val
	return m
}

func TestMapToCPUUsage_MultipleResultsUsesFirst(t *testing.T) {
	// Payload with multiple results — should use the first one.
	msg := dtx.Message{
		Payload: []interface{}{
			[]interface{}{
				validResultMap(),
				map[string]interface{}{"extra": "ignored"},
			},
		},
	}
	result, err := mapToCPUUsage(msg)
	require.NoError(t, err)
	assert.Equal(t, uint64(8), result.CPUCount)
}

func TestMapToCPUUsage_MultiElementPayloadUsesFirst(t *testing.T) {
	// Payload with multiple top-level elements — should use the first one (relaxed check).
	msg := dtx.Message{
		Payload: []interface{}{
			[]interface{}{validResultMap()},
			"extra element",
		},
	}
	result, err := mapToCPUUsage(msg)
	require.NoError(t, err)
	assert.Equal(t, uint64(8), result.CPUCount)
}
