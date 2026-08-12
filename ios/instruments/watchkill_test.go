package instruments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldKill(t *testing.T) {
	blocked := map[string]bool{"Preferences": true, "MobilePhone": true}

	tests := []struct {
		name         string
		notification map[string]interface{}
		wantOK       bool
		wantProcess  string
		wantPid      uint64
	}{
		{
			name: "blocked app launching",
			notification: map[string]interface{}{
				"appName":           "Preferences",
				"pid":               uint64(1079),
				"state_description": "Running",
			},
			wantOK:      true,
			wantProcess: "Preferences",
			wantPid:     1079,
		},
		{
			name: "accepts numeric pid types",
			notification: map[string]interface{}{
				"appName":           "MobilePhone",
				"pid":               int(42),
				"state_description": "Running",
			},
			wantOK:      true,
			wantProcess: "MobilePhone",
			wantPid:     42,
		},
		{
			name: "not a blocked app",
			notification: map[string]interface{}{
				"appName":           "MobileSafari",
				"pid":               uint64(1036),
				"state_description": "Running",
			},
			wantOK: false,
		},
		{
			name: "blocked app but only backgrounded",
			notification: map[string]interface{}{
				"appName":           "Preferences",
				"pid":               uint64(1079),
				"state_description": "Suspended",
			},
			wantOK: false,
		},
		{
			name: "blocked app terminating",
			notification: map[string]interface{}{
				"appName":           "Preferences",
				"pid":               uint64(1079),
				"state_description": "Terminated",
			},
			wantOK: false,
		},
		{
			name:         "missing fields",
			notification: map[string]interface{}{"state_description": "Running"},
			wantOK:       false,
		},
		{
			name: "unparseable pid",
			notification: map[string]interface{}{
				"appName":           "Preferences",
				"pid":               "not-a-number",
				"state_description": "Running",
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, pid, ok := shouldKill(tt.notification, blocked)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantProcess, name)
				assert.Equal(t, tt.wantPid, pid)
			}
		})
	}
}
