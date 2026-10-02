package imagemounter

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

// ddiRepo hosts the signed developer disk images for iOS 11.4 - 16.7. It is the
// same source pymobiledevice3 downloads from and, unlike the older mirror, also
// carries the last classic images (15.8, 16.7). Versions the repo does not have
// (pre-11.4 and patch releases like 14.7.1) stay on the older mirror.
const ddiRepo = "https://raw.githubusercontent.com/doronz88/DeveloperDiskImage/main/DeveloperDiskImages"

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
		"11.4 (15F5037c)": ddiRepo + "/11.4",
		"12.0 (16A5288q)": ddiRepo + "/12.0",
		"12.1 (16B5059d)": ddiRepo + "/12.1",
		"12.2 (16E5191d)": ddiRepo + "/12.2",
		"12.3 (16F148)":   ddiRepo + "/12.3",
		"12.4":            ddiRepo + "/12.4",
		"13.0":            ddiRepo + "/13.0",
		"13.1":            ddiRepo + "/13.1",
		"13.2":            ddiRepo + "/13.2",
		"13.3":            ddiRepo + "/13.3",
		"13.4":            ddiRepo + "/13.4",
		"13.5":            ddiRepo + "/13.5",
		"13.6":            ddiRepo + "/13.6",
		"13.7":            ddiRepo + "/13.7",
		"14.0":            ddiRepo + "/14.0",
		"14.1":            ddiRepo + "/14.1",
		"14.2":            ddiRepo + "/14.2",
		"14.3":            ddiRepo + "/14.3",
		"14.4":            ddiRepo + "/14.4",
		"14.5":            ddiRepo + "/14.5",
		"14.6":            ddiRepo + "/14.6",
		"14.7":            ddiRepo + "/14.7",
		"14.7.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/14.7.1",
		"14.8":            ddiRepo + "/14.8",
		"15.0":            ddiRepo + "/15.0",
		"15.1":            ddiRepo + "/15.1",
		"15.2":            ddiRepo + "/15.2",
		"15.3":            ddiRepo + "/15.3",
		"15.3.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.3.1",
		"15.4":            ddiRepo + "/15.4",
		"15.5":            ddiRepo + "/15.5",
		"15.6":            ddiRepo + "/15.6",
		"15.6.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/15.6.1",
		"15.7":            ddiRepo + "/15.7",
		"15.8":            ddiRepo + "/15.8",
		"16.0":            ddiRepo + "/16.0",
		"16.1":            ddiRepo + "/16.1",
		"16.2":            ddiRepo + "/16.2",
		"16.3":            ddiRepo + "/16.3",
		"16.4":            ddiRepo + "/16.4",
		"16.4.1":          "https://github.com/mspvirajpatel/Xcode_Developer_Disk_Images/blob/master/Developer%20Disk%20Image/16.4.1",
		"16.5":            ddiRepo + "/16.5",
		"16.6":            ddiRepo + "/16.6",
		"16.7":            ddiRepo + "/16.7",
	}

	availableVersions = []string{"4.2", "4.3", "5.0", "5.1", "6.0", "6.1", "7.0", "7.1", "8.0", "8.1", "8.2", "8.3", "8.4 (12H141)", "9.0 (13A340)", "9.1 (13B5110e)", "9.2 (13C75)", "9.3 (13E230)", "10.0 (14A345)", "10.1 (14B72)", "10.2 (14C5062c)", "10.3 (14E269)", "11.0 (15A372)", "11.1 (15B87)", "11.2 (15C5092b)", "11.3 (15E5178d)", "11.4 (15F5037c)", "12.0 (16A5288q)", "12.1 (16B5059d)", "12.2 (16E5191d)", "12.3 (16F148)", "12.4", "13.0", "13.1", "13.2", "13.3", "13.4", "13.5", "13.6", "13.7", "14.0", "14.1", "14.2", "14.3", "14.4", "14.5", "14.6", "14.7", "14.7.1", "14.8", "15.0", "15.1", "15.2", "15.3.1", "15.3", "15.4", "15.5", "15.6", "15.6.1", "15.7", "15.8", "16.0", "16.1", "16.2", "16.3", "16.4", "16.4.1", "16.5", "16.6", "16.7"}
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
	imageFileName, err := safeJoin(baseDir, xcode15_4_ddi+".zip")
	if err != nil {
		return "", err
	}
	extractedPath, err := safeJoin(baseDir, xcode15_4_ddi)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(imageFileName); err == nil {
		golog.Info("using image archive already present in basedir", "module", logModule, "path", imageFileName)
		if _, _, err := ios.Unzip(imageFileName, extractedPath); err == nil {
			return filepath.Join(extractedPath, "Restore"), nil
		}
		golog.Warn("extracting present image archive failed, downloading again", "module", logModule, "path", imageFileName)
	}
	downloadUrl := fmt.Sprintf("%s%s%s", devicebox, xcode15_4_ddi, ".zip")
	golog.Info("downloading image", "module", logModule, "url", downloadUrl, "path", imageFileName)
	err = downloadFile(imageFileName, downloadUrl)
	if err != nil {
		return "", err
	}
	_, _, err = ios.Unzip(imageFileName, extractedPath)
	if err != nil {
		return "", fmt.Errorf("Download17Plus: error extracting image %s %w", imageFileName, err)
	}

	return filepath.Join(extractedPath, "Restore"), nil
}

// safeJoin joins elems onto baseDir and guarantees that the result cannot
// escape baseDir: every element must be a local path fragment (no absolute
// path, no '..' traversal) and the cleaned result must still be inside the
// cleaned base directory. Use it for every filesystem path that is built from
// values not fully controlled by go-ios itself.
func safeJoin(baseDir string, elems ...string) (string, error) {
	for _, elem := range elems {
		if !filepath.IsLocal(elem) {
			return "", fmt.Errorf("safeJoin: path element '%s' would escape base directory '%s'", elem, baseDir)
		}
	}
	cleanedBase := filepath.Clean(baseDir)
	joined := filepath.Join(append([]string{cleanedBase}, elems...)...)
	if joined != cleanedBase && !strings.HasPrefix(joined, cleanedBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("safeJoin: path '%s' would escape base directory '%s'", joined, baseDir)
	}
	return joined, nil
}

// findPersonalizedDDI looks for a personalized developer disk image (iOS 17+)
// that is already present in baseDir and returns its 'Restore' directory. The
// layout is recognized by structure (a BuildManifest.plist inside a Restore
// directory) rather than by name, so images downloaded under a different name,
// by an older go-ios or manually, are found without any network access.
//
// Selection is deterministic: an image whose containing directory (the parent
// of 'Restore') is exactly the pinned DDI name wins, then one whose containing
// directory name contains the pinned name (e.g. a manual rename), and only then
// the lexicographically first structurally-valid image. The preference matches
// on the directory name alone, not the whole path, so an unrelated ancestor
// that happens to contain the pinned string cannot hijack the choice.
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
	sort.Strings(restoreDirs)
	// ddiName returns the name of the directory that holds the 'Restore'
	// directory, which is what go-ios names after the DDI.
	ddiName := func(restoreDir string) string {
		return filepath.Base(filepath.Dir(restoreDir))
	}
	for _, dir := range restoreDirs {
		if ddiName(dir) == xcode15_4_ddi {
			return dir, true
		}
	}
	for _, dir := range restoreDirs {
		if strings.Contains(ddiName(dir), xcode15_4_ddi) {
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
	// version comes from the fixed availableVersions catalog, but validate it
	// before using it as a path component anyway.
	if version == "" || !filepath.IsLocal(version) {
		return "", fmt.Errorf("downloadImageForVersion: no usable image version for ios version '%s'", productVersion)
	}
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
	golog.Info("thank you github.com/doronz88 and github.com/mspvirajpatel for making these images available :-)", "module", logModule, "udid", udid)
	versionPath, err := safeJoin(baseDir, versionDir)
	if err != nil {
		return "", err
	}
	downloadUrl := versionMap[version] + "/" + imageFile + "?raw=true"
	imageFileName := filepath.Join(versionPath, imageFile)

	signatureDownloadUrl := versionMap[version] + "/" + signatureFile + "?raw=true"
	signatureFileName := filepath.Join(versionPath, signatureFile)
	err = os.MkdirAll(versionPath, 0o755)
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
