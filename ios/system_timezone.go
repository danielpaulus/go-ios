package ios

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// paths the system records its timezone in. Variables so tests can redirect them.
var (
	localtimePath = "/etc/localtime"
	timezonePath  = "/etc/timezone"
)

// SystemTimezone returns the host timezone as an IANA timezone database name
// (e.g. "America/Chicago"), which is what lockdown expects for the "TimeZone"
// value. It falls back to "UTC" when no valid IANA name can be determined.
//
// time.Local.String() cannot be used for this: it returns Go's zone name
// ("Local", "UTC" or an abbreviation like "CEST"), never a zoneinfo path, so
// on macOS it yields the invalid timezone "Local".
func SystemTimezone() string {
	if tz, ok := os.LookupEnv("TZ"); ok {
		// Go treats an explicitly empty TZ as UTC.
		tz = strings.TrimPrefix(tz, ":")
		if tz == "" {
			return "UTC"
		}
		if isIANAZone(tz) {
			return tz
		}
	}
	// macOS and most Linux distributions symlink /etc/localtime into the
	// zoneinfo database, so the link target carries the IANA name.
	for _, resolve := range []func(string) (string, error){os.Readlink, filepath.EvalSymlinks} {
		target, err := resolve(localtimePath)
		if err != nil {
			continue
		}
		if name := zoneNameFromPath(target); name != "" {
			return name
		}
	}
	// Debian-based images often ship a plain /etc/localtime copy plus this file.
	if b, err := os.ReadFile(timezonePath); err == nil {
		if name := strings.TrimSpace(string(b)); isIANAZone(name) {
			return name
		}
	}
	return "UTC"
}

// zoneNameFromPath extracts the IANA name from a zoneinfo file path such as
// "/usr/share/zoneinfo/America/Chicago" or, on macOS,
// "/var/db/timezone/tz/2026c.1.0/zoneinfo/Europe/Berlin".
func zoneNameFromPath(path string) string {
	segments := strings.Split(filepath.ToSlash(path), "/")
	for i := len(segments) - 1; i >= 0; i-- {
		// the database directory is "zoneinfo", but macOS also uses
		// versioned variants like "zoneinfo.default".
		if !strings.HasPrefix(segments[i], "zoneinfo") {
			continue
		}
		name := strings.Join(segments[i+1:], "/")
		if isIANAZone(name) {
			return name
		}
		return ""
	}
	return ""
}

func isIANAZone(name string) bool {
	if name == "" || name == "Local" {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}
