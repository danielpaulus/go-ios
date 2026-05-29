//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// TestPcap captures packets for a few seconds (pcap runs until killed and
// writes a dump-<ts>.pcap into its working dir), then asserts a valid pcap file
// was produced. pcapd is a lockdown service, so no tunnel is required.
func TestPcap(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		dir := harness.StreamInTempDir(t, udid, 6*time.Second, "pcap")

		caps, err := filepath.Glob(filepath.Join(dir, "*.pcap"))
		if err != nil || len(caps) == 0 {
			t.Fatalf("pcap: no .pcap file produced in %s: %v", dir, err)
		}
		// 24 bytes is the libpcap global header; a valid capture is at least that.
		if fi, err := os.Stat(caps[0]); err != nil || fi.Size() < 24 {
			t.Fatalf("pcap: %s is not a valid capture: %v size=%v", caps[0], err, fi)
		}
	})
}
