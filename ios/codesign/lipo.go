package codesign

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"

	"howett.net/plist"
)

const executableKey = "CFBundleExecutable"

const lipo = "/usr/bin/lipo"

// CheckLipo check if lipo works properly
func CheckLipo() error {
	cmd := exec.Command(lipo, "-info", lipo)
	_, err := cmd.CombinedOutput()
	return err
}

// IsSimulatorApp returns true if one of the architectures equals "x86_64".
// It returns false otherwise.
func IsSimulatorApp(architectures []string) bool {
	for _, arch := range architectures {
		if arch == "x86_64" {
			return true
		}
	}
	return false
}

// ExtractArchitectures takes a directory where it can find a Info.plist and an executable file to check.
// Usually that will be the .app folder.
// It will parse the Info.plist to find the executable file and run lipo -info against it to extract
// architectures.
// It returns an string array with the parsed architectures contained.
func ExtractArchitectures(binDir string) ([]string, error) {
	binFile, err := getExecutable(binDir)
	if err != nil {
		return []string{}, err
	}
	cmd := exec.Command(lipo, "-info", path.Join(binDir, binFile))
	output, err := cmd.CombinedOutput()
	architectures := strings.TrimSpace(string(output))

	if strings.Contains(architectures, "is architecture: ") {
		splitted := strings.Split(architectures, "is architecture: ")
		if len(splitted) != 2 {
			return []string{}, fmt.Errorf("architectures could not be found in lipo output: %s", architectures)
		}
		return []string{splitted[1]}, nil
	}

	splitted := strings.Split(architectures, "are: ")
	if len(splitted) != 2 {
		return []string{}, fmt.Errorf("architectures could not be found in lipo output: %s", architectures)
	}
	architectures = splitted[1]

	return strings.Split(architectures, " "), err
}

func getExecutable(binDir string) (string, error) {
	plistBytes, err := ioutil.ReadFile(path.Join(binDir, infoPlist))
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	_, err = plist.Unmarshal(plistBytes, &data)
	if err != nil {
		return "", err
	}
	if val, ok := data[executableKey]; ok {
		return val.(string), nil
	}
	return "", fmt.Errorf("%s not in Info.plist: %+v", executableKey, data)
}
