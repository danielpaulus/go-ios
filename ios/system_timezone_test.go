package ios

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZoneNameFromPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"linux zoneinfo", "/usr/share/zoneinfo/America/Chicago", "America/Chicago"},
		{"macos versioned zoneinfo", "/private/var/db/timezone/tz/2026c.1.0/zoneinfo/Europe/Berlin", "Europe/Berlin"},
		{"macos default zoneinfo", "/usr/share/zoneinfo.default/Europe/Berlin", "Europe/Berlin"},
		{"single segment zone", "/usr/share/zoneinfo/UTC", "UTC"},
		{"three segment zone", "/usr/share/zoneinfo/America/Argentina/Salta", "America/Argentina/Salta"},
		{"no zoneinfo segment", "/etc/localtime", ""},
		{"unknown zone", "/usr/share/zoneinfo/Mars/Olympus_Mons", ""},
		{"empty", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, zoneNameFromPath(tc.path))
		})
	}
}

func TestSystemTimezoneFromEnv(t *testing.T) {
	// point the file lookups at nothing so only TZ can answer
	withPaths(t, filepath.Join(t.TempDir(), "missing"), filepath.Join(t.TempDir(), "missing"))

	t.Run("valid TZ is used", func(t *testing.T) {
		t.Setenv("TZ", "America/Chicago")
		assert.Equal(t, "America/Chicago", SystemTimezone())
	})
	t.Run("TZ path prefix is stripped", func(t *testing.T) {
		t.Setenv("TZ", ":America/Chicago")
		assert.Equal(t, "America/Chicago", SystemTimezone())
	})
	t.Run("empty TZ means UTC", func(t *testing.T) {
		t.Setenv("TZ", "")
		assert.Equal(t, "UTC", SystemTimezone())
	})
	t.Run("invalid TZ falls back", func(t *testing.T) {
		t.Setenv("TZ", "Mars/Olympus_Mons")
		assert.Equal(t, "UTC", SystemTimezone())
	})
}

func TestSystemTimezoneFromFiles(t *testing.T) {
	t.Setenv("TZ", "Mars/Olympus_Mons") // invalid, so the file lookups decide
	dir := t.TempDir()

	t.Run("localtime symlink into zoneinfo", func(t *testing.T) {
		link := filepath.Join(dir, "localtime")
		zone := filepath.Join(dir, "zoneinfo", "Europe", "Berlin")
		require.NoError(t, os.MkdirAll(filepath.Dir(zone), 0o755))
		require.NoError(t, os.WriteFile(zone, nil, 0o644))
		require.NoError(t, os.Symlink(zone, link))
		withPaths(t, link, filepath.Join(dir, "missing"))

		assert.Equal(t, "Europe/Berlin", SystemTimezone())
	})

	t.Run("timezone file when localtime is a plain copy", func(t *testing.T) {
		tzFile := filepath.Join(dir, "timezone")
		require.NoError(t, os.WriteFile(tzFile, []byte("America/Chicago\n"), 0o644))
		withPaths(t, filepath.Join(dir, "missing"), tzFile)

		assert.Equal(t, "America/Chicago", SystemTimezone())
	})

	t.Run("nothing resolvable falls back to UTC", func(t *testing.T) {
		withPaths(t, filepath.Join(dir, "missing"), filepath.Join(dir, "missing"))
		assert.Equal(t, "UTC", SystemTimezone())
	})
}

// TestSystemTimezoneIsIANA guards the actual bug this replaced: time.Local.String()
// returns "Local" on macOS, which is not a timezone any device accepts.
func TestSystemTimezoneIsIANA(t *testing.T) {
	assert.True(t, isIANAZone(SystemTimezone()), "SystemTimezone must return a loadable IANA name")
}

func withPaths(t *testing.T, localtime, timezone string) {
	t.Helper()
	origLocaltime, origTimezone := localtimePath, timezonePath
	localtimePath, timezonePath = localtime, timezone
	t.Cleanup(func() { localtimePath, timezonePath = origLocaltime, origTimezone })
}
