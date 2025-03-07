package codesign_test

import (
	"os"
	"path"
	"testing"

	"github.com/danielpaulus/go-ios/ios/codesign"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateKeychain(t *testing.T) {
	directory, err := os.MkdirTemp("", "codesign-test")
	if err != nil {
		log.Fatalf("Failed with %+v", err)
	}
	defer os.RemoveAll(directory)
	keychain := path.Join(directory, "test.keychain")

	err = codesign.CreateKeychain(keychain)
	if assert.NoError(t, err) {
		//will fail if the file does not exist
		info, err := os.Stat(keychain)
		if assert.NoError(t, err) {
			assert.False(t, info.IsDir())
		}
	}
}

func TestInstallCertificate(t *testing.T) {
	directory, err := os.MkdirTemp("", "codesign-test")
	if err != nil {
		log.Fatalf("Failed with %+v", err)
	}
	defer os.RemoveAll(directory)
	keychain := path.Join(directory, "test.keychain")
	err = codesign.CreateKeychain(keychain)
	codesign.UnlockKeychain(keychain)
	codesign.DisableTimeoutForKeychain(keychain)

	certsha1, certpath := extractFixtureCertificate(directory)

	assert.False(t, codesign.KeychainHasCertificate(keychain, certsha1))
	codesign.AddX509CertificateToKeychain(keychain, certpath)
	assert.True(t, codesign.KeychainHasCertificate(keychain, certsha1))

}

func extractFixtureCertificate(tempdir string) (string, string) {
	profileAndCertificate, err := codesign.ParseProfile("fixtures/test.mobileprovision")
	if err != nil {
		log.Fatal(err)
	}
	certsha1 := profileAndCertificate.CertificateSha1
	certPath := path.Join(tempdir, "test.p12")
	err = os.WriteFile(certPath, profileAndCertificate.P12Bytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return certsha1, certPath
}

// Be careful with these tests and the underlying code, they use the actual keychain searchlist.
// If you mess it up, you might remove all your system passwords until you restore
// the keychain search list. So if you work on this code, be sure to call
// "security list-keychain" first and write your current list down somewhere :-)
func TestGetAndSetChangesNothing(t *testing.T) {
	originalList, err := codesign.GetKeychainSearchList()
	if err != nil {
		log.Fatalf("Test failed getting keychain list %+v", err)
	}
	err = codesign.SetKeychainSearchList(originalList)
	if err != nil {
		log.Fatalf("Failed setting keychain with %+v", err)
	}
	listAfterTesting, err := codesign.GetKeychainSearchList()
	if err != nil {
		log.Fatalf("Test failed getting keychain list %+v", err)
	}
	assert.ElementsMatch(t, originalList, listAfterTesting)
}

func TestAddRemoveKeychain(t *testing.T) {
	originalList, err := codesign.GetKeychainSearchList()
	if err != nil {
		log.Fatalf("Test failed getting keychain list %+v", err)
	}

	randomPath := "/Library/test/keychain.keychain"
	err = codesign.AddKeychainToSearchList(randomPath)
	if err != nil {
		log.Fatalf("Test failed adding keychain list %+v", err)
	}

	listWithRandomPath, err := codesign.GetKeychainSearchList()
	if err != nil {
		log.Fatalf("Test failed getting keychain list %+v", err)
	}
	assert.Contains(t, listWithRandomPath, randomPath)

	err = codesign.RemoveFromKeychainSearchList(randomPath)
	if err != nil {
		log.Fatalf("Test failed removing from keychain list %+v", err)
	}

	listAfterTesting, err := codesign.GetKeychainSearchList()
	if err != nil {
		log.Fatalf("Test failed getting keychain list %+v", err)
	}
	assert.ElementsMatch(t, originalList, listAfterTesting)
}

func TestRemoveNonPresentItemDoesNothing(t *testing.T) {
	originalList, err := codesign.GetKeychainSearchList()
	if err != nil {
		log.Fatalf("Test failed getting keychain list %+v", err)
	}
	codesign.RemoveFromKeychainSearchList("not in the list because invalid filepath")

	listAfterTesting, err := codesign.GetKeychainSearchList()
	if err != nil {
		log.Fatalf("Test failed getting keychain list %+v", err)
	}
	assert.ElementsMatch(t, originalList, listAfterTesting)
}
