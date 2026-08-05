package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"testing"

	"golang.org/x/crypto/chacha20poly1305"
)

// fakeXpcConn is an in-memory xpcConn that captures what the host sends and
// replays a single scripted reply. It lets us drive the real createTunnelListener
// (and the real cipherStream/controlChannelReadWriter) without a device.
type fakeXpcConn struct {
	reply map[string]interface{}
}

func (f *fakeXpcConn) Send(data map[string]interface{}, flags ...uint32) error {
	return nil
}

func (f *fakeXpcConn) ReceiveOnClientServerStream() (map[string]interface{}, error) {
	return f.reply, nil
}

// newTunnelServiceForTest wires a tunnelService whose cipher decrypts a reply
// produced by sealPayload below. Both client and server AEAD use the same key so
// the test can seal a device reply the host's serverCipher can open.
func newTunnelServiceForTest(t *testing.T, payload map[string]interface{}) *tunnelService {
	t.Helper()
	key := make([]byte, chacha20poly1305.KeySize)
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		t.Fatalf("failed to create AEAD: %v", err)
	}

	// The host reads device replies with serverCipher at the current nonce.
	// createTunnelListener writes one message before reading: write() derives the
	// nonce from sequence (0) and only afterwards increments sequence, so the
	// reply is opened at the all-zero nonce. Seal the reply with that same nonce.
	nonce := make([]byte, aead.NonceSize())
	binary.LittleEndian.PutUint64(nonce[0:8], 0)

	marshalled, err := json.Marshal(map[string]interface{}{
		"response": map[string]interface{}{
			"_1": map[string]interface{}{
				"createListener": payload,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	encrypted := aead.Seal(nil, nonce, marshalled, nil)

	reply := map[string]interface{}{
		"value": map[string]interface{}{
			"message": map[string]interface{}{
				"streamEncrypted": map[string]interface{}{
					"_0": encrypted,
				},
			},
		},
	}

	conn := &fakeXpcConn{reply: reply}
	control := newControlChannelReadWriter(conn)
	return &tunnelService{
		controlChannel: control,
		cipher:         newCipherStream(control, aead, aead),
	}
}

// TestCreateTunnelListenerMissingPort ensures a malformed createListener reply
// that omits "port" returns a clean error instead of panicking on a bare type
// assertion (TUN-01).
func TestCreateTunnelListenerMissingPort(t *testing.T) {
	svc := newTunnelServiceForTest(t, map[string]interface{}{
		// no "port" key, but a valid devicePublicKey so we reach the port check
		"devicePublicKey": "AAAA",
	})

	_, err := svc.createTunnelListener()
	if err == nil {
		t.Fatal("expected an error for a createListener reply missing 'port', got nil")
	}
}

// TestCreateTunnelListenerNonNumericPort ensures a non-numeric "port" value
// returns a clean error instead of panicking.
func TestCreateTunnelListenerNonNumericPort(t *testing.T) {
	svc := newTunnelServiceForTest(t, map[string]interface{}{
		"port":            "not-a-number",
		"devicePublicKey": "AAAA",
	})

	_, err := svc.createTunnelListener()
	if err == nil {
		t.Fatal("expected an error for a createListener reply with a non-numeric 'port', got nil")
	}
}
