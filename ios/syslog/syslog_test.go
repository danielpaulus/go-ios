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

func TestDecodeVis(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "vis-escaped CJK decodes to UTF-8",
			in:   `indexPath \M-f\M^H\M^V model \M-d\M-8\M-:\M-g\M-)\M-: - done`,
			want: "indexPath 或 model 为空 - done",
		},
		{
			name: "meta escape",
			in:   `\M-f`,
			want: "\xe6",
		},
		{
			name: "meta-control escape",
			in:   `\M^H`,
			want: "\x88",
		},
		{
			name: "meta-control DEL variant",
			in:   `\M^?`,
			want: "\xff",
		},
		{
			name: "plain ASCII passes through",
			in:   "hello world [123] <Notice>: nothing to decode",
			want: "hello world [123] <Notice>: nothing to decode",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "escaped backslash decodes to single backslash",
			in:   `path C:\\Users\\M-f`,
			want: `path C:\Users\M-f`,
		},
		{
			name: "lone trailing backslash stays",
			in:   `abc\`,
			want: `abc\`,
		},
		{
			name: "truncated meta escape stays",
			in:   `abc\M-`,
			want: `abc\M-`,
		},
		{
			name: "truncated meta-control escape stays",
			in:   `abc\M^`,
			want: `abc\M^`,
		},
		{
			name: "unknown escape stays",
			in:   `abc\q def\Mx`,
			want: `abc\q def\Mx`,
		},
		{
			name: "control escape is left encoded",
			in:   `esc \^[ bell \^G`,
			want: `esc \^[ bell \^G`,
		},
		{
			name: "raw UTF-8 passes through unchanged",
			in:   "已经是 UTF-8 了",
			want: "已经是 UTF-8 了",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, syslog.DecodeVis(tt.in))
		})
	}
}

func TestParserWithDecodedVisLine(t *testing.T) {
	parse := syslog.Parser()
	line := syslog.DecodeVis(`Jan 2 15:04:05 iPhone MyApp[42] <Error>: indexPath \M-f\M^H\M^V model \M-d\M-8\M-:\M-g\M-)\M-:`)
	e, err := parse(line)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.Equal(t, "MyApp", e.Process)
	assert.Equal(t, "42", e.PID)
	assert.Equal(t, "indexPath 或 model 为空", e.Message)
}
