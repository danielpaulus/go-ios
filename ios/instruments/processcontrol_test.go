package instruments

import (
	"errors"
	"testing"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

// stubSleep replaces the retry sleep with a no-op that records the backoffs, so
// retry tests don't actually wait. It restores the original on cleanup.
func stubSleep(t *testing.T) *[]time.Duration {
	t.Helper()
	var slept []time.Duration
	orig := launchRetrySleep
	launchRetrySleep = func(d time.Duration) { slept = append(slept, d) }
	t.Cleanup(func() { launchRetrySleep = orig })
	return &slept
}

func transientErrMsg() dtx.Message {
	return dtx.Message{
		PayloadHeader: dtx.PayloadHeader{MessageType: dtx.DtxTypeError},
		Payload: []interface{}{nskeyedarchiver.NSError{
			ErrorCode: transientLaunchErrorCode,
			Domain:    deviceProcessControlDomain,
			UserInfo:  map[string]interface{}{"NSLocalizedDescription": "Request to launch com.apple.Preferences failed."},
		}},
	}
}

func okMsg(pid uint64) dtx.Message {
	return dtx.Message{
		PayloadHeader: dtx.PayloadHeader{MessageType: dtx.ResponseWithReturnValueInPayload},
		Payload:       []interface{}{pid},
	}
}

// a transient launch reply, as MethodCall actually returns it: the wrapped error
// AND the message carrying the structured NSError.
func transientLaunch() (dtx.Message, error) {
	return transientErrMsg(), errors.New("failed invoking method 'launchSuspendedProcess...' with error: Request to launch failed. (Error code: 2, Domain: com.apple.dt.deviceprocesscontrolservice)")
}

func TestStartProcessRetry_SucceedsAfterTransientFailures(t *testing.T) {
	slept := stubSleep(t)
	calls := 0
	pid, err := startProcessWithRetry("com.apple.Preferences", func() (dtx.Message, error) {
		calls++
		if calls < 3 {
			return transientLaunch()
		}
		return okMsg(4242), nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if pid != 4242 {
		t.Fatalf("pid = %d, want 4242", pid)
	}
	if calls != 3 {
		t.Fatalf("launch called %d times, want 3", calls)
	}
	// Two retries → two backoffs, growing linearly.
	if len(*slept) != 2 || (*slept)[0] != launchRetryBackoff || (*slept)[1] != 2*launchRetryBackoff {
		t.Fatalf("backoffs = %v, want [%v %v]", *slept, launchRetryBackoff, 2*launchRetryBackoff)
	}
}

func TestStartProcessRetry_GivesUpAfterMaxAttempts(t *testing.T) {
	stubSleep(t)
	calls := 0
	_, err := startProcessWithRetry("com.apple.Preferences", func() (dtx.Message, error) {
		calls++
		return transientLaunch()
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != launchMaxAttempts {
		t.Fatalf("launch called %d times, want %d", calls, launchMaxAttempts)
	}
}

func TestStartProcessRetry_NoRetryOnNonTransientError(t *testing.T) {
	stubSleep(t)
	calls := 0
	otherErr := dtx.Message{
		PayloadHeader: dtx.PayloadHeader{MessageType: dtx.DtxTypeError},
		Payload: []interface{}{nskeyedarchiver.NSError{
			ErrorCode: 4, // not the transient launch code
			Domain:    "com.apple.dt.deviceprocesscontrolservice",
		}},
	}
	_, err := startProcessWithRetry("com.example.notinstalled", func() (dtx.Message, error) {
		calls++
		return otherErr, errors.New("Failed starting process")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("non-transient error must not retry: launch called %d times, want 1", calls)
	}
}

func TestStartProcessRetry_NoRetryOnTransportError(t *testing.T) {
	stubSleep(t)
	calls := 0
	_, err := startProcessWithRetry("com.apple.Preferences", func() (dtx.Message, error) {
		calls++
		return dtx.Message{}, errors.New("connection reset") // no NSError payload
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("transport error must not retry: launch called %d times, want 1", calls)
	}
}

func TestStartProcessRetry_SucceedsFirstTry(t *testing.T) {
	slept := stubSleep(t)
	calls := 0
	pid, err := startProcessWithRetry("com.apple.Preferences", func() (dtx.Message, error) {
		calls++
		return okMsg(7), nil
	})
	if err != nil || pid != 7 || calls != 1 {
		t.Fatalf("pid=%d err=%v calls=%d, want 7/nil/1", pid, err, calls)
	}
	if len(*slept) != 0 {
		t.Fatalf("no retries expected, slept %v", *slept)
	}
}

func TestIsTransientLaunchError(t *testing.T) {
	cases := []struct {
		name string
		msg  dtx.Message
		want bool
	}{
		{"transient code 2", transientErrMsg(), true},
		{"wrong code", dtx.Message{Payload: []interface{}{nskeyedarchiver.NSError{ErrorCode: 3, Domain: deviceProcessControlDomain}}}, false},
		{"wrong domain", dtx.Message{Payload: []interface{}{nskeyedarchiver.NSError{ErrorCode: 2, Domain: "com.apple.something.else"}}}, false},
		{"empty payload", dtx.Message{}, false},
		{"non-nserror payload", dtx.Message{Payload: []interface{}{uint64(123)}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isTransientLaunchError(c.msg); got != c.want {
				t.Fatalf("isTransientLaunchError = %v, want %v", got, c.want)
			}
		})
	}
}
