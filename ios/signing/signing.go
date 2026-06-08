package signing

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aluedeke/go-codesign/pkg/codesign"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"software.sslmate.com/src/go-pkcs12"
)

const logModule = "go-ios/signing"

type PrepareAndSignOptions struct {
	AppPath     string
	OutputPath  string
	BundleID    string
	BundleName  string
	ProfileName string
	DeviceName  string
	P12Password string
	P12Output   string
	ProfileOut  string
	Credentials AppStoreConnectCredentials
	Device      ios.DeviceEntry
}

type PrepareAssetsOptions struct {
	BundleID    string
	BundleName  string
	ProfileName string
	DeviceName  string
	P12Password string
	P12Output   string
	ProfileOut  string
	Credentials AppStoreConnectCredentials
	Device      ios.DeviceEntry
	// RevokeExisting revokes every existing iOS Development certificate before
	// creating a new one. Apple allows only one current certificate of that type,
	// so without this a second create fails with a 409. Intended for a dedicated
	// (CI) account where the signing identity is disposable.
	RevokeExisting bool
	// CertificateID, when set, provisions a profile against an existing
	// certificate (by App Store Connect resource id) instead of creating a new
	// one. No certificate is created or revoked and no P12 is written — the caller
	// already holds the matching P12. This lets CI mint per-bundle profiles in
	// parallel against one shared, pre-provisioned certificate.
	CertificateID string
}

type PrepareAndSignResult struct {
	OutputPath  string
	P12Path     string
	ProfilePath string
	BundleID    string
}

type PrepareAssetsResult struct {
	P12Path     string
	ProfilePath string
	BundleID    string
	// CertificateID is the App Store Connect resource id of the certificate used
	// (created, or reused via PrepareAssetsOptions.CertificateID). Store it so
	// later profile-only provisioning can reference the same certificate.
	CertificateID string
}

type PrepareCertificateOptions struct {
	P12Password string
	P12Output   string
	Credentials AppStoreConnectCredentials
	// RevokeExisting revokes every existing iOS Development certificate first; see
	// PrepareAssetsOptions.RevokeExisting.
	RevokeExisting bool
}

type PrepareCertificateResult struct {
	P12Path       string
	CertificateID string
}

type SignWithFilesOptions struct {
	AppPath     string
	OutputPath  string
	BundleID    string
	P12Path     string
	P12Password string
	ProfilePath string
}

type SignWithFilesResult struct {
	OutputPath string
	BundleID   string
}

func PrepareAndSign(ctx context.Context, opts PrepareAndSignOptions) (PrepareAndSignResult, error) {
	if opts.AppPath == "" {
		return PrepareAndSignResult{}, fmt.Errorf("app path is required")
	}
	if opts.OutputPath == "" {
		opts.OutputPath = defaultSignedOutputPath(opts.AppPath)
	}
	if opts.P12Output == "" {
		opts.P12Output = strings.TrimSuffix(opts.OutputPath, filepath.Ext(opts.OutputPath)) + ".p12"
	}
	if opts.ProfileOut == "" {
		opts.ProfileOut = strings.TrimSuffix(opts.OutputPath, filepath.Ext(opts.OutputPath)) + ".mobileprovision"
	}
	if opts.BundleID == "" {
		bundleID, err := getBundleID(opts.AppPath)
		if err != nil {
			return PrepareAndSignResult{}, err
		}
		opts.BundleID = bundleID
	}
	assets, err := PrepareSigningAssets(ctx, PrepareAssetsOptions{
		BundleID:    opts.BundleID,
		BundleName:  opts.BundleName,
		ProfileName: opts.ProfileName,
		DeviceName:  opts.DeviceName,
		P12Password: opts.P12Password,
		P12Output:   opts.P12Output,
		ProfileOut:  opts.ProfileOut,
		Credentials: opts.Credentials,
		Device:      opts.Device,
	})
	if err != nil {
		return PrepareAndSignResult{}, err
	}
	p12, err := os.ReadFile(assets.P12Path)
	if err != nil {
		return PrepareAndSignResult{}, fmt.Errorf("failed reading generated P12: %w", err)
	}
	profile, err := os.ReadFile(assets.ProfilePath)
	if err != nil {
		return PrepareAndSignResult{}, fmt.Errorf("failed reading generated provisioning profile: %w", err)
	}

	if err := ResignApp(opts.AppPath, opts.OutputPath, p12, opts.P12Password, profile, opts.BundleID); err != nil {
		return PrepareAndSignResult{}, err
	}
	return PrepareAndSignResult{
		OutputPath:  opts.OutputPath,
		P12Path:     assets.P12Path,
		ProfilePath: assets.ProfilePath,
		BundleID:    opts.BundleID,
	}, nil
}

func PrepareSigningAssets(ctx context.Context, opts PrepareAssetsOptions) (PrepareAssetsResult, error) {
	if opts.BundleID == "" {
		return PrepareAssetsResult{}, fmt.Errorf("bundle ID is required")
	}
	if opts.P12Output == "" {
		opts.P12Output = opts.BundleID + ".p12"
	}
	if opts.ProfileOut == "" {
		opts.ProfileOut = opts.BundleID + ".mobileprovision"
	}
	if opts.ProfileName == "" {
		opts.ProfileName = fmt.Sprintf("go-ios %s %s", opts.BundleID, time.Now().UTC().Format("20060102150405"))
	}

	client := NewAppStoreConnectClient(opts.Credentials)

	bundleResourceID, err := client.EnsureBundleID(ctx, opts.BundleID, opts.BundleName)
	if err != nil {
		return PrepareAssetsResult{}, fmt.Errorf("failed ensuring bundle ID: %w", err)
	}
	deviceResourceID, err := client.EnsureDevice(ctx, opts.Device.Properties.SerialNumber, opts.DeviceName)
	if err != nil {
		return PrepareAssetsResult{}, fmt.Errorf("failed ensuring device: %w", err)
	}

	result := PrepareAssetsResult{ProfilePath: opts.ProfileOut, BundleID: opts.BundleID, CertificateID: opts.CertificateID}

	// Reuse mode: provision a profile against an existing certificate. No
	// certificate is created/revoked and no P12 is written.
	if opts.CertificateID == "" {
		privateKey, csrPEM, err := GenerateCertificateRequest("go-ios " + opts.BundleID)
		if err != nil {
			return PrepareAssetsResult{}, err
		}
		if opts.RevokeExisting {
			if err := revokeAllDevelopmentCertificates(ctx, client); err != nil {
				return PrepareAssetsResult{}, fmt.Errorf("failed revoking existing certificates: %w", err)
			}
		}
		certificateResourceID, certDER, err := client.CreateCertificate(ctx, csrPEM)
		if err != nil {
			return PrepareAssetsResult{}, fmt.Errorf("failed creating certificate: %w", err)
		}
		result.CertificateID = certificateResourceID

		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return PrepareAssetsResult{}, fmt.Errorf("failed parsing Apple certificate: %w", err)
		}
		p12, err := pkcs12.Encode(rand.Reader, privateKey, cert, nil, opts.P12Password)
		if err != nil {
			return PrepareAssetsResult{}, fmt.Errorf("failed encoding P12: %w", err)
		}
		if err := writeFile(opts.P12Output, p12, 0600); err != nil {
			return PrepareAssetsResult{}, err
		}
		result.P12Path = opts.P12Output
	}

	profile, err := client.CreateDevelopmentProfile(ctx, opts.ProfileName, bundleResourceID, result.CertificateID, deviceResourceID)
	if err != nil {
		return PrepareAssetsResult{}, fmt.Errorf("failed creating provisioning profile: %w", err)
	}
	if err := writeFile(opts.ProfileOut, profile, 0644); err != nil {
		return PrepareAssetsResult{}, err
	}
	return result, nil
}

// PrepareCertificate creates one iOS Development certificate and writes its P12
// (certificate + private key). Unlike PrepareSigningAssets it needs no device,
// bundle id, or provisioning profile — it is the account-wide half of signing,
// used to mint the shared identity that profile-only provisioning
// (PrepareAssetsOptions.CertificateID) then reuses. Because it needs no device it
// can run on a hosted CI runner.
func PrepareCertificate(ctx context.Context, opts PrepareCertificateOptions) (PrepareCertificateResult, error) {
	if opts.P12Output == "" {
		opts.P12Output = "identity.p12"
	}
	privateKey, csrPEM, err := GenerateCertificateRequest("go-ios signing identity")
	if err != nil {
		return PrepareCertificateResult{}, err
	}
	client := NewAppStoreConnectClient(opts.Credentials)
	if opts.RevokeExisting {
		if err := revokeAllDevelopmentCertificates(ctx, client); err != nil {
			return PrepareCertificateResult{}, fmt.Errorf("failed revoking existing certificates: %w", err)
		}
	}
	certificateResourceID, certDER, err := client.CreateCertificate(ctx, csrPEM)
	if err != nil {
		return PrepareCertificateResult{}, fmt.Errorf("failed creating certificate: %w", err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return PrepareCertificateResult{}, fmt.Errorf("failed parsing Apple certificate: %w", err)
	}
	p12, err := pkcs12.Encode(rand.Reader, privateKey, cert, nil, opts.P12Password)
	if err != nil {
		return PrepareCertificateResult{}, fmt.Errorf("failed encoding P12: %w", err)
	}
	if err := writeFile(opts.P12Output, p12, 0600); err != nil {
		return PrepareCertificateResult{}, err
	}
	return PrepareCertificateResult{P12Path: opts.P12Output, CertificateID: certificateResourceID}, nil
}

// revokeAllDevelopmentCertificates revokes every IOS_DEVELOPMENT certificate on
// the account. Apple permits only one current certificate of this type, and a
// leftover one (whatever its name) blocks creating a new one, so a reliable
// "create a fresh certificate" needs to clear them all first. Only use this on a
// dedicated signing account — it does not discriminate by name.
func revokeAllDevelopmentCertificates(ctx context.Context, client *AppStoreConnectClient) error {
	certs, err := client.ListDevelopmentCertificates(ctx)
	if err != nil {
		return err
	}
	for _, cert := range certs {
		name, _ := cert.Attributes["name"].(string)
		displayName, _ := cert.Attributes["displayName"].(string)
		golog.Info("revoking existing iOS Development certificate",
			"module", logModule, "id", cert.ID, "name", name, "displayName", displayName)
		if err := client.RevokeCertificate(ctx, cert.ID); err != nil {
			return fmt.Errorf("revoking certificate %s: %w", cert.ID, err)
		}
	}
	return nil
}

func SignWithFiles(opts SignWithFilesOptions) (SignWithFilesResult, error) {
	if opts.AppPath == "" {
		return SignWithFilesResult{}, fmt.Errorf("app path is required")
	}
	if opts.P12Path == "" {
		return SignWithFilesResult{}, fmt.Errorf("P12 path is required")
	}
	if opts.ProfilePath == "" {
		return SignWithFilesResult{}, fmt.Errorf("provisioning profile path is required")
	}
	if opts.OutputPath == "" {
		opts.OutputPath = defaultSignedOutputPath(opts.AppPath)
	}
	if opts.BundleID == "" {
		bundleID, err := getBundleID(opts.AppPath)
		if err != nil {
			return SignWithFilesResult{}, err
		}
		opts.BundleID = bundleID
	}
	p12, err := os.ReadFile(opts.P12Path)
	if err != nil {
		return SignWithFilesResult{}, fmt.Errorf("failed reading P12: %w", err)
	}
	profile, err := os.ReadFile(opts.ProfilePath)
	if err != nil {
		return SignWithFilesResult{}, fmt.Errorf("failed reading provisioning profile: %w", err)
	}
	if err := ResignApp(opts.AppPath, opts.OutputPath, p12, opts.P12Password, profile, opts.BundleID); err != nil {
		return SignWithFilesResult{}, err
	}
	return SignWithFilesResult{OutputPath: opts.OutputPath, BundleID: opts.BundleID}, nil
}

func ResignApp(inputPath, outputPath string, p12 []byte, password string, profile []byte, bundleID string) error {
	isIPA := strings.EqualFold(filepath.Ext(inputPath), ".ipa")
	var appPath string
	var tempDir string
	var cleanup bool
	defer func() {
		if cleanup {
			_ = os.RemoveAll(tempDir)
		}
	}()

	if isIPA {
		var err error
		tempDir, err = codesign.ExtractIPA(inputPath)
		if err != nil {
			return fmt.Errorf("failed extracting IPA: %w", err)
		}
		cleanup = true
		appPath, err = codesign.FindAppBundle(tempDir)
		if err != nil {
			return fmt.Errorf("failed finding app bundle in IPA: %w", err)
		}
	} else {
		var err error
		tempDir, err = os.MkdirTemp("", "go-ios-sign-*")
		if err != nil {
			return fmt.Errorf("failed creating temp dir: %w", err)
		}
		cleanup = true
		appPath = filepath.Join(tempDir, filepath.Base(inputPath))
		if err := codesign.CopyAppBundle(inputPath, appPath); err != nil {
			return fmt.Errorf("failed copying app bundle: %w", err)
		}
	}

	if err := codesign.Resign(codesign.ResignOptions{
		AppPath:             appPath,
		P12Data:             p12,
		P12Password:         password,
		ProvisioningProfile: profile,
		NewBundleID:         bundleID,
	}); err != nil {
		return fmt.Errorf("failed signing app: %w", err)
	}

	if isIPA {
		if err := codesign.RepackageIPA(tempDir, outputPath); err != nil {
			return fmt.Errorf("failed repackaging IPA: %w", err)
		}
		return nil
	}
	if err := os.RemoveAll(outputPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed removing existing output app: %w", err)
	}
	if err := codesign.CopyAppBundle(appPath, outputPath); err != nil {
		return fmt.Errorf("failed writing signed app bundle: %w", err)
	}
	return nil
}

func getBundleID(appPath string) (string, error) {
	if strings.EqualFold(filepath.Ext(appPath), ".ipa") {
		tmpDir, err := codesign.ExtractIPA(appPath)
		if err != nil {
			return "", fmt.Errorf("failed extracting IPA: %w", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()
		bundlePath, err := codesign.FindAppBundle(tmpDir)
		if err != nil {
			return "", fmt.Errorf("failed finding app bundle in IPA: %w", err)
		}
		return codesign.GetAppBundleID(bundlePath)
	}
	return codesign.GetAppBundleID(appPath)
}

func defaultSignedOutputPath(appPath string) string {
	ext := filepath.Ext(appPath)
	if ext == "" {
		return appPath + "-signed"
	}
	return strings.TrimSuffix(appPath, ext) + "-signed" + ext
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed creating output directory: %w", err)
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed writing %s: %w", path, err)
	}
	return nil
}
