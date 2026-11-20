package api

import (
	"context"

	"github.com/danielpaulus/go-ios/ios/signing"
)

// Signer is the seam between the REST signing endpoints and whatever code
// actually talks to App Store Connect and rewrites app signatures. The
// endpoints in sign_endpoints.go depend ONLY on this interface, so the concrete
// implementation is swappable:
//
//   - The default wired below is signingAdapter, a thin wrapper over the current
//     ios/signing package.
//   - The in-repo ios/codesign package from branch daniel/codesign-command is the
//     intended future implementation. When it lands, provide a second adapter
//     that satisfies Signer over ios/codesign and swap defaultSigner (or make it
//     injectable) — no endpoint or test change is required.
//
// Keeping the surface small (three operations) is deliberate: it is the minimum
// the endpoints need and the maximum a swapped-in implementation must provide.
type Signer interface {
	// PrepareCertificate creates one signing certificate via App Store Connect
	// and returns its P12 bytes (certificate + private key) plus the ASC
	// resource id. Device-free: it is the account-wide half of signing.
	PrepareCertificate(ctx context.Context, creds ASCCredentials, opts PrepareCertOptions) (CertResult, error)

	// PrepareProvisioning creates a bundle id, a development certificate (unless
	// reusing one) and a development provisioning profile, returning the profile
	// bytes and — when a certificate was created — the P12 bytes. Needs a device
	// udid so Apple can register it against the profile.
	PrepareProvisioning(ctx context.Context, creds ASCCredentials, opts PrepareProvisioningOptions) (AssetsResult, error)

	// SignApp resigns the app/IPA at inputPath using the given P12 and
	// provisioning profile files, writing the signed artifact to a path it
	// chooses and returns. It never mutates the input.
	SignApp(ctx context.Context, opts SignAppOptions) (SignResult, error)
}

// ASCCredentials carries the App Store Connect API key material. The private key
// is the raw bytes of the .p8 file. Never log this struct.
type ASCCredentials struct {
	KeyID      string
	IssuerID   string
	PrivateKey []byte
}

// PrepareCertOptions are the inputs to PrepareCertificate.
type PrepareCertOptions struct {
	P12Password    string
	P12Output      string
	RevokeExisting bool
}

// CertResult is the output of PrepareCertificate: the P12 bytes and the ASC
// certificate resource id.
type CertResult struct {
	P12           []byte
	CertificateID string
}

// PrepareProvisioningOptions are the inputs to PrepareProvisioning.
type PrepareProvisioningOptions struct {
	BundleID       string
	BundleName     string
	ProfileName    string
	DeviceName     string
	DeviceUDID     string
	P12Password    string
	RevokeExisting bool
	// CertificateID, when set, provisions a profile against an existing
	// certificate instead of creating a new one; no P12 is produced.
	CertificateID string
}

// AssetsResult is the output of PrepareProvisioning. P12 is nil in reuse mode
// (CertificateID set on the request).
type AssetsResult struct {
	P12           []byte
	Profile       []byte
	BundleID      string
	CertificateID string
}

// SignAppOptions are the inputs to SignApp. P12 and Profile are the raw file
// bytes; P12Password may be empty. BundleID is optional (derived from the app
// when empty).
type SignAppOptions struct {
	InputPath   string
	P12         []byte
	P12Password string
	Profile     []byte
	BundleID    string
}

// SignResult is the output of SignApp: the path of the signed artifact and the
// resolved bundle id.
type SignResult struct {
	OutputPath string
	BundleID   string
}

// signingAdapter is the default Signer, implemented over the current ios/signing
// package. It exists so the endpoints don't import ios/signing directly and so a
// future ios/codesign implementation can replace it without touching endpoints.
type signingAdapter struct{}

// Compile-time assertion that the concrete adapter satisfies the interface.
var _ Signer = signingAdapter{}

// defaultSigner is the Signer the endpoints use. Swap this (or make it
// injectable) to move to ios/codesign later.
var defaultSigner Signer = signingAdapter{}

func (signingAdapter) toCreds(creds ASCCredentials) signing.AppStoreConnectCredentials {
	return signing.AppStoreConnectCredentials{
		KeyID:      creds.KeyID,
		IssuerID:   creds.IssuerID,
		PrivateKey: creds.PrivateKey,
	}
}

func (a signingAdapter) PrepareCertificate(ctx context.Context, creds ASCCredentials, opts PrepareCertOptions) (CertResult, error) {
	res, err := signing.PrepareCertificate(ctx, signing.PrepareCertificateOptions{
		P12Password:    opts.P12Password,
		P12Output:      opts.P12Output,
		RevokeExisting: opts.RevokeExisting,
		Credentials:    a.toCreds(creds),
	})
	if err != nil {
		return CertResult{}, err
	}
	p12, err := readAndRemove(res.P12Path)
	if err != nil {
		return CertResult{}, err
	}
	return CertResult{P12: p12, CertificateID: res.CertificateID}, nil
}

func (a signingAdapter) PrepareProvisioning(ctx context.Context, creds ASCCredentials, opts PrepareProvisioningOptions) (AssetsResult, error) {
	res, err := signing.PrepareSigningAssets(ctx, signing.PrepareAssetsOptions{
		BundleID:       opts.BundleID,
		BundleName:     opts.BundleName,
		ProfileName:    opts.ProfileName,
		DeviceName:     opts.DeviceName,
		P12Password:    opts.P12Password,
		RevokeExisting: opts.RevokeExisting,
		CertificateID:  opts.CertificateID,
		Credentials:    a.toCreds(creds),
		Device:         deviceEntryForUDID(opts.DeviceUDID),
	})
	if err != nil {
		return AssetsResult{}, err
	}
	out := AssetsResult{BundleID: res.BundleID, CertificateID: res.CertificateID}
	profile, err := readAndRemove(res.ProfilePath)
	if err != nil {
		return AssetsResult{}, err
	}
	out.Profile = profile
	// In reuse mode PrepareSigningAssets writes no P12.
	if res.P12Path != "" {
		p12, err := readAndRemove(res.P12Path)
		if err != nil {
			return AssetsResult{}, err
		}
		out.P12 = p12
	}
	return out, nil
}

func (a signingAdapter) SignApp(ctx context.Context, opts SignAppOptions) (SignResult, error) {
	// ResignApp works directly from in-memory p12/profile bytes, so no key
	// material is ever written to disk here. We choose the output path alongside
	// the input so it stays inside the caller's per-request temp dir.
	outputPath := signedOutputPath(opts.InputPath)
	if err := signing.ResignApp(opts.InputPath, outputPath, opts.P12, opts.P12Password, opts.Profile, opts.BundleID); err != nil {
		return SignResult{}, err
	}
	return SignResult{OutputPath: outputPath, BundleID: opts.BundleID}, nil
}
