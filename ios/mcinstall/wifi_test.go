package mcinstall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXmlEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ampersand", "a&b", "a&amp;b"},
		{"less than", "a<b", "a&lt;b"},
		{"greater than", "a>b", "a&gt;b"},
		{"double quote", `a"b`, "a&quot;b"},
		{"single quote", "a'b", "a&apos;b"},
		{"all entities", `<>&"'`, "&lt;&gt;&amp;&quot;&apos;"},
		{"no special chars", "plain", "plain"},
		{"empty", "", ""},
		{
			"injection attempt",
			`</string><key>evil</key><string>`,
			"&lt;/string&gt;&lt;key&gt;evil&lt;/key&gt;&lt;string&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, xmlEscape(tt.input))
		})
	}
}

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain alphanumeric", "abc123", "abc123"},
		{"spaces replaced", "my ssid", "my-ssid"},
		{"special chars replaced", "a<>&\"'b", "a-b"},
		{"leading and trailing specials trimmed", "--abc--", "abc"},
		{"only specials", "<>&", ""},
		{"mixed", "Wi-Fi @ Home!", "Wi-Fi-Home"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeIdentifier(tt.input))
		})
	}
}
