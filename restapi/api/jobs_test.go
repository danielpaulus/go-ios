package api

import (
	"errors"
	"testing"
	"time"
)

func TestJobLifecycleSucceed(t *testing.T) {
	j := jobs.create("unittest", "UDID-A", func() error { return nil })
	if j.view().Status != jobRunning {
		t.Fatalf("new job should be running")
	}
	j.finish("done", nil)
	v := j.view()
	if v.Status != jobSucceeded || v.Result != "done" || v.FinishedAt == nil {
		t.Fatalf("unexpected terminal state: %+v", v)
	}
	// finishing again must not flip the state.
	j.finish(nil, errors.New("late"))
	if j.view().Status != jobSucceeded {
		t.Fatalf("terminal job state must be immutable")
	}
}

func TestJobStopIsTerminalAndNotRelabeled(t *testing.T) {
	cancelled := false
	j := jobs.create("unittest", "UDID-B", func() error { cancelled = true; return nil })
	ok, err := jobs.stop(j.ID)
	if !ok || err != nil {
		t.Fatalf("stop failed: ok=%v err=%v", ok, err)
	}
	if !cancelled {
		t.Fatalf("stop func was not invoked")
	}
	if j.view().Status != jobStopped {
		t.Fatalf("job should be stopped, got %s", j.view().Status)
	}
	// A late finish (e.g. the cancelled goroutine returning ctx.Canceled) must
	// not relabel a stopped job as failed.
	j.finish(nil, errors.New("context canceled"))
	if j.view().Status != jobStopped {
		t.Fatalf("stopped job must stay stopped, got %s", j.view().Status)
	}
}

func TestListForUDIDIsolatesDevices(t *testing.T) {
	a := jobs.create("unittest", "UDID-C", nil)
	jobs.create("unittest", "UDID-D", nil)
	list := jobs.listForUDID("UDID-C")
	found := false
	for _, v := range list {
		if v.UDID != "UDID-C" {
			t.Fatalf("listForUDID leaked another device's job: %s", v.UDID)
		}
		if v.ID == a.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected job %s in UDID-C list", a.ID)
	}
}

func TestJobLogSnapshotSubscribeAndClose(t *testing.T) {
	l := newJobLog()
	l.Write([]byte("line1\n"))
	if snap := l.snapshot(); len(snap) != 1 || snap[0] != "line1\n" {
		t.Fatalf("snapshot wrong: %#v", snap)
	}

	ch, unsub := l.subscribe()
	defer unsub()
	l.Write([]byte("line2\n"))
	select {
	case got := <-ch:
		if got != "line2\n" {
			t.Fatalf("subscriber got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive live line")
	}

	// close ends the stream.
	l.close()
	if _, ok := <-ch; ok {
		t.Fatal("channel should be closed after jobLog.close")
	}
	// writes after close are dropped, not panics.
	l.Write([]byte("ignored\n"))
}

func TestJobLogSubscribeAfterCloseReturnsClosedChannel(t *testing.T) {
	l := newJobLog()
	l.close()
	ch, _ := l.subscribe()
	if _, ok := <-ch; ok {
		t.Fatal("subscribing to a closed log must return a closed channel")
	}
}
