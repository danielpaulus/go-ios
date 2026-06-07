package signing

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const appStoreConnectAPI = "https://api.appstoreconnect.apple.com"

type AppStoreConnectCredentials struct {
	KeyID      string
	IssuerID   string
	PrivateKey []byte
}

func LoadAppStoreConnectCredentials(keyID, issuerID, privateKeyPath string) (AppStoreConnectCredentials, error) {
	if keyID == "" {
		keyID = os.Getenv("GO_IOS_ASC_KEY_ID")
	}
	if issuerID == "" {
		issuerID = os.Getenv("GO_IOS_ASC_ISSUER_ID")
	}
	if privateKeyPath == "" {
		privateKeyPath = os.Getenv("GO_IOS_ASC_PRIVATE_KEY")
	}
	if keyID == "" {
		return AppStoreConnectCredentials{}, fmt.Errorf("--asc-key-id is required or set GO_IOS_ASC_KEY_ID")
	}
	if issuerID == "" {
		return AppStoreConnectCredentials{}, fmt.Errorf("--asc-issuer-id is required or set GO_IOS_ASC_ISSUER_ID")
	}
	if privateKeyPath == "" {
		return AppStoreConnectCredentials{}, fmt.Errorf("--asc-private-key is required or set GO_IOS_ASC_PRIVATE_KEY")
	}
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return AppStoreConnectCredentials{}, fmt.Errorf("failed to read App Store Connect private key: %w", err)
	}
	return AppStoreConnectCredentials{KeyID: keyID, IssuerID: issuerID, PrivateKey: privateKey}, nil
}

type AppStoreConnectClient struct {
	client      *http.Client
	creds       AppStoreConnectCredentials
	token       string
	tokenExpiry time.Time
}

func NewAppStoreConnectClient(creds AppStoreConnectCredentials) *AppStoreConnectClient {
	return &AppStoreConnectClient{
		client: &http.Client{Timeout: 60 * time.Second},
		creds:  creds,
	}
}

func (c *AppStoreConnectClient) EnsureBundleID(ctx context.Context, identifier, name string) (string, error) {
	if name == "" {
		name = defaultBundleName(identifier)
	}
	var list jsonAPIListResponse
	q := url.Values{}
	q.Set("filter[identifier]", identifier)
	q.Set("filter[platform]", "IOS")
	q.Set("limit", "1")
	if err := c.do(ctx, http.MethodGet, "/v1/bundleIds?"+q.Encode(), nil, &list); err != nil {
		return "", err
	}
	if len(list.Data) > 0 {
		return list.Data[0].ID, nil
	}

	req := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "bundleIds",
			"attributes": map[string]interface{}{
				"identifier": identifier,
				"name":       name,
				"platform":   "IOS",
			},
		},
	}
	var resp jsonAPIResourceResponse
	if err := c.do(ctx, http.MethodPost, "/v1/bundleIds", req, &resp); err != nil {
		return "", err
	}
	return resp.Data.ID, nil
}

func defaultBundleName(identifier string) string {
	replacer := strings.NewReplacer(".", " ", "-", " ", "_", " ")
	name := strings.Join(strings.Fields(replacer.Replace(identifier)), " ")
	if name == "" {
		return "go ios app"
	}
	return name
}

func (c *AppStoreConnectClient) EnsureDevice(ctx context.Context, udid, name string) (string, error) {
	if name == "" {
		name = "go-ios " + udid
	}
	var list jsonAPIListResponse
	q := url.Values{}
	q.Set("filter[udid]", udid)
	q.Set("limit", "1")
	if err := c.do(ctx, http.MethodGet, "/v1/devices?"+q.Encode(), nil, &list); err != nil {
		return "", err
	}
	if len(list.Data) > 0 {
		return list.Data[0].ID, nil
	}

	req := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "devices",
			"attributes": map[string]interface{}{
				"name":     name,
				"platform": "IOS",
				"udid":     udid,
			},
		},
	}
	var resp jsonAPIResourceResponse
	if err := c.do(ctx, http.MethodPost, "/v1/devices", req, &resp); err != nil {
		return "", err
	}
	return resp.Data.ID, nil
}

func (c *AppStoreConnectClient) CreateCertificate(ctx context.Context, csrPEM []byte) (string, []byte, error) {
	req := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "certificates",
			"attributes": map[string]interface{}{
				"certificateType": "IOS_DEVELOPMENT",
				"csrContent":      string(csrPEM),
			},
		},
	}
	var resp jsonAPIResourceResponse
	if err := c.do(ctx, http.MethodPost, "/v1/certificates", req, &resp); err != nil {
		return "", nil, err
	}
	content, _ := resp.Data.Attributes["certificateContent"].(string)
	if content == "" {
		return "", nil, fmt.Errorf("certificate response did not include certificateContent")
	}
	certDER, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", nil, fmt.Errorf("failed decoding certificateContent: %w", err)
	}
	return resp.Data.ID, certDER, nil
}

func (c *AppStoreConnectClient) CreateDevelopmentProfile(ctx context.Context, name, bundleID, certID, deviceID string) ([]byte, error) {
	req := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "profiles",
			"attributes": map[string]interface{}{
				"name":        name,
				"profileType": "IOS_APP_DEVELOPMENT",
			},
			"relationships": map[string]interface{}{
				"bundleId": map[string]interface{}{
					"data": map[string]string{"type": "bundleIds", "id": bundleID},
				},
				"certificates": map[string]interface{}{
					"data": []map[string]string{{"type": "certificates", "id": certID}},
				},
				"devices": map[string]interface{}{
					"data": []map[string]string{{"type": "devices", "id": deviceID}},
				},
			},
		},
	}
	var resp jsonAPIResourceResponse
	if err := c.do(ctx, http.MethodPost, "/v1/profiles", req, &resp); err != nil {
		return nil, err
	}
	content, _ := resp.Data.Attributes["profileContent"].(string)
	if content == "" {
		return nil, fmt.Errorf("profile response did not include profileContent")
	}
	profile, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("failed decoding profileContent: %w", err)
	}
	return profile, nil
}

func (c *AppStoreConnectClient) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	token, err := c.bearerToken()
	if err != nil {
		return err
	}

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, appStoreConnectAPI+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("%s %s failed: %s: %s", method, path, resp.Status, summarizeAppleError(respBody))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("failed decoding Apple API response: %w", err)
	}
	return nil
}

func (c *AppStoreConnectClient) bearerToken() (string, error) {
	if c.token != "" && time.Now().Before(c.tokenExpiry.Add(-time.Minute)) {
		return c.token, nil
	}
	token, expiry, err := makeJWT(c.creds)
	if err != nil {
		return "", err
	}
	c.token = token
	c.tokenExpiry = expiry
	return token, nil
}

func makeJWT(creds AppStoreConnectCredentials) (string, time.Time, error) {
	block, _ := pem.Decode(creds.PrivateKey)
	if block == nil {
		return "", time.Time{}, fmt.Errorf("failed to decode App Store Connect private key PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse App Store Connect private key: %w", err)
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", time.Time{}, fmt.Errorf("App Store Connect private key must be an EC private key")
	}

	now := time.Now().UTC()
	expiry := now.Add(20 * time.Minute)
	header := map[string]string{"alg": "ES256", "kid": creds.KeyID, "typ": "JWT"}
	claims := map[string]interface{}{
		"iss": creds.IssuerID,
		"iat": now.Unix(),
		"exp": expiry.Unix(),
		"aud": "appstoreconnect-v1",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	sum := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, ecKey, sum[:])
	if err != nil {
		return "", time.Time{}, err
	}
	signature := joseECDSASignature(r, s, ecKey)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), expiry, nil
}

func joseECDSASignature(r, s *big.Int, key *ecdsa.PrivateKey) []byte {
	size := (key.Curve.Params().BitSize + 7) / 8
	out := make([]byte, size*2)
	r.FillBytes(out[:size])
	s.FillBytes(out[size:])
	return out
}

func GenerateCertificateRequest(commonName string) (*rsa.PrivateKey, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generating signing private key: %w", err)
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkixName(commonName),
	}, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generating certificate signing request: %w", err)
	}
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER})
	return privateKey, csrPEM, nil
}

func pkixName(commonName string) pkix.Name {
	if commonName == "" {
		commonName = "go-ios development"
	}
	return pkix.Name{CommonName: commonName}
}

type jsonAPIResourceResponse struct {
	Data jsonAPIResource `json:"data"`
}

type jsonAPIListResponse struct {
	Data []jsonAPIResource `json:"data"`
}

type jsonAPIResource struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
}

type appleErrorResponse struct {
	Errors []struct {
		Status string `json:"status"`
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

func summarizeAppleError(body []byte) string {
	var errResp appleErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && len(errResp.Errors) > 0 {
		parts := make([]string, 0, len(errResp.Errors))
		for _, appleErr := range errResp.Errors {
			parts = append(parts, strings.TrimSpace(appleErr.Title+": "+appleErr.Detail))
		}
		return strings.Join(parts, "; ")
	}
	return strings.TrimSpace(string(body))
}
