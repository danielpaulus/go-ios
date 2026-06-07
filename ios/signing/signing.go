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

	privateKey, csrPEM, err := GenerateCertificateRequest("go-ios " + opts.BundleID)
	if err != nil {
		return PrepareAssetsResult{}, err
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
	certificateResourceID, certDER, err := client.CreateCertificate(ctx, csrPEM)
	if err != nil {
		return PrepareAssetsResult{}, fmt.Errorf("failed creating certificate: %w", err)
	}
	profile, err := client.CreateDevelopmentProfile(ctx, opts.ProfileName, bundleResourceID, certificateResourceID, deviceResourceID)
	if err != nil {
		return PrepareAssetsResult{}, fmt.Errorf("failed creating provisioning profile: %w", err)
	}

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
	if err := writeFile(opts.ProfileOut, profile, 0644); err != nil {
		return PrepareAssetsResult{}, err
	}
	return PrepareAssetsResult{
		P12Path:     opts.P12Output,
		ProfilePath: opts.ProfileOut,
		BundleID:    opts.BundleID,
	}, nil
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
