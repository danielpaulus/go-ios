package testmanagerd

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Helper function to create mock data and parse the .xctestrun file format v1
func parseXCTestRunFileVersion1(t *testing.T) schemeData {
	// Act: parse version 1 of xctestrun file
	xcTestRunData, err := parseFile("testdata/format_version_1.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	return xcTestRunData[0]
}

func TestTestHostBundleIdentifier(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, "com.example.myApp", xcTestRunData.TestHostBundleIdentifier, "TestHostBundleIdentifier mismatch")
}

func TestTestBundlePath(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, "__TESTHOST__/PlugIns/RunnerTests.xctest", xcTestRunData.TestBundlePath, "TestBundlePath mismatch")
}

func TestEnvironmentVariables(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, map[string]any{
		"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"TERM":                            "dumb",
	}, xcTestRunData.EnvironmentVariables, "EnvironmentVariables mismatch")
}

func TestTestingEnvironmentVariables(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, map[string]any{
		"DYLD_INSERT_LIBRARIES": "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
		"XCInjectBundleInto":    "unused",
		"Test":                  "xyz",
	}, xcTestRunData.TestingEnvironmentVariables, "TestingEnvironmentVariables mismatch")
}

func TestCommandLineArguments(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, []string{}, xcTestRunData.CommandLineArguments, "CommandLineArguments mismatch")
}

func TestOnlyTestIdentifiers(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, []string{
		"TestClass1/testMethod1",
		"TestClass2/testMethod1",
	}, xcTestRunData.OnlyTestIdentifiers, "OnlyTestIdentifiers mismatch")
}

func TestSkipTestIdentifiers(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, []string{
		"TestClass1/testMethod2",
		"TestClass2/testMethod2",
	}, xcTestRunData.SkipTestIdentifiers, "SkipTestIdentifiers mismatch")
}

func TestIsUITestBundle(t *testing.T) {
	xcTestRunData := parseXCTestRunFileVersion1(t)
	assert.Equal(t, true, xcTestRunData.IsUITestBundle, "IsUITestBundle mismatch")
}

func TestParseXCTestRunFormatV2ThrowsErrorForMultipleTestConfigurations(t *testing.T) {
	// Act: Use the codec to parse the temp file
	_, err := parseFile("testdata/contains_invalid_test_configuration.xctestrun")
	// Assert the Error Message
	assert.Equal(t, "The .xctestrun file you provided contained 2 entries in the TestConfiguration list. This list should contain exactly 1 entry. Please revisit your test configuration so that it only contains one entry.", err.Error(), "Error Message mismatch")
}

// Helper function to create testConfig from parsed mock data using .xctestrun file format v1
func createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t *testing.T) (TestConfig, ios.DeviceEntry, *TestListener) {
	// Arrange: Create parsed XCTestRunData using the helper function
	xcTestRunData := parseXCTestRunFileVersion1(t)

	// Mock dependencies
	mockDevice := ios.DeviceEntry{
		DeviceID: 8110,
	}
	mockListener := &TestListener{}

	// Act: Convert XCTestRunData to TestConfig
	testConfig, err := xcTestRunData.buildTestConfig(mockDevice, mockListener)

	// Assert: Validate the returned TestConfig
	assert.NoError(t, err, "Error converting to TestConfig")

	return testConfig, mockDevice, mockListener
}

func TestConfigTestRunnerBundleId(t *testing.T) {
	testConfig, _, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, "com.example.myApp", testConfig.TestRunnerBundleId, "TestRunnerBundleId mismatch")
}

func TestConfigXctestConfigName(t *testing.T) {
	testConfig, _, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, "RunnerTests.xctest", testConfig.XctestConfigName, "XctestConfigName mismatch")
}

func TestConfigCommandLineArguments(t *testing.T) {
	testConfig, _, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, []string{}, testConfig.Args, "data mismatch")
}

func TestConfigEnvironmentVariables(t *testing.T) {
	testConfig, _, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, map[string]any{
		"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"TERM":                            "dumb",
		"DYLD_INSERT_LIBRARIES":           "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
		"XCInjectBundleInto":              "unused",
		"Test":                            "xyz",
	}, testConfig.Env, "EnvironmentVariables mismatch")
}

func TestConfigTestsToRun(t *testing.T) {
	testConfig, _, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, []string{
		"TestClass1/testMethod1",
		"TestClass2/testMethod1",
	}, testConfig.TestsToRun, "TestsToRun mismatch")
}

func TestConfigTestsToSkip(t *testing.T) {
	testConfig, _, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, []string{
		"TestClass1/testMethod2",
		"TestClass2/testMethod2",
	}, testConfig.TestsToSkip, "TestsToSkip mismatch")
}

func TestConfigXcTest(t *testing.T) {
	testConfig, _, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, false, testConfig.XcTest, "XcTest mismatch")
}

func TestConfigDevice(t *testing.T) {
	testConfig, mockDevice, _ := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, mockDevice, testConfig.Device, "Device mismatch")
}

func TestConfigListener(t *testing.T) {
	testConfig, _, mockListener := createTestConfigFromParsedMockDataUsingXCTestRunFileV1(t)
	assert.Equal(t, mockListener, testConfig.Listener, "Listener mismatch")
}

// Test XCTest Config Parsing with format version 2

func TestTestHostBundleIdentifier_XCTestRunFileVersion2_XCTest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCTest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xctestTarget := testTargets[0]

	// Assert
	assert.Equal(t, "saucelabs.FakeCounterApp", xctestTarget.TestHostBundleIdentifier, "TestHostBundleIdentifier mismatch")
}

func TestTestBundlePath_XCTestRunFileVersion2_XCTest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCTest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xctestTarget := testTargets[0]

	// Assert
	assert.Equal(t, "__TESTHOST__/PlugIns/FakeCounterAppTests.xctest", xctestTarget.TestBundlePath, "TestBundlePath mismatch")
}

func TestEnvironmentVariables_XCTestRunFileVersion2_XCTest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCTest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xctestTarget := testTargets[0]

	// Assert
	assert.Equal(t, map[string]any{
		"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"TERM":                            "dumb",
	}, xctestTarget.EnvironmentVariables, "EnvironmentVariables mismatch")
}

func TestTestingEnvironmentVariables_XCTestRunFileVersion2_XCTest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCTest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xctestTarget := testTargets[0]

	// Assert
	assert.Equal(t, map[string]any{
		"DYLD_INSERT_LIBRARIES": "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
		"XCInjectBundleInto":    "unused",
	}, xctestTarget.TestingEnvironmentVariables, "TestingEnvironmentVariables mismatch")
}

func TestCommandLineArguments_XCTestRunFileVersion2_XCTest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCTest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xctestTarget := testTargets[0]

	// Assert
	assert.Equal(t, []string{}, xctestTarget.CommandLineArguments, "CommandLineArguments mismatch")
}

func TestSkipTestIdentifiers_XCTestRunFileVersion2_XCTest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCTest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xctestTarget := testTargets[0]

	// Assert
	assert.Equal(t, []string{
		"SkippedTests", "SkippedTests/testThatAlwaysFailsAndShouldBeSkipped",
	}, xctestTarget.SkipTestIdentifiers, "SkipTestIdentifiers mismatch")
}

func TestIsUITestBundle_XCTestRunFileVersion2_XCTest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCTest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xctestTarget := testTargets[0]

	// Assert
	assert.Equal(t, false, xctestTarget.IsUITestBundle, "IsUITestBundle mismatch")
}

// Test XCUITest Config Parsing with format version 2

func TestTestHostBundleIdentifier_XCTestRunFileVersion2_XCUITest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCUITest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xcUITestTarget := testTargets[1]

	// Assert
	assert.Equal(t, "saucelabs.FakeCounterAppUITests.xctrunner", xcUITestTarget.TestHostBundleIdentifier, "TestHostBundleIdentifier mismatch")
}

func TestTestBundlePath_XCTestRunFileVersion2_XCUITest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCUITest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xcUITestTarget := testTargets[1]

	// Assert
	assert.Equal(t, "__TESTHOST__/PlugIns/FakeCounterAppUITests.xctest", xcUITestTarget.TestBundlePath, "TestBundlePath mismatch")
}

func TestEnvironmentVariables_XCTestRunFileVersion2_XCUITest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCUITest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xcUITestTarget := testTargets[1]

	// Assert
	assert.Equal(t, map[string]any{
		"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"TERM":                            "dumb",
	}, xcUITestTarget.EnvironmentVariables, "EnvironmentVariables mismatch")
}

func TestTestingEnvironmentVariables_XCTestRunFileVersion2_XCUITest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCUITest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xcUITestTarget := testTargets[1]

	// Assert
	assert.Equal(t, map[string]any{}, xcUITestTarget.TestingEnvironmentVariables, "TestingEnvironmentVariables mismatch")
}

func TestCommandLineArguments_XCTestRunFileVersion2_XCUITest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCUITest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xcUITestTarget := testTargets[1]

	// Assert
	assert.Equal(t, []string{}, xcUITestTarget.CommandLineArguments, "CommandLineArguments mismatch")
}

func TestIsUITestBundle_XCTestRunFileVersion2_XCUITest(t *testing.T) {
	// Arrange: parse version 2 of xctestrun file and get the XCUITest target from the 'Test Target' array.
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xcUITestTarget := testTargets[1]

	// Assert
	assert.Equal(t, true, xcUITestTarget.IsUITestBundle, "IsUITestBundle mismatch")
}

// Helper function to create testConfig from parsed mock data using .xctestrun file format v2
// If includeUITest is true, it returns a UI test configuration.
// If includeUITest is false, it returns a non-UI test configuration.
func buildTestConfigXcTestRunFileV2(t *testing.T) (TestConfig, ios.DeviceEntry, *TestListener) {
	// Arrange: Create parsed XCTestRunData using the helper function
	testTargets, err := parseFile("testdata/format_version_2.xctestrun")
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	xcUITestTarget := testTargets[0]

	// Assert
	// Mock dependencies
	mockDevice := ios.DeviceEntry{
		DeviceID: 8110,
	}
	mockListener := &TestListener{}

	// Act: Convert XCTestRunData to TestConfig
	testConfig, err := xcUITestTarget.buildTestConfig(mockDevice, mockListener)

	// Assert: Validate the returned TestConfig
	assert.NoError(t, err, "Error converting to TestConfig")

	return testConfig, mockDevice, mockListener
}

// Test Building Test Config for XCTest using xctestrun format version 2
func TestConfigTestRunnerBundleId_XCTestRunFileVersion2_XCTest(t *testing.T) {
	testConfig, _, _ := buildTestConfigXcTestRunFileV2(t)
	assert.Equal(t, "saucelabs.FakeCounterApp", testConfig.TestRunnerBundleId, "TestRunnerBundleId mismatch")
}

func TestConfigXctestConfigName_XCTestRunFileVersion2_XCTest(t *testing.T) {
	testConfig, _, _ := buildTestConfigXcTestRunFileV2(t)
	assert.Equal(t, "FakeCounterAppTests.xctest", testConfig.XctestConfigName, "XctestConfigName mismatch")
}

func TestConfigCommandLineArguments_XCTestRunFileVersion2_XCTest(t *testing.T) {
	testConfig, _, _ := buildTestConfigXcTestRunFileV2(t)
	assert.Equal(t, []string{}, testConfig.Args, "data mismatch")
}

func TestConfigEnvironmentVariables_XCTestRunFileVersion2_XCTest(t *testing.T) {
	testConfig, _, _ := buildTestConfigXcTestRunFileV2(t)
	assert.Equal(t, map[string]any{}, testConfig.Env, "EnvironmentVariables mismatch")
}

func TestConfigTestsToSkip_XCTestRunFileVersion2_XCTest(t *testing.T) {
	testConfig, _, _ := buildTestConfigXcTestRunFileV2(t)
	assert.Equal(t, []string{
		"SkippedTests", "SkippedTests/testThatAlwaysFailsAndShouldBeSkipped",
	}, testConfig.TestsToSkip, "TestsToSkip mismatch")
}

func TestConfigXcTest_XCTestRunFileVersion2_XCTest(t *testing.T) {
	testConfig, _, _ := buildTestConfigXcTestRunFileV2(t)
	assert.Equal(t, true, testConfig.XcTest, "XcTest mismatch")
}
