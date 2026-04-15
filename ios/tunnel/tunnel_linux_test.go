//go:build linux

package tunnel

import "testing"

func TestCapNetAdminBit(t *testing.T) {
	// CAP_NET_ADMIN is bit 12 per include/uapi/linux/capability.h.
	if capNetAdmin != 12 {
		t.Errorf("capNetAdmin = %d, want 12", capNetAdmin)
	}
}

func TestParseCapEff(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantCaps  uint64
		wantFound bool
	}{
		{
			name: "typical status block, no caps",
			status: "Name:\tbash\n" +
				"Umask:\t0022\n" +
				"State:\tS (sleeping)\n" +
				"CapInh:\t0000000000000000\n" +
				"CapPrm:\t0000000000000000\n" +
				"CapEff:\t0000000000000000\n" +
				"CapBnd:\t000001ffffffffff\n",
			wantCaps:  0x0,
			wantFound: true,
		},
		{
			name:      "only CAP_NET_ADMIN set (bit 12 = 0x1000)",
			status:    "CapEff:\t0000000000001000\n",
			wantCaps:  0x1000,
			wantFound: true,
		},
		{
			name:      "full capability set",
			status:    "CapEff:\t000001ffffffffff\n",
			wantCaps:  0x000001ffffffffff,
			wantFound: true,
		},
		{
			name:      "spaces instead of tabs after colon",
			status:    "CapEff:    0000000000001000\n",
			wantCaps:  0x1000,
			wantFound: true,
		},
		{
			name:      "line without trailing newline",
			status:    "CapEff:\t0000000000000400",
			wantCaps:  0x400,
			wantFound: true,
		},
		{
			name: "CapEff in the middle of other lines",
			status: "Name:\ttest\n" +
				"CapEff:\t0000000000000400\n" +
				"Seccomp:\t0\n",
			wantCaps:  0x400,
			wantFound: true,
		},
		{
			name:      "missing CapEff line",
			status:    "Name:\tbash\nUmask:\t0022\n",
			wantCaps:  0,
			wantFound: false,
		},
		{
			name:      "empty input",
			status:    "",
			wantCaps:  0,
			wantFound: false,
		},
		{
			name:      "malformed hex",
			status:    "CapEff:\txyzzy\n",
			wantCaps:  0,
			wantFound: false,
		},
		{
			name:      "substring match does not confuse (CapEffective is not CapEff:)",
			status:    "CapEffective:\t0000000000001000\nCapEff:\t0000000000000000\n",
			wantCaps:  0,
			wantFound: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotCaps, gotFound := parseCapEff(tc.status)
			if gotCaps != tc.wantCaps || gotFound != tc.wantFound {
				t.Errorf("parseCapEff(%q) = (%#x, %v), want (%#x, %v)",
					tc.status, gotCaps, gotFound, tc.wantCaps, tc.wantFound)
			}
		})
	}
}

// TestHasCapNetAdminDoesNotPanic is a smoke test: hasCapNetAdmin reads the
// live /proc/self/status and should return some bool without crashing.
// The actual value depends on how the test runner is invoked.
func TestHasCapNetAdminDoesNotPanic(t *testing.T) {
	_ = hasCapNetAdmin()
}

func TestCheckPermissionsSmoke(t *testing.T) {
	// Smoke test: CheckPermissions returns either nil (if running as root or
	// with CAP_NET_ADMIN) or a non-nil error. Either is valid; we just want
	// to make sure it doesn't panic and the error (if any) is descriptive.
	err := CheckPermissions()
	if err != nil && err.Error() == "" {
		t.Error("CheckPermissions returned non-nil error with empty message")
	}
}
