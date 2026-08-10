package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios/golog"
)

// logModule is the module attribute attached to every golog line emitted by the
// REST API, matching the repo-wide "module", logModule convention.
const logModule = "go-ios/restapi"

// Job lifecycle states.
const (
	jobRunning   = "running"
	jobSucceeded = "succeeded"
	jobFailed    = "failed"
	jobStopped   = "stopped"
)

// maxJobLogLines bounds how much output a single job retains in memory.
const maxJobLogLines = 5000

// Job is a long-running operation (a test run, a port-forward, …) started via
// the REST API. Its logs are captured on a dedicated per-job sink so concurrent
// jobs never interleave, and can be streamed independently.
type Job struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	UDID       string     `json:"udid"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Error      string     `json:"error,omitempty"`
	Result     any        `json:"result,omitempty"`

	mu   sync.Mutex
	stop func() error
	log  *jobLog
}

// jobView is a lock-free, JSON-safe snapshot of a Job.
type jobView struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	UDID       string     `json:"udid"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Error      string     `json:"error,omitempty"`
	Result     any        `json:"result,omitempty"`
}

func (j *Job) view() jobView {
	j.mu.Lock()
	defer j.mu.Unlock()
	return jobView{
		ID: j.ID, Kind: j.Kind, UDID: j.UDID, Status: j.Status,
		StartedAt: j.StartedAt, FinishedAt: j.FinishedAt, Error: j.Error, Result: j.Result,
	}
}

// finish records terminal success/failure. It is a no-op if the job was already
// stopped, so a cancellation doesn't get re-labelled as a failure.
func (j *Job) finish(result any, err error) {
	j.mu.Lock()
	if j.Status != jobRunning {
		j.mu.Unlock()
		return
	}
	now := time.Now()
	j.FinishedAt = &now
	if err != nil {
		j.Status = jobFailed
		j.Error = err.Error()
	} else {
		j.Status = jobSucceeded
		j.Result = result
	}
	status := j.Status
	j.mu.Unlock()

	golog.Info("job finished", "module", logModule, "udid", j.UDID, "job", j.ID, "kind", j.Kind, "status", status)
	j.log.close()
}

// jobManager is a process-wide, in-memory registry of jobs.
type jobManager struct {
	mu      sync.Mutex
	jobs    map[string]*Job
	counter int
}

var jobs = &jobManager{jobs: map[string]*Job{}}

// create registers a new running job. stop is invoked to cancel it.
func (m *jobManager) create(kind, udid string, stop func() error) *Job {
	m.mu.Lock()
	m.counter++
	id := fmt.Sprintf("%s-%d", kind, m.counter)
	j := &Job{ID: id, Kind: kind, UDID: udid, Status: jobRunning, StartedAt: time.Now(), stop: stop, log: newJobLog()}
	m.jobs[id] = j
	m.mu.Unlock()

	golog.Info("job started", "module", logModule, "udid", udid, "job", id, "kind", kind)
	return j
}

func (m *jobManager) get(id string) (*Job, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	return j, ok
}

// listForUDID returns the jobs belonging to a device.
func (m *jobManager) listForUDID(udid string) []jobView {
	m.mu.Lock()
	all := make([]*Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		all = append(all, j)
	}
	m.mu.Unlock()

	out := make([]jobView, 0, len(all))
	for _, j := range all {
		if j.UDID == udid {
			out = append(out, j.view())
		}
	}
	return out
}

// stop cancels a running job. It is safe to call on an already-terminal job.
func (m *jobManager) stop(id string) (bool, error) {
	j, ok := m.get(id)
	if !ok {
		return false, nil
	}
	j.mu.Lock()
	if j.Status != jobRunning {
		j.mu.Unlock()
		return true, nil
	}
	now := time.Now()
	j.Status = jobStopped
	j.FinishedAt = &now
	stop := j.stop
	j.mu.Unlock()

	golog.Info("job stopped", "module", logModule, "udid", j.UDID, "job", j.ID, "kind", j.Kind)
	var err error
	if stop != nil {
		err = stop()
	}
	j.log.close()
	return true, err
}

// jobLog is a per-job, streamable log sink. It stores a bounded history and
// fans out new lines to any live subscribers (the /jobs/:id/logs stream).
type jobLog struct {
	mu     sync.Mutex
	lines  []string
	subs   map[chan string]struct{}
	closed bool
}

func newJobLog() *jobLog {
	return &jobLog{subs: make(map[chan string]struct{})}
}

// Write implements io.Writer so a testmanagerd TestListener can log into it.
func (l *jobLog) Write(p []byte) (int, error) {
	s := string(p)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return len(p), nil
	}
	l.lines = append(l.lines, s)
	if len(l.lines) > maxJobLogLines {
		l.lines = l.lines[len(l.lines)-maxJobLogLines:]
	}
	for ch := range l.subs {
		select {
		case ch <- s:
		default: // drop for a slow subscriber rather than block the job
		}
	}
	return len(p), nil
}

// snapshot returns the buffered history so far.
func (l *jobLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.lines))
	copy(out, l.lines)
	return out
}

// subscribe returns a channel of future log lines and an unsubscribe func. If
// the log is already closed the channel is closed immediately.
func (l *jobLog) subscribe() (chan string, func()) {
	ch := make(chan string, 256)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		close(ch)
		return ch, func() {}
	}
	l.subs[ch] = struct{}{}
	return ch, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if _, ok := l.subs[ch]; ok {
			delete(l.subs, ch)
			close(ch)
		}
	}
}

// close ends all subscriber streams. Called when the job reaches a terminal state.
func (l *jobLog) close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.closed = true
	for ch := range l.subs {
		delete(l.subs, ch)
		close(ch)
	}
}
