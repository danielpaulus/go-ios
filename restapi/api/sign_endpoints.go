package api

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/gin-gonic/gin"
	"software.sslmate.com/src/go-pkcs12"
)

// Signing / codesigning REST endpoints.
//
// SECRETS: these handlers move App Store Connect .p8 keys, P12 identities and
// their passwords through memory and short-lived temp files. NONE of that
// material is ever logged — logs carry only non-secret metadata (bundle id,
// artifact size, cert resource id). Uploads are bounded by maxUploadBytes and
// written to a per-request temp dir that is always removed on return.
//
// The endpoints depend only on the Signer interface (see signing_adapter.go);
// the concrete implementation is swappable for ios/codesign later.

// Signing-specific sentinel errors.
var (
	errMissingASCKey    = errors.New("missing required multipart 'asc-private-key' (.p8 App Store Connect key)")
	errMissingASCFields = errors.New("both 'asc-key-id' and 'asc-issuer-id' are required")
	errMissingIPA       = errors.New("missing required multipart 'ipa' (app or .ipa upload)")
	errMissingP12File   = errors.New("missing required multipart 'p12file'")
	errMissingSignProf  = errors.New("missing required multipart 'profile' (mobileprovision)")
	errMissingSupCert   = errors.New("missing required multipart 'cert' (supervision identity: DER/PEM/P12)")
)

// registerSignHostRoutes registers host-local (device-free) codesigning routes
// under /api/v1/sign plus the host-level supervision cert creation route. These
// need no device.
func registerSignHostRoutes(router *gin.RouterGroup) {
	sign := router.Group("/sign")
	sign.POST("/certificate", SignCertificate)
	sign.POST("/provision", SignProvision)
	sign.POST("/app", SignApp)

	// Supervision-cert generation, mirroring `ios prepare create-cert`. Device-
	// free, so it lives at the host level next to the other prepare routes.
	router.POST("/prepare/create-cert", PrepareCreateCert)
}

// registerSignDeviceRoutes registers device-scoped preparation routes under
// /device/:udid.
func registerSignDeviceRoutes(device *gin.RouterGroup) {
	device.POST("/prepare", PrepareDevice)
}

// SignCertificate creates one App Store Connect signing certificate and returns
// its P12 (certificate + private key) as a downloadable application/x-pkcs12
// file. The generated P12 password is echoed in the X-P12-Password response
// header (it is client-supplied, or empty). Device-free.
// @Summary Create a signing certificate (App Store Connect)
// @Accept multipart/form-data
// @Produce application/x-pkcs12
// @Param asc-private-key formData file true ".p8 App Store Connect API key"
// @Param asc-key-id formData string true "App Store Connect key id"
// @Param asc-issuer-id formData string true "App Store Connect issuer id"
// @Param revoke-existing formData bool false "revoke existing iOS Development certificates first"
// @Param p12password formData string false "password to protect the generated P12"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sign/certificate [post]
func SignCertificate(c *gin.Context) {
	creds, err := readASCCredentials(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	p12password := c.PostForm("p12password")
	revoke := parseBoolForm(c.PostForm("revoke-existing"))

	res, err := defaultSigner.PrepareCertificate(c.Request.Context(), creds, PrepareCertOptions{
		P12Password:    p12password,
		RevokeExisting: revoke,
	})
	if err != nil {
		RespondError(c, http.StatusBadGateway, err)
		return
	}
	golog.Info("created signing certificate", "module", logModule, "certificateId", res.CertificateID, "p12Bytes", len(res.P12))
	c.Header("X-Certificate-Id", res.CertificateID)
	c.Header("X-P12-Password", p12password)
	c.Data(http.StatusOK, "application/x-pkcs12", res.P12)
}

// SignProvision creates a bundle id, development certificate and provisioning
// profile via App Store Connect and returns both artifacts base64-encoded in a
// JSON envelope (so a single response can carry the .mobileprovision and .p12).
// Device-free at the host level: the target device udid is supplied as a form
// field so Apple can register it against the profile.
// @Summary Create a provisioning profile + P12 (App Store Connect)
// @Accept multipart/form-data
// @Produce json
// @Param asc-private-key formData file true ".p8 App Store Connect API key"
// @Param asc-key-id formData string true "App Store Connect key id"
// @Param asc-issuer-id formData string true "App Store Connect issuer id"
// @Param bundleid formData string true "app bundle identifier"
// @Param udid formData string true "target device udid"
// @Param bundlename formData string false "bundle display name"
// @Param profilename formData string false "provisioning profile name"
// @Param devicename formData string false "device display name"
// @Param certificate-id formData string false "reuse an existing certificate (no new P12)"
// @Param revoke-existing formData bool false "revoke existing certificates first"
// @Param p12password formData string false "password to protect the generated P12"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sign/provision [post]
func SignProvision(c *gin.Context) {
	creds, err := readASCCredentials(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	bundleID := strings.TrimSpace(c.PostForm("bundleid"))
	if bundleID == "" {
		RespondError(c, http.StatusBadRequest, errMissingBundleID)
		return
	}
	udid := strings.TrimSpace(c.PostForm("udid"))
	p12password := c.PostForm("p12password")

	res, err := defaultSigner.PrepareProvisioning(c.Request.Context(), creds, PrepareProvisioningOptions{
		BundleID:       bundleID,
		BundleName:     c.PostForm("bundlename"),
		ProfileName:    c.PostForm("profilename"),
		DeviceName:     c.PostForm("devicename"),
		DeviceUDID:     udid,
		P12Password:    p12password,
		RevokeExisting: parseBoolForm(c.PostForm("revoke-existing")),
		CertificateID:  strings.TrimSpace(c.PostForm("certificate-id")),
	})
	if err != nil {
		RespondError(c, http.StatusBadGateway, err)
		return
	}
	golog.Info("created provisioning assets", "module", logModule, "udid", udid, "bundleId", res.BundleID,
		"certificateId", res.CertificateID, "profileBytes", len(res.Profile), "p12Bytes", len(res.P12))

	resp := gin.H{
		"bundleId":              res.BundleID,
		"certificateId":         res.CertificateID,
		"mobileprovisionBase64": encodeBase64(res.Profile),
	}
	if res.P12 != nil {
		resp["p12Base64"] = encodeBase64(res.P12)
		resp["p12Password"] = p12password
	}
	c.JSON(http.StatusOK, resp)
}

// SignApp resigns an uploaded app/IPA with an uploaded P12 identity and
// provisioning profile, streaming the signed IPA back as application/octet-
// stream. Uploads go to a per-request temp dir that is removed on return.
//
// NOTE: this is synchronous for v1. Large IPAs make it a prime candidate for the
// async jobs subsystem (jobs.go) — resign in a background job and let the client
// poll/download — but that is out of scope here.
// @Summary Resign an app/IPA with a P12 + provisioning profile
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param ipa formData file true "app or .ipa to resign"
// @Param p12file formData file true "signing identity (.p12)"
// @Param profile formData file true "provisioning profile (.mobileprovision)"
// @Param p12password formData string false "P12 password"
// @Param bundleid formData string false "override bundle id"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sign/app [post]
func SignApp(c *gin.Context) {
	tmpDir, cleanup, err := newRequestTempDir("go-ios-sign-app-")
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer cleanup()

	ipaName, _ := formFileName(c, "ipa")
	ext := filepath.Ext(ipaName)
	if ext == "" {
		ext = ".ipa"
	}
	ipaPath, err := saveFormFileToDir(c, "ipa", tmpDir, "input"+ext)
	if err != nil {
		RespondError(c, http.StatusBadRequest, errMissingIPA)
		return
	}
	p12, err := readFormFile(c, "p12file")
	if err != nil {
		RespondError(c, http.StatusBadRequest, errMissingP12File)
		return
	}
	profile, err := readFormFile(c, "profile")
	if err != nil {
		RespondError(c, http.StatusBadRequest, errMissingSignProf)
		return
	}

	res, err := defaultSigner.SignApp(c.Request.Context(), SignAppOptions{
		InputPath:   ipaPath,
		P12:         p12,
		P12Password: c.PostForm("p12password"),
		Profile:     profile,
		BundleID:    strings.TrimSpace(c.PostForm("bundleid")),
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	info, err := os.Stat(res.OutputPath)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	golog.Info("resigned app", "module", logModule, "bundleId", res.BundleID, "signedBytes", info.Size())
	downloadName := "signed" + filepath.Ext(res.OutputPath)
	c.Header("Content-Disposition", "attachment; filename=\""+downloadName+"\"")
	c.File(res.OutputPath)
}

// PrepareCreateCert generates a self-signed supervision identity (CLI: ios
// prepare create-cert) and returns it as a JSON envelope containing the DER and
// PEM certificate and private key, base64-encoded. Device-free.
// @Summary Generate a device supervision certificate
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /prepare/create-cert [post]
func PrepareCreateCert(c *gin.Context) {
	cert, err := ios.CreateDERFormattedSupervisionCert()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	golog.Info("generated supervision certificate", "module", logModule, "certDerBytes", len(cert.CertDER))
	c.JSON(http.StatusOK, gin.H{
		"certDerBase64":       encodeBase64(cert.CertDER),
		"certPem":             string(cert.CertPEM),
		"privateKeyDerBase64": encodeBase64(cert.PrivateKeyDER),
		"privateKeyPem":       string(cert.PrivateKeyPEM),
	})
}

// PrepareDevice runs the device preparation/provisioning flow (CLI: ios
// prepare). Send multipart/form-data. To supervise the device, include a "cert"
// file (DER/PEM/P12 supervision identity) and optional "p12password"; without a
// cert the device is prepared without supervision. Optional fields: repeated
// "skip" values (see /prepare/skip-options), "orgname", "locale", "lang".
// @Summary Prepare (and optionally supervise) a device
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param cert formData file false "supervision identity (DER/PEM/P12)"
// @Param p12password formData string false "P12 password (if cert is a P12)"
// @Param skip formData []string false "setup panes to skip"
// @Param orgname formData string false "supervision organization name"
// @Param locale formData string false "device locale (default en_US)"
// @Param lang formData string false "device language (default en)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/prepare [post]
func PrepareDevice(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	var certDER []byte
	if raw, err := readFormFile(c, "cert"); err == nil && len(raw) > 0 {
		der, derErr := extractSupervisionCertDER(raw, c.PostForm("p12password"))
		if derErr != nil {
			RespondError(c, http.StatusBadRequest, derErr)
			return
		}
		certDER = der
	}

	skip := c.PostFormArray("skip")
	orgname := c.PostForm("orgname")
	locale := c.PostForm("locale")
	lang := c.PostForm("lang")

	golog.Info("preparing device", "module", logModule, "udid", device.Properties.SerialNumber,
		"supervise", certDER != nil, "skipCount", len(skip))
	if err := mcinstall.Prepare(device, skip, certDER, orgname, locale, lang); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "prepared", "supervised": certDER != nil})
}

// --- helpers ---------------------------------------------------------------

// readASCCredentials reads the App Store Connect .p8 key and its id fields from a
// multipart request. The private key bytes are held only in memory and never
// logged.
func readASCCredentials(c *gin.Context) (ASCCredentials, error) {
	keyID := strings.TrimSpace(c.PostForm("asc-key-id"))
	issuerID := strings.TrimSpace(c.PostForm("asc-issuer-id"))
	if keyID == "" || issuerID == "" {
		return ASCCredentials{}, errMissingASCFields
	}
	key, err := readFormFile(c, "asc-private-key")
	if err != nil || len(key) == 0 {
		return ASCCredentials{}, errMissingASCKey
	}
	return ASCCredentials{KeyID: keyID, IssuerID: issuerID, PrivateKey: key}, nil
}

// parseBoolForm interprets a form field as a boolean, treating parse failures and
// empty values as false.
func parseBoolForm(v string) bool {
	if v == "" {
		return false
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return false
	}
	return b
}

// newRequestTempDir creates a per-request temp dir and returns a cleanup func
// that removes it (logging, never failing, on error).
func newRequestTempDir(prefix string) (string, func(), error) {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			golog.Warn("failed to remove request temp dir", "module", logModule, "path", dir, "err", err.Error())
		}
	}
	return dir, cleanup, nil
}

// saveFormFileToDir reads a bounded multipart file field and writes it into dir
// under name, returning the written path.
func saveFormFileToDir(c *gin.Context, field, dir, name string) (string, error) {
	data, err := readFormFile(c, field)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errUploadTooLarge // treated as invalid by callers mapping to 400
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return path, nil
}

// formFileName returns the client-supplied filename for a multipart file field.
func formFileName(c *gin.Context, field string) (string, error) {
	_, hdr, err := c.Request.FormFile(field)
	if err != nil {
		return "", err
	}
	return hdr.Filename, nil
}

// readAndRemove reads a file fully then removes it. Used to pull generated
// artifacts (P12/profile) off disk and immediately delete the on-disk copy so
// key material does not linger.
func readAndRemove(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if rmErr := os.Remove(path); rmErr != nil {
		golog.Warn("failed to remove generated artifact", "module", logModule, "path", path, "err", rmErr.Error())
	}
	return data, nil
}

// signedOutputPath derives a "-signed" sibling path for the given input,
// preserving the extension so IPAs stay .ipa.
func signedOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return inputPath + "-signed"
	}
	return strings.TrimSuffix(inputPath, ext) + "-signed" + ext
}

// deviceEntryForUDID builds a minimal DeviceEntry carrying just the udid, which
// is all PrepareSigningAssets needs to register the device with Apple.
func deviceEntryForUDID(udid string) ios.DeviceEntry {
	return ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}}
}

// encodeBase64 is a tiny wrapper so artifact-encoding call sites read clearly.
func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// errUnparseableCert is returned when a supervision identity upload is not
// recognizable as DER, PEM, or (with a password) P12.
var errUnparseableCert = errors.New("unable to parse supervision certificate: not valid DER, PEM, or PKCS12")

// extractSupervisionCertDER normalizes an uploaded supervision identity
// (raw DER, PEM, PEM-with-metadata, or P12 when a password is given) to raw DER,
// which is what mcinstall.Prepare requires. The p12password, if present, is used
// only to decrypt and is never logged.
func extractSupervisionCertDER(raw []byte, p12password string) ([]byte, error) {
	// Raw DER.
	if _, err := x509.ParseCertificate(raw); err == nil {
		return raw, nil
	}
	// PEM, possibly with leading metadata (e.g. OpenSSL "Bag Attributes").
	if start := bytes.Index(raw, []byte("-----BEGIN CERTIFICATE-----")); start != -1 {
		if block, _ := pem.Decode(raw[start:]); block != nil && block.Type == "CERTIFICATE" {
			if _, err := x509.ParseCertificate(block.Bytes); err == nil {
				return block.Bytes, nil
			}
		}
	}
	// P12 (needs a password).
	if p12password != "" {
		if _, cert, err := pkcs12.Decode(raw, p12password); err == nil && cert != nil {
			return cert.Raw, nil
		}
	}
	return nil, errUnparseableCert
}
