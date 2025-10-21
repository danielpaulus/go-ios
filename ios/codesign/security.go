package codesign

import (
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

const securityPath = "/usr/bin/security"

//KeychainPassword contains some password for the keychain created and unlocked.
//The password can be publicly visible because the keychain created here
//only will contain our signing
//certificate, which is included in github and in the filesystem anyway.
//The certificate is even contained in the installed application, so anyone
//with access to our devices(all customers) could potentially extract it.
const KeychainPassword = "keychain-pwd"

//CreateKeychain creates a new keychain file at the specified path by invoking
//the "security create-keychain" command.
func CreateKeychain(path string) error {
	_, err := executeSecurity("create-keychain", "-p", KeychainPassword, path)
	return err
}

//AddKeychainToSearchList adds a new keychain path to the current
//search list by getting all entries first, adding the given path and then
//setting the new list. This has a slight probability of a race condition, so use with care
//and never invoke this concurrently!.
func AddKeychainToSearchList(path string) error {
	keychainSearchList, err := GetKeychainSearchList()
	if err != nil {
		return err
	}
	keychainSearchList = append(keychainSearchList, path)
	return SetKeychainSearchList(keychainSearchList)
}

//RemoveFromKeychainSearchList remove an entry from the keychainSearchList only if it is present.
//If the element is not in the list, nothing will happen
func RemoveFromKeychainSearchList(path string) error {
	keychainSearchList, err := GetKeychainSearchList()
	if err != nil {
		return err
	}
	index := -1
	for i, entry := range keychainSearchList {
		//for some paths, the security command changes directory automatically
		//f.ex. temp dirs will be changed from /var/.. to /private/var which causes removing to become a non op
		if strings.Contains(entry, path) {
			index = i
		}
	}
	if index != -1 {
		newKeychainList := append(keychainSearchList[:index], keychainSearchList[index+1:]...)
		return SetKeychainSearchList(newKeychainList)
	}
	log.Warn("tried to remove a non existing keychain, could be a bug")
	return nil

}

//SetKeychainSearchList sets the current keychain search list using the
//security list-kechain -s [keychain1] [keychain2]...  command
// (!!)Be careful to only use this in combination with GetKeychainSearchList as it
//completely replaces the current list instead of just adding a new entry.
func SetKeychainSearchList(entries []string) error {
	cmd := []string{"list-keychain", "-s"}

	cmd = append(cmd, entries...)
	output, err := executeSecurity(cmd...)
	log.Debug(output)
	return err
}

//GetKeychainSearchList parses the output of security list-keychain to a
//string slice containing all the entries
func GetKeychainSearchList() ([]string, error) {
	output, err := executeSecurity("list-keychain")
	if err != nil {
		return []string{}, err
	}
	output = strings.ReplaceAll(output, "\"", "")
	list := strings.Split(output, "\n")
	length := len(list)
	if list[length-1] == "" {
		list = list[:length-1]
	}
	for i, s := range list {
		list[i] = strings.TrimSpace(s)
	}
	return list, nil
}

//UnlockKeychain unlocks the keychain so we can use it for signing and installing a certificate.
//Don't forget to disable the timeout too so it does not lock itself again.
func UnlockKeychain(path string) error {
	_, err := executeSecurity("unlock-keychain", "-p", KeychainPassword, path)
	return err
}

//DisableTimeoutForKeychain sets the timeout to no timeout by invoking "security set-keychain-settings".
// "security set-keychain-settings -h" explains why this works if you do not specify a '-t' switch like so:
// -t  Timeout in seconds (omitting this option specifies "no timeout")
func DisableTimeoutForKeychain(path string) error {
	_, err := executeSecurity("set-keychain-settings", path)
	return err
}

//AddX509CertificateToKeychain installs a x509 based certificate extracted directly from a
//MobileProvisioningProfile into the given keychain.
func AddX509CertificateToKeychain(keychain string, certificate string) error {
	_, err := executeSecurity("import", certificate, "-k", keychain, "-P", P12Password, "-T", codesignPath)
	return err
}

//KeychainHasCertificate looks for the sha1hash to be present in the given keychain.
//It uses "security find-certificate -Z keychainpath" which prints cert output and SHA1 hash.
func KeychainHasCertificate(keychain string, sha1hash string) bool {
	output, err := executeSecurity("find-certificate", "-Z", keychain)
	//will also err if the keychain is empty
	if err != nil {
		return false
	}
	return strings.Contains(output, strings.ToUpper(sha1hash))
}

func executeSecurity(args ...string) (string, error) {
	cmd := exec.Command(securityPath, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"cmd": cmd, "output": string(output)}).Errorf("security failed")
	}
	log.WithFields(log.Fields{"cmd": cmd, "output": string(output)}).Debugf("security invoked")
	return string(output), err
}
