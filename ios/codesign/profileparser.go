package codesign

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pkcs12"

	"github.com/fullsailor/pkcs7"
	plist "howett.net/plist"
)

// ProfileAndCertificate contains a profiles raw bytes,
// a parsed MobileProvisioningProfile struct to access the fields,
// the p12 sha1 fingerpringt, x509.Certificate and the raw p12 bytes
// belonging to this profile.
type ProfileAndCertificate struct {
	RawData                   []byte
	MobileProvisioningProfile MobileProvisioningProfile
	CertificateSha1           string
	SigningCert               *x509.Certificate
	P12Bytes                  []byte
}

// MobileProvisioningProfile is an exact representation of a *.mobileprovision plist
type MobileProvisioningProfile struct {
	AppIDName                   string
	ApplicationIdentifierPrefix []string
	CreationDate                time.Time
	Platform                    []string
	IsXcodeManaged              bool
	DeveloperCertificates       [][]byte
	Entitlements                map[string]interface{}
	ExpirationDate              time.Time
	Name                        string
	ProvisionedDevices          []string
	TeamIdentifier              []string
	TeamName                    string
	TimeToLive                  int
	UUID                        string
	Version                     int
}

const P12Password = "a"

// FindProfileForDevice finds the correct profile for a given device udid out of an array
// of profiles and returns the index of the correct profile or -1 if the device is not in any of them
func FindProfileForDevice(udid string, profileAndCertificates []ProfileAndCertificate) int {
	for profileIndex, profileAndCertificate := range profileAndCertificates {
		for _, profileUdid := range profileAndCertificate.MobileProvisioningProfile.ProvisionedDevices {
			if profileUdid == udid {
				return profileIndex
			}
		}
	}
	return -1
}

func verifyP12CertIsInProfile(p12cert *x509.Certificate, certificates []*x509.Certificate) bool {
	if len(certificates) == 0 {
		return false
	}
	p12certHash := getSha1Fingerprint(p12cert)
	for _, cert := range certificates {
		if p12certHash == getSha1Fingerprint(cert) {
			return true
		}
	}
	return false
}

// IsEnterpriseProfile returns true if there is an enterprise profile at profilePath and
// false otherwise.
func IsEnterpriseProfile(profilePath string) bool {
	profileBytes, err := os.ReadFile(profilePath)
	if err != nil {
		return false
	}
	p7, err := pkcs7.Parse(profileBytes)
	if err != nil {
		return false
	}

	decoder := plist.NewDecoder(bytes.NewReader(p7.Content))

	var profile map[string]interface{}
	err = decoder.Decode(&profile)
	if err != nil {
		return false
	}
	if val, ok := profile["ProvisionsAllDevices"]; ok {
		return val.(bool)
	}
	return false
}

// ParseProfiles looks for *.mobileprovision in the given path and parses each of them.
// It returns an error if the path does not contain any profiles.
func ParseProfiles(profilesPath string) ([]ProfileAndCertificate, error) {
	result := []ProfileAndCertificate{}
	profiles, err := filepath.Glob(path.Join(profilesPath, "*.mobileprovision"))
	if err != nil {
		return result, err
	}
	for _, file := range profiles {
		log.Infof("parsing profile '%s'", file)
		profile, err := ParseProfile(file)
		if err != nil {
			return result, err
		}
		result = append(result, profile)

	}
	if len(result) == 0 {
		return result, fmt.Errorf("no profiles found in path %s", profilesPath)
	}
	return result, nil
}

// ParseProfile extracts the plist from a pkcs7 signed mobileprovision file.
// It decodes the plist into a go struct. Additionally a p12 certificate
// must be present next to the profile with the same filename.
// Example: test.mobileprovision and test.p12 must both be present or the parser will fail.
// The parser also checks if the p12 certificate is contained in the profile to prevent errors.
// It returns a ProfileAndCertificate struct containing everything needed for signing.
func ParseProfile(profilePath string) (ProfileAndCertificate, error) {
	profileBytes, err := ioutil.ReadFile(profilePath)
	if err != nil {
		return ProfileAndCertificate{}, err
	}
	p12bytes, err := os.ReadFile(strings.Replace(profilePath, ".mobileprovision", ".p12", 1))
	if err != nil {
		return ProfileAndCertificate{}, fmt.Errorf("Failed reading p12 file for %s with err: %+v", profilePath, err)
	}

	_, cert, err := pkcs12.Decode(p12bytes, P12Password)
	if err != nil {
		return ProfileAndCertificate{}, fmt.Errorf("Failed parsing p12 certificate with: %+v", err)
	}

	p7, err := pkcs7.Parse(profileBytes)
	if err != nil {
		return ProfileAndCertificate{}, err
	}

	decoder := plist.NewDecoder(bytes.NewReader(p7.Content))

	var profile MobileProvisioningProfile
	err = decoder.Decode(&profile)

	parsedDeveloperCertificates := make([]*x509.Certificate, len(profile.DeveloperCertificates))

	for i, certBytes := range profile.DeveloperCertificates {
		cert, err := x509.ParseCertificate(certBytes)
		parsedDeveloperCertificates[i] = cert
		if err != nil {
			return ProfileAndCertificate{}, err
		}
	}

	if !verifyP12CertIsInProfile(cert, parsedDeveloperCertificates) {
		return ProfileAndCertificate{}, fmt.Errorf("p12 certificate is not contained in provisioning profile, wrong profile file for this p12")
	}

	return ProfileAndCertificate{MobileProvisioningProfile: profile,
		RawData:         profileBytes,
		CertificateSha1: getSha1Fingerprint(cert),
		P12Bytes:        p12bytes,
		SigningCert:     cert,
	}, err
}

func getSha1Fingerprint(cert *x509.Certificate) string {
	fp := sha1.Sum(cert.Raw)
	return fmt.Sprintf("%x", fp)
}
