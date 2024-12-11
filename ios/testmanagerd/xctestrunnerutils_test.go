package testmanagerd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseXCTestRunFormatVersion1(t *testing.T) {
	// Arrange: Create a temporary .xctestrun file with mock data
	tempFile, err := os.CreateTemp("", "testfile*.xctestrun")
	assert.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tempFile.Name()) // Cleanup after test

	xcTestRunFileFormatVersion1 := `
		<?xml version="1.0" encoding="UTF-8"?>
		<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
		<plist version="1.0">
			<dict>
				<key>RunnerTests</key>
				<dict>
					<key>BlueprintName</key>
					<string>RunnerTests</string>
					<key>BlueprintProviderName</key>
					<string>Runner</string>
					<key>BlueprintProviderRelativePath</key>
					<string>Runner.xcodeproj</string>
					<key>BundleIdentifiersForCrashReportEmphasis</key>
					<array>
						<string>com.example.myApp</string>
						<string>com.example.myApp.RunnerTests</string>
					</array>
					<key>CommandLineArguments</key>
					<array/>
					<key>DefaultTestExecutionTimeAllowance</key>
					<integer>600</integer>
					<key>DependentProductPaths</key>
					<array>
						<string>__TESTROOT__/Release-iphoneos/Runner.app</string>
						<string>__TESTROOT__/Release-iphoneos/Runner.app/PlugIns/RunnerTests.xctest</string>
					</array>
					<key>DiagnosticCollectionPolicy</key>
					<integer>1</integer>
					<key>EnvironmentVariables</key>
					<dict>
						<key>APP_DISTRIBUTOR_ID_OVERRIDE</key>
						<string>com.apple.AppStore</string>
						<key>OS_ACTIVITY_DT_MODE</key>
						<string>YES</string>
						<key>SQLITE_ENABLE_THREAD_ASSERTIONS</key>
						<string>1</string>
						<key>TERM</key>
						<string>dumb</string>
					</dict>
					<key>IsAppHostedTestBundle</key>
					<true/>
					<key>ParallelizationEnabled</key>
					<true/>
					<key>PreferredScreenCaptureFormat</key>
					<string>screenRecording</string>
					<key>ProductModuleName</key>
					<string>RunnerTests</string>
					<key>RunOrder</key>
					<integer>0</integer>
					<key>SystemAttachmentLifetime</key>
					<string>deleteOnSuccess</string>
					<key>TestBundlePath</key>
					<string>__TESTHOST__/PlugIns/RunnerTests.xctest</string>
					<key>TestHostBundleIdentifier</key>
					<string>com.example.myApp</string>
					<key>TestHostPath</key>
					<string>__TESTROOT__/Release-iphoneos/Runner.app</string>
					<key>TestLanguage</key>
					<string></string>
					<key>TestRegion</key>
					<string></string>
					<key>TestTimeoutsEnabled</key>
					<false/>
					<key>TestingEnvironmentVariables</key>
					<dict>
						<key>DYLD_INSERT_LIBRARIES</key>
						<string>__TESTHOST__/Frameworks/libXCTestBundleInject.dylib</string>
						<key>XCInjectBundleInto</key>
						<string>unused</string>
						<key>Test</key>
						<string>xyz</string>
					</dict>
					<key>ToolchainsSettingValue</key>
					<array/>
					<key>UserAttachmentLifetime</key>
					<string>deleteOnSuccess</string>
					<key>OnlyTestIdentifiers</key>
					<array>
						<string>TestClass1/testMethod1</string>
						<string>TestClass2/testMethod1</string>
					</array>
					<key>SkipTestIdentifiers</key>
					<array>
						<string>TestClass1/testMethod2</string>
						<string>TestClass2/testMethod2</string>
					</array>
					<key>IsUITestBundle</key>
					<true/>
				</dict>
				<key>__xctestrun_metadata__</key>
				<dict>
					<key>ContainerInfo</key>
					<dict>
						<key>ContainerName</key>
						<string>Runner</string>
						<key>SchemeName</key>
						<string>Runner</string>
					</dict>
					<key>FormatVersion</key>
					<integer>1</integer>
				</dict>
			</dict>
		</plist>
	`
	_, err = tempFile.WriteString(xcTestRunFileFormatVersion1)
	assert.NoError(t, err, "Failed to write mock data to temp file")
	tempFile.Close()

	// Act: Use the codec to parse the temp file
	codec := NewXCTestRunCodec()
	data, err := codec.ParseFile(tempFile.Name())

	// Print the parsed data before asserting
	fmt.Printf("Parsed Data: %+v\n", data)

	// Assert: Verify the parsed data
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	assert.NotNil(t, data, "Parsed data should not be nil")

	// Assert TestHostBundleIdentifier value
	assert.Equal(t, "com.example.myApp", data.TestConfig.TestHostBundleIdentifier, "TestHostBundleIdentifier mismatch")

	// Assert TestBundlePath value
	assert.Equal(t, "__TESTHOST__/PlugIns/RunnerTests.xctest", data.TestConfig.TestBundlePath, "TestBundlePath mismatch")

	// Assert EnvironmentVariables values
	assert.Equal(t, map[string]any{
		"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"TERM":                            "dumb",
	}, data.TestConfig.EnvironmentVariables, "EnvironmentVariables mismatch")

	// Assert TestingEnvironmentVariables values
	assert.Equal(t, map[string]any{
		"DYLD_INSERT_LIBRARIES": "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
		"XCInjectBundleInto":    "unused",
		"Test":                  "xyz",
	}, data.TestConfig.TestingEnvironmentVariables, "TestingEnvironmentVariables mismatch")

	// Assert CommandLineArguments values
	assert.Equal(t, []string{}, data.TestConfig.CommandLineArguments, "CommandLineArguments mismatch")

	// Assert OnlyTestIdentifiers values
	assert.Equal(t, []string{
		"TestClass1/testMethod1",
		"TestClass2/testMethod1",
	}, data.TestConfig.OnlyTestIdentifiers, "OnlyTestIdentifiers mismatch")

	// Assert SkipTestIdentifiers values
	assert.Equal(t, []string{
		"TestClass1/testMethod2",
		"TestClass2/testMethod2",
	}, data.TestConfig.SkipTestIdentifiers, "SkipTestIdentifiers mismatch")

	// Assert XCTestRunMetadata values
	assert.Equal(t, 1, data.XCTestRunMetadata.FormatVersion, "FormatVersion mismatch")
}

func TestParseXCTestRunNotSupportedForFormatVersionOtherThanOne(t *testing.T) {
	// Arrange: Create a temporary .xctestrun file with mock data
	tempFile, err := os.CreateTemp("", "testfile*.xctestrun")
	assert.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tempFile.Name()) // Cleanup after test

	xcTestRunFileFormatVersion2 := `
		<?xml version="1.0" encoding="UTF-8"?>
		<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
		<plist version="1.0">
		<dict>
			<key>__xctestrun_metadata__</key>
			<dict>
				<key>FormatVersion</key>
				<integer>2</integer>
			</dict>
		</dict>
		</plist>
	`
	_, err = tempFile.WriteString(xcTestRunFileFormatVersion2)
	assert.NoError(t, err, "Failed to write mock data to temp file")
	tempFile.Close()

	// Act: Use the codec to parse the temp file
	codec := NewXCTestRunCodec()
	_, err = codec.ParseFile(tempFile.Name())

	// Assert the Error Message
	assert.Equal(t, "go-ios currently only supports .xctestrun files in formatVersion 1: The formatVersion of your xctestrun file is 2, feel free to open an issue in https://github.com/danielpaulus/go-ios/issues to add support", err.Error(), "Error Message mismatch")
}
