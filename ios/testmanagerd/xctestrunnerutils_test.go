package testmanagerd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseXCTestRun(t *testing.T) {
	// Arrange: Create a temporary .xctestrun file with mock data
	tempFile, err := os.CreateTemp("", "testfile*.xctestrun")
	assert.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tempFile.Name()) // Cleanup after test

	mockData := `
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
					</dict>
					<key>ToolchainsSettingValue</key>
					<array/>
					<key>UserAttachmentLifetime</key>
					<string>deleteOnSuccess</string>
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
	_, err = tempFile.WriteString(mockData)
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

	// Check specific fields in the parsed data
	assert.Equal(t, "com.example.myApp", data.RunnerTests.TestHostBundleIdentifier, "TestHostBundleIdentifier mismatch")
	assert.Equal(t, "__TESTHOST__/PlugIns/RunnerTests.xctest", data.RunnerTests.TestBundlePath, "TestBundlePath mismatch")

	// Assert XCTestRunMetadata values
	assert.Equal(t, 1, data.XCTestRunMetadata.FormatVersion, "FormatVersion mismatch")
}
