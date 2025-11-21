package testmanagerd

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"howett.net/plist"
)

// xctestrunutils provides utilities for parsing `.xctestrun` files.
// It simplifies the extraction of test configurations and metadata into structured objects (`xCTestRunData`),
// enabling efficient setup for iOS test execution.
//
// Features:
// - Parses `.xctestrun` files to extract test metadata and configurations.
// - Supports building `TestConfig` objects for test execution.

// schemeData represents the structure of a scheme-specific test configuration
type schemeData struct {
	TestHostBundleIdentifier        string
	TestBundlePath                  string
	SkipTestIdentifiers             []string
	OnlyTestIdentifiers             []string
	IsUITestBundle                  bool
	CommandLineArguments            []string
	EnvironmentVariables            map[string]any
	TestingEnvironmentVariables     map[string]any
	UITargetAppEnvironmentVariables map[string]any
	UITargetAppPath                 string
}

type testConfiguration struct {
	Name        string       `plist:"Name"`
	TestTargets []schemeData `plist:"TestTargets"`
}

func (data schemeData) buildTestConfig(device ios.DeviceEntry, listener *TestListener, installedApps []installationproxy.AppInfo) (TestConfig, error) {
	testsToRun := data.OnlyTestIdentifiers
	testsToSkip := data.SkipTestIdentifiers

	testEnv := make(map[string]any)
	var bundleId string

	if data.IsUITestBundle {
		maps.Copy(testEnv, data.EnvironmentVariables)
		maps.Copy(testEnv, data.TestingEnvironmentVariables)
		maps.Copy(testEnv, data.UITargetAppEnvironmentVariables)
		// Only call getBundleID if :
		// - allAps is provided
		// - UITargetAppPath is populated since it can be empty for UI tests in some edge cases
		if len(data.UITargetAppPath) > 0 && installedApps != nil {
			bundleId = getBundleId(installedApps, data.UITargetAppPath)
		}
	}

	// Extract only the file name
	var testBundlePath = filepath.Base(data.TestBundlePath)

	// Build the TestConfig object from parsed data
	testConfig := TestConfig{
		BundleId:           bundleId,
		TestRunnerBundleId: data.TestHostBundleIdentifier,
		XctestConfigName:   testBundlePath,
		Args:               data.CommandLineArguments,
		Env:                testEnv,
		TestsToRun:         testsToRun,
		TestsToSkip:        testsToSkip,
		XcTest:             !data.IsUITestBundle,
		Device:             device,
		Listener:           listener,
	}

	return testConfig, nil
}

// parseFile reads the .xctestrun file and decodes it into a map
func parseFile(filePath string) ([]testConfiguration, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []testConfiguration{}, fmt.Errorf("failed to open xctestrun file: %w", err)
	}
	defer file.Close()
	return decode(file)
}

// decode decodes the binary xctestrun content into the xCTestRunData struct
func decode(r io.Reader) ([]testConfiguration, error) {
	// Read the entire content once
	xctestrunFileContent, err := io.ReadAll(r)
	if err != nil {
		return []testConfiguration{}, fmt.Errorf("unable to read xctestrun content: %w", err)
	}

	// First, we only parse the version property of the xctestrun file. The rest of the parsing depends on this version.
	version, err := getFormatVersion(xctestrunFileContent)
	if err != nil {
		return []testConfiguration{}, err
	}

	switch version {
	case 1:
		return parseVersion1(xctestrunFileContent)
	case 2:
		return parseVersion2(xctestrunFileContent)
	default:
		return []testConfiguration{}, fmt.Errorf("the provided .xctestrun format version %d is not supported", version)
	}
}

// Helper method to get the format version of the xctestrun file
func getFormatVersion(xctestrunFileContent []byte) (int, error) {

	type xCTestRunMetadata struct {
		Metadata struct {
			Version int `plist:"FormatVersion"`
		} `plist:"__xctestrun_metadata__"`
	}

	var metadata xCTestRunMetadata
	if _, err := plist.Unmarshal(xctestrunFileContent, &metadata); err != nil {
		return 0, fmt.Errorf("failed to parse format version: %w", err)
	}

	return metadata.Metadata.Version, nil
}

func parseVersion1(xctestrunFile []byte) ([]testConfiguration, error) {
	// xctestrun files in version 1 use a dynamic key for the pListRoot of the TestConfig. As in the 'key' for the TestConfig is the name
	// of the app. This forces us to iterate over the root of the plist, instead of using a static struct to decode the xctestrun file.
	var pListRoot map[string]interface{}
	if _, err := plist.Unmarshal(xctestrunFile, &pListRoot); err != nil {
		return []testConfiguration{}, fmt.Errorf("failed to unmarshal plist: %w", err)
	}

	for key, value := range pListRoot {
		// Skip the metadata object
		if key == "__xctestrun_metadata__" {
			continue
		}

		// Attempt to convert to schemeData
		schemeMap, ok := value.(map[string]interface{})
		if !ok {
			continue // Skip if not a valid scheme map
		}

		// Parse the scheme into schemeData and update the TestConfig
		var schemeParsed schemeData
		schemeBuf := new(bytes.Buffer)
		encoder := plist.NewEncoder(schemeBuf)
		if err := encoder.Encode(schemeMap); err != nil {
			return []testConfiguration{}, fmt.Errorf("failed to encode scheme %s: %w", key, err)
		}

		// Decode the plist buffer into schemeData
		decoder := plist.NewDecoder(bytes.NewReader(schemeBuf.Bytes()))
		if err := decoder.Decode(&schemeParsed); err != nil {
			return []testConfiguration{}, fmt.Errorf("failed to decode scheme %s: %w", key, err)
		}
		// Convert the return type to table of testConfiguration
		return []testConfiguration{{
			Name:        "", // No specific name available, leaving it empty
			TestTargets: []schemeData{schemeParsed},
		}}, nil
	}
	return []testConfiguration{}, nil
}

func parseVersion2(content []byte) ([]testConfiguration, error) {
	type xCTestRunVersion2 struct {
		ContainerInfo struct {
			ContainerName string `plist:"ContainerName"`
		} `plist:"ContainerInfo"`
		TestConfigurations []testConfiguration `plist:"TestConfigurations"`
	}

	var testConfigs xCTestRunVersion2
	if _, err := plist.Unmarshal(content, &testConfigs); err != nil {
		return []testConfiguration{}, fmt.Errorf("failed to parse format version: %w", err)
	}

	// Check if TestConfigurations is empty
	if len(testConfigs.TestConfigurations) == 0 {
		return []testConfiguration{}, fmt.Errorf("The .xctestrun file you provided does not contain any test configurations. Please check your test setup and ensure it includes at least one test configuration.")
	}

	// Return a table of TestConfigurations
	return testConfigs.TestConfigurations, nil
}

func getBundleId(installedApps []installationproxy.AppInfo, uiTargetAppPath string) string {
	var appNameWithSuffix = filepath.Base(uiTargetAppPath)
	var uiTargetAppName = strings.TrimSuffix(appNameWithSuffix, ".app")
	for _, app := range installedApps {
		if app.CFBundleName() == uiTargetAppName {
			return app.CFBundleIdentifier()
		}
	}
	return ""
}
