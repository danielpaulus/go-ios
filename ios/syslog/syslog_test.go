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
			name: "meta boundary payloads",
			in:   `\M-!\M-~\M^@\M^_`,
			want: "\xa1\xfe\x80\x9f",
		},
		{
			name: "meta escape with backslash payload",
			in:   `\M-\`,
			want: "\xdc",
		},
		{
			name: "octal escape for 0xA0 continuation byte decodes",
			in:   `voil\M-C\240!`,
			want: "voilà!",
		},
		{
			name: "octal escape decodes full high-byte sequence",
			in:   `\346\210\226`,
			want: "或",
		},
		{
			name: "octal escape for backslash decodes",
			in:   `path C:\134Users\134M-f`,
			want: `path C:\Users\M-f`,
		},
		{
			name: "control-valued octal escapes stay encoded",
			in:   `nul \000 esc \033 del \177`,
			want: `nul \000 esc \033 del \177`,
		},
		{
			name: "out-of-range octal escape stays",
			in:   `\777`,
			want: `\777`,
		},
		{
			name: "truncated octal escape stays",
			in:   `abc\24`,
			want: `abc\24`,
		},
		{
			name: "two octal digits then non-octal stays",
			in:   `abc\24x`,
			want: `abc\24x`,
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
			name: "meta escape with non-graphic payload stays",
			in:   "a\\M- b\\M-\tc\\M-\nd",
			want: "a\\M- b\\M-\tc\\M-\nd",
		},
		{
			name: "meta-control escape with invalid payload stays",
			in:   `\M^a \M^0`,
			want: `\M^a \M^0`,
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
