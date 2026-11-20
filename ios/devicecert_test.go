package ios

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pairing is what pairing a device leaves: the host's pair record, and the device's key and the
// certificate the host issued for it.
type pairing struct {
	record    PairRecord
	deviceKey *rsa.PrivateKey
	deviceDER []byte
}

func newPairing(t *testing.T) pairing {
	t.Helper()
	deviceKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	require.NoError(t, err)
	devicePublicKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&deviceKey.PublicKey)})
	rootCert, hostCert, deviceCert, _, hostKey, err := createRootCertificate(devicePublicKey)
	require.NoError(t, err)
	block, _ := pem.Decode(deviceCert)
	require.NotNil(t, block)
	return pairing{
		record:    PairRecord{HostCertificate: hostCert, HostPrivateKey: hostKey, DeviceCertificate: deviceCert, RootCertificate: rootCert},
		deviceKey: deviceKey,
		deviceDER: block.Bytes,
	}
}

// selfSigned is a certificate for a key of its own, as something that is not the device would have.
func selfSigned(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	require.NoError(t, err)
	template := x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{}, NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// peer is the device side of a lockdown TLS session: it presents cert, and reports what the host
// sent once the handshake is done.
type peer struct {
	received chan []byte
}

func servePeer(t *testing.T, cert tls.Certificate) (*DeviceConnection, *peer) {
	t.Helper()
	host, device := net.Pipe()
	t.Cleanup(func() { _ = host.Close(); _ = device.Close() })
	p := &peer{received: make(chan []byte, 1)}
	go func() {
		s := tls.Server(device, &tls.Config{Certificates: []tls.Certificate{cert}})
		if err := s.Handshake(); err != nil {
			close(p.received)
			return
		}
		b, _ := io.ReadAll(s)
		p.received <- b
	}()
	return NewDeviceConnectionWithConn(host), p
}

// handshakeDone closes the host's side and reports whether the device's side completed the
// handshake, and what it received then.
func (p *peer) handshakeDone(t *testing.T, conn *DeviceConnection) ([]byte, bool) {
	t.Helper()
	_ = conn.Close()
	select {
	case b, ok := <-p.received:
		return b, ok
	case <-time.After(10 * time.Second):
		t.Fatal("the device's side did not end")
		return nil, false
	}
}

func TestEnableSessionSslAcceptsThePairedDevice(t *testing.T) {
	p := newPairing(t)
	conn, device := servePeer(t, tls.Certificate{Certificate: [][]byte{p.deviceDER}, PrivateKey: p.deviceKey})

	require.NoError(t, conn.EnableSessionSsl(p.record))
	require.NoError(t, conn.Send([]byte("hello")))
	got, ok := device.handshakeDone(t, conn)
	assert.True(t, ok)
	assert.Equal(t, []byte("hello"), got)
}

func TestEnableSessionSslRefusesAnotherKey(t *testing.T) {
	// Something on the network that answers in the device's place, with a certificate of its own.
	p := newPairing(t)
	conn, device := servePeer(t, selfSigned(t))

	err := conn.EnableSessionSsl(p.record)
	assert.ErrorIs(t, err, ErrDeviceCertificateMismatch)
	_, done := device.handshakeDone(t, conn)
	assert.False(t, done, "the handshake completed for something that is not the device")
}

func TestEnableSessionSslRefusesTheDeviceCertificateWithoutItsKey(t *testing.T) {
	// The device's certificate is no secret: it travels in every handshake. Presenting it
	// without the device's private key fails the handshake's proof of possession.
	p := newPairing(t)
	other := selfSigned(t)
	conn, device := servePeer(t, tls.Certificate{Certificate: [][]byte{p.deviceDER}, PrivateKey: other.PrivateKey})

	assert.Error(t, conn.EnableSessionSsl(p.record))
	_, done := device.handshakeDone(t, conn)
	assert.False(t, done, "the handshake completed without the device's private key")
}

func TestEnableSessionSslHandshakeOnlyRefusesAnotherKey(t *testing.T) {
	// DTX services use TLS for the handshake only, and are checked the same way.
	p := newPairing(t)
	conn, device := servePeer(t, selfSigned(t))

	assert.ErrorIs(t, conn.EnableSessionSslHandshakeOnly(p.record), ErrDeviceCertificateMismatch)
	_, done := device.handshakeDone(t, conn)
	assert.False(t, done, "the handshake completed for something that is not the device")
}

func TestEnableSessionSslWithoutDeviceCertificate(t *testing.T) {
	// A pair record without a DeviceCertificate has nothing to check against: any certificate is
	// accepted, as before.
	p := newPairing(t)
	p.record.DeviceCertificate = nil
	conn, device := servePeer(t, selfSigned(t))

	require.NoError(t, conn.EnableSessionSsl(p.record))
	_, done := device.handshakeDone(t, conn)
	assert.True(t, done)
}

func TestEnableSessionSslInvalidDeviceCertificate(t *testing.T) {
	p := newPairing(t)
	p.record.DeviceCertificate = []byte("not a certificate")
	conn, device := servePeer(t, tls.Certificate{Certificate: [][]byte{p.deviceDER}, PrivateKey: p.deviceKey})

	err := conn.EnableSessionSsl(p.record)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrDeviceCertificateMismatch))
	_, done := device.handshakeDone(t, conn)
	assert.False(t, done, "a handshake with an invalid DeviceCertificate in the pair record")
}

func TestDeviceCertificateVerifierTakesDER(t *testing.T) {
	p := newPairing(t)
	verify, err := deviceCertificateVerifier(p.deviceDER)
	require.NoError(t, err)
	assert.NoError(t, verify([][]byte{p.deviceDER}, nil))
	assert.ErrorIs(t, verify(nil, nil), ErrDeviceCertificateMismatch)
	assert.ErrorIs(t, verify([][]byte{selfSigned(t).Certificate[0]}, nil), ErrDeviceCertificateMismatch)
}
