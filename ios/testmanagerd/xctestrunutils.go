package testmanagerd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

type XCTestRunData struct {
	RunnerTests struct {
		TestHostBundleIdentifier    string            `plist:"TestHostBundleIdentifier"`
		TestBundlePath              string            `plist:"TestBundlePath"`
		SkipTestIdentifiers         []string          `plist:"SkipTestIdentifiers"`
		OnlyTestIdentifiers         []string          `plist:"OnlyTestIdentifiers"`
		CommandLineArguments        []string          `plist:"CommandLineArguments"`
		EnvironmentVariables        map[string]string `plist:"EnvironmentVariables"`
		TestingEnvironmentVariables struct {
			DYLD_INSERT_LIBRARIES string `plist:"DYLD_INSERT_LIBRARIES"`
			XCInjectBundleInto    string `plist:"XCInjectBundleInto"`
		} `plist:"TestingEnvironmentVariables"`
	} `plist:"RunnerTests"`
	XCTestRunMetadata struct {
		FormatVersion int `plist:"FormatVersion"`
	} `plist:"__xctestrun_metadata__"`
}

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

	xctestRunDate, err := codec.Decode(file)

	// Verify that the FormatVersion is 1
	if xctestRunDate.XCTestRunMetadata.FormatVersion != 1 {
		log.Errorf("Invalid FormatVersion in .xctestrun file: got %d, expected 1", xctestRunDate.XCTestRunMetadata.FormatVersion)
		return xctestRunDate, fmt.Errorf("go-ios currently only supports .xctestrun files in formatVersion 1: "+
			"The formatVersion of your xctestrun file is %d, feel free to open an issue in https://github.com/danielpaulus/go-ios/issues to "+
			"add support", xctestRunDate.XCTestRunMetadata.FormatVersion)
	}

	return xctestRunDate, err
}

// Decode reads and decodes the binary xctestrun content from a reader
func (codec XCTestRunCodec) Decode(r io.Reader) (XCTestRunData, error) {
	if r == nil {
		return XCTestRunData{}, errors.New("reader was nil")
	}

	var result XCTestRunData
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r) // Read the entire file content
	if err != nil {
		return XCTestRunData{}, fmt.Errorf("failed to read file content: %w", err)
	}

	// Decode the xctestrun content into the struct
	_, err = plist.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return XCTestRunData{}, fmt.Errorf("failed to decode xctestrun content: %w", err)
	}

	log.Tracef("Successfully parsed .xctestrun file: %v", result)
	return result, nil
}
