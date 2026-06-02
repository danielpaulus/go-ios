//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
	"github.com/google/gopacket/pcapgo"
)

// TestPcap captures packets for a few seconds (pcap runs until killed and writes
// a dump-<ts>.pcap into its working dir), then parses it with pcapgo and asserts
// at least one packet was captured.
func TestPcap(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		dir := harness.StreamInTempDir(t, udid, 6*time.Second, "pcap")

		caps, err := filepath.Glob(filepath.Join(dir, "*.pcap"))
		if err != nil || len(caps) == 0 {
			t.Fatalf("pcap: no .pcap file produced in %s: %v", dir, err)
		}

		f, err := os.Open(caps[0])
		if err != nil {
			t.Fatalf("pcap: open %s: %v", caps[0], err)
		}
		defer f.Close()

		r, err := pcapgo.NewReader(f)
		if err != nil {
			t.Fatalf("pcap: %s is not a valid pcap file: %v", caps[0], err)
		}
		n := 0
		for {
			if _, _, err := r.ReadPacketData(); err != nil {
				break
			}
			n++
		}
		if n < 1 {
			t.Fatalf("pcap: captured 0 packets in %s", caps[0])
		}
	})
}
