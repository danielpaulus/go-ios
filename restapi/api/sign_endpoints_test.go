package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
)

// mockSigner is a device-free, impl-free Signer used to prove the endpoints
// parse multipart correctly, pass the right args, and stream/return artifacts.
// It records what it was called with (secrets included) so tests can assert both
// correct wiring and that the handlers never leak those secrets into responses
// or logs.
type mockSigner struct {
	certResult   CertResult
	assetsResult AssetsResult
	signResult   SignResult
	err          error

	gotCreds        ASCCredentials
	gotCertOpts     PrepareCertOptions
	gotProvOpts     PrepareProvisioningOptions
	gotSignOpts     SignAppOptions
	signInputExists bool
}

func (m *mockSigner) PrepareCertificate(_ context.Context, creds ASCCredentials, opts PrepareCertOptions) (CertResult, error) {
	m.gotCreds = creds
	m.gotCertOpts = opts
	return m.certResult, m.err
}

func (m *mockSigner) PrepareProvisioning(_ context.Context, creds ASCCredentials, opts PrepareProvisioningOptions) (AssetsResult, error) {
	m.gotCreds = creds
	m.gotProvOpts = opts
	return m.assetsResult, m.err
}

func (m *mockSigner) SignApp(_ context.Context, opts SignAppOptions) (SignResult, error) {
	m.gotSignOpts = opts
	// Record whether the uploaded input actually landed on disk in a temp dir.
	if _, err := os.Stat(opts.InputPath); err == nil {
		m.signInputExists = true
	}
	return m.signResult, m.err
}

// withSigner swaps defaultSigner for the duration of a test.
func withSigner(t *testing.T, s Signer) {
	t.Helper()
	prev := defaultSigner
	defaultSigner = s
	t.Cleanup(func() { defaultSigner = prev })
}

// signMultipart builds a multipart request context for a signing endpoint.
func signMultipart(t *testing.T, fields map[string]string, files map[string][]byte) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range files {
		fw, err := mw.CreateFormFile(name, name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	mw.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	r := httptest.NewRequest(http.MethodPost, "/x", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	c.Request = r
	return w, c
}

// TestSignCertificate_Success proves the certificate endpoint parses the ASC
// credentials + fields, calls the Signer with them, and streams back the P12.
func TestSignCertificate_Success(t *testing.T) {
	m := &mockSigner{certResult: CertResult{P12: []byte("P12BYTES"), CertificateID: "CERT123"}}
	withSigner(t, m)

	secretKey := []byte("-----BEGIN PRIVATE KEY-----\nSECRET\n-----END PRIVATE KEY-----")
	w, c := signMultipart(t,
		map[string]string{"asc-key-id": "KID", "asc-issuer-id": "IID", "revoke-existing": "true", "p12password": "hunter2"},
		map[string][]byte{"asc-private-key": secretKey},
	)
	SignCertificate(c)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	if got := w.Body.Bytes(); !bytes.Equal(got, []byte("P12BYTES")) {
		t.Fatalf("body = %q, want P12 bytes", got)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/x-pkcs12" {
		t.Fatalf("content-type = %q", ct)
	}
	if w.Header().Get("X-Certificate-Id") != "CERT123" {
		t.Fatalf("missing cert id header")
	}
	// Signer received the exact credentials + options.
	if m.gotCreds.KeyID != "KID" || m.gotCreds.IssuerID != "IID" || !bytes.Equal(m.gotCreds.PrivateKey, secretKey) {
		t.Fatalf("signer got wrong creds: %+v", m.gotCreds)
	}
	if !m.gotCertOpts.RevokeExisting || m.gotCertOpts.P12Password != "hunter2" {
		t.Fatalf("signer got wrong opts: %+v", m.gotCertOpts)
	}
	// The private key bytes and password must never appear in the response body.
	assertNoSecrets(t, w.Body.String(), "SECRET", "hunter2")
}

// TestSignCertificate_MissingKey rejects a request without the .p8 upload.
func TestSignCertificate_MissingKey(t *testing.T) {
	m := &mockSigner{}
	withSigner(t, m)
	w, c := signMultipart(t, map[string]string{"asc-key-id": "K", "asc-issuer-id": "I"}, nil)
	SignCertificate(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

// TestSignCertificate_MissingFields rejects a request without key/issuer ids.
func TestSignCertificate_MissingFields(t *testing.T) {
	m := &mockSigner{}
	withSigner(t, m)
	w, c := signMultipart(t, nil, map[string][]byte{"asc-private-key": []byte("k")})
	SignCertificate(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

// TestSignProvision_Success proves the provision endpoint wires bundle/udid and
// returns base64 profile + p12 in JSON, redacting secrets.
func TestSignProvision_Success(t *testing.T) {
	m := &mockSigner{assetsResult: AssetsResult{
		P12: []byte("P12"), Profile: []byte("PROFILE"), BundleID: "com.example.app", CertificateID: "C1",
	}}
	withSigner(t, m)

	w, c := signMultipart(t,
		map[string]string{
			"asc-key-id": "KID", "asc-issuer-id": "IID",
			"bundleid": "com.example.app", "udid": "UDID-1", "p12password": "pw",
		},
		map[string][]byte{"asc-private-key": []byte("SECRETKEY")},
	)
	SignProvision(c)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad json: %v (%s)", err, w.Body.String())
	}
	if resp["bundleId"] != "com.example.app" || resp["certificateId"] != "C1" {
		t.Fatalf("wrong metadata: %+v", resp)
	}
	if dec, _ := base64.StdEncoding.DecodeString(resp["mobileprovisionBase64"]); string(dec) != "PROFILE" {
		t.Fatalf("profile round-trip failed: %q", resp["mobileprovisionBase64"])
	}
	if dec, _ := base64.StdEncoding.DecodeString(resp["p12Base64"]); string(dec) != "P12" {
		t.Fatalf("p12 round-trip failed")
	}
	if m.gotProvOpts.BundleID != "com.example.app" || m.gotProvOpts.DeviceUDID != "UDID-1" {
		t.Fatalf("signer got wrong prov opts: %+v", m.gotProvOpts)
	}
	assertNoSecrets(t, w.Body.String(), "SECRETKEY")
}

// TestSignProvision_MissingBundle rejects a request without a bundle id.
func TestSignProvision_MissingBundle(t *testing.T) {
	withSigner(t, &mockSigner{})
	w, c := signMultipart(t,
		map[string]string{"asc-key-id": "K", "asc-issuer-id": "I", "udid": "U"},
		map[string][]byte{"asc-private-key": []byte("k")},
	)
	SignProvision(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

// TestSignProvision_ReuseModeNoP12 verifies reuse mode (nil P12) omits p12 keys.
func TestSignProvision_ReuseModeNoP12(t *testing.T) {
	m := &mockSigner{assetsResult: AssetsResult{Profile: []byte("P"), BundleID: "b", CertificateID: "C1"}}
	withSigner(t, m)
	w, c := signMultipart(t,
		map[string]string{"asc-key-id": "K", "asc-issuer-id": "I", "bundleid": "b", "certificate-id": "C1"},
		map[string][]byte{"asc-private-key": []byte("k")},
	)
	SignProvision(c)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", w.Code, w.Body.String())
	}
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["p12Base64"]; ok {
		t.Fatalf("reuse mode should not return a p12")
	}
	if m.gotProvOpts.CertificateID != "C1" {
		t.Fatalf("certificate-id not forwarded")
	}
}

// TestSignApp_Success proves the app endpoint writes the upload to a temp dir,
// calls the Signer with that path, streams the signed artifact, and cleans up.
func TestSignApp_Success(t *testing.T) {
	// The signed artifact the mock "produces" — write it so c.File can stream it.
	signed, err := os.CreateTemp(t.TempDir(), "signed-*.ipa")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := signed.WriteString("SIGNED-IPA"); err != nil {
		t.Fatal(err)
	}
	signed.Close()

	m := &mockSigner{signResult: SignResult{OutputPath: signed.Name(), BundleID: "com.x"}}
	withSigner(t, m)

	w, c := signMultipart(t,
		map[string]string{"p12password": "pw", "bundleid": "com.x"},
		map[string][]byte{"ipa": []byte("RAW-IPA"), "p12file": []byte("P12"), "profile": []byte("PROF")},
	)
	SignApp(c)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", w.Code, w.Body.String())
	}
	if w.Body.String() != "SIGNED-IPA" {
		t.Fatalf("body = %q, want streamed signed artifact", w.Body.String())
	}
	if !m.signInputExists {
		t.Fatalf("upload was not written to disk before signing")
	}
	if !bytes.Equal(m.gotSignOpts.P12, []byte("P12")) || !bytes.Equal(m.gotSignOpts.Profile, []byte("PROF")) {
		t.Fatalf("signer got wrong bytes: %+v", m.gotSignOpts)
	}
	if m.gotSignOpts.P12Password != "pw" || m.gotSignOpts.BundleID != "com.x" {
		t.Fatalf("signer got wrong fields: %+v", m.gotSignOpts)
	}
	// The per-request temp dir (and thus the uploaded input) must be gone.
	if _, err := os.Stat(m.gotSignOpts.InputPath); !os.IsNotExist(err) {
		t.Fatalf("temp upload not cleaned up: %v", err)
	}
	assertNoSecrets(t, w.Body.String(), "pw")
}

// TestSignApp_MissingUploads rejects requests missing required file fields.
func TestSignApp_MissingUploads(t *testing.T) {
	withSigner(t, &mockSigner{})
	// Missing ipa.
	w, c := signMultipart(t, nil, map[string][]byte{"p12file": []byte("x"), "profile": []byte("y")})
	SignApp(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing ipa: got %d, want 400", w.Code)
	}
	// Missing p12file.
	w2, c2 := signMultipart(t, nil, map[string][]byte{"ipa": []byte("x"), "profile": []byte("y")})
	SignApp(c2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("missing p12file: got %d, want 400", w2.Code)
	}
}

// TestPrepareCreateCert generates a supervision cert end-to-end (device-free,
// no mock: it exercises the real ios crypto helper) and checks the envelope.
func TestPrepareCreateCert(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/prepare/create-cert", nil)
	PrepareCreateCert(c)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 (%s)", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	for _, k := range []string{"certDerBase64", "certPem", "privateKeyDerBase64", "privateKeyPem"} {
		if resp[k] == "" {
			t.Fatalf("missing field %q", k)
		}
	}
	if !strings.Contains(resp["certPem"], "CERTIFICATE") {
		t.Fatalf("certPem not PEM: %q", resp["certPem"])
	}
	if _, err := base64.StdEncoding.DecodeString(resp["certDerBase64"]); err != nil {
		t.Fatalf("certDerBase64 not base64: %v", err)
	}
}

// TestPrepareDevice_BadCert rejects an unparseable supervision cert before any
// device I/O (so this stays device-free).
func TestPrepareDevice_BadCert(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w, c := signMultipart(t, map[string]string{"orgname": "Acme"}, map[string][]byte{"cert": []byte("not-a-cert")})
	c.Set(IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: "UDID"}})
	PrepareDevice(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400 (%s)", w.Code, w.Body.String())
	}
}

// TestExtractSupervisionCertDER covers DER, PEM and rejection paths of the
// upload normalizer without needing a device.
func TestExtractSupervisionCertDER(t *testing.T) {
	if _, err := extractSupervisionCertDER([]byte("garbage"), ""); err == nil {
		t.Fatalf("expected error for garbage input")
	}
	// A real DER round-trips: reuse the supervision cert generator.
	cert, err := ios.CreateDERFormattedSupervisionCert()
	if err != nil {
		t.Fatal(err)
	}
	der, err := extractSupervisionCertDER(cert.CertDER, "")
	if err != nil || !bytes.Equal(der, cert.CertDER) {
		t.Fatalf("DER passthrough failed: err=%v", err)
	}
	pemDer, err := extractSupervisionCertDER(cert.CertPEM, "")
	if err != nil || !bytes.Equal(pemDer, cert.CertDER) {
		t.Fatalf("PEM extraction failed: err=%v", err)
	}
}

// TestSigningAdapterSatisfiesSigner is a compile-time check that the concrete
// adapter (and the wired default) implement Signer.
func TestSigningAdapterSatisfiesSigner(t *testing.T) {
	var _ Signer = signingAdapter{}
	var _ Signer = defaultSigner
}

// assertNoSecrets fails if any secret substring leaks into the given text.
func assertNoSecrets(t *testing.T, text string, secrets ...string) {
	t.Helper()
	for _, s := range secrets {
		if s != "" && strings.Contains(text, s) {
			t.Fatalf("secret %q leaked into output: %s", s, text)
		}
	}
}
