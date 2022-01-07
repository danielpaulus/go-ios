package ios

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"math/big"
	"time"
)

//This code could be a little nicer
func createRootCertificate(publicKeyBytes []byte) ([]byte, []byte, []byte, []byte, []byte, error) {
	reader := rand.Reader
	bitSize := 2048

	rootKeyPair, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return nil, nil, nil, nil, nil, err
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
		return nil, nil, nil, nil, nil, err
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
		return nil, nil, nil, nil, nil, err
	}
	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

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
	hostKeyPair, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	hostdigestString, err := computeSKIKey(&hostKeyPair.PublicKey)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	hostCertTemplate.Extensions = append(hostCertTemplate.Extensions, pkix.Extension{
		Id:    []int{2, 5, 29, 14},
		Value: hostdigestString,
	},
	)

	hostCertBytes, err := x509.CreateCertificate(rand.Reader, &hostCertTemplate, rootCert, &hostKeyPair.PublicKey, rootKeyPair)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	deviceCertBytes, err := createDeviceCert(publicKeyBytes, rootCert, rootKeyPair)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return certBytesToPEM(rootCertBytes), certBytesToPEM(hostCertBytes), certBytesToPEM(deviceCertBytes), savePEMKey(rootKeyPair), savePEMKey(hostKeyPair), nil
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
