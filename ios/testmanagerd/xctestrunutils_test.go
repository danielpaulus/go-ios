package testmanagerd

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/stretchr/testify/assert"
)

// Helper function to create mock data and parse the .xctestrun file
func setupParsing(t *testing.T, filePath string) []testConfiguration {
	// Act: parse version 1 of xctestrun file
	xcTestRunData, err := parseFile(filePath)
	assert.NoError(t, err, "Failed to parse .xctestrun file")
	return xcTestRunData
}

// Test Parsing an xctestrun file with format version 1
func TestParsingV1(t *testing.T) {
	xcTestRunData := setupParsing(t, "testdata/format_version_1.xctestrun")
	var expectedTestConfigurations = []testConfiguration{
		{
			Name: "",
			TestTargets: []schemeData{
				{
					TestHostBundleIdentifier: "com.example.myApp",
					TestBundlePath:           "__TESTHOST__/PlugIns/RunnerTests.xctest",
					EnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
						"OS_ACTIVITY_DT_MODE":             "YES",
						"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
						"TERM":                            "dumb",
					},
					TestingEnvironmentVariables: map[string]any{
						"DYLD_INSERT_LIBRARIES": "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
						"XCInjectBundleInto":    "unused",
						"Test":                  "xyz",
					},
					CommandLineArguments: []string{},
					OnlyTestIdentifiers: []string{
						"TestClass1/testMethod1",
						"TestClass2/testMethod1",
					},
					SkipTestIdentifiers: []string{
						"TestClass1/testMethod2",
						"TestClass2/testMethod2",
					},
					IsUITestBundle: true,
				},
			},
		},
	}
	assert.Equal(t, expectedTestConfigurations, xcTestRunData)
}

// Test Building test configs from a parsed xctestrun file with format version 1
func TestBuildTestConfigV1(t *testing.T) {
	// Arrange: Create parsed XCTestRunData using the helper function
	testConfigurations := setupParsing(t, "testdata/format_version_1.xctestrun")

	// Mock dependencies
	mockDevice := ios.DeviceEntry{
		DeviceID: 8110,
	}
	mockListener := &TestListener{}

	// Act: Convert testConfigSpecification to TestConfig
	var testConfigs []TestConfig
	for _, testConfigSpecification := range testConfigurations {
		for _, r := range testConfigSpecification.TestTargets {
			tc, _ := r.buildTestConfig(mockDevice, mockListener, nil)
			testConfigs = append(testConfigs, tc)
		}
	}

	var expected = []TestConfig{
		{
			TestRunnerBundleId: "com.example.myApp",
			XctestConfigName:   "RunnerTests.xctest",
			Args:               []string{},
			Env: map[string]any{
				"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
				"OS_ACTIVITY_DT_MODE":             "YES",
				"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
				"TERM":                            "dumb",
				"DYLD_INSERT_LIBRARIES":           "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
				"XCInjectBundleInto":              "unused",
				"Test":                            "xyz",
			},
			TestsToRun: []string{
				"TestClass1/testMethod1",
				"TestClass2/testMethod1",
			},
			TestsToSkip: []string{
				"TestClass1/testMethod2",
				"TestClass2/testMethod2",
			},
			XcTest:   false,
			Device:   mockDevice,
			Listener: mockListener,
		},
	}
	assert.Equal(t, expected, testConfigs)
}

// Test Parsing an xctestrun file with format version 2
func TestParsingV2(t *testing.T) {
	testTargets := setupParsing(t, "testdata/format_version_2.xctestrun")

	var expectedTestConfigurations = []testConfiguration{
		{
			Name: "Test Scheme Action",
			TestTargets: []schemeData{
				{
					TestHostBundleIdentifier: "saucelabs.FakeCounterApp",
					TestBundlePath:           "__TESTHOST__/PlugIns/FakeCounterAppTests.xctest",
					EnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
						"OS_ACTIVITY_DT_MODE":             "YES",
						"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
						"TERM":                            "dumb",
					},
					TestingEnvironmentVariables: map[string]any{
						"DYLD_INSERT_LIBRARIES": "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
						"XCInjectBundleInto":    "unused",
					},
					CommandLineArguments: []string{},
					OnlyTestIdentifiers:  nil,
					SkipTestIdentifiers: []string{
						"SkippedTests", "SkippedTests/testThatAlwaysFailsAndShouldBeSkipped",
					},
					IsUITestBundle:                  false,
					UITargetAppEnvironmentVariables: nil,
					UITargetAppPath:                 "",
				},
				{
					TestHostBundleIdentifier: "saucelabs.FakeCounterAppUITests.xctrunner",
					TestBundlePath:           "__TESTHOST__/PlugIns/FakeCounterAppUITests.xctest",
					EnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
						"OS_ACTIVITY_DT_MODE":             "YES",
						"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
						"TERM":                            "dumb",
					},
					TestingEnvironmentVariables: map[string]any{},
					CommandLineArguments:        []string{},
					OnlyTestIdentifiers:         nil,
					SkipTestIdentifiers:         nil,
					IsUITestBundle:              true,
					UITargetAppEnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE": "com.apple.AppStore",
					},
					UITargetAppPath: "__TESTROOT__/Debug-iphoneos/FakeCounterApp.app",
				},
			},
		},
	}

	assert.Equal(t, expectedTestConfigurations, testTargets)
}

// Test Parsing an xctestrun file with format version 2
func TestParsingV2_Multiple(t *testing.T) {
	testTargets := setupParsing(t, "testdata/format_version_2_multiple.xctestrun")

	var expectedTestConfigurations = []testConfiguration{
		{
			Name: "TestCounterApp_1",
			TestTargets: []schemeData{
				{
					TestHostBundleIdentifier: "saucelabs.FakeCounterAppUITests.xctrunner",
					TestBundlePath:           "__TESTHOST__/PlugIns/FakeCounterAppUITests.xctest",
					EnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
						"OS_ACTIVITY_DT_MODE":             "YES",
						"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
						"TERM":                            "dumb",
					},
					TestingEnvironmentVariables: map[string]any{},
					CommandLineArguments:        []string{},
					OnlyTestIdentifiers:         nil,
					SkipTestIdentifiers:         nil,
					IsUITestBundle:              true,
					UITargetAppEnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE": "com.apple.AppStore",
					},
					UITargetAppPath: "__TESTROOT__/Debug-iphoneos/FakeCounterApp.app",
				},
				{
					TestHostBundleIdentifier: "saucelabs.FakeCounterApp",
					TestBundlePath:           "__TESTHOST__/PlugIns/FakeCounterAppTests.xctest",
					EnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
						"OS_ACTIVITY_DT_MODE":             "YES",
						"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
						"TERM":                            "dumb",
					},
					TestingEnvironmentVariables: map[string]any{
						"DYLD_INSERT_LIBRARIES": "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
						"XCInjectBundleInto":    "unused",
					},
					CommandLineArguments: []string{},
					OnlyTestIdentifiers:  nil,
					SkipTestIdentifiers: []string{
						"SkippedTests", "SkippedTests/testThatAlwaysFailsAndShouldBeSkipped",
					},
					IsUITestBundle:                  false,
					UITargetAppEnvironmentVariables: nil,
					UITargetAppPath:                 "",
				},
			},
		},
		{
			Name: "TestDuplicateApp_2",
			TestTargets: []schemeData{
				{
					TestHostBundleIdentifier: "saucelabs.FakeCounterDuplicateApp",
					TestBundlePath:           "__TESTHOST__/PlugIns/FakeCounterDuplicateAppTests.xctest",
					SkipTestIdentifiers:      nil,
					OnlyTestIdentifiers:      nil,
					IsUITestBundle:           false,
					CommandLineArguments:     []string{},
					EnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
						"OS_ACTIVITY_DT_MODE":             "YES",
						"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
						"TERM":                            "dumb",
					},
					TestingEnvironmentVariables: map[string]any{
						"DYLD_INSERT_LIBRARIES": "__TESTHOST__/Frameworks/libXCTestBundleInject.dylib",
						"XCInjectBundleInto":    "unused",
					},
					UITargetAppEnvironmentVariables: nil,
					UITargetAppPath:                 "",
				},
				{
					TestHostBundleIdentifier: "saucelabs.FakeCounterDuplicateAppUITests.xctrunner",
					TestBundlePath:           "__TESTHOST__/PlugIns/FakeCounterDuplicateAppUITests.xctest",
					SkipTestIdentifiers:      nil,
					OnlyTestIdentifiers:      nil,
					IsUITestBundle:           true,
					CommandLineArguments:     []string{},
					EnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
						"OS_ACTIVITY_DT_MODE":             "YES",
						"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
						"TERM":                            "dumb",
					},
					TestingEnvironmentVariables: map[string]any{},
					UITargetAppEnvironmentVariables: map[string]any{
						"APP_DISTRIBUTOR_ID_OVERRIDE": "com.apple.AppStore",
					},
					UITargetAppPath: "__TESTROOT__/Debug-iphoneos/FakeCounterDuplicateApp.app",
				},
			},
		},
	}

	assert.Equal(t, expectedTestConfigurations, testTargets)
}

func TestBuildTestConfigV2(t *testing.T) {
	testConfigurations := setupParsing(t, "testdata/format_version_2.xctestrun")

	// Mock dependencies
	mockDevice := ios.DeviceEntry{
		DeviceID: 8110,
	}
	mockListener := &TestListener{}
	// Build allApps mock data to verify the getBundleID function
	allAppsMockData := []installationproxy.AppInfo{
		{
			"CFBundleName":       "FakeCounterApp",
			"CFBundleIdentifier": "saucelabs.FakeCounterApp",
		},
	}
	var testConfigs []TestConfig
	for _, testConfigSpecification := range testConfigurations {
		for _, r := range testConfigSpecification.TestTargets {
			tc, _ := r.buildTestConfig(mockDevice, mockListener, allAppsMockData)
			testConfigs = append(testConfigs, tc)
		}
	}

	var expected = []TestConfig{
		{
			TestRunnerBundleId: "saucelabs.FakeCounterApp",
			XctestConfigName:   "FakeCounterAppTests.xctest",
			Args:               []string{},
			Env:                map[string]any{},
			TestsToRun:         nil,
			TestsToSkip:        []string{"SkippedTests", "SkippedTests/testThatAlwaysFailsAndShouldBeSkipped"},
			XcTest:             true,
			Device:             mockDevice,
			Listener:           mockListener,
		},
		{
			BundleId:           "saucelabs.FakeCounterApp",
			TestRunnerBundleId: "saucelabs.FakeCounterAppUITests.xctrunner",
			XctestConfigName:   "FakeCounterAppUITests.xctest",
			Args:               []string{},
			Env: map[string]any{
				"APP_DISTRIBUTOR_ID_OVERRIDE":     "com.apple.AppStore",
				"OS_ACTIVITY_DT_MODE":             "YES",
				"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
				"TERM":                            "dumb",
			},
			TestsToRun:  nil,
			TestsToSkip: nil,
			XcTest:      false,
			Device:      mockDevice,
			Listener:    mockListener,
		},
	}
	assert.Equal(t, expected, testConfigs)
}

// Test When we use an invalid xctestrun file.
func TestParseXCTestRunFormatV2ThrowsErrorEmptyTestConfigurations(t *testing.T) {
	// Act: Use the codec to parse the temp file
	_, err := parseFile("testdata/contains_invalid_test_configuration.xctestrun")
	// Assert the Error Message
	assert.Equal(t, "The .xctestrun file you provided does not contain any test configurations. Please check your test setup and ensure it includes at least one test configuration.", err.Error(), "Error Message mismatch")
}
