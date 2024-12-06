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
		TestHostBundleIdentifier    string `plist:"TestHostBundleIdentifier"`
		TestBundlePath              string `plist:"TestBundlePath"`
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

	return codec.Decode(file)
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
