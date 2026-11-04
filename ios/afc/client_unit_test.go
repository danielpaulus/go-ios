package afc

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// rwc adapts a bytes.Buffer (read side) and discards writes so we can drive a
// Client without a real device connection.
type rwc struct {
	r io.Reader
}

func (c *rwc) Read(p []byte) (int, error)  { return c.r.Read(p) }
func (c *rwc) Write(p []byte) (int, error) { return len(p), nil }
func (c *rwc) Close() error                { return nil }

// encodeAfcPacket builds a raw AFC packet for the given operation and payload.
func encodeAfcPacket(op opcode, headerPayload, payload []byte) []byte {
	thisLen := headerSize + uint64(len(headerPayload))
	h := header{
		Magic:     magic,
		EntireLen: thisLen + uint64(len(payload)),
		ThisLen:   thisLen,
		PacketNum: 0,
		Operation: op,
	}
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, h)
	buf.Write(headerPayload)
	buf.Write(payload)
	return buf.Bytes()
}

// TestFileReadOversizedPayload verifies that File.Read never returns a count
// larger than len(p), even when the device sends more bytes than were
// requested (which would violate the io.Reader contract).
func TestFileReadOversizedPayload(t *testing.T) {
	payload := []byte("0123456789") // 10 bytes
	raw := encodeAfcPacket(fileRead, nil, payload)

	f := &File{
		client: &Client{connection: &rwc{r: bytes.NewReader(raw)}},
		handle: 1,
	}

	p := make([]byte, 4) // smaller than the payload
	n, err := f.Read(p)
	assert.NoError(t, err)
	assert.LessOrEqual(t, n, len(p))
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("0123"), p)
}

// TestFileReadEOF verifies that an empty payload is reported as io.EOF.
func TestFileReadEOF(t *testing.T) {
	raw := encodeAfcPacket(fileRead, nil, nil)

	f := &File{
		client: &Client{connection: &rwc{r: bytes.NewReader(raw)}},
		handle: 1,
	}

	p := make([]byte, 8)
	n, err := f.Read(p)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

// TestAfcErrorDescriptiveMessage verifies afcError.Error() includes the
// human-readable description for a known code.
func TestAfcErrorDescriptiveMessage(t *testing.T) {
	err := afcError{code: errObjectNotFound}
	assert.Contains(t, err.Error(), "ObjectNotFound")
	assert.Contains(t, err.Error(), "8")
}

// encodeFileInfoPayload builds the payload of an AFC GET_FILE_INFO reply: a
// series of null-terminated key/value pairs, in the order a device sends them.
func encodeFileInfoPayload(tokens ...string) []byte {
	var buf bytes.Buffer
	for _, tok := range tokens {
		buf.WriteString(tok)
		buf.WriteByte(0)
	}
	return buf.Bytes()
}

// statAgainstReply drives Client.Stat against a canned GET_FILE_INFO reply, so
// the test covers the real parsing path instead of only the parser.
func statAgainstReply(t *testing.T, p string, payload []byte) FileInfo {
	t.Helper()
	raw := encodeAfcPacket(fileInfo, nil, payload)
	c := &Client{connection: &rwc{r: bytes.NewReader(raw)}}
	info, err := c.Stat(p)
	assert.NoError(t, err)
	return info
}

// TestStatParsesTimestamps verifies st_mtime and st_birthtime reach FileInfo.
// Both are nanoseconds since the Unix epoch; iOS sends them for every DCIM
// entry, and a client that drops them cannot tell when a file was written
// without reading it.
func TestStatParsesTimestamps(t *testing.T) {
	payload := encodeFileInfoPayload(
		"st_size", "12345678",
		"st_blocks", "24416",
		"st_nlink", "1",
		"st_ifmt", "S_IFREG",
		"st_mtime", "1762240518123456789",
		"st_birthtime", "1761000000000000000",
		"st_mode", "100644",
	)

	info := statAgainstReply(t, "/DCIM/100APPLE/IMG_0001.HEIC", payload)

	assert.Equal(t, "IMG_0001.HEIC", info.Name)
	assert.Equal(t, FileType("S_IFREG"), info.Type)
	assert.False(t, info.IsDir())
	assert.Equal(t, int64(12345678), info.Size)
	assert.Equal(t, uint32(0o100644), info.Mode)
	assert.True(t, info.Mtime.Equal(time.Unix(0, 1762240518123456789).UTC()), "mtime = %v", info.Mtime)
	assert.True(t, info.Birthtime.Equal(time.Unix(0, 1761000000000000000).UTC()), "birthtime = %v", info.Birthtime)
}

// TestStatTimestampsAreUTC verifies the decoded times do not depend on the
// host's time zone: the AFC reply carries none.
func TestStatTimestampsAreUTC(t *testing.T) {
	payload := encodeFileInfoPayload("st_mtime", "1762240518123456789")

	info := statAgainstReply(t, "/DCIM/100APPLE/IMG_0001.HEIC", payload)

	assert.Equal(t, time.UTC, info.Mtime.Location())
}

// TestStatWithoutTimestampsKeepsZeroTime verifies a reply that omits the
// timestamp keys leaves them at the zero time, so callers can tell "unknown"
// from a real date.
func TestStatWithoutTimestampsKeepsZeroTime(t *testing.T) {
	payload := encodeFileInfoPayload("st_size", "42", "st_ifmt", "S_IFDIR")

	info := statAgainstReply(t, "/DCIM/100APPLE", payload)

	assert.True(t, info.Mtime.IsZero())
	assert.True(t, info.Birthtime.IsZero())
	assert.True(t, info.IsDir())
	assert.Equal(t, int64(42), info.Size)
}

// TestStatIgnoresMalformedTimestamp verifies a timestamp the client cannot
// parse costs only that field: the rest of the reply is still returned.
func TestStatIgnoresMalformedTimestamp(t *testing.T) {
	payload := encodeFileInfoPayload(
		"st_size", "42",
		"st_mtime", "not-a-number",
		"st_birthtime", "",
	)

	info := statAgainstReply(t, "/DCIM/100APPLE/IMG_0002.MOV", payload)

	assert.True(t, info.Mtime.IsZero())
	assert.True(t, info.Birthtime.IsZero())
	assert.Equal(t, int64(42), info.Size)
}

// TestParseFileInfoUnknownKeys verifies fields the client does not model are
// skipped instead of ending the parse.
func TestParseFileInfoUnknownKeys(t *testing.T) {
	payload := encodeFileInfoPayload(
		"st_some_future_field", "whatever",
		"st_size", "7",
	)

	info := parseFileInfo(payload)

	assert.Equal(t, int64(7), info.Size)
}

// TestParseFileInfoTruncatedPayload verifies a reply that ends after a key
// (no value, no null terminator) is not parsed past the end.
func TestParseFileInfoTruncatedPayload(t *testing.T) {
	payload := encodeFileInfoPayload("st_size", "7")
	payload = append(payload, []byte("st_mtime")...) // truncated: no value

	info := parseFileInfo(payload)

	assert.Equal(t, int64(7), info.Size)
	assert.True(t, info.Mtime.IsZero())
}

// TestAfcTimestampBoundaries pins the edges of the timestamp decoding: 0 is
// the Unix epoch rather than "unknown", so a file the device reports at the
// epoch stays distinguishable from one it says nothing about.
func TestAfcTimestampBoundaries(t *testing.T) {
	epoch := time.Unix(0, 0).UTC()
	tests := []struct {
		name     string
		value    string
		want     time.Time
		wantZero bool
	}{
		{name: "epoch", value: "0", want: epoch},
		{name: "one nanosecond", value: "1", want: time.Unix(0, 1).UTC()},
		{name: "negative", value: "-1", wantZero: true},
		{name: "overflows int64", value: "9223372036854775808", wantZero: true},
		{name: "empty", value: "", wantZero: true},
		{name: "not a number", value: "not-a-number", wantZero: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := afcTimestamp(tt.value)
			assert.Equal(t, tt.wantZero, got.IsZero(), "afcTimestamp(%q) = %v", tt.value, got)
			assert.True(t, got.Equal(tt.want), "afcTimestamp(%q) = %v, want %v", tt.value, got, tt.want)
			assert.Equal(t, time.UTC, got.Location())
		})
	}
}

// TestStatEpochTimestamps verifies a reply that reports 0 reaches the caller
// as the epoch, not as the zero time that means "the device did not say".
func TestStatEpochTimestamps(t *testing.T) {
	payload := encodeFileInfoPayload("st_mtime", "0", "st_birthtime", "0")

	info := statAgainstReply(t, "/DCIM/100APPLE/IMG_0001.HEIC", payload)

	epoch := time.Unix(0, 0).UTC()
	assert.False(t, info.Mtime.IsZero())
	assert.True(t, info.Mtime.Equal(epoch), "mtime = %v, want epoch", info.Mtime)
	assert.True(t, info.Birthtime.Equal(epoch), "birthtime = %v, want epoch", info.Birthtime)
}

// TestStatParsesDeviceReply runs the parser over a reply captured verbatim from
// a device (iPhone16,2 on iOS 26.4.1, /DCIM/173APPLE/IMG_3794.HEIC), so the fixture
// reflects what AFC actually sends rather than only what the format allows.
//
// Worth noting for anyone editing this file: a real reply carries st_size,
// st_blocks, st_nlink, st_ifmt, st_mtime and st_birthtime, and no st_mode — so
// FileInfo.Mode is 0 on a real device even though the parser handles the key.
func TestStatParsesDeviceReply(t *testing.T) {
	payload := []byte("st_size\x004199202\x00" +
		"st_blocks\x008208\x00" +
		"st_nlink\x001\x00" +
		"st_ifmt\x00S_IFREG\x00" +
		"st_mtime\x001775095043590371464\x00" +
		"st_birthtime\x001775095043579017563\x00")

	info := statAgainstReply(t, "/DCIM/173APPLE/IMG_3794.HEIC", payload)

	assert.Equal(t, "IMG_3794.HEIC", info.Name)
	assert.Equal(t, FileType("S_IFREG"), info.Type)
	assert.False(t, info.IsDir())
	assert.Equal(t, int64(4199202), info.Size)

	wantMtime := time.Unix(0, 1775095043590371464).UTC()
	wantBirth := time.Unix(0, 1775095043579017563).UTC()
	assert.True(t, info.Mtime.Equal(wantMtime), "mtime = %v, want %v", info.Mtime, wantMtime)
	assert.True(t, info.Birthtime.Equal(wantBirth), "birthtime = %v, want %v", info.Birthtime, wantBirth)
	assert.Equal(t, 2026, info.Mtime.Year())
	assert.Equal(t, time.April, info.Mtime.Month())
	assert.Equal(t, 2, info.Mtime.Day())

	assert.Zero(t, info.Mode, "the device sends no st_mode")
}
