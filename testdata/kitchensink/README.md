# KitchenSink XCUITest bundle

A tiny SwiftUI host app plus an XCUITest target used by the e2e test
`test/e2e/kitchensink_junit_test.go` to verify go-ios's `ios runtest
--junit-output` (PR #799 / `feat/issue-286-junit-report`) against a *real*
XCTest run on the device farm.

The built, **unsigned** runner is checked in as `testdata/kitchensink-runner.zip`
(like `testdata/wda.ipa`). The e2e re-signs and installs it. You only need to
rebuild if you change the Swift sources.

## Bundle identifiers

| Bundle | ID |
| --- | --- |
| Host app (`KitchenSink.app`) | `com.deviceboxhq.goios.kitchensink` |
| UI-test bundle (`KitchenSinkUITests.xctest`) | `com.deviceboxhq.goios.kitchensink` |
| Test runner (`KitchenSinkUITests-Runner.app`) | `com.deviceboxhq.goios.kitchensink.xctrunner` |

The runner id is the xctest bundle id + Xcode's automatic `.xctrunner` suffix.
The runner `.app` inside the zip is `KitchenSinkUITests-Runner.app` and contains
`PlugIns/KitchenSinkUITests.xctest`.

`ios runtest` is invoked with:

```
ios runtest \
  --test-runner-bundle-id=com.deviceboxhq.goios.kitchensink.xctrunner \
  --xctest-config=KitchenSinkUITests.xctest \
  --junit-output=<path>/report.xml
```

Note: **no `--bundle-id`.** The `.xctest` bundle is packaged inside the runner,
so the runner is its own test host — there is no separate app-under-test to
install. If you pass `--bundle-id=<host app>`, testmanagerd tries to look up that
app among the installed apps and, when it isn't installed, the run yields an
empty report. The tests here are pure asserts that never drive
`XCUIApplication`, so a target app is unnecessary. (Only the runner is checked in
/ installed; the `KitchenSink.app` host target exists just to let Xcode generate
the UI-test runner.)

## Test outcomes (what each method produces in the JUnit report)

`OutcomeTests` (→ one `<testsuite name="OutcomeTests">`):

| Method | Outcome | JUnit element |
| --- | --- | --- |
| `testPasses` | pass (sleeps 0.3s) | plain `<testcase>`, no child |
| `testFails` | fail (sleeps 0.5s, `XCTAssertEqual(1,2,"kitchen sink failure message")`) | `<failure message="…kitchen sink failure message…">` + `OutcomeTests.swift:<line>` chardata |
| `testExpectedFailure` | expected failure (`XCTExpectFailure{ XCTFail("boom") }`) | `<skipped message="expected failure…">` |
| `testSkipped` | skipped (`throw XCTSkip("skipped on purpose")`) | `<skipped>` mentioning "skipped on purpose" |
| `testWithAttachment` | pass, two `XCTAttachment(string:)` | plain `<testcase>`, report intact |
| `testStalls` | **skipped by default**; sleeps 600s only when `KITCHENSINK_ENABLE_STALL=1` is passed via `--env` (kept OFF so the e2e stays stable) |

`SecondSuiteTests` (→ one `<testsuite name="SecondSuiteTests">`):

| Method | Outcome | JUnit element |
| --- | --- | --- |
| `testAlsoPasses` | pass (sleeps 0.2s) | plain `<testcase>`, no child |

Two XCTestCase subclasses ⇒ two `<testsuite>` elements aggregated under one
`<testsuites>`.

## Rebuilding the runner

Requires Xcode with an iOS device SDK. Build **unsigned** (the e2e re-signs):

```
cd testdata/kitchensink
xcodebuild build \
  -project KitchenSink.xcodeproj \
  -target KitchenSink -target KitchenSinkUITests \
  -sdk iphoneos -arch arm64 \
  CONFIGURATION_BUILD_DIR="$PWD/build/Debug-iphoneos" \
  CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""
```

Then repackage (drop the `.dSYM` so the zip stays small):

```
cd build/Debug-iphoneos
rm -rf KitchenSinkUITests-Runner.app/PlugIns/*.dSYM
rm -f ../../../kitchensink-runner.zip
zip -qry ../../../kitchensink-runner.zip KitchenSinkUITests-Runner.app
```

The zip must contain `KitchenSinkUITests-Runner.app/PlugIns/KitchenSinkUITests.xctest/KitchenSinkUITests`.

### Why `-target` instead of `build-for-testing -scheme … -destination generic/platform=iOS`

`build-for-testing` resolves a run destination, and when a device advertising an
iOS version whose platform component isn't installed is connected, Xcode marks
both the device and the `generic/platform=iOS` placeholder ineligible ("iOS X is
not installed"), so no destination is found. Building the two targets directly
with `-sdk iphoneos -arch arm64` sidesteps destination resolution and produces
the identical `-Runner.app` (Xcode's UI-testing product type still generates the
runner and copies the XCTest frameworks). If your machine has the matching
platform component, the standard `build-for-testing` command works too.
