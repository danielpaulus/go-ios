package instruments

import (
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

// validProcEntry returns a well-formed process map as nskeyedarchiver would
// legitimately produce it, with pid supplied as the given (numeric) value.
func validProcEntry(pid interface{}) map[string]interface{} {
	return map[string]interface{}{
		"isApplication": true,
		"name":          "SpringBoard",
		"pid":           pid,
		"realAppName":   "/Applications/SpringBoard.app/SpringBoard",
		"startDate":     nskeyedarchiver.NSDate{Timestamp: time.Unix(1000, 0)},
	}
}

// TestMapToProcInfo_ValidUnchanged pins the happy-path values so the guards
// cannot silently alter behavior for well-formed replies.
func TestMapToProcInfo_ValidUnchanged(t *testing.T) {
	in := []interface{}{validProcEntry(uint64(42))}
	got, err := mapToProcInfo(in)
	if err != nil {
		t.Fatalf("unexpected error for valid input: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 process, got %d", len(got))
	}
	p := got[0]
	if !p.IsApplication || p.Name != "SpringBoard" || p.Pid != 42 ||
		p.RealAppName != "/Applications/SpringBoard.app/SpringBoard" ||
		!p.StartDate.Equal(time.Unix(1000, 0)) {
		t.Fatalf("valid ProcessInfo changed: %+v", p)
	}
}

// TestMapToProcInfo_NumericPidKinds ensures a pid delivered as a non-uint64
// numeric kind (int64/float64), which nskeyedarchiver legitimately produces,
// succeeds via toUint64 instead of panicking.
func TestMapToProcInfo_NumericPidKinds(t *testing.T) {
	cases := map[string]interface{}{
		"int64":   int64(7),
		"float64": float64(7),
		"uint64":  uint64(7),
	}
	for name, pid := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := mapToProcInfo([]interface{}{validProcEntry(pid)})
			if err != nil {
				t.Fatalf("expected success for %s pid, got error: %v", name, err)
			}
			if got[0].Pid != 7 {
				t.Fatalf("expected pid 7, got %d", got[0].Pid)
			}
		})
	}
}

// TestMapToProcInfo_MalformedReturnsError ensures malformed entries yield a
// clean error instead of panicking.
func TestMapToProcInfo_MalformedReturnsError(t *testing.T) {
	nilIsApp := validProcEntry(uint64(1))
	nilIsApp["isApplication"] = nil

	nonMap := []interface{}{"not-a-map"}

	badPid := validProcEntry("not-a-number")

	cases := map[string][]interface{}{
		"nil isApplication": {nilIsApp},
		"non-map entry":     nonMap,
		"non-numeric pid":   {badPid},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := mapToProcInfo(in)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", name)
			}
		})
	}
}
