package main

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/stretchr/testify/assert"
)

func TestMatchesKillTarget(t *testing.T) {
	processNames := map[string]string{
		"Preferences": "com.apple.Preferences",
		"MobileMail":  "", // --process target, no bundleID
	}

	tests := []struct {
		name         string
		p            instruments.ProcessInfo
		pid          uint64
		wantBundleID string
		wantByName   bool
		wantByPid    bool
	}{
		{
			name:         "matches by bundleID-resolved name",
			p:            instruments.ProcessInfo{Name: "Preferences", Pid: 1184},
			pid:          0,
			wantBundleID: "com.apple.Preferences",
			wantByName:   true,
		},
		{
			name:       "matches by plain --process name",
			p:          instruments.ProcessInfo{Name: "MobileMail", Pid: 900},
			pid:        0,
			wantByName: true,
		},
		{
			name: "no match, unrelated process, no --pid given",
			p:    instruments.ProcessInfo{Name: "Springboard", Pid: 100},
			pid:  0,
		},
		{
			name:      "matches by explicit pid",
			p:         instruments.ProcessInfo{Name: "Springboard", Pid: 100},
			pid:       100,
			wantByPid: true,
		},
		{
			// Regression: pid 0 is the Mach kernel's real, legitimate pid in
			// the process list. An absent --pid (which zero-values to 0) must
			// never be treated as "matches pid 0" — otherwise every kill call
			// without --pid would also try to kill the kernel entry.
			name: "pid 0 in process list never matches an absent --pid",
			p:    instruments.ProcessInfo{Name: "Mach Kernel", Pid: 0},
			pid:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundleID, byName, byPid := matchesKillTarget(tt.p, processNames, tt.pid)
			assert.Equal(t, tt.wantBundleID, bundleID)
			assert.Equal(t, tt.wantByName, byName)
			assert.Equal(t, tt.wantByPid, byPid)
		})
	}
}

func TestShouldWatchKill(t *testing.T) {
	processNames := map[string]string{
		"Preferences": "com.apple.Preferences",
		"MobilePhone": "com.apple.mobilephone",
	}

	tests := []struct {
		name         string
		event        instruments.AppStateEvent
		wantBundleID string
		wantOK       bool
	}{
		{
			name:         "blocked app launching",
			event:        instruments.AppStateEvent{ProcessName: "Preferences", Pid: 1079, State: "Running"},
			wantBundleID: "com.apple.Preferences",
			wantOK:       true,
		},
		{
			name:  "not a watched app",
			event: instruments.AppStateEvent{ProcessName: "MobileSafari", Pid: 1036, State: "Running"},
		},
		{
			name:  "watched app but only backgrounded",
			event: instruments.AppStateEvent{ProcessName: "Preferences", Pid: 1079, State: "Suspended"},
		},
		{
			name:  "watched app terminating",
			event: instruments.AppStateEvent{ProcessName: "Preferences", Pid: 1079, State: "Terminated"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundleID, ok := shouldWatchKill(tt.event, processNames)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantBundleID, bundleID)
		})
	}
}
