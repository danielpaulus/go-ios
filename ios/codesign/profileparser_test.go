package codesign_test

import (
	"log"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios/codesign"
	"github.com/stretchr/testify/assert"
)

func TestDirWithoutProfiles(t *testing.T) {
	_, err := codesign.ParseProfiles(".")
	assert.Error(t, err)
}

func TestP12NotInProfile(t *testing.T) {
	_, err := codesign.ParseProfiles("fixtures/profile_notmatching_cert")
	assert.Equal(t, "p12 certificate is not contained in provisioning profile, wrong profile file for this p12", err.Error())
}

func TestEnterpriseProfileDetection(t *testing.T) {
	shouldBeTrue := codesign.IsEnterpriseProfile("fixtures/enterpriseprofile/embedded.mobileprovision")
	assert.True(t, shouldBeTrue)
	shouldBeFalse := codesign.IsEnterpriseProfile("fixtures/test.mobileprovision")
	assert.False(t, shouldBeFalse)
}

func TestParsing(t *testing.T) {
	profileAndCertificates, err := codesign.ParseProfiles("fixtures")
	if err != nil {
		log.Fatalf("failed finding profiles %+v", err)
	}
	assert.Equal(t, 2, len(profileAndCertificates))

	profileAndCertificate := profileAndCertificates[1]
	profile := profileAndCertificate.MobileProvisioningProfile
	if assert.NoError(t, err) {

		assert.Equal(t, time.Date(2021, 10, 21, 06, 52, 23, 0, time.UTC), profile.ExpirationDate)
	}

}

func TestFindDeviceInProfile(t *testing.T) {
	profileAndCertificates, err := codesign.ParseProfiles("fixtures")
	if err != nil {
		log.Fatalf("failed finding profiles %+v", err)
	}
	for profileIndex, profileAndCertificate := range profileAndCertificates {
		for _, udid := range profileAndCertificate.MobileProvisioningProfile.ProvisionedDevices {
			foundIndex := codesign.FindProfileForDevice(udid, profileAndCertificates)
			assert.Equal(t, profileIndex, foundIndex)
		}
	}
	assert.Equal(t, -1, codesign.FindProfileForDevice("not contained", profileAndCertificates))
}
