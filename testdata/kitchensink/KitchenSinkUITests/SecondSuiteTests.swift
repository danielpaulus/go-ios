import XCTest

// SecondSuiteTests exists purely so the report contains a second <testsuite>,
// letting the e2e test assert that go-ios emits one <testsuite> per XCTestCase
// subclass and that the top-level <testsuites> aggregates across both.
final class SecondSuiteTests: XCTestCase {

    func testAlsoPasses() throws {
        XCTAssertTrue(true)
        Thread.sleep(forTimeInterval: 0.2)
    }
}
