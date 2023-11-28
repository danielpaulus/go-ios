package tunnel

import (
	"crypto/cipher"
	"encoding/binary"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type controlChannelCodec struct {
	seqNr uint64
}

func newControlChannelCodec() *controlChannelCodec {
	return &controlChannelCodec{seqNr: 1}
}

func (c *controlChannelCodec) Encode(message map[string]interface{}) map[string]interface{} {
	e := map[string]interface{}{
		"mangledTypeName": "RemotePairing.ControlChannelMessageEnvelope",
		"value": map[string]interface{}{
			"message":        message,
			"originatedBy":   "host",
			"sequenceNumber": c.seqNr,
		},
	}
	c.seqNr += 1
	log.WithField("seq", c.seqNr).Trace("enc: updated sequence number")
	return e
}

func (c *controlChannelCodec) Decode(p map[string]interface{}) (map[string]interface{}, error) {
	value, err := getChildMap(p, "value")
	if err != nil {
		return nil, err
	}
	//if seqNr, ok := value["sequenceNumber"].(uint64); ok {
	//	c.seqNr = seqNr + 1
	//	log.WithField("seq", c.seqNr).Trace("dec: updated sequence number")
	//} else {
	//
	//}
	return getChildMap(value, "message")
}

func EncodeEvent(c *controlChannelCodec, event eventCodec) map[string]interface{} {
	return c.Encode(map[string]interface{}{
		"plain": map[string]interface{}{
			"_0": map[string]interface{}{
				"event": map[string]interface{}{
					"_0": event.Encode(),
				},
			},
		},
	})
}

func EncodeStreamEncrypted(c *controlChannelCodec, ciph cipher.AEAD, s *cipherStream, payload map[string]interface{}) (map[string]interface{}, error) {
	encrypted, err := s.Encrypt2(ciph, payload)
	if err != nil {
		return nil, err
	}
	return c.Encode(map[string]interface{}{
		"streamEncrypted": map[string]interface{}{
			"_0": encrypted,
		},
	}), nil
}

func DecodeStreamEncrypted(c *controlChannelCodec, ciph cipher.AEAD, s *cipherStream, msg map[string]interface{}) (map[string]interface{}, error) {
	m, err := c.Decode(msg)
	if err != nil {
		return nil, err
	}
	encr, err := getChildMap(m, "streamEncrypted")
	if err != nil {
		return nil, err
	}
	if data, ok := encr["_0"].([]byte); ok {
		plain, err := s.Decrypt(ciph, data)
		if err != nil {
			return nil, err
		}
		res := make(map[string]interface{})
		err = json.Unmarshal(plain, &res)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("could not find encrypted data")
	}
}

func DecodeEvent(c *controlChannelCodec, m map[string]interface{}, event eventCodec) error {
	msg, err := c.Decode(m)
	if err != nil {
		return err
	}
	e, err := getChildMap(msg, "plain", "_0", "event", "_0")
	if err != nil {
		return err
	}
	return event.Decode(e)
}

func EncodeRequest(c *controlChannelCodec, r map[string]interface{}) map[string]interface{} {
	return c.Encode(map[string]interface{}{
		"plain": map[string]interface{}{
			"_0": map[string]interface{}{
				"request": map[string]interface{}{
					"_0": r,
				},
			},
		},
	})
}

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

func (p pairVerifyFailed) Decode(e map[string]interface{}) error {
	return nil
}

func getChildMap(m map[string]interface{}, keys ...string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return m, nil
	}
	if c, ok := m[keys[0]].(map[string]interface{}); ok {
		return getChildMap(c, keys[1:]...)
	} else {
		return nil, fmt.Errorf("something went wrong")
	}
}

type cipherStream struct {
	sequence uint64
	nonce    []byte
}

func (e *cipherStream) Encrypt(c cipher.AEAD, p []byte) []byte {
	e.nonce = e.createNonce(c)
	encrypted := c.Seal(nil, e.nonce, p, nil)
	e.sequence += 1
	return encrypted
}

func (e *cipherStream) Encrypt2(c cipher.AEAD, plain map[string]interface{}) ([]byte, error) {
	p, err := json.Marshal(plain)
	if err != nil {
		return nil, err
	}
	e.nonce = e.createNonce(c)
	encrypted := c.Seal(nil, e.nonce, p, nil)
	e.sequence += 1
	return encrypted, nil
}

func (e *cipherStream) Decrypt(c cipher.AEAD, p []byte) ([]byte, error) {
	return c.Open(nil, e.nonce, p, nil)
}

func (e *cipherStream) createNonce(c cipher.AEAD) []byte {
	return append(binary.LittleEndian.AppendUint64(nil, e.sequence), make([]byte, 4)...)
}
