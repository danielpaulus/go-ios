package testmanagerd

import (
	"bytes"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
	"io"
	"maps"
	"os"
	"path/filepath"
)

// xctestrunutils provides utilities for parsing `.xctestrun` files with FormatVersion 1.
// It simplifies the extraction of test configurations and metadata into structured objects (`xCTestRunData`),
// enabling efficient setup for iOS test execution.
//
// Features:
// - Parses `.xctestrun` files to extract test metadata and configurations.
// - Supports building `TestConfig` objects for test execution.
//
// Note: Only `.xctestrun` files with `FormatVersion` 1 are supported. For other versions,
// contributions or requests for support can be made in the relevant GitHub repository.

// xCTestRunData represents the structure of an .xctestrun file
type xCTestRunData struct {
	TestConfig schemeData `plist:"-"`
}

// schemeData represents the structure of a scheme-specific test configuration
type schemeData struct {
	TestHostBundleIdentifier    string
	TestBundlePath              string
	SkipTestIdentifiers         []string
	OnlyTestIdentifiers         []string
	IsUITestBundle              bool
	CommandLineArguments        []string
	EnvironmentVariables        map[string]any
	TestingEnvironmentVariables map[string]any
}

// All 'xctestrun' files are versioned. The parsing of the 'xctestrun' file depends on the version. This struct helps with reading the version.
type xCTestRunFormatVersion struct {
	XCTestRunMetadata struct {
		FormatVersion int `plist:"FormatVersion"`
	} `plist:"__xctestrun_metadata__"`
}

func (data xCTestRunData) buildTestConfig(device ios.DeviceEntry, listener *TestListener) (TestConfig, error) {
	testsToRun := data.TestConfig.OnlyTestIdentifiers
	testsToSkip := data.TestConfig.SkipTestIdentifiers

	testEnv := make(map[string]any)
	if data.TestConfig.IsUITestBundle {
		maps.Copy(testEnv, data.TestConfig.EnvironmentVariables)
		maps.Copy(testEnv, data.TestConfig.TestingEnvironmentVariables)
	}

	// Extract only the file name
	var testBundlePath = filepath.Base(data.TestConfig.TestBundlePath)

	// Build the TestConfig object from parsed data
	testConfig := TestConfig{
		TestRunnerBundleId: data.TestConfig.TestHostBundleIdentifier,
		XctestConfigName:   testBundlePath,
		Args:               data.TestConfig.CommandLineArguments,
		Env:                testEnv,
		TestsToRun:         testsToRun,
		TestsToSkip:        testsToSkip,
		XcTest:             !data.TestConfig.IsUITestBundle,
		Device:             device,
		Listener:           listener,
	}

	return testConfig, nil
}

// parseFile reads the .xctestrun file and decodes it into a map
func parseFile(filePath string) (xCTestRunData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return xCTestRunData{}, fmt.Errorf("failed to open xctestrun file: %w", err)
	}
	defer file.Close()
	return decode(file)
}

// decode decodes the binary xctestrun content into the xCTestRunData struct
func decode(r io.Reader) (xCTestRunData, error) {
	// Read the entire content once
	content, err := io.ReadAll(r)
	if err != nil {
		return xCTestRunData{}, fmt.Errorf("failed to read content: %w", err)
	}

	// First, we only parse the version property of the xctestrun file. The rest of the parsing depends on this version.
	var xctestRunVersion xCTestRunFormatVersion
	_, err = plist.Unmarshal(content, &xctestRunVersion)
	if err != nil {
		return xCTestRunData{}, fmt.Errorf("xctestrun file did not contain a format version, inside the '__xctestrun_metadata__': %w", err)
	}

	if xctestRunVersion.XCTestRunMetadata.FormatVersion != 1 {
		return xCTestRunData{}, fmt.Errorf("go-ios currently only supports .xctestrun files in formatVersion 1: "+
			"The formatVersion of your xctestrun file is %d, feel free to open an issue in https://github.com/danielpaulus/go-ios/issues to "+
			"add support", xctestRunVersion.XCTestRunMetadata.FormatVersion)
	}

	if xctestRunVersion.XCTestRunMetadata.FormatVersion == 1 {
		return parseVersion1(content)
	}
	return xCTestRunData{}, nil
}

func parseVersion1(content []byte) (xCTestRunData, error) {
	// xctestrun files in version 1 use a dynamic key for the pListRoot of the TestConfig. We therefore need to jump through some hoopes when
	// parsing the file. Here we need to iterate over the root objects of the xctestrun file and find the actual test config.
	var pListRoot map[string]interface{}
	if _, err := plist.Unmarshal(content, &pListRoot); err != nil {
		return xCTestRunData{}, fmt.Errorf("failed to unmarshal plist: %w", err)
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
			return xCTestRunData{}, fmt.Errorf("failed to encode scheme %s: %w", key, err)
		}

		// Decode the plist buffer into schemeData
		decoder := plist.NewDecoder(bytes.NewReader(schemeBuf.Bytes()))
		if err := decoder.Decode(&schemeParsed); err != nil {
			return xCTestRunData{}, fmt.Errorf("failed to decode scheme %s: %w", key, err)
		}
		return xCTestRunData{TestConfig: schemeParsed}, nil
	}
	return xCTestRunData{}, nil
}
