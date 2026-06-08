//go:build e2e

package preios17_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// TestAccessibilityAudit launches a content-rich system app and runs the
// accessibility audit against it, asserting the device reports at least one
// issue and that each carries an issue type. Exercises the iOS 14 audit path
// (named case IDs via deviceBeginAuditCaseIDs:).
func TestAccessibilityAudit(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		harness.AuditAfterLaunch(t, udid, knownSystemApp)
	})
}
