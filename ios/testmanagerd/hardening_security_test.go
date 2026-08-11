package testmanagerd

import (
	"io"
	"os"
	"testing"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

// TMGR-01 (HIGH-12): a proxyDispatcher built with a nil testListener whose
// Dispatch body panics must not let a SECOND, unrecovered panic escape from the
// recover handler (which historically nil-dereferenced dispatcher.testListener
// as its very first statement). A valid-but-listener-less callback like
// "_XCT_didFinishExecutingTestPlan" panics inside the switch body; the recover
// must swallow it without touching the nil listener.
func TestDispatchNilListenerRecoverIsNilSafe(t *testing.T) {
	dispatcher := proxyDispatcher{id: "test-nil-listener"} // no testListener, no channel

	m := dtx.Message{
		Payload: []interface{}{"_XCT_didFinishExecutingTestPlan"},
	}

	// Without the fix, the panic inside Dispatch triggers the recover handler,
	// whose first statement nil-derefs dispatcher.testListener and re-panics,
	// escaping Dispatch and crashing the process. With the fix this returns
	// cleanly.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Dispatch let a panic escape with a nil test listener: %v", r)
		}
	}()

	dispatcher.Dispatch(m)
}

// TMGR-02 (HIGH-13): sending the testBundleReady selector to a dispatcher whose
// testBundleReadyChannel is nil must not block. Historically Dispatch did an
// unconditional `p.testBundleReadyChannel <- m`, which blocks forever on a nil
// channel. With the fix it is a guarded non-blocking send and returns promptly.
func TestDispatchTestBundleReadyNilChannelDoesNotBlock(t *testing.T) {
	dispatcher := proxyDispatcher{id: "test-nil-channel", testListener: noopTestListener()} // nil testBundleReadyChannel

	m := dtx.Message{
		Payload: []interface{}{"_XCT_testBundleReadyWithProtocolVersion:minimumVersion:"},
	}

	done := make(chan struct{})
	go func() {
		dispatcher.Dispatch(m)
		close(done)
	}()

	select {
	case <-done:
		// returned promptly, good
	case <-time.After(2 * time.Second):
		t.Fatal("Dispatch blocked sending testBundleReady on a nil channel")
	}
}

// TMGR-02 companion: a full but unread buffered channel (cap 1) must also not
// block Dispatch. The cap-1 channel's only reader is dead code, so a second
// testBundleReady historically deadlocked. With the non-blocking send the
// duplicate is dropped instead.
func TestDispatchTestBundleReadyFullChannelDoesNotBlock(t *testing.T) {
	ch := make(chan dtx.Message, 1)
	ch <- dtx.Message{} // fill it so a further send would block
	dispatcher := proxyDispatcher{id: "test-full-channel", testBundleReadyChannel: ch, testListener: noopTestListener()}

	m := dtx.Message{
		Payload: []interface{}{"_XCT_testBundleReadyWithProtocolVersion:minimumVersion:"},
	}

	done := make(chan struct{})
	go func() {
		dispatcher.Dispatch(m)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Dispatch blocked sending testBundleReady on a full channel")
	}
}

// TMGR-03 (MED-4): an activity-record callback (class/method "none") arriving
// before any test suite has started must not nil-deref runningTestSuite. Apple
// genuinely reports attachments under class "none".
func TestTestCaseFinishedNoneWithoutRunningSuite(t *testing.T) {
	listener := NewTestListener(io.Discard, io.Discard, os.TempDir())

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("testCaseFinished panicked without a running test suite: %v", r)
		}
	}()

	// Fresh listener: runningTestSuite is nil. Historically this dereferenced
	// nil via ts.TestCases.
	listener.testCaseFinished("none", "none", nskeyedarchiver.XCActivityRecord{
		Title:        "activity",
		ActivityType: "userDefined",
	})
}

// LOW-3 (TMGR-04): testBundleReady must return an error instead of panicking on
// a message with fewer than two arguments.
func TestTestBundleReadyTooFewArguments(t *testing.T) {
	ch := make(chan dtx.Message, 1)
	ch <- dtx.Message{} // zero-value Auxiliary -> GetArguments() returns nothing
	xide := XCTestManager_IDEInterface{testBundleReadyChannel: ch}

	_, _, err := xide.testBundleReady()
	if err == nil {
		t.Fatal("testBundleReady must return an error when there are fewer than 2 arguments")
	}
}

// LOW-3 (TMGR-04): testBundleReady must return an error instead of panicking on
// a message whose arguments are present but not valid uint64 version values.
func TestTestBundleReadyWrongArgumentTypes(t *testing.T) {
	// Build an Auxiliary with two byte-slice arguments that are NOT valid
	// nskeyedarchived uint64 values, then round-trip through the encoder so the
	// dtx.PrimitiveDictionary.values slice is populated the way a decoded
	// device message would be.
	aux := dtx.NewPrimitiveDictionary()
	aux.AddBytes([]byte("not a valid archive"))
	aux.AddBytes([]byte("not a valid archive either"))
	auxBytes, err := aux.ToBytes()
	if err != nil {
		t.Fatalf("failed serializing auxiliary: %v", err)
	}
	decodedAux := dtx.DecodeAuxiliary(auxBytes)
	if len(decodedAux.GetArguments()) != 2 {
		t.Fatalf("expected 2 decoded arguments, got %d", len(decodedAux.GetArguments()))
	}

	ch := make(chan dtx.Message, 1)
	ch <- dtx.Message{Auxiliary: decodedAux}
	xide := XCTestManager_IDEInterface{testBundleReadyChannel: ch}

	_, _, err = xide.testBundleReady()
	if err == nil {
		t.Fatal("testBundleReady must return an error when arguments are not valid uint64 versions")
	}
}
