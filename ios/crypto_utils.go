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
	"fmt"
	"math/big"
	"regexp"
	"strings"
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

	digestString, _ := computeSKIKey(&rootKeyPair.PublicKey)

	rootCertTemplate.ExtraExtensions = []pkix.Extension{
		{
			Id:    []int{2, 5, 29, 14},
			Value: []byte(digestString),
		}}

	rootCert, err := x509.CreateCertificate(rand.Reader, &rootCertTemplate, &rootCertTemplate, &rootKeyPair.PublicKey, rootKeyPair)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	hostCertTemplate := x509.Certificate{
		SerialNumber:       &b,
		Subject:            pkix.Name{},
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(10, 0, 0),
		SignatureAlgorithm: x509.SHA1WithRSA,
		KeyUsage:           x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		//ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	hostCertTemplate.ExtraExtensions = []pkix.Extension{
		{
			Id:    []int{2, 5, 29, 14},
			Value: (digestString),
		}}
	block, _ := pem.Decode([]byte(publicKeyBytes))

	if block == nil {
		return nil, nil, nil, nil, nil, errors.New("failed to parse PEM block containing the public key")
	}

	var devicePublicKey rsa.PublicKey
	_, err1 := asn1.Unmarshal(block.Bytes, &devicePublicKey)
	if err1 != nil {
		return nil, nil, nil, nil, nil, err1
	}

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
	digestString2, _ := computeSKIKey(&devicePublicKey)
	deviceCertTemplate.ExtraExtensions = []pkix.Extension{
		{
			Id:    []int{2, 5, 29, 14},
			Value: (digestString2),
		}}

	hostCert, err := x509.CreateCertificate(rand.Reader, &hostCertTemplate, &hostCertTemplate, &rootKeyPair.PublicKey, rootKeyPair)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	deviceCert, err := x509.CreateCertificate(rand.Reader, &deviceCertTemplate, &deviceCertTemplate, &devicePublicKey, rootKeyPair)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return certBytesToPEM(rootCert), certBytesToPEM(hostCert), certBytesToPEM(deviceCert), savePEMKey(rootKeyPair), savePEMKey(rootKeyPair), nil

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

	digestString := toHexString(pubHash[:])

	return []byte(digestString), nil
}

func toHexString(bytes []byte) string {
	digestString := fmt.Sprintf("%x", bytes)
	if len(digestString)%2 == 1 {
		digestString = "0" + digestString
	}
	re := regexp.MustCompile("..")
	digestString = strings.TrimRight(re.ReplaceAllString(digestString, "$0:"), ":")
	digestString = strings.ToUpper(digestString)
	return digestString
}

type subjectPublicKeyInfo struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

func certBytesToPEM(certBytes []byte) []byte {
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	return pemCert
}

func savePEMKey(key *rsa.PrivateKey) []byte {
	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(privateKey)
}
