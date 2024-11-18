package imagemounter

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
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
	devicebox     = "https://deviceboxhq.com/"
	xcode15_4_ddi = "ddi-15F31d"
)

func MatchAvailable(version string) string {
	log.Debugf("device version: %s ", version)
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
	log.Debugf("device version: %s bestMatch: %s", version, bestMatch)

	return bestMatchString
}

func Download17Plus(baseDir string, version *semver.Version) (string, error) {
	downloadUrl := fmt.Sprintf("%s%s%s", devicebox, xcode15_4_ddi, ".zip")
	log.Infof("device iOS version: %s, getting developer image: %s", version.String(), downloadUrl)

	imageDownloaded, err := validateBaseDirAndLookForImage(baseDir, xcode15_4_ddi)
	if err != nil {
		return "", err
	}
	if imageDownloaded != "" {
		log.Infof("using already downloaded image: %s", imageDownloaded)
		return path.Join(imageDownloaded, "Restore"), err
	}
	imageFileName := path.Join(baseDir, xcode15_4_ddi+".zip")
	extractedPath := path.Join(baseDir, xcode15_4_ddi)
	log.Infof("downloading '%s' to path '%s'", downloadUrl, imageFileName)
	err = downloadFile(imageFileName, downloadUrl)
	if err != nil {
		return "", err
	}
	_, _, err = ios.Unzip(imageFileName, extractedPath)
	if err != nil {
		return "", fmt.Errorf("Download17Plus: error extracting image %s %w", imageFileName, err)
	}

	return path.Join(extractedPath, "Restore"), nil
}

func DownloadImageFor(device ios.DeviceEntry, baseDir string) (string, error) {
	allValues, err := ios.GetValues(device)
	if err != nil {
		return "", err
	}
	parsedVersion, err := semver.NewVersion(allValues.Value.ProductVersion)
	if err != nil {
		return "", fmt.Errorf("DownloadImageFor: failed parsing ios productversion: '%s' with %w", allValues.Value.ProductVersion, err)
	}
	if parsedVersion.GreaterThan(ios.IOS17()) || parsedVersion.Equal(ios.IOS17()) {
		return Download17Plus(baseDir, parsedVersion)
	}
	version := MatchAvailable(allValues.Value.ProductVersion)
	log.Infof("device iOS version: %s, getting developer image for iOS %s", allValues.Value.ProductVersion, version)
	var imageToFind string
	switch runtime.GOOS {
	case "windows":
		imageToFind = fmt.Sprintf("%s\\%s", version, imageFile)
	default:
		imageToFind = fmt.Sprintf("%s/%s", version, imageFile)
	}
	imageDownloaded, err := validateBaseDirAndLookForImage(baseDir, imageToFind)
	if err != nil {
		return "", err
	}
	if imageDownloaded != "" {
		log.Infof("%s already downloaded from https://github.com/mspvirajpatel/", imageDownloaded)
		return imageDownloaded, nil
	}
	downloadUrl := ""
	log.Infof("downloading from: %s", downloadUrl)
	log.Info("thank you github.com/mspvirajpatel for making these images available :-)")
	versionDir := strings.Split(version, " (")[0]
	downloadUrl = versionMap[version] + "/" + imageFile + "?raw=true"
	imageFileName := path.Join(baseDir, versionDir, imageFile)

	signatureDownloadUrl := versionMap[version] + "/" + signatureFile + "?raw=true"
	signatureFileName := path.Join(baseDir, versionDir, signatureFile)
	err = os.Mkdir(path.Join(baseDir, versionDir), 0o755)
	if err != nil {
		return "", err
	}
	log.Infof("downloading '%s' to path '%s'", downloadUrl, imageFileName)
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
	defer dirHandle.Close()
	if err != nil {
		err := os.MkdirAll(baseDir, 0o777)
		if err != nil {
			return "", err
		}
		return "", nil
	}

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
