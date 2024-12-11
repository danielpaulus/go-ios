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

// XCTestRunData represents the structure of an .xctestrun file
type XCTestRunData struct {
	TestConfig        SchemeData        `plist:"-"`
	XCTestRunMetadata XCTestRunMetadata `plist:"__xctestrun_metadata__"`
}

func (data XCTestRunData) toTestConfig(device ios.DeviceEntry, listener *TestListener) (TestConfig, error) {
	testsToRun := data.TestConfig.OnlyTestIdentifiers
	testsToSkip := data.TestConfig.SkipTestIdentifiers

	testEnv := make(map[string]any)
	maps.Copy(testEnv, data.TestConfig.EnvironmentVariables)
	maps.Copy(testEnv, data.TestConfig.TestingEnvironmentVariables)

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

// SchemeData represents the structure of a scheme-specific test configuration
type SchemeData struct {
	TestHostBundleIdentifier    string         `plist:"TestHostBundleIdentifier"`
	TestBundlePath              string         `plist:"TestBundlePath"`
	SkipTestIdentifiers         []string       `plist:"SkipTestIdentifiers"`
	OnlyTestIdentifiers         []string       `plist:"OnlyTestIdentifiers"`
	IsUITestBundle              bool           `plist:"IsUITestBundle"`
	CommandLineArguments        []string       `plist:"CommandLineArguments"`
	EnvironmentVariables        map[string]any `plist:"EnvironmentVariables"`
	TestingEnvironmentVariables map[string]any `plist:"TestingEnvironmentVariables"`
}

// XCTestRunMetadata contains metadata about the .xctestrun file
type XCTestRunMetadata struct {
	FormatVersion int `plist:"FormatVersion"`
}

// XCTestRunCodec is a utility for parsing .xctestrun files with FormatVersion 1.
// It extracts test configurations and metadata, providing a structured SchemeData object
// that contains all the necessary information required to execute a test.
// This includes details like test bundle paths, environment variables, command-line arguments,
// and other configuration settings essential for running tests.

// XCTestRunCodec handles encoding and decoding operations for .xctestrun files
type XCTestRunCodec struct{}

// NewXCTestRunCodec creates a new instance of XCTestRunCodec
func NewXCTestRunCodec() XCTestRunCodec {
	return XCTestRunCodec{}
}

// ParseFile reads the .xctestrun file and decodes it into a map
func (codec XCTestRunCodec) ParseFile(filePath string) (XCTestRunData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return XCTestRunData{}, fmt.Errorf("failed to open xctestrun file: %w", err)
	}
	defer file.Close()
	return codec.Decode(file)
}

// Decode decodes the binary xctestrun content into the XCTestRunData struct
func (codec XCTestRunCodec) Decode(r io.Reader) (XCTestRunData, error) {
	// Read the entire content once
	content, err := io.ReadAll(r)
	if err != nil {
		return XCTestRunData{}, fmt.Errorf("failed to read content: %w", err)
	}

	// Use a single map for initial parsing
	var rawData map[string]interface{}
	if _, err := plist.Unmarshal(content, &rawData); err != nil {
		return XCTestRunData{}, fmt.Errorf("failed to unmarshal plist: %w", err)
	}

	result := XCTestRunData{
		TestConfig: SchemeData{}, // Initialize TestConfig
	}

	// Parse metadata
	metadataMap, ok := rawData["__xctestrun_metadata__"].(map[string]interface{})
	if !ok {
		return XCTestRunData{}, errors.New("invalid or missing __xctestrun_metadata__")
	}

	// Direct decoding of metadata to avoid additional conversion
	switch v := metadataMap["FormatVersion"].(type) {
	case int:
		result.XCTestRunMetadata.FormatVersion = v
	case uint64:
		result.XCTestRunMetadata.FormatVersion = int(v)
	default:
		return XCTestRunData{}, fmt.Errorf("unexpected FormatVersion type: %T", metadataMap["FormatVersion"])
	}

	// Verify FormatVersion
	if result.XCTestRunMetadata.FormatVersion != 1 {
		return result, fmt.Errorf("go-ios currently only supports .xctestrun files in formatVersion 1: "+
			"The formatVersion of your xctestrun file is %d, feel free to open an issue in https://github.com/danielpaulus/go-ios/issues to "+
			"add support", result.XCTestRunMetadata.FormatVersion)
	}

	// Parse test schemes
	if err := codec.parseTestSchemes(rawData, &result.TestConfig); err != nil {
		return XCTestRunData{}, err
	}

	return result, nil
}

// parseTestSchemes extracts and parses test schemes from the raw data
func (codec XCTestRunCodec) parseTestSchemes(rawData map[string]interface{}, scheme *SchemeData) error {
	// Dynamically find and parse test schemes
	for key, value := range rawData {
		// Skip metadata key
		if key == "__xctestrun_metadata__" {
			continue
		}

		// Attempt to convert to SchemeData
		schemeMap, ok := value.(map[string]interface{})
		if !ok {
			continue // Skip if not a valid scheme map
		}

		// Parse the scheme into SchemeData and update the TestConfig
		var schemeParsed SchemeData
		schemeBuf := new(bytes.Buffer)
		encoder := plist.NewEncoder(schemeBuf)
		if err := encoder.Encode(schemeMap); err != nil {
			return fmt.Errorf("failed to encode scheme %s: %w", key, err)
		}

		// Decode the plist buffer into SchemeData
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
