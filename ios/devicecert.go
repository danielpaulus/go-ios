package ios

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// ErrDeviceCertificateMismatch means the peer of a lockdown TLS session presented a certificate
// without the public key of the pair record's DeviceCertificate: it is not the device that was
// paired, but something that answered in its place.
var ErrDeviceCertificateMismatch = errors.New("the device's TLS certificate does not match the pair record")

// deviceCertificateVerifier returns a tls.Config.VerifyPeerCertificate func that accepts the peer
// only when its certificate has the public key of deviceCert, the pair record's DeviceCertificate
// (PEM or DER). The host issued that certificate for the key the device sent when it paired, and
// the TLS handshake proves that the peer holds the private half, which only the device has. The
// certificate is issued by the host's own root, so there is no chain to verify; the key is what
// counts.
//
// A pair record without a DeviceCertificate gets nil, which accepts any certificate as before; one
// whose DeviceCertificate cannot be parsed is an error.
func deviceCertificateVerifier(deviceCert []byte) (func(rawCerts [][]byte, _ [][]*x509.Certificate) error, error) {
	if len(deviceCert) == 0 {
		return nil, nil
	}
	want, err := parseCertificate(deviceCert)
	if err != nil {
		return nil, fmt.Errorf("deviceCertificateVerifier: invalid DeviceCertificate in the pair record: %w", err)
	}
	wantKey, ok := want.PublicKey.(interface{ Equal(crypto.PublicKey) bool })
	if !ok {
		return nil, fmt.Errorf("deviceCertificateVerifier: unsupported DeviceCertificate key type %T", want.PublicKey)
	}
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("%w: the peer sent no certificate", ErrDeviceCertificateMismatch)
		}
		got, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("%w: %v", ErrDeviceCertificateMismatch, err)
		}
		if !wantKey.Equal(got.PublicKey) {
			return ErrDeviceCertificateMismatch
		}
		return nil
	}, nil
}

// parseCertificate parses a certificate in PEM, as pair records store it, or DER.
func parseCertificate(b []byte) (*x509.Certificate, error) {
	if block, _ := pem.Decode(b); block != nil {
		b = block.Bytes
	}
	return x509.ParseCertificate(b)
}
