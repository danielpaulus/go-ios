package syslog

import "strings"

// DecodeVis reverses the BSD vis(3) encoding that syslog_relay applies to
// non-ASCII bytes before sending log lines. Without decoding, multi-byte
// UTF-8 characters (e.g. CJK) show up as escape sequences like
// "\M-f\M^H\M^V" instead of readable text.
//
// Only the escapes syslog_relay emits for high bytes and backslashes are
// decoded:
//
//	\M-x  -> 0x80 | x          (meta)
//	\M^X  -> 0x80 | (X ^ 0x40) (meta + control)
//	\ddd  -> octal byte value  (vis octal-encodes 0xA0 as \240 and, in
//	                            current Apple libc, backslash as \134)
//	\\    -> \                 (escaped backslash, older vis variants)
//
// Octal escapes are only decoded to high bytes (>= 0x80) or printable ASCII;
// control-valued octal escapes, control-character escapes (\^X) and anything
// unrecognized are left untouched, so decoding never reintroduces raw control
// bytes (e.g. ESC) into terminal output. Lone or truncated backslash
// sequences pass through unchanged.
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
		case len(rest) >= 3 && rest[0] == 'M' && rest[1] == '-' && isGraph(rest[2]):
			b.WriteByte(rest[2] | 0x80)
			i += 4
		case len(rest) >= 3 && rest[0] == 'M' && rest[1] == '^' && isMetaCtrl(rest[2]):
			b.WriteByte((rest[2] ^ 0x40) | 0x80)
			i += 4
		case len(rest) >= 3 && isOctal(rest[0]) && isOctal(rest[1]) && isOctal(rest[2]) &&
			decodableOctal(octalValue(rest[0], rest[1], rest[2])):
			b.WriteByte(byte(octalValue(rest[0], rest[1], rest[2])))
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

// isGraph reports whether c is a graphic ASCII character — the only payloads
// vis emits after "\M-".
func isGraph(c byte) bool {
	return c > ' ' && c < 0x7f
}

// isMetaCtrl reports whether c is a payload vis emits after "\M^":
// '@'..'_' for control values 0x00-0x1F, or '?' for DEL.
func isMetaCtrl(c byte) bool {
	return (c >= '@' && c <= '_') || c == '?'
}

func isOctal(c byte) bool {
	return c >= '0' && c <= '7'
}

func octalValue(a, b, c byte) int {
	return int(a-'0')<<6 | int(b-'0')<<3 | int(c-'0')
}

// decodableOctal reports whether an octal escape value is safe to decode:
// a high byte (part of multi-byte UTF-8) or printable ASCII. Control values
// (and out-of-range ones like \777) stay encoded.
func decodableOctal(v int) bool {
	return (v >= 0x80 && v <= 0xff) || (v >= 0x20 && v < 0x7f)
}
