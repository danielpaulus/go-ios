//go:build e2e

package e2e_test

import "testing"

// TestProfileAddRemove is skipped: installing/removing a configuration profile
// on an UNSUPERVISED device requires manual approval in Settings, so it cannot
// run unattended. It also needs a .mobileconfig fixture. Enable on a supervised
// device with a test profile committed to the repo.
func TestProfileAddRemove(t *testing.T) {
	t.Skip("profile add/remove needs a supervised device and a .mobileconfig fixture")
}
