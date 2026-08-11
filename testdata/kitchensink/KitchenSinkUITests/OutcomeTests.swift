import XCTest

// OutcomeTests exercises every JUnit outcome go-ios's ios/junit mapping cares
// about: pass, fail (with a message + file:line), expected failure, XCTSkip, and
// a passing test that produces attachments. Each test sleeps a distinct amount
// so the emitted <testcase time="..."> durations are nonzero and distinguishable.
final class OutcomeTests: XCTestCase {

    func testPasses() throws {
        XCTAssertTrue(true)
        Thread.sleep(forTimeInterval: 0.3)
    }

    func testFails() throws {
        Thread.sleep(forTimeInterval: 0.5)
        XCTAssertEqual(1, 2, "kitchen sink failure message")
    }

    func testExpectedFailure() throws {
        XCTExpectFailure("known bug") {
            XCTFail("boom")
        }
    }

    func testSkipped() throws {
        throw XCTSkip("skipped on purpose")
    }

    func testWithAttachment() throws {
        let a1 = XCTAttachment(string: "log data")
        a1.name = "primary-log"
        a1.lifetime = .keepAlways
        add(a1)
        let a2 = XCTAttachment(string: "second attachment payload")
        a2.name = "secondary-log"
        a2.lifetime = .keepAlways
        add(a2)
        XCTAssertTrue(true)
    }

    // testStalls is a stalling test that is only enabled when
    // KITCHENSINK_ENABLE_STALL=1 is passed via --env. It is OFF by default so the
    // e2e run stays stable; the stalled->error mapping is covered by junit unit
    // tests. When enabled it sleeps long enough to trip testmanagerd's stall
    // detection.
    func testStalls() throws {
        guard ProcessInfo.processInfo.environment["KITCHENSINK_ENABLE_STALL"] == "1" else {
            throw XCTSkip("stall test disabled (set KITCHENSINK_ENABLE_STALL=1 to enable)")
        }
        Thread.sleep(forTimeInterval: 600)
    }
}
