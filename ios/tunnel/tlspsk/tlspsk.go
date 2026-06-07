// Package tlspsk implements the bare minimum of a TLS 1.2 client needed to talk
// to Apple's modern (iOS 18.2+) CoreDevice tunnel listener, which negotiates the
// plain pre-shared-key cipher suite TLS_PSK_WITH_AES_256_GCM_SHA384 (0x00A9).
//
// Go's standard crypto/tls deliberately omits all TLS_PSK_WITH_* suites, and the
// maintained third-party libraries either lack the GCM PSK suite or are
// DTLS-only, so go-ios carries this focused client. It implements exactly one
// cipher suite, client side only, with no certificate, ECDHE or signature
// handling — plain PSK is the simplest TLS key exchange: all security comes from
// the 32-byte shared secret. See RFC 4279 (PSK), RFC 5487 (PSK+GCM, SHA-384 PRF),
// RFC 5246 (TLS 1.2) and RFC 5288 (AES-GCM record layer).
//
// It is validated offline against `openssl s_server -psk … -cipher
// PSK-AES256-GCM-SHA384 -tls1_2` and against real devices behind the e2e tag.
package tlspsk

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"net"
	"time"
)

const (
	recordChangeCipherSpec uint8 = 20
	recordAlert            uint8 = 21
	recordHandshake        uint8 = 22
	recordApplicationData  uint8 = 23

	versionTLS12 uint16 = 0x0303

	hsClientHello       uint8 = 1
	hsServerHello       uint8 = 2
	hsCertificate       uint8 = 11
	hsServerKeyExchange uint8 = 12
	hsServerHelloDone   uint8 = 14
	hsClientKeyExchange uint8 = 16
	hsFinished          uint8 = 20

	// cipherPSKWithAES256GCMSHA384 is TLS_PSK_WITH_AES_256_GCM_SHA384 (RFC 5487).
	cipherPSKWithAES256GCMSHA384 uint16 = 0x00A9

	gcmTagSize       = 16
	gcmExplicitNonce = 8
	gcmFixedIVLen    = 4
	aes256KeyLen     = 32
	masterSecretLen  = 48
	verifyDataLen    = 12
	maxRecordPayload = 16384
)

// prfHash is the PRF/transcript hash for the 0x00A9 suite: SHA-384.
func prfHash() hash.Hash { return sha512.New384() }

// Client performs a TLS 1.2 plain-PSK handshake over conn using psk as the
// pre-shared key (empty PSK identity, no certificate verification) and returns a
// net.Conn that transparently encrypts/decrypts application data. The returned
// conn takes ownership of the underlying conn.
func Client(conn net.Conn, psk []byte) (net.Conn, error) {
	c := &pskConn{conn: conn, psk: psk}
	if err := c.handshake(); err != nil {
		conn.Close()
		return nil, err
	}
	return c, nil
}

type pskConn struct {
	conn net.Conn
	psk  []byte

	writeAEAD, readAEAD cipher.AEAD
	writeIV, readIV     [gcmFixedIVLen]byte
	writeSeq, readSeq   uint64
	writeEnc, readEnc   bool

	// rawBuf holds bytes read from the socket but not yet parsed into a record.
	rawBuf bytes.Buffer
	// hsBuf holds decoded plaintext handshake bytes spanning record boundaries.
	hsBuf bytes.Buffer
	// appBuf holds decrypted application data not yet returned by Read.
	appBuf bytes.Buffer

	transcript bytes.Buffer
}

// ---- handshake ----

func (c *pskConn) handshake() error {
	clientRandom := make([]byte, 32)
	if _, err := rand.Read(clientRandom); err != nil {
		return err
	}
	if err := c.writeHandshake(c.buildClientHello(clientRandom)); err != nil {
		return fmt.Errorf("tlspsk: write ClientHello: %w", err)
	}

	var serverRandom []byte
	for {
		msgType, body, err := c.readHandshakeMsg()
		if err != nil {
			return fmt.Errorf("tlspsk: read handshake: %w", err)
		}
		switch msgType {
		case hsServerHello:
			sr, err := parseServerHello(body)
			if err != nil {
				return err
			}
			serverRandom = sr
		case hsCertificate, hsServerKeyExchange:
			// plain PSK: no certificate is sent; a ServerKeyExchange, if present,
			// only carries a PSK identity hint, which we ignore.
		case hsServerHelloDone:
			goto donewithserverflight
		default:
			return fmt.Errorf("tlspsk: unexpected handshake message type %d", msgType)
		}
	}
donewithserverflight:
	if serverRandom == nil {
		return errors.New("tlspsk: server never sent ServerHello")
	}

	// ClientKeyExchange: plain PSK body is just the (empty) PSK identity.
	cke := []byte{0x00, 0x00}
	if err := c.writeHandshake(handshakeMsg(hsClientKeyExchange, cke)); err != nil {
		return fmt.Errorf("tlspsk: write ClientKeyExchange: %w", err)
	}

	// Derive keys (RFC 4279 premaster secret + TLS 1.2 PRF key schedule).
	pms := pskPremasterSecret(c.psk)
	master := prf12(pms, "master secret", append(append([]byte{}, clientRandom...), serverRandom...), masterSecretLen)
	c.setupKeys(master, clientRandom, serverRandom)

	// ChangeCipherSpec (plaintext), then encrypted Finished.
	if err := c.writeRecord(recordChangeCipherSpec, []byte{0x01}); err != nil {
		return err
	}
	c.writeEnc = true
	clientFinished := handshakeMsg(hsFinished, c.verifyData(master, "client finished"))
	if err := c.writeHandshake(clientFinished); err != nil {
		return fmt.Errorf("tlspsk: write Finished: %w", err)
	}

	// Server ChangeCipherSpec, then encrypted Finished.
	ct, _, err := c.readRecord()
	if err != nil {
		return err
	}
	if ct != recordChangeCipherSpec {
		return fmt.Errorf("tlspsk: expected ChangeCipherSpec, got record type %d", ct)
	}
	c.readEnc = true
	// The server Finished verify_data is computed over the transcript up to but
	// NOT including the server Finished message itself, so compute it before
	// readHandshakeMsg appends that message to the transcript.
	expected := c.verifyData(master, "server finished")
	msgType, body, err := c.readHandshakeMsg()
	if err != nil {
		return fmt.Errorf("tlspsk: read server Finished: %w", err)
	}
	if msgType != hsFinished {
		return fmt.Errorf("tlspsk: expected server Finished, got %d", msgType)
	}
	if !hmac.Equal(body, expected) {
		return errors.New("tlspsk: server Finished verify_data mismatch")
	}
	return nil
}

func (c *pskConn) buildClientHello(clientRandom []byte) []byte {
	var b bytes.Buffer
	b.Write([]byte{byte(versionTLS12 >> 8), byte(versionTLS12 & 0xff)})
	b.Write(clientRandom)
	b.WriteByte(0) // session_id length
	// cipher_suites: just our one suite.
	b.Write([]byte{0x00, 0x02, byte(cipherPSKWithAES256GCMSHA384 >> 8), byte(cipherPSKWithAES256GCMSHA384)})
	b.Write([]byte{0x01, 0x00}) // compression_methods: 1 method, null
	// no extensions
	return handshakeMsg(hsClientHello, b.Bytes())
}

func parseServerHello(body []byte) (serverRandom []byte, err error) {
	// version(2) random(32) session_id_len(1) session_id suite(2) compression(1)
	if len(body) < 35 {
		return nil, errors.New("tlspsk: ServerHello too short")
	}
	serverRandom = append([]byte{}, body[2:34]...)
	sidLen := int(body[34])
	off := 35 + sidLen
	if len(body) < off+3 {
		return nil, errors.New("tlspsk: ServerHello truncated at cipher suite")
	}
	suite := binary.BigEndian.Uint16(body[off : off+2])
	if suite != cipherPSKWithAES256GCMSHA384 {
		return nil, fmt.Errorf("tlspsk: server selected unsupported cipher suite 0x%04x", suite)
	}
	return serverRandom, nil
}

// ---- key schedule ----

// pskPremasterSecret builds the RFC 4279 §2 premaster secret for plain PSK:
// uint16(len) || zeros(len) || uint16(len) || psk.
func pskPremasterSecret(psk []byte) []byte {
	n := len(psk)
	pms := make([]byte, 0, 2+n+2+n)
	var l [2]byte
	binary.BigEndian.PutUint16(l[:], uint16(n))
	pms = append(pms, l[:]...)
	pms = append(pms, make([]byte, n)...)
	pms = append(pms, l[:]...)
	pms = append(pms, psk...)
	return pms
}

func (c *pskConn) setupKeys(master, clientRandom, serverRandom []byte) {
	// key_block = PRF(master, "key expansion", server_random + client_random).
	// AEAD suite: no MAC keys; client/server write keys (32) + fixed IVs (4).
	need := 2*aes256KeyLen + 2*gcmFixedIVLen
	kb := prf12(master, "key expansion", append(append([]byte{}, serverRandom...), clientRandom...), need)
	clientKey := kb[0:aes256KeyLen]
	serverKey := kb[aes256KeyLen : 2*aes256KeyLen]
	off := 2 * aes256KeyLen
	copy(c.writeIV[:], kb[off:off+gcmFixedIVLen])
	copy(c.readIV[:], kb[off+gcmFixedIVLen:off+2*gcmFixedIVLen])
	c.writeAEAD = newGCM(clientKey)
	c.readAEAD = newGCM(serverKey)
}

func newGCM(key []byte) cipher.AEAD {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err) // key length is a compile-time constant
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	return aead
}

func (c *pskConn) verifyData(master []byte, label string) []byte {
	h := prfHash()
	h.Write(c.transcript.Bytes())
	return prf12(master, label, h.Sum(nil), verifyDataLen)
}

// ---- record layer ----

func (c *pskConn) writeHandshake(msg []byte) error {
	c.transcript.Write(msg)
	return c.writeRecord(recordHandshake, msg)
}

func (c *pskConn) writeRecord(contentType uint8, payload []byte) error {
	var fragment []byte
	if c.writeEnc && contentType != recordChangeCipherSpec {
		var explicit [gcmExplicitNonce]byte
		binary.BigEndian.PutUint64(explicit[:], c.writeSeq)
		nonce := append(append([]byte{}, c.writeIV[:]...), explicit[:]...)
		aad := makeAAD(c.writeSeq, contentType, len(payload))
		sealed := c.writeAEAD.Seal(nil, nonce, payload, aad)
		fragment = append(explicit[:], sealed...)
		c.writeSeq++
	} else {
		fragment = payload
	}
	hdr := []byte{contentType, byte(versionTLS12 >> 8), byte(versionTLS12 & 0xff), byte(len(fragment) >> 8), byte(len(fragment))}
	if _, err := c.conn.Write(append(hdr, fragment...)); err != nil {
		return err
	}
	return nil
}

// readRecord reads one TLS record, decrypting it if the read side is encrypted,
// and returns its content type and plaintext payload.
func (c *pskConn) readRecord() (uint8, []byte, error) {
	hdr, err := c.readN(5)
	if err != nil {
		return 0, nil, err
	}
	contentType := hdr[0]
	length := int(binary.BigEndian.Uint16(hdr[3:5]))
	if length > maxRecordPayload+2048 {
		return 0, nil, fmt.Errorf("tlspsk: oversized record %d", length)
	}
	fragment, err := c.readN(length)
	if err != nil {
		return 0, nil, err
	}
	if c.readEnc && contentType != recordChangeCipherSpec {
		if len(fragment) < gcmExplicitNonce+gcmTagSize {
			return 0, nil, errors.New("tlspsk: short encrypted record")
		}
		explicit := fragment[:gcmExplicitNonce]
		nonce := append(append([]byte{}, c.readIV[:]...), explicit...)
		ciphertext := fragment[gcmExplicitNonce:]
		aad := makeAAD(c.readSeq, contentType, len(ciphertext)-gcmTagSize)
		plain, err := c.readAEAD.Open(nil, nonce, ciphertext, aad)
		if err != nil {
			return 0, nil, fmt.Errorf("tlspsk: record decrypt failed: %w", err)
		}
		c.readSeq++
		fragment = plain
	}
	if contentType == recordAlert {
		if len(fragment) >= 2 {
			return 0, nil, fmt.Errorf("tlspsk: received alert level=%d description=%d", fragment[0], fragment[1])
		}
		return 0, nil, errors.New("tlspsk: received alert")
	}
	return contentType, fragment, nil
}

// readHandshakeMsg returns the next handshake message (type and body),
// reassembling across records.
func (c *pskConn) readHandshakeMsg() (uint8, []byte, error) {
	for c.hsBuf.Len() < 4 {
		if err := c.fillHandshakeBuf(); err != nil {
			return 0, nil, err
		}
	}
	header := c.hsBuf.Bytes()[:4]
	msgType := header[0]
	bodyLen := int(header[1])<<16 | int(header[2])<<8 | int(header[3])
	for c.hsBuf.Len() < 4+bodyLen {
		if err := c.fillHandshakeBuf(); err != nil {
			return 0, nil, err
		}
	}
	full := make([]byte, 4+bodyLen)
	_, _ = io.ReadFull(&c.hsBuf, full)
	c.transcript.Write(full)
	return msgType, full[4:], nil
}

func (c *pskConn) fillHandshakeBuf() error {
	ct, payload, err := c.readRecord()
	if err != nil {
		return err
	}
	if ct != recordHandshake {
		return fmt.Errorf("tlspsk: expected handshake record, got content type %d", ct)
	}
	c.hsBuf.Write(payload)
	return nil
}

// readN returns exactly n bytes from the connection, buffering any surplus.
func (c *pskConn) readN(n int) ([]byte, error) {
	for c.rawBuf.Len() < n {
		tmp := make([]byte, 4096)
		m, err := c.conn.Read(tmp)
		if m > 0 {
			c.rawBuf.Write(tmp[:m])
		}
		if err != nil {
			if c.rawBuf.Len() >= n {
				break
			}
			return nil, err
		}
	}
	out := make([]byte, n)
	_, _ = io.ReadFull(&c.rawBuf, out)
	return out, nil
}

func makeAAD(seq uint64, contentType uint8, plaintextLen int) []byte {
	aad := make([]byte, 13)
	binary.BigEndian.PutUint64(aad[0:8], seq)
	aad[8] = contentType
	aad[9] = byte(versionTLS12 >> 8)
	aad[10] = byte(versionTLS12 & 0xff)
	binary.BigEndian.PutUint16(aad[11:13], uint16(plaintextLen))
	return aad
}

func handshakeMsg(msgType uint8, body []byte) []byte {
	out := make([]byte, 4+len(body))
	out[0] = msgType
	out[1] = byte(len(body) >> 16)
	out[2] = byte(len(body) >> 8)
	out[3] = byte(len(body))
	copy(out[4:], body)
	return out
}

// ---- TLS 1.2 PRF (SHA-384) ----

func prf12(secret []byte, label string, seed []byte, length int) []byte {
	labelSeed := make([]byte, 0, len(label)+len(seed))
	labelSeed = append(labelSeed, label...)
	labelSeed = append(labelSeed, seed...)
	out := make([]byte, length)
	pHash(out, secret, labelSeed)
	return out
}

func pHash(result, secret, seed []byte) {
	h := hmac.New(sha512.New384, secret)
	h.Write(seed)
	a := h.Sum(nil)
	for len(result) > 0 {
		h.Reset()
		h.Write(a)
		h.Write(seed)
		b := h.Sum(nil)
		n := copy(result, b)
		result = result[n:]
		h.Reset()
		h.Write(a)
		a = h.Sum(nil)
	}
}

// ---- net.Conn ----

func (c *pskConn) Read(p []byte) (int, error) {
	for c.appBuf.Len() == 0 {
		ct, payload, err := c.readRecord()
		if err != nil {
			return 0, err
		}
		switch ct {
		case recordApplicationData:
			c.appBuf.Write(payload)
		case recordHandshake:
			// ignore post-handshake messages (e.g. HelloRequest); not expected here
		default:
			return 0, fmt.Errorf("tlspsk: unexpected record type %d during Read", ct)
		}
	}
	return c.appBuf.Read(p)
}

func (c *pskConn) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		chunk := p
		if len(chunk) > maxRecordPayload {
			chunk = chunk[:maxRecordPayload]
		}
		if err := c.writeRecord(recordApplicationData, chunk); err != nil {
			return written, err
		}
		written += len(chunk)
		p = p[len(chunk):]
	}
	return written, nil
}

func (c *pskConn) Close() error                       { return c.conn.Close() }
func (c *pskConn) LocalAddr() net.Addr                { return c.conn.LocalAddr() }
func (c *pskConn) RemoteAddr() net.Addr               { return c.conn.RemoteAddr() }
func (c *pskConn) SetDeadline(t time.Time) error      { return c.conn.SetDeadline(t) }
func (c *pskConn) SetReadDeadline(t time.Time) error  { return c.conn.SetReadDeadline(t) }
func (c *pskConn) SetWriteDeadline(t time.Time) error { return c.conn.SetWriteDeadline(t) }
