/*
This package contains everything needed to sign ios apps, parse, validate and generate provisioning profiles and certificates.
*/
package codesign

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const codesignPath = "/usr/bin/codesign"

// In iOS apps you will find three types of directories that need signing applied.
// They either end with .app, .appex (app extensions) or .xctest.
const (
	appSuffix          = ".app"
	appExtensionSuffix = ".appex"
	xctestSuffix       = ".xctest"
)

// SigningConfig contains the CertSha1 of the certificate that will be used for signing.
// EntitlementsFilePath points to a plist file containing the entitlements extracted from
// the correct mobileprovisioning profile.
// KeychainPath contains the path to the keychain that contains the signing certificate.
type SigningConfig struct {
	CertSha1             string
	EntitlementsFilePath string
	KeychainPath         string
	ProfileBytes         []byte
}

func Resign(udid string, ipaFile *os.File, s SigningWorkspace) error {
	if runtime.GOOS != "darwin" {
		return errors.New("Resign: can only resign on macOS for now.")
	}
	if udid == "" {
		return errors.New("udid is empty")
	}
	info, err := ipaFile.Stat()
	if err != nil {
		return fmt.Errorf("Resign: could not get file info: %w", err)
	}
	_, directory, err := ExtractIpa(ipaFile, info.Size())
	if err != nil {
		return fmt.Errorf("Resign: could not extract ipa: %w", err)
	}
	defer os.RemoveAll(directory)

	index := FindProfileForDevice(udid, s.profiles)

	if index == -1 {
		return fmt.Errorf("Resign: could not find profile for device %s", udid)
	}

	appFolder, err := FindAppFolder(directory)
	if err != nil {
		return fmt.Errorf("Resign: could not find .app folder in extracted ipa payload folder: %w", err)
	}

	archs, err := ExtractArchitectures(appFolder)
	if err != nil {
		return fmt.Errorf("Resign: could not determine build architecture of build, run 'lipo -info appDir/appExecutable' to debug: %w", err)
	}
	if IsSimulatorApp(archs) {
		return errors.New("Resign: cannot resign simulator app")
	}

	err = Sign(directory, s.GetConfig(index))
	if err != nil {
		return fmt.Errorf("Resign: could not sign app: %w", err)
	}

	w, err := os.Create(ipaFile.Name())
	CompressToIpa(directory, w)
	return nil

}

// RemoveSignature executes "codesign --remove-signature" for the given path.
// this is mostly needed for removing the signature after wrapping a simulator app.
func RemoveSignature(dir string) error {
	cmd := exec.Command(codesignPath, "--remove-signature", dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "cmd": cmd, "output": string(output)}).Errorf("error removing signature with codesign")
		return err
	}
	log.WithFields(log.Fields{"cmd": cmd, "output": string(output)}).Debugf("codesign invoked")
	return err
}

// SignDylib can be used to codesign a dylib library.
// This is only used for signing the injected dylib during wrapping.
func SignDylib(sdkPath string, config SigningConfig) error {
	return exeuteCodesignFramework(sdkPath, config)
}

// Sign uses the cert, entitlements and keychain from the SigningConf to codesign the unzipped app
// in the root path. Root needs to be a directory named 'Payload' with all the app contents inside of it.
// Then the filetree will be walked and all the frameworks and app folders will be codesigned.
func Sign(root string, config SigningConfig) error {
	if !strings.HasSuffix(root, "Payload") {
		root = path.Join(root, "Payload")
	}
	rootPathInfo, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !rootPathInfo.IsDir() {
		return errors.New("does not exist:" + root)
	}
	dirs, err := findAppDirs(root)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		err := signFrameworks(dir, config)
		if err != nil {
			return fmt.Errorf("error signing frameworks %s err:%w", dir, err)
		}
		err = signAppDir(dir, config)
		if err != nil {
			return fmt.Errorf("error signing appDir %s err:%w", dir, err)
		}
	}

	return nil
}

// Verify runs "codesign -vv --deep" to verbosely verify recursively the given path is properly signed.
func Verify(path string) error {
	cmd := exec.Command(codesignPath, "-vv", "--deep", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"path": path, "error": err, "cmd": cmd, "output": string(output)}).Infof("codesign invoked")
		return err
	}
	return nil
}

func signFrameworks(root string, config SigningConfig) error {
	frameworksPath := path.Join(root, "Frameworks")
	//it is a recursive call, if there are no more frameworks found, we just return nil here
	if _, err := os.Stat(frameworksPath); os.IsNotExist(err) {
		return nil
	}
	files, err := os.ReadDir(frameworksPath)
	if err != nil {
		return err
	}
	//Now recursively look into each child to find other Frameworks directories deeper
	//in the file tree and sign them. To get valid overall signatures, of course the
	//Frameworks at the leaf level of the file tree must be signed first.
	// Afterwards sign the current Frameworks directory.
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".framework") {
			fullpath := path.Join(frameworksPath, file.Name())
			err := signFrameworks(fullpath, config)
			if err != nil {
				return fmt.Errorf("signing Frameworks had err:%w", err)
			}
			err = exeuteCodesignFramework(fullpath, config)
			if err != nil {
				return fmt.Errorf("running codesign on frameworks had err:%w", err)
			}
		}
	}
	return nil
}

func exeuteCodesignFramework(path string, config SigningConfig) error {
	cmd := exec.Command(codesignPath, "-vv", "--keychain", config.KeychainPath, "--deep", "--force", "--sign", config.CertSha1, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "cmd": cmd, "output": string(output)}).Errorf("codesign invoked")
		return err
	}
	log.WithFields(log.Fields{"cmd": cmd, "output": string(output)}).Debugf("codesign invoked")
	return err
}

func signAppDir(appPath string, config SigningConfig) error {
	if shouldReplaceProfile(appPath) {
		target := path.Join(appPath, "embedded.mobileprovision")
		err := os.WriteFile(target, config.ProfileBytes, 0644)
		if err != nil {
			return fmt.Errorf("failed replacing embedded.mobileprovision profile in %s with %w", appPath, err)
		}
	}
	cmd := exec.Command(codesignPath, "-vv", "--keychain", config.KeychainPath, "--deep", "--force", "--sign", config.CertSha1, "--entitlements", config.EntitlementsFilePath, appPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.WithFields(log.Fields{"error": err, "cmd": cmd, "output": string(output)}).Errorf("codesign failed")
		return err
	}
	log.WithFields(log.Fields{"cmd": cmd, "output": string(output)}).Debugf("codesign invoked")
	return err
}

func findAppDirs(root string) ([]string, error) {
	allFiles, err := GetFiles(root)
	if err != nil {
		return []string{}, err
	}
	appDirs := []string{}
	for _, file := range allFiles {
		if isDirWithApp(file) {
			appDirs = append(appDirs, file)
		}
	}
	return reverse(appDirs), nil
}

func isDirWithApp(dir string) bool {
	return strings.HasSuffix(dir, appSuffix) || strings.HasSuffix(dir, xctestSuffix) || strings.HasSuffix(dir, appExtensionSuffix)
}

func shouldReplaceProfile(dir string) bool {
	return strings.HasSuffix(dir, appSuffix) || strings.HasSuffix(dir, appExtensionSuffix)
}

// GetFiles performs a walk to recursively find all files and directories in the given root path.
// It returns a list of all files omitting the root path itself.
func GetFiles(root string) ([]string, error) {
	walkStart := time.Now()
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking file tree for path: %s error: %w", path, err)
		}
		if path == root {
			return nil
		}
		files = append(files, path)
		return nil
	})
	walkDuration := time.Since(walkStart)
	log.Infof("Walk duration: %v", walkDuration)
	return files, err
}

func reverse(a []string) []string {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
	return a
}
