package testmanagerd

import (
	"bytes"
	"errors"
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
	TestConfig        schemeData        `plist:"-"`
	XCTestRunMetadata xCTestRunMetadata `plist:"__xctestrun_metadata__"`
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

// XCTestRunMetadata contains metadata about the .xctestrun file
type xCTestRunMetadata struct {
	FormatVersion int `plist:"FormatVersion"`
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

	// Use a single map for initial parsing
	var rawData map[string]interface{}
	if _, err := plist.Unmarshal(content, &rawData); err != nil {
		return xCTestRunData{}, fmt.Errorf("failed to unmarshal plist: %w", err)
	}

	result := xCTestRunData{
		TestConfig: schemeData{}, // Initialize TestConfig
	}

	// Parse metadata
	metadataMap, ok := rawData["__xctestrun_metadata__"].(map[string]interface{})
	if !ok {
		return xCTestRunData{}, errors.New("invalid or missing __xctestrun_metadata__")
	}

	// Direct decoding of metadata to avoid additional conversion
	switch v := metadataMap["FormatVersion"].(type) {
	case int:
		result.XCTestRunMetadata.FormatVersion = v
	case uint64:
		result.XCTestRunMetadata.FormatVersion = int(v)
	default:
		return xCTestRunData{}, fmt.Errorf("unexpected FormatVersion type: %T", metadataMap["FormatVersion"])
	}

	// Verify FormatVersion
	if result.XCTestRunMetadata.FormatVersion != 1 {
		return result, fmt.Errorf("go-ios currently only supports .xctestrun files in formatVersion 1: "+
			"The formatVersion of your xctestrun file is %d, feel free to open an issue in https://github.com/danielpaulus/go-ios/issues to "+
			"add support", result.XCTestRunMetadata.FormatVersion)
	}

	// Parse test schemes
	if err := parseTestSchemes(rawData, &result.TestConfig); err != nil {
		return xCTestRunData{}, err
	}

	return result, nil
}

// parseTestSchemes extracts and parses test schemes from the raw data
func parseTestSchemes(rawData map[string]interface{}, scheme *schemeData) error {
	// Dynamically find and parse test schemes
	for key, value := range rawData {
		// Skip metadata key
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
			return fmt.Errorf("failed to encode scheme %s: %w", key, err)
		}

		// Decode the plist buffer into schemeData
		decoder := plist.NewDecoder(bytes.NewReader(schemeBuf.Bytes()))
		if err := decoder.Decode(&schemeParsed); err != nil {
			return fmt.Errorf("failed to decode scheme %s: %w", key, err)
		}

		// Store the scheme in the result TestConfig
		*scheme = schemeParsed
		break // Only one scheme expected, break after the first valid scheme
	}

	return nil
}
