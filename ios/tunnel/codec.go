package tunnel

import (
	"crypto/cipher"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

type eventCodec interface {
	Encode() map[string]interface{}
	Decode(e map[string]interface{}) error
}

type pairingData struct {
	data            []byte
	kind            string
	sendingHost     string
	startNewSession bool
}

func (p *pairingData) Encode() map[string]interface{} {
	return map[string]interface{}{
		"pairingData": map[string]interface{}{
			"_0": map[string]interface{}{
				"data":            p.data,
				"kind":            p.kind,
				"sendingHost":     p.sendingHost,
				"startNewSession": p.startNewSession,
			},
		},
	}
}

func (p *pairingData) Decode(e map[string]interface{}) error {
	pd, err := getChildMap(e, "pairingData", "_0")
	if err != nil {
		return err
	}
	if data, ok := pd["data"].([]byte); ok {
		p.data = data
	}
	if kind, ok := pd["kind"].(string); ok {
		p.kind = kind
	}
	if startNewSession, ok := pd["startNewSession"].(bool); ok {
		p.startNewSession = startNewSession
	}
	if sendingHost, ok := pd["sendingHost"].(string); ok {
		p.sendingHost = sendingHost
	}
	return nil
}

type pairVerifyFailed struct {
}

func (p pairVerifyFailed) Encode() map[string]interface{} {
	return map[string]interface{}{
		"pairVerifyFailed": map[string]interface{}{},
	}
}

func (p pairVerifyFailed) Decode(_ map[string]interface{}) error {
	return nil
}

func getChildMap(m map[string]interface{}, keys ...string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return m, nil
	}
	k := keys[0]
	if c, ok := m[k].(map[string]interface{}); ok {
		return getChildMap(c, keys[1:]...)
	}
	return nil, fmt.Errorf("getChildMap: could not find entry for '%s'", k)
}

type xpcConn interface {
	Send(data map[string]interface{}, flags ...uint32) error
	ReceiveOnClientServerStream() (map[string]interface{}, error)
}

// controlChannelReadWriter encodes messages into the 'RemotePairing.ControlChannelMessageEnvelope'
// format for RemoteXPC connections
// There are three types of payload put into this envelope
//  1. Requests
//     the only request used here is to initiate a handshake
//  2. Events
//     they are sent in both directions using the same format. In most cases events also carry another payload
//     that is encoded using a type-length-value (TLV) format
//  3. Encrypted Streams
//     these messages contain an encrypted payload using an AEAD format.
//     They are implemented in a separate type cipherStream here.
type controlChannelReadWriter struct {
	seqNr uint64
	conn  xpcConn
}

func newControlChannelReadWriter(conn xpcConn) *controlChannelReadWriter {
	return &controlChannelReadWriter{
		seqNr: 1,
		conn:  conn,
	}
}

// writeEvent wraps an event into a 'RemotePairing.ControlChannelMessageEnvelope' and transfers it on the RemoteXPC
// connection
func (c *controlChannelReadWriter) writeEvent(e eventCodec) error {
	encoded := map[string]interface{}{
		"plain": map[string]interface{}{
			"_0": map[string]interface{}{
				"event": map[string]interface{}{
					"_0": e.Encode(),
				},
			},
		},
	}
	return c.write(encoded)
}

// readEvent unwraps an event from a 'RemotePairing.ControlChannelMessageEnvelope'
func (c *controlChannelReadWriter) readEvent(e eventCodec) error {
	m, err := c.read()
	if err != nil {
		return fmt.Errorf("readEvent: failed to read message: %w", err)
	}
	event, err := getChildMap(m, "plain", "_0", "event", "_0")
	if err != nil {
		return fmt.Errorf("readEvent: failed to get event payload: %w", err)
	}
	return e.Decode(event)
}

func (c *controlChannelReadWriter) writeRequest(req map[string]interface{}) error {
	err := c.write(map[string]interface{}{
		"plain": map[string]interface{}{
			"_0": map[string]interface{}{
				"request": map[string]interface{}{
					"_0": req,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("writeRequest: failed to write message: %w", err)
	}
	return nil
}

func (c *controlChannelReadWriter) write(message map[string]interface{}) error {
	e := map[string]interface{}{
		"mangledTypeName": "RemotePairing.ControlChannelMessageEnvelope",
		"value": map[string]interface{}{
			"message":        message,
			"originatedBy":   "host",
			"sequenceNumber": c.seqNr,
		},
	}
	c.seqNr += 1
	err := c.conn.Send(e)
	if err != nil {
		return fmt.Errorf("write: failed to send message: %w", err)
	}
	return nil
}

func (c *controlChannelReadWriter) read() (map[string]interface{}, error) {
	p, err := c.conn.ReceiveOnClientServerStream()
	if err != nil {
		return nil, err
	}
	value, err := getChildMap(p, "value")
	if err != nil {
		return nil, err
	}

	return getChildMap(value, "message")
}

// cipherStream encrypts and decrypts payloads embedded into 'RemotePairing.ControlChannelMessageEnvelope' messages
// It uses an authenticated encryption with associated (AEAD) format where the nonce is a counter starting
// at zero for the first message. There is always a message from the host to the device, and one from the device to
// the host. This message pair uses the same nonce before that counter is increased for the next message from the host
// to the device
type cipherStream struct {
	controlChannel *controlChannelReadWriter
	clientCipher   cipher.AEAD
	serverCipher   cipher.AEAD
	nonce          []byte
	sequence       uint64
}

func newCipherStream(controlChannel *controlChannelReadWriter, clientCipher, serverCipher cipher.AEAD) *cipherStream {
	return &cipherStream{
		controlChannel: controlChannel,
		clientCipher:   clientCipher,
		serverCipher:   serverCipher,
		nonce:          make([]byte, clientCipher.NonceSize()),
		sequence:       0,
	}
}

func (c *cipherStream) write(p map[string]interface{}) error {
	c.updateNonce()
	marshalled, err := json.Marshal(p)
	if err != nil {
		return err
	}
	encrypted := c.clientCipher.Seal(nil, c.nonce, marshalled, nil)
	c.sequence += 1
	return c.controlChannel.write(map[string]interface{}{
		"streamEncrypted": map[string]interface{}{
			"_0": encrypted,
		},
	})
}

func (c *cipherStream) read(p *map[string]interface{}) error {
	m, err := c.controlChannel.read()
	if err != nil {
		return err
	}
	if streamEncr, err := getChildMap(m, "streamEncrypted"); err == nil {
		if cip, ok := streamEncr["_0"].([]byte); ok {
			plain, err := c.serverCipher.Open(nil, c.nonce, cip, nil)
			if err != nil {
				return err
			}
			return json.Unmarshal(plain, p)
		}
	}
	return fmt.Errorf("not implemented")
}

func (c *cipherStream) updateNonce() {
	b := binary.LittleEndian.AppendUint64(nil, c.sequence)
	copy(c.nonce[0:8], b)
}
