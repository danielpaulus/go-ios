package remote

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// setFastRunnerTimings shrinks the supervisor's backoff/poll/timeout knobs so
// the state machine turns over in milliseconds, and returns a restore func.
func setFastRunnerTimings() func() {
	oldTimeout, oldMin, oldMax := runnerStartTimeout, runnerBackoffMin, runnerBackoffMax
	oldStable, oldPoll := runnerStableAfter, runnerHealthPoll
	runnerStartTimeout = 200 * time.Millisecond
	runnerBackoffMin = 5 * time.Millisecond
	runnerBackoffMax = 20 * time.Millisecond
	runnerStableAfter = 100 * time.Millisecond
	runnerHealthPoll = 5 * time.Millisecond
	return func() {
		runnerStartTimeout, runnerBackoffMin, runnerBackoffMax = oldTimeout, oldMin, oldMax
		runnerStableAfter, runnerHealthPoll = oldStable, oldPoll
	}
}

// fakeProc is a controllable runnerProc: it "runs" until die is closed (or Stop
// is called), letting tests drive death/respawn deterministically.
type fakeProc struct {
	pid  int
	die  chan struct{}
	once sync.Once
}

func newFakeProc(pid int) *fakeProc { return &fakeProc{pid: pid, die: make(chan struct{})} }

func (p *fakeProc) Wait() error { <-p.die; return nil }
func (p *fakeProc) Stop()       { p.kill() }
func (p *fakeProc) Pid() int    { return p.pid }
func (p *fakeProc) kill()       { p.once.Do(func() { close(p.die) }) }

// fakeSpawner hands out fakeProcs and reports readiness from an atomic flag.
type fakeSpawner struct {
	mu      sync.Mutex
	procs   []*fakeProc
	spawns  int32
	ready   atomic.Bool
	spawned chan *fakeProc
}

func newFakeSpawner() *fakeSpawner {
	return &fakeSpawner{spawned: make(chan *fakeProc, 16)}
}

func (f *fakeSpawner) spawn(ctx context.Context) (runnerProc, error) {
	n := atomic.AddInt32(&f.spawns, 1)
	p := newFakeProc(int(n))
	f.mu.Lock()
	f.procs = append(f.procs, p)
	f.mu.Unlock()
	f.spawned <- p
	return p, nil
}

func (f *fakeSpawner) healthy(ctx context.Context) bool { return f.ready.Load() }

func (f *fakeSpawner) spawnCount() int { return int(atomic.LoadInt32(&f.spawns)) }

// waitState polls the supervisor state until it matches want or the deadline
// passes.
func waitState(t *testing.T, s *runnerSupervisor, want runnerState) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if s.State() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("state=%q, want %q", s.State(), want)
}

// TestSupervisorRespawnsOnDeath verifies the core self-heal: a runner that dies
// is respawned and becomes ready again without intervention.
func TestSupervisorRespawnsOnDeath(t *testing.T) {
	// Speed up the state machine for the test.
	restore := setFastRunnerTimings()
	defer restore()

	sp := newFakeSpawner()
	sp.ready.Store(true) // health answers immediately
	s := newRunnerSupervisor("udid-1", sp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.run(ctx)

	first := <-sp.spawned
	waitState(t, s, runnerReady)
	if sp.spawnCount() != 1 {
		t.Fatalf("spawnCount=%d, want 1", sp.spawnCount())
	}

	// Simulate the intermittent testmanagerd DTX EOF death.
	first.kill()

	// It must respawn and become ready again on its own.
	<-sp.spawned
	waitState(t, s, runnerReady)
	if sp.spawnCount() < 2 {
		t.Fatalf("spawnCount=%d, want >=2 after death", sp.spawnCount())
	}
}

// TestSupervisorShutdownStopsRunner verifies ctx cancellation terminates the
// child and leaves the supervisor in runnerDown (no orphan).
func TestSupervisorShutdownStopsRunner(t *testing.T) {
	restore := setFastRunnerTimings()
	defer restore()

	sp := newFakeSpawner()
	sp.ready.Store(true)
	s := newRunnerSupervisor("udid-1", sp)

	ctx, cancel := context.WithCancel(context.Background())
	go s.run(ctx)

	p := <-sp.spawned
	waitState(t, s, runnerReady)

	cancel()
	// The child's Wait must unblock via Stop; supervisor must reach Down.
	waitState(t, s, runnerDown)
	select {
	case <-p.die:
	default:
		t.Fatalf("runner was not stopped on shutdown")
	}
}

// TestInputReady503WhenNotReady verifies input endpoints return 503 with the
// documented JSON body while the runner is not ready, and proceed only when it
// is. It exercises the HTTP surface without a real device by driving state.
func TestInputReady503WhenNotReady(t *testing.T) {
	sp := newFakeSpawner()
	s := &Server{udid: "udid-1", driver: DriverDeviceKit, supervisor: newRunnerSupervisor("udid-1", sp)}

	// starting -> 503
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tap", strings.NewReader(`{"x":0.5,"y":0.5}`))
	s.handleTap(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, want 503 while starting", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "input runner starting") || !strings.Contains(body, `"runnerState":"starting"`) {
		t.Fatalf("503 body = %q, want error + runnerState", body)
	}

	// restarting -> 503 with that state
	s.supervisor.setState(runnerRestarting)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/tap", strings.NewReader(`{"x":0.5,"y":0.5}`))
	s.handleTap(rec, req)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"runnerState":"restarting"`) {
		t.Fatalf("status=%d body=%q, want 503 restarting", rec.Code, rec.Body.String())
	}
}

// TestStatusEndpoint reports the runner state and uses 503 when not ready, 200
// when ready.
func TestStatusEndpoint(t *testing.T) {
	sp := newFakeSpawner()
	s := &Server{udid: "u", driver: DriverDeviceKit, supervisor: newRunnerSupervisor("u", sp)}

	rec := httptest.NewRecorder()
	s.handleStatus(rec, httptest.NewRequest(http.MethodGet, "/status", nil))
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"runnerState":"starting"`) {
		t.Fatalf("status=%d body=%q, want 503 starting", rec.Code, rec.Body.String())
	}

	s.supervisor.setState(runnerReady)
	rec = httptest.NewRecorder()
	s.handleStatus(rec, httptest.NewRequest(http.MethodGet, "/status", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"runnerState":"ready"`) {
		t.Fatalf("status=%d body=%q, want 200 ready", rec.Code, rec.Body.String())
	}
}

// TestUnmanagedRunnerAlwaysReady verifies that without a supervisor (externally
// managed runner), input is not gated (state reports ready).
func TestUnmanagedRunnerAlwaysReady(t *testing.T) {
	s := &Server{udid: "u", driver: DriverDeviceKit} // no supervisor
	if got := s.runnerStateString(); got != runnerReady {
		t.Fatalf("unmanaged runnerState=%q, want ready", got)
	}
	rec := httptest.NewRecorder()
	if !s.inputReady(rec) {
		t.Fatalf("inputReady=false for unmanaged runner, want true")
	}
}

// TestExecSpawnerHealthy exercises the production health probe against a test
// server, covering the 2xx and error paths.
func TestExecSpawnerHealthy(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()

	e := &execRunnerSpawner{healthURL: ok.URL + "/health"}
	if !e.healthy(context.Background()) {
		t.Fatalf("healthy=false for 200 endpoint")
	}
	e.healthURL = bad.URL + "/health"
	if e.healthy(context.Background()) {
		t.Fatalf("healthy=true for 500 endpoint")
	}
	e.healthURL = "http://127.0.0.1:1/health" // nothing listening
	if e.healthy(context.Background()) {
		t.Fatalf("healthy=true for unreachable endpoint")
	}
}
