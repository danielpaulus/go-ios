package ios

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"go.mozilla.org/pkcs7"
)

const bitSize = 2048

func createRootCertificate(publicKeyBytes []byte) ([]byte, []byte, []byte, []byte, []byte, error) {
	rootKeyPair, rootCertBytes, rootCert, err := createRootCert()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	hostKeyPair, hostCertBytes, err := createHostCert(rootCert, rootKeyPair)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	deviceCertBytes, err := createDeviceCert(publicKeyBytes, rootCert, rootKeyPair)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return certBytesToPEM(rootCertBytes), certBytesToPEM(hostCertBytes), certBytesToPEM(deviceCertBytes), savePEMKey(rootKeyPair), savePEMKey(hostKeyPair), nil
}

func createRootCert() (*rsa.PrivateKey, []byte, *x509.Certificate, error) {
	rootKeyPair, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, nil, nil, err
	}
	var b big.Int
	b.SetInt64(0)
	rootCertTemplate := x509.Certificate{
		SerialNumber:          &b,
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SignatureAlgorithm:    x509.SHA1WithRSA,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	digestString, err := computeSKIKey(&rootKeyPair.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// reminder: you cannot use the ExtraExtentions field here because for some reason
	// that created invalid certificates that throw errors when I try to parse them
	// with golang.
	rootCertTemplate.Extensions = append(rootCertTemplate.Extensions, pkix.Extension{
		Id:    []int{2, 5, 29, 14},
		Value: digestString,
	})

	rootCertBytes, err := x509.CreateCertificate(rand.Reader, &rootCertTemplate, &rootCertTemplate, &rootKeyPair.PublicKey, rootKeyPair)
	if err != nil {
		return nil, nil, nil, err
	}
	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	return rootKeyPair, rootCertBytes, rootCert, nil
}

func createHostCert(rootCert *x509.Certificate, rootKeyPair *rsa.PrivateKey) (*rsa.PrivateKey, []byte, error) {
	var b big.Int
	b.SetInt64(0)
	hostCertTemplate := x509.Certificate{
		SerialNumber:          &b,
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SignatureAlgorithm:    x509.SHA1WithRSA,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	hostKeyPair, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, nil, err
	}

	hostdigestString, err := computeSKIKey(&hostKeyPair.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	hostCertTemplate.Extensions = append(hostCertTemplate.Extensions, pkix.Extension{
		Id:    []int{2, 5, 29, 14},
		Value: hostdigestString,
	},
	)

	hostCertBytes, err := x509.CreateCertificate(rand.Reader, &hostCertTemplate, rootCert, &hostKeyPair.PublicKey, rootKeyPair)
	if err != nil {
		return nil, nil, err
	}
	return hostKeyPair, hostCertBytes, nil
}

func createDeviceCert(publicKeyBytes []byte, rootCert *x509.Certificate, rootKeyPair *rsa.PrivateKey) ([]byte, error) {
	block, _ := pem.Decode(publicKeyBytes)

	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	var devicePublicKey rsa.PublicKey
	_, err := asn1.Unmarshal(block.Bytes, &devicePublicKey)
	if err != nil {
		return nil, err
	}
	var b big.Int
	b.SetInt64(0)
	deviceCertTemplate := x509.Certificate{
		SerialNumber:          &b,
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SignatureAlgorithm:    x509.SHA1WithRSA,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	devicePublicKeyDigest, err := computeSKIKey(&devicePublicKey)
	if err != nil {
		return nil, err
	}

	deviceCertTemplate.Extensions = append(deviceCertTemplate.Extensions, pkix.Extension{
		Id:    []int{2, 5, 29, 14},
		Value: devicePublicKeyDigest,
	},
	)

	deviceCertBytes, err := x509.CreateCertificate(rand.Reader, &deviceCertTemplate, rootCert, &devicePublicKey, rootKeyPair)
	if err != nil {
		return nil, err
	}
	return deviceCertBytes, nil
}

type subjectPublicKeyInfo struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

func computeSKIKey(pub *rsa.PublicKey) ([]byte, error) {
	encodedPub, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}

	var subPKI subjectPublicKeyInfo
	_, err = asn1.Unmarshal(encodedPub, &subPKI)
	if err != nil {
		return nil, err
	}

	pubHash := sha1.Sum(subPKI.SubjectPublicKey.Bytes)
	return pubHash[:], nil
}

func certBytesToPEM(certBytes []byte) []byte {
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	return pemCert
}

func savePEMKey(key *rsa.PrivateKey) []byte {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	privateKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	return privateKeyPem
}

func Sign(challengeBytes []byte, cert *x509.Certificate, supervisedPrivateKey interface{}) ([]byte, error) {
	sd, err := pkcs7.NewSignedData(challengeBytes)
	if err != nil {
		return []byte{}, err
	}

	err = sd.AddSigner(cert, supervisedPrivateKey.(crypto.Signer), pkcs7.SignerInfoConfig{})
	if err != nil {
		return []byte{}, err
	}

	return sd.Finish()
}

// CaCertificate is a simple struct to hold a x509 cert and privateKey in DER and PEM formats
// as well as the CER should you need it.
type CaCertificate struct {
	CertDER       []byte
	PrivateKeyDER []byte
	Csr           string
	CertPEM       []byte
	PrivateKeyPEM []byte
}

// CreateDERFormattedSupervisionCert is a convenience function to generate DER and PEM formatted private key and cert to be used
// for device supervision.
// It basically does the same as these openSSL commands:
//
//	openssl genrsa -des3 -out domain.key 2048
//	openssl req -key domain.key -new -out domain.csr
//	openssl x509 -signkey domain.key -in domain.csr -req -days 365 -out domain.crt
//	openssl x509 -in domain.crt -outform der -out domain.der
//	and returns the resulting certs in a CaCertificate struct.
//
// If you need p12 files, please save the PEMs to files and run this:
// openssl pkcs12 -export -inkey supervision-private-key.pem -in supervision-cert.pem -out certificate.p12 -password pass:a
func CreateDERFormattedSupervisionCert() (*CaCertificate, error) {
	// step: generate a keypair
	keys, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, fmt.Errorf("unable to genarate private keys, error: %s", err)
	}

	// step: generate a csr template
	csrTemplate := x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA512WithRSA,
		ExtraExtensions: []pkix.Extension{
			{
				Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
				Critical: true,
			},
		},
	}
	// step: generate the csr request
	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	if err != nil {
		return nil, err
	}
	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	// step: generate a serial number
	serial, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	// step: create the request template
	template := x509.Certificate{
		SerialNumber: serial,
		// Subject:               names,
		NotBefore:             now.Add(-10 * time.Minute).UTC(),
		NotAfter:              now.Add(time.Hour * 24 * 365 * 10).UTC(),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	// step: sign the certificate authority
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &keys.PublicKey, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate, error: %s", err)
	}

	return &CaCertificate{
		CertDER:       certificate,
		PrivateKeyDER: x509.MarshalPKCS1PrivateKey(keys),
		CertPEM:       certBytesToPEM(certificate),
		PrivateKeyPEM: savePEMKey(keys),
		Csr:           string(csr),
	}, nil
}
