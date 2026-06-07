package tlspsk

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestHandshakeAgainstOpenSSL validates the hand-rolled TLS 1.2 plain-PSK client
// against a real OpenSSL server speaking exactly the suite iOS uses
// (PSK-AES256-GCM-SHA384). It is skipped where openssl is absent or lacks that
// PSK suite (e.g. macOS LibreSSL); run it on a host with OpenSSL (Linux).
func TestHandshakeAgainstOpenSSL(t *testing.T) {
	openssl, err := exec.LookPath("openssl")
	if err != nil {
		t.Skip("openssl not found")
	}
	if out, err := exec.Command(openssl, "ciphers", "PSK-AES256-GCM-SHA384").CombinedOutput(); err != nil || !strings.Contains(string(out), "PSK-AES256-GCM-SHA384") {
		t.Skipf("openssl lacks PSK-AES256-GCM-SHA384 (%v): %s", err, out)
	}

	psk := make([]byte, 32)
	if _, err := rand.Read(psk); err != nil {
		t.Fatal(err)
	}
	pskHex := hex.EncodeToString(psk)

	// Grab a free port.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	addr := net.JoinHostPort("127.0.0.1", itoa(port))

	// -rev echoes each received line reversed; empty psk_identity matches our
	// empty client identity (as iOS uses).
	srv := exec.Command(openssl, "s_server",
		"-accept", itoa(port),
		"-psk", pskHex,
		"-psk_identity", "",
		"-cipher", "PSK-AES256-GCM-SHA384",
		"-tls1_2", "-nocert", "-quiet", "-rev",
	)
	var srvErr bytes.Buffer
	srv.Stderr = &srvErr
	if err := srv.Start(); err != nil {
		t.Fatalf("start openssl s_server: %v", err)
	}
	defer func() {
		_ = srv.Process.Kill()
		_ = srv.Wait()
	}()

	// Wait for the server to accept connections.
	var raw net.Conn
	for i := 0; i < 50; i++ {
		raw, err = net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if raw == nil {
		t.Fatalf("could not connect to s_server: %v (stderr: %s)", err, srvErr.String())
	}

	conn, err := Client(raw, psk)
	if err != nil {
		t.Fatalf("PSK handshake failed: %v (server stderr: %s)", err, srvErr.String())
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read: %v (server stderr: %s)", err, srvErr.String())
	}
	got := strings.TrimSpace(string(buf[:n]))
	if got != "olleh" {
		t.Fatalf("expected reversed echo %q, got %q", "olleh", got)
	}
}

// itoa avoids strconv import churn for a single small int.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [12]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}
