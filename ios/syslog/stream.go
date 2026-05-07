package syslog

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Format selects how each log line is rendered before being written.
type Format int

const (
	// FormatRaw writes each line verbatim, with no JSON envelope.
	FormatRaw Format = iota
	// FormatJSON wraps each line in a single-field JSON object: {"msg":"…"}.
	// Equivalent to the previous default output of `ios syslog`.
	FormatJSON
	// FormatParsedJSON parses the syslog line and emits one JSON object per
	// line with timestamp, device, process, pid, level, message fields.
	FormatParsedJSON
)

// StreamConfig configures a Streamer. Zero-valued fields take sensible
// defaults; the empty StreamConfig{} produces FormatJSON output with a 64 KiB
// buffer flushed every 100 ms and no level filtering.
type StreamConfig struct {
	// Format controls the output rendering.
	Format Format
	// LevelFilter, if non-nil, restricts output to lines whose ASL level
	// (e.g. Notice, Error, Fault) is a key in this set. Keys must be lower-
	// case; ParseLevelFilter does the case-folding for you. nil = no filter.
	LevelFilter map[string]bool
	// BufferSize sets the size of the bufio.Writer wrapping the destination.
	// 0 means 64 KiB.
	BufferSize int
	// FlushInterval is the time-bounded fallback flush cadence so low-rate
	// streams stay responsive; 0 means 100 ms.
	FlushInterval time.Duration
}

// LineReader is the slice of *Connection's API the Streamer needs. Mocking
// this interface makes Streamer trivially testable without a device.
type LineReader interface {
	ReadLogMessageBytes() ([]byte, error)
}

// Streamer drains lines from a LineReader, optionally filters by level,
// renders them per the configured Format, and writes them to a destination.
// One Streamer is single-use: call Run, return, discard.
type Streamer struct {
	src       LineReader
	cfg       StreamConfig
	formatter byteFormatter
}

// NewStreamer returns a Streamer for the given source. The source typically
// comes from syslog.New(...) but any LineReader works.
func NewStreamer(src LineReader, cfg StreamConfig) *Streamer {
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 64 * 1024
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 100 * time.Millisecond
	}
	s := &Streamer{src: src, cfg: cfg}
	switch cfg.Format {
	case FormatRaw:
		s.formatter = rawByteFormatter
	case FormatParsedJSON:
		s.formatter = newParsedJSONFormatter()
	default:
		s.formatter = newLegacyJSONFormatter()
	}
	return s
}

// Run streams to w until the source returns an error, the writer reports
// a closed pipe (treated as clean shutdown), or ctx is cancelled. The bufio
// buffer is always flushed on return. Returns nil for the clean-shutdown
// cases (ctx cancellation, EPIPE, io.EOF) and the underlying error otherwise.
func (s *Streamer) Run(ctx context.Context, w io.Writer) error {
	silenceSIGPIPE()

	out := bufio.NewWriterSize(w, s.cfg.BufferSize)
	defer out.Flush()

	// If ctx is cancelable and the source is closeable, close the source on
	// cancel so the blocking ReadLogMessageBytes returns. We unconditionally
	// stop the watcher in either Run-exit path.
	if ctx != nil {
		closer, _ := s.src.(io.Closer)
		if closer != nil {
			watcherDone := make(chan struct{})
			defer close(watcherDone)
			go func() {
				select {
				case <-ctx.Done():
					_ = closer.Close()
				case <-watcherDone:
				}
			}()
		}
	}

	lastFlush := time.Now()
	for {
		line, readErr := s.src.ReadLogMessageBytes()
		if readErr != nil {
			if ctx != nil && ctx.Err() != nil {
				return nil
			}
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return fmt.Errorf("read syslog: %w", readErr)
		}
		line = trimSyslogTerminators(line)

		if s.cfg.LevelFilter != nil && !lineMatchesLevel(line, s.cfg.LevelFilter) {
			continue
		}

		if err := s.formatter(line, out); err != nil {
			if isClosedPipe(err) {
				return nil
			}
			return fmt.Errorf("format syslog line: %w", err)
		}
		if err := out.WriteByte('\n'); err != nil {
			if isClosedPipe(err) {
				return nil
			}
			return fmt.Errorf("write syslog line: %w", err)
		}
		if now := time.Now(); now.Sub(lastFlush) > s.cfg.FlushInterval {
			if err := out.Flush(); err != nil {
				if isClosedPipe(err) {
					return nil
				}
				return fmt.Errorf("flush syslog output: %w", err)
			}
			lastFlush = now
		}
	}
}

// trimSyslogTerminators strips the ReadSlice-included delimiter byte and any
// preceding newline. Returns the original slice when no terminator is present.
func trimSyslogTerminators(line []byte) []byte {
	n := len(line)
	if n > 0 && line[n-1] == 0x00 {
		n--
	}
	if n > 0 && line[n-1] == 0x0A {
		n--
	}
	return line[:n]
}

func isClosedPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE)
}

// silenceSIGPIPE converts SIGPIPE-on-stdout-close into an EPIPE returned by
// Write/Flush so callers can treat it as a clean shutdown rather than dying
// with a stack trace. Idempotent — safe to call repeatedly.
var sigpipeOnce sync.Once

func silenceSIGPIPE() {
	sigpipeOnce.Do(func() {
		signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
	})
}

// --- byte formatters -------------------------------------------------------

// byteFormatter renders one log line directly into a buffered writer (without
// the trailing newline — that is appended by Streamer.Run).
type byteFormatter func(line []byte, w *bufio.Writer) error

func rawByteFormatter(line []byte, w *bufio.Writer) error {
	_, err := w.Write(line)
	return err
}

// newLegacyJSONFormatter returns a formatter producing {"msg":"…"} per line.
// The fast path (no characters needing JSON escape — typical for syslog) is
// allocation-free; the slow path defers to encoding/json for correct escapes.
func newLegacyJSONFormatter() byteFormatter {
	const prefix = `{"msg":"`
	const suffix = `"}`
	return func(line []byte, w *bufio.Writer) error {
		if jsonNeedsNoEscape(line) {
			if _, err := w.WriteString(prefix); err != nil {
				return err
			}
			if _, err := w.Write(line); err != nil {
				return err
			}
			_, err := w.WriteString(suffix)
			return err
		}
		// Slow path. The string conversion is unavoidable here because
		// json.Marshal accepts a string. Allocations on this branch are
		// acceptable — it fires only on lines with control chars or quotes.
		escaped, _ := json.Marshal(string(line))
		if _, err := w.WriteString(`{"msg":`); err != nil {
			return err
		}
		if _, err := w.Write(escaped); err != nil {
			return err
		}
		return w.WriteByte('}')
	}
}

// newParsedJSONFormatter returns a formatter that parses each syslog line and
// emits the structured LogEntry as JSON.
func newParsedJSONFormatter() byteFormatter {
	parser := Parser()
	return func(line []byte, w *bufio.Writer) error {
		s := string(line)
		entry, err := parser(s)
		if err != nil {
			// Error envelope: {"msg": <line>, "error": <err>}.
			msg, _ := json.Marshal(s)
			errStr, _ := json.Marshal(err.Error())
			if _, e := w.WriteString(`{"msg":`); e != nil {
				return e
			}
			if _, e := w.Write(msg); e != nil {
				return e
			}
			if _, e := w.WriteString(`,"error":`); e != nil {
				return e
			}
			if _, e := w.Write(errStr); e != nil {
				return e
			}
			return w.WriteByte('}')
		}
		encoded, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		_, err = w.Write(encoded)
		return err
	}
}

// jsonNeedsNoEscape reports whether b can be embedded verbatim inside a JSON
// string literal. JSON requires escaping for double-quote, backslash, and any
// control character below 0x20. Bytes >= 0x80 (UTF-8 continuation/start)
// are legal as-is.
func jsonNeedsNoEscape(b []byte) bool {
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c < 0x20 || c == '"' || c == '\\' {
			return false
		}
	}
	return true
}

// --- level filter ---------------------------------------------------------

// ParseLevelFilter parses a comma-separated list of ASL log levels into a
// case-folded set suitable for StreamConfig.LevelFilter. Whitespace around
// items is trimmed. Empty input yields nil (no filtering).
func ParseLevelFilter(csv string) map[string]bool {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil
	}
	out := make(map[string]bool)
	for _, l := range strings.Split(csv, ",") {
		l = strings.TrimSpace(l)
		if l != "" {
			out[strings.ToLower(l)] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// lineMatchesLevel reports whether a syslog line's ASL level token is in the
// allowed set. Locates the level via the `>: ` separator that terminates it
// (more reliable than scanning for `<` since message bodies often contain
// angle brackets), then walks back to the preceding `<`. Malformed lines
// without a recognisable level are passed through (returns true) rather than
// silently dropped.
func lineMatchesLevel(line []byte, allowed map[string]bool) bool {
	end := bytes.Index(line, []byte(">: "))
	if end < 0 {
		return true
	}
	start := -1
	for i := end - 1; i >= 0; i-- {
		if line[i] == '<' {
			start = i
			break
		}
	}
	if start < 0 {
		return true
	}
	level := line[start+1 : end]
	// Case-fold without allocating: ASL level tokens are short ASCII.
	var buf [16]byte
	if len(level) > len(buf) {
		return true
	}
	for i, c := range level {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		buf[i] = c
	}
	return allowed[string(buf[:len(level)])]
}
