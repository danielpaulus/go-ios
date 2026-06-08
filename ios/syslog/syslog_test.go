package syslog_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	parse := syslog.Parser()

	tests := []struct {
		name      string
		line      string
		expectErr bool
		process   string
		pid       string
		level     string
		message   string
	}{
		{
			name:    "valid line parses",
			line:    "Jan 2 15:04:05 iPhone SpringBoard[123] <Notice>: hello world",
			process: "SpringBoard",
			pid:     "123",
			level:   "Notice",
			message: "hello world",
		},
		{
			name:      "malformed line returns error",
			line:      "this is not a syslog line",
			expectErr: true,
		},
		{
			name:      "empty line returns error",
			line:      "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				e, err := parse(tt.line)
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, e)
					return
				}
				require.NoError(t, err)
				require.NotNil(t, e)
				assert.Equal(t, tt.process, e.Process)
				assert.Equal(t, tt.pid, e.PID)
				assert.Equal(t, tt.level, e.Level)
				assert.Equal(t, tt.message, e.Message)
			})
		})
	}
}
