package remote

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios/golog"
)

// runnerState is the lifecycle of the supervised DeviceKit input runner. Input
// endpoints only proceed when the state is runnerReady; otherwise they return
// 503 so the browser can retry instead of the shelled `ios ui` command failing
// with connection-refused.
type runnerState string

const (
	// runnerStarting: the runner child has been spawned but its RPC endpoint has
	// not answered yet (first boot is ~25s on device).
	runnerStarting runnerState = "starting"
	// runnerReady: the runner's health endpoint answered; input works.
	runnerReady runnerState = "ready"
	// runnerRestarting: the runner died and we are backing off before respawning
	// (or waiting for the respawned child to become ready again).
	runnerRestarting runnerState = "restarting"
	// runnerDown: supervision has stopped (server shutting down).
	runnerDown runnerState = "down"
)

// Timing knobs for the supervisor. They are vars (not consts) only so tests can
// shrink them; production never mutates them.
var (
	// runnerStartTimeout bounds how long we wait for a freshly spawned runner to
	// answer its health endpoint before declaring the attempt failed and
	// respawning. First boot is ~25s; give generous headroom.
	runnerStartTimeout = 60 * time.Second
	// runnerBackoffMin/Max cap the respawn backoff so we never spin-loop hot.
	runnerBackoffMin = 1 * time.Second
	runnerBackoffMax = 10 * time.Second
	// runnerStableAfter is how long a runner must stay up before we treat it as
	// healthy and reset the backoff to the minimum.
	runnerStableAfter = 60 * time.Second
	// runnerHealthPoll is the interval between health probes while waiting for a
	// runner to become ready.
	runnerHealthPoll = 1 * time.Second
)

// runnerProc is a spawned runner child process. The concrete implementation
// wraps os/exec; tests substitute a fake so the supervisor state machine can be
// exercised without a real device.
type runnerProc interface {
	// Wait blocks until the process exits and returns its exit error (nil on a
	// clean exit).
	Wait() error
	// Stop terminates the process (best-effort; used on shutdown).
	Stop()
	// Pid returns the OS process id (0 if unknown), for logging.
	Pid() int
}

// runnerSpawner starts a runner child and reports whether its health endpoint
// is answering. It is the single seam the supervisor uses so tests can inject a
// fake runner without shelling out to `ios ui run devicekit`.
type runnerSpawner interface {
	// spawn starts the runner child. ctx cancellation must terminate the child.
	spawn(ctx context.Context) (runnerProc, error)
	// healthy reports whether the runner's RPC/health endpoint answers.
	healthy(ctx context.Context) bool
}

// runnerSupervisor spawns, monitors and auto-respawns the DeviceKit input
// runner so input self-heals across the intermittent testmanagerd DTX EOF
// disconnects that kill the runner. It exposes the current lifecycle state so
// input handlers can 503 during recovery and the UI can show "reconnecting".
type runnerSupervisor struct {
	udid    string
	spawner runnerSpawner

	mu    sync.RWMutex
	state runnerState
}

func newRunnerSupervisor(udid string, spawner runnerSpawner) *runnerSupervisor {
	return &runnerSupervisor{udid: udid, spawner: spawner, state: runnerStarting}
}

// State returns the current runner lifecycle state.
func (s *runnerSupervisor) State() runnerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *runnerSupervisor) setState(st runnerState) {
	s.mu.Lock()
	s.state = st
	s.mu.Unlock()
}

// run supervises the runner until ctx is cancelled, respawning it with capped
// exponential backoff whenever it dies. It blocks; callers run it in its own
// goroutine. On return the state is runnerDown.
func (s *runnerSupervisor) run(ctx context.Context) {
	backoff := runnerBackoffMin
	attempt := 0
	for {
		if ctx.Err() != nil {
			s.setState(runnerDown)
			return
		}
		attempt++
		startedAt := time.Now()
		s.superviseOnce(ctx, attempt)
		if ctx.Err() != nil {
			s.setState(runnerDown)
			return
		}
		// A runner that stayed up long enough is considered healthy: reset the
		// backoff and attempt counter so a later isolated death recovers fast.
		if time.Since(startedAt) >= runnerStableAfter {
			backoff = runnerBackoffMin
			attempt = 0
		}
		s.setState(runnerRestarting)
		golog.Warn("input runner restarting after backoff", "module", logModule,
			"udid", s.udid, "attempt", attempt, "backoff", backoff.String())
		select {
		case <-ctx.Done():
			s.setState(runnerDown)
			return
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > runnerBackoffMax {
			backoff = runnerBackoffMax
		}
	}
}

// superviseOnce spawns a single runner, waits for it to become ready (updating
// state), then blocks until it exits — logging spawn/ready/death. It returns
// when the child has exited or ctx is cancelled.
func (s *runnerSupervisor) superviseOnce(ctx context.Context, attempt int) {
	proc, err := s.spawner.spawn(ctx)
	if err != nil {
		golog.Warn("input runner spawn failed", "module", logModule, "udid", s.udid,
			"attempt", attempt, "error", err)
		return
	}
	golog.Info("input runner spawned", "module", logModule, "udid", s.udid,
		"pid", proc.Pid(), "attempt", attempt)

	// Wait for the child in one place; readiness polling and death detection both
	// observe this single channel.
	exited := make(chan error, 1)
	go func() { exited <- proc.Wait() }()

	spawnedAt := time.Now()
	ready, died := s.awaitReady(ctx, exited)
	if died {
		// Runner exited before it ever became ready; its death was already logged.
		return
	}
	if ready {
		s.setState(runnerReady)
		golog.Info("input runner ready", "module", logModule, "udid", s.udid,
			"pid", proc.Pid(), "elapsed", time.Since(spawnedAt).Round(time.Millisecond).String())
	}

	// Block until the runner exits (or ctx cancelled). On shutdown, stop it.
	select {
	case werr := <-exited:
		logRunnerExit(s.udid, proc.Pid(), werr)
	case <-ctx.Done():
		proc.Stop()
		<-exited
		golog.Info("input runner stopped for shutdown", "module", logModule, "udid", s.udid, "pid", proc.Pid())
	}
}

// awaitReady polls the runner health endpoint until it answers, the startup
// timeout elapses, the runner exits early, or ctx is cancelled. It returns
// (ready, died): ready=true when the endpoint became healthy, died=true when the
// runner exited before becoming ready (in which case the exit was logged here).
func (s *runnerSupervisor) awaitReady(ctx context.Context, exited <-chan error) (ready, died bool) {
	deadline := time.Now().Add(runnerStartTimeout)
	ticker := time.NewTicker(runnerHealthPoll)
	defer ticker.Stop()
	for {
		if s.spawner.healthy(ctx) {
			return true, false
		}
		select {
		case werr := <-exited:
			logRunnerExit(s.udid, 0, werr)
			return false, true
		case <-ctx.Done():
			return false, false
		case <-ticker.C:
			if time.Now().After(deadline) {
				golog.Warn("input runner did not become ready before timeout", "module", logModule,
					"udid", s.udid, "timeout", runnerStartTimeout.String())
				return false, false
			}
		}
	}
}

// logRunnerExit logs a runner death with its exit reason/return code.
func logRunnerExit(udid string, pid int, werr error) {
	rc := 0
	if werr != nil {
		var ee *exec.ExitError
		if errors.As(werr, &ee) {
			rc = ee.ExitCode()
		} else {
			rc = -1
		}
	}
	golog.Warn("input runner exited", "module", logModule, "udid", udid,
		"pid", pid, "rc", rc, "reason", reasonOf(werr))
}

func reasonOf(err error) string {
	if err == nil {
		return "clean exit"
	}
	return err.Error()
}

// --- exec-backed spawner (production) ---

// execRunnerSpawner spawns `ios ui run devicekit --udid=<udid>` as a child of the
// current `ios` binary and probes the DeviceKit health endpoint for readiness.
type execRunnerSpawner struct {
	iosBinary string
	udid      string
	healthURL string
	stdout    *os.File
	stderr    *os.File
}

func (e *execRunnerSpawner) spawn(ctx context.Context) (runnerProc, error) {
	cmd := exec.CommandContext(ctx, e.iosBinary, "ui", "run", "devicekit", "--udid="+e.udid)
	// Inherit stdout/stderr so the runner's logs (including "ui run ready" and
	// the DTX EOF death) land in the same stream as `ios remote`.
	cmd.Stdout = e.stdout
	cmd.Stderr = e.stderr
	// Give the child its own process group so Stop can signal it directly and it
	// isn't collaterally killed by a Ctrl-C delivered to our group.
	configureRunnerProcAttr(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &execRunnerProc{cmd: cmd}, nil
}

func (e *execRunnerSpawner) healthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.healthURL, nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

type execRunnerProc struct {
	cmd  *exec.Cmd
	once sync.Once
}

func (p *execRunnerProc) Wait() error { return p.cmd.Wait() }

func (p *execRunnerProc) Stop() {
	p.once.Do(func() {
		if p.cmd.Process == nil {
			return
		}
		terminateRunnerProc(p.cmd)
	})
}

func (p *execRunnerProc) Pid() int {
	if p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}
