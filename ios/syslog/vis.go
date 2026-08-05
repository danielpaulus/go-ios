package syslog

import "strings"

// DecodeVis reverses the BSD vis(3) encoding that syslog_relay applies to
// non-ASCII bytes before sending log lines. Without decoding, multi-byte
// UTF-8 characters (e.g. CJK) show up as escape sequences like
// "\M-f\M^H\M^V" instead of readable text.
//
// Only the escapes syslog_relay emits for high bytes are decoded:
//
//	\M-x  -> 0x80 | x          (meta)
//	\M^X  -> 0x80 | (X ^ 0x40) (meta + control)
//	\\    -> \                 (escaped backslash)
//
// Control-character escapes (\^X) and anything unrecognized are left
// untouched, so decoding never reintroduces raw control bytes (e.g. ESC)
// into terminal output. Lone or truncated backslash sequences pass through
// unchanged.
func DecodeVis(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			i++
			continue
		}
		rest := s[i+1:]
		switch {
		case len(rest) >= 3 && rest[0] == 'M' && rest[1] == '-':
			b.WriteByte(rest[2] | 0x80)
			i += 4
		case len(rest) >= 3 && rest[0] == 'M' && rest[1] == '^':
			b.WriteByte((rest[2] ^ 0x40) | 0x80)
			i += 4
		case len(rest) >= 1 && rest[0] == '\\':
			b.WriteByte('\\')
			i += 2
		default:
			// lone backslash or unrecognized escape: keep as-is
			b.WriteByte('\\')
			i++
		}
	}
	return b.String()
}
