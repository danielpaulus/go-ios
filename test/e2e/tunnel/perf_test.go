//go:build e2e

package tunnel_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// perfStreamWindow is how long each syslog strand streams — long enough to
// overlap the file transfer and the screenshot loop so the workloads contend.
const perfStreamWindow = 10 * time.Second

// largeFileSize is the size of the file pushed and pulled back over the tunnel.
const largeFileSize = 8 << 20 // 8 MiB

// TestTunnelLoad drives a mixed, concurrent workload through the iOS 17+ tunnel
// — an 8 MiB file push+pull round-trip, ten sequential screenshots, and two
// syslog streams, all at the same time — as a regression guard for tunnel
// performance and stability under load. Each strand is a parallel subtest, so
// the strands genuinely overlap and a failure is attributed to the right one.
// Throughput is logged (see `go test -v`) but deliberately not asserted, since
// absolute numbers vary by device and host.
func TestTunnelLoad(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		// Each strand is a parallel subtest, so once started they overlap and
		// contend for the tunnel. The small per-strand start delays stagger the
		// initial RemoteXPC connection setups: iOS limits how many connections
		// can be *established* at the same instant, so a cold burst of four
		// simultaneous setups is unreliable (unrelated to tunnel throughput).
		// Once established the strands stream/transfer concurrently.
		strand := func(name string, delay time.Duration, fn func(*testing.T, string)) {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				time.Sleep(delay)
				fn(t, udid)
			})
		}
		strand("syslog-a", 0, loadSyslogStream)
		strand("file-roundtrip", 1*time.Second, loadFileRoundTrip)
		strand("syslog-b", 2*time.Second, loadSyslogStream)
		strand("screenshots", 3*time.Second, func(t *testing.T, udid string) { loadScreenshots(t, udid, 10) })
	})
}

// loadFileRoundTrip pushes a multi-MiB random file to the device's temp area
// over the tunnel, pulls it back, and asserts the bytes survive the round trip —
// the strongest check that the tunnel data plane neither corrupts nor drops
// bytes under load.
func loadFileRoundTrip(t *testing.T, udid string) {
	t.Helper()
	payload := make([]byte, largeFileSize)
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("generate payload: %v", err)
	}
	local := filepath.Join(t.TempDir(), "perf-upload.bin")
	if err := os.WriteFile(local, payload, 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}
	const remote = "go-ios-e2e-perf.bin" // fixed name: overwritten each run, no accumulation

	pushStart := time.Now()
	runIOSForDevice(t, udid, "file", "push", "--temp", "--local="+local, "--remote="+remote)
	pushDur := time.Since(pushStart)

	back := filepath.Join(t.TempDir(), "perf-download.bin")
	pullStart := time.Now()
	runIOSForDevice(t, udid, "file", "pull", "--temp", "--remote="+remote, "--local="+back)
	pullDur := time.Since(pullStart)

	got, err := os.ReadFile(back)
	if err != nil {
		t.Fatalf("read pulled file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("file round-trip corrupted: pushed %d bytes, pulled %d bytes with differing content", len(payload), len(got))
	}
	mib := float64(largeFileSize) / (1 << 20)
	t.Logf("file round-trip %.0f MiB: push %s (%.1f MiB/s), pull %s (%.1f MiB/s)",
		mib, pushDur.Round(time.Millisecond), mib/pushDur.Seconds(),
		pullDur.Round(time.Millisecond), mib/pullDur.Seconds())
}

// loadScreenshots captures n screenshots sequentially over the tunnel and
// asserts each is a valid, non-empty PNG.
func loadScreenshots(t *testing.T, udid string, n int) {
	t.Helper()
	dir := t.TempDir()
	var totalBytes int64
	start := time.Now()
	for i := 0; i < n; i++ {
		out := filepath.Join(dir, fmt.Sprintf("shot-%d.png", i))
		runIOSForDevice(t, udid, "screenshot", "--output="+out)

		f, err := os.Open(out)
		if err != nil {
			t.Fatalf("screenshot %d: output not created: %v", i, err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatalf("screenshot %d: not a valid PNG: %v", i, err)
		}
		if b := img.Bounds(); b.Dx() <= 0 || b.Dy() <= 0 {
			t.Fatalf("screenshot %d: invalid dimensions %dx%d", i, b.Dx(), b.Dy())
		}
		if fi, err := os.Stat(out); err == nil {
			totalBytes += fi.Size()
		}
	}
	t.Logf("%d screenshots in %s (%d bytes total)", n, time.Since(start).Round(time.Millisecond), totalBytes)
}

// loadSyslogStream streams syslog for perfStreamWindow and asserts it produced
// well-formed JSON log lines while the rest of the load was running.
func loadSyslogStream(t *testing.T, udid string) {
	t.Helper()
	out := harness.StreamSmoke(t, udid, perfStreamWindow, "syslog")
	assertJSONLogLine(t, out, "msg")
	t.Logf("syslog streamed %d bytes in %s", len(out), perfStreamWindow)
}
