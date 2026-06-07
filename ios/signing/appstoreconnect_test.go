package signing

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
)

func TestMakeJWT(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	token, expiry, err := makeJWT(AppStoreConnectCredentials{
		KeyID:      "ABC123DEFG",
		IssuerID:   "00000000-0000-0000-0000-000000000000",
		PrivateKey: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(token, ".")) != 3 {
		t.Fatalf("token should have three JWT segments: %q", token)
	}
	if expiry.IsZero() {
		t.Fatal("expiry was not set")
	}
}

func TestGenerateCertificateRequest(t *testing.T) {
	key, csrPEM, err := GenerateCertificateRequest("go-ios test")
	if err != nil {
		t.Fatal(err)
	}
	if key == nil {
		t.Fatal("private key was nil")
	}
	block, _ := pem.Decode(csrPEM)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		t.Fatalf("unexpected CSR PEM block: %#v", block)
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if csr.Subject.CommonName != "go-ios test" {
		t.Fatalf("common name = %q", csr.Subject.CommonName)
	}
}
