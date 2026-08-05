package imagemounter

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

var (
	versionMap = map[string]string{
		"4.2":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/4.2",
		"4.3":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/4.3",
		"5.0":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/5.0",
		"5.1":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/5.1",
		"6.0":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/6.0",
		"6.1":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/6.1",
		"7.0":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/7.0",
		"7.1":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/7.1",
		"8.0":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/8.0",
		"8.1":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/8.1",
		"8.2":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/8.2",
		"8.3":             "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/8.3",
		"8.4 (12H141)":    "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/8.4%20(12H141)",
		"9.0 (13A340)":    "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/9.0%20(13A340)",
		"9.1 (13B5110e)":  "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/9.1%20(13B5110e)",
		"9.2 (13C75)":     "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/9.2%20(13C75)",
		"9.3 (13E230)":    "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/9.3%20(13E230)",
		"10.0 (14A345)":   "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/10.0%20(14A345)",
		"10.1 (14B72)":    "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/10.1%20(14B72)",
		"10.2 (14C5062c)": "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/10.2%20(14C5062c)",
		"10.3 (14E269)":   "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/10.3%20(14E269)",
		"11.0 (15A372)":   "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/11.0%20(15A372)",
		"11.1 (15B87)":    "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/11.1%20(15B87)",
		"11.2 (15C5092b)": "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/11.2%20(15C5092b)",
		"11.3 (15E5178d)": "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/11.3%20(15E5178d)",
		"11.4 (15F5037c)": "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/11.4%20(15F5037c)",
		"12.0 (16A5288q)": "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/12.0%20(16A5288q)",
		"12.1 (16B5059d)": "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/12.1%20(16B5059d)",
		"12.2 (16E5191d)": "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/12.2%20(16E5191d)",
		"12.3 (16F148)":   "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/12.3%20(16F148)",
		"12.4":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/12.4",
		"13.0":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/13.0",
		"13.1":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/13.1",
		"13.2":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/13.2",
		"13.3":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/13.3",
		"13.4":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/13.4",
		"13.5":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/13.5",
		"13.7":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/13.7",
		"14.0":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.0",
		"14.1":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.1",
		"14.2":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.2",
		"14.4":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.4",
		"14.5":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.5",
		"14.6":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.6",
		"14.7":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.7",
		"14.7.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.7.1",
		"14.8":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.8",
		"15.0":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.0",
		"15.1":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.1",
		"15.2":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.2",
		"15.3":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.3",
		"15.3.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.3.1",
		"15.4":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.4",
		"15.5":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.5",
		"15.6":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.6",
		"15.6.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.6.1",
		"15.7":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.7",
		"16.0":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.0",
		"16.1":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.1",
		"16.2":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.2",
		"16.3":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.3",
		"16.4":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.4",
		"16.4.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.4.1",
		"16.5":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.5",
		"16.6":            "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.6",
	}

	availableVersions = []string{"4.2", "4.3", "5.0", "5.1", "6.0", "6.1", "7.0", "7.1", "8.0", "8.1", "8.2", "8.3", "8.4 (12H141)", "9.0 (13A340)", "9.1 (13B5110e)", "9.2 (13C75)", "9.3 (13E230)", "10.0 (14A345)", "10.1 (14B72)", "10.2 (14C5062c)", "10.3 (14E269)", "11.0 (15A372)", "11.1 (15B87)", "11.2 (15C5092b)", "11.3 (15E5178d)", "11.4 (15F5037c)", "12.0 (16A5288q)", "12.1 (16B5059d)", "12.2 (16E5191d)", "12.3 (16F148)", "12.4", "13.0", "13.1", "13.2", "13.3", "13.4", "13.5", "13.7", "14.0", "14.1", "14.2", "14.4", "14.5", "14.6", "14.7", "14.7.1", "14.8", "15.0", "15.1", "15.2", "15.3.1", "15.3", "15.4", "15.5", "15.6", "15.6.1", "15.7", "16.0", "16.1", "16.2", "16.3", "16.4", "16.4.1", "16.5", "16.6"}
)

const (
	imageFile     = "DeveloperDiskImage.dmg"
	signatureFile = "DeveloperDiskImage.dmg.signature"
	// iOS 17+ universal personalized developer disk image hosted on deviceboxhq.
	// Bump this when a newer DDI is published there (was ddi-15F31d).
	xcode15_4_ddi = "ddi-17E5179g"
	// buildManifestFile identifies the personalized developer disk image
	// layout (iOS 17+): a 'Restore' directory containing a BuildManifest.plist.
	buildManifestFile = "BuildManifest.plist"
)

// devicebox hosts the personalized developer disk image. It is a var so unit
// tests can point it at a local test server.
var devicebox = "https://deviceboxhq.com/"

func MatchAvailable(version string) string {
	golog.Debug("matching available image for device version", "module", logModule, "version", version)
	requestedVersionParsed := semver.MustParse(version)
	var bestMatch *semver.Version = nil
	var bestMatchString string

	for _, availableVersion := range availableVersions {
		parsedAV := semver.MustParse(strings.Split(availableVersion, " (")[0])
		if parsedAV.Equal(requestedVersionParsed) {
			return availableVersion
		}
		if bestMatch == nil {
			bestMatch = parsedAV
			bestMatchString = availableVersion
			continue
		}
		if parsedAV.GreaterThan(bestMatch) && (parsedAV.LessThan(requestedVersionParsed)) {
			bestMatch = parsedAV
			bestMatchString = availableVersion
		}
	}
	golog.Debug("matched available image", "module", logModule, "version", version, "bestMatch", bestMatch)

	return bestMatchString
}

func Download17Plus(baseDir string, version *semver.Version) (string, error) {
	golog.Info("getting developer image", "module", logModule, "version", version.String())
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", err
	}
	if restoreDir, ok := findPersonalizedDDI(baseDir); ok {
		golog.Info("using personalized developer disk image already present in basedir", "module", logModule, "path", restoreDir)
		return restoreDir, nil
	}
	imageFileName := path.Join(baseDir, xcode15_4_ddi+".zip")
	extractedPath := path.Join(baseDir, xcode15_4_ddi)
	if _, err := os.Stat(imageFileName); err == nil {
		golog.Info("using image archive already present in basedir", "module", logModule, "path", imageFileName)
		if _, _, err := ios.Unzip(imageFileName, extractedPath); err == nil {
			return path.Join(extractedPath, "Restore"), nil
		}
		golog.Warn("extracting present image archive failed, downloading again", "module", logModule, "path", imageFileName)
	}
	downloadUrl := fmt.Sprintf("%s%s%s", devicebox, xcode15_4_ddi, ".zip")
	golog.Info("downloading image", "module", logModule, "url", downloadUrl, "path", imageFileName)
	err := downloadFile(imageFileName, downloadUrl)
	if err != nil {
		return "", err
	}
	_, _, err = ios.Unzip(imageFileName, extractedPath)
	if err != nil {
		return "", fmt.Errorf("Download17Plus: error extracting image %s %w", imageFileName, err)
	}

	return path.Join(extractedPath, "Restore"), nil
}

// findPersonalizedDDI looks for a personalized developer disk image (iOS 17+)
// that is already present in baseDir and returns its 'Restore' directory. The
// layout is recognized by structure (a BuildManifest.plist inside a Restore
// directory) rather than by name, so images downloaded under a different name,
// by an older go-ios or manually, are found without any network access. When
// multiple images are present, the one matching the pinned DDI is preferred.
func findPersonalizedDDI(baseDir string) (string, bool) {
	var restoreDirs []string
	err := filepath.Walk(baseDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() != buildManifestFile {
			return nil
		}
		if dir := filepath.Dir(p); filepath.Base(dir) == "Restore" {
			restoreDirs = append(restoreDirs, dir)
		}
		return nil
	})
	if err != nil || len(restoreDirs) == 0 {
		return "", false
	}
	for _, dir := range restoreDirs {
		if strings.Contains(dir, xcode15_4_ddi) {
			return dir, true
		}
	}
	return restoreDirs[0], true
}

func DownloadImageFor(device ios.DeviceEntry, baseDir string) (string, error) {
	allValues, err := ios.GetValues(device)
	if err != nil {
		return "", err
	}
	return downloadImageForVersion(baseDir, allValues.Value.ProductVersion, device.Properties.SerialNumber)
}

// downloadImageForVersion resolves the developer disk image for the given
// product version, using an image already present in baseDir when possible and
// downloading only on a cache miss.
func downloadImageForVersion(baseDir string, productVersion string, udid string) (string, error) {
	parsedVersion, err := semver.NewVersion(productVersion)
	if err != nil {
		return "", fmt.Errorf("downloadImageForVersion: failed parsing ios productversion: '%s' with %w", productVersion, err)
	}
	if parsedVersion.GreaterThan(ios.IOS17()) || parsedVersion.Equal(ios.IOS17()) {
		return Download17Plus(baseDir, parsedVersion)
	}
	version := MatchAvailable(productVersion)
	golog.Info("getting developer image", "module", logModule, "udid", udid, "version", productVersion, "imageVersion", version)
	versionDir := strings.Split(version, " (")[0]
	// Look for a present image before any network access, both in the full
	// '<version> (<build>)' layout used by manual downloads mirroring the
	// upstream repository and in the short '<version>' layout that go-ios
	// itself uses when downloading.
	candidateDirs := []string{version}
	if versionDir != version {
		candidateDirs = append(candidateDirs, versionDir)
	}
	for _, dir := range candidateDirs {
		imageDownloaded, err := validateBaseDirAndLookForImage(baseDir, filepath.Join(dir, imageFile))
		if err != nil {
			return "", err
		}
		if imageDownloaded != "" {
			golog.Info("using image already present in basedir", "module", logModule, "udid", udid, "path", imageDownloaded)
			return imageDownloaded, nil
		}
	}
	golog.Info("thank you github.com/mspvirajpatel for making these images available :-)", "module", logModule, "udid", udid)
	downloadUrl := versionMap[version] + "/" + imageFile + "?raw=true"
	imageFileName := path.Join(baseDir, versionDir, imageFile)

	signatureDownloadUrl := versionMap[version] + "/" + signatureFile + "?raw=true"
	signatureFileName := path.Join(baseDir, versionDir, signatureFile)
	err = os.MkdirAll(path.Join(baseDir, versionDir), 0o755)
	if err != nil {
		return "", err
	}
	golog.Info("downloading image", "module", logModule, "udid", udid, "url", downloadUrl, "path", imageFileName)
	err = downloadFile(imageFileName, downloadUrl)
	if err != nil {
		return "", err
	}

	err = downloadFile(signatureFileName, signatureDownloadUrl)
	if err != nil {
		return "", err
	}

	return imageFileName, nil
}

func findImage(dir string, imageToFind string) (string, error) {
	var imageWeFound string
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, imageToFind) {
				imageWeFound = path
			}
			return nil
		})
	if err != nil {
		return "", err
	}
	if imageWeFound != "" {
		return imageWeFound, nil
	}
	return "", fmt.Errorf("image not found")
}

func validateBaseDirAndLookForImage(baseDir string, imageToFind string) (string, error) {
	dirHandle, err := os.Open(baseDir)
	if err != nil {
		err := os.MkdirAll(baseDir, 0o777)
		if err != nil {
			return "", err
		}
		return "", nil
	}
	defer dirHandle.Close()

	dmgPath, err := findImage(baseDir, imageToFind)
	if err != nil {
		return "", nil
	}

	return dmgPath, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// PS: Taken from golangcode.com
func downloadFile(filepath string, url string) error {
	c := &http.Client{
		Timeout:   2 * time.Minute,
		Transport: http.DefaultTransport,
	}
	// Get the data
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloadFile: unexpected status code %d downloading %s", resp.StatusCode, url)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
