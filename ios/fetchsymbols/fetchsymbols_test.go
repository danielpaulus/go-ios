package fetchsymbols

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestEnvelopeRoundtrip encodes the DSCFilePaths request the way ListFiles sends it
// and verifies it decodes back to the same envelope, side-channel UUID included.
func TestRequestEnvelopeRoundtrip(t *testing.T) {
	sideChannel := uuid.New()
	body := map[string]interface{}{
		"XPCDictionary_sideChannel": sideChannel,
		"DSCFilePaths":              []interface{}{},
	}

	buf := bytes.NewBuffer(nil)
	err := xpc.EncodeMessage(buf, xpc.Message{
		Flags: xpc.AlwaysSetFlag | xpc.DataFlag | xpc.HeartbeatRequestFlag,
		Body:  body,
		Id:    1,
	})
	require.NoError(t, err)

	decoded, err := xpc.DecodeMessage(buf)
	require.NoError(t, err)
	assert.Equal(t, xpc.AlwaysSetFlag|xpc.DataFlag|xpc.HeartbeatRequestFlag, decoded.Flags)
	assert.Equal(t, body, decoded.Body)
}

// TestFileCountResponse decodes a fixture of the first reply (file count) and parses it.
func TestFileCountResponse(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	err := xpc.EncodeMessage(buf, xpc.Message{
		Flags: xpc.AlwaysSetFlag | xpc.DataFlag | xpc.HeartbeatReplyFlag,
		Body:  map[string]interface{}{"DSCFilePaths": uint64(2)},
	})
	require.NoError(t, err)

	decoded, err := xpc.DecodeMessage(buf)
	require.NoError(t, err)
	count, err := parseFileCount(decoded.Body)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestParseFileCount(t *testing.T) {
	tests := []struct {
		name    string
		resp    map[string]interface{}
		want    int
		wantErr bool
	}{
		{name: "uint64 count", resp: map[string]interface{}{"DSCFilePaths": uint64(3)}, want: 3},
		{name: "int64 count", resp: map[string]interface{}{"DSCFilePaths": int64(1)}, want: 1},
		{name: "zero files", resp: map[string]interface{}{"DSCFilePaths": uint64(0)}, want: 0},
		{name: "missing key", resp: map[string]interface{}{}, wantErr: true},
		{name: "wrong type", resp: map[string]interface{}{"DSCFilePaths": "3"}, wantErr: true},
		{name: "negative count", resp: map[string]interface{}{"DSCFilePaths": int64(-1)}, wantErr: true},
		{name: "implausible count", resp: map[string]interface{}{"DSCFilePaths": uint64(maxFileCount + 1)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFileCount(tt.resp)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestFileMetadataResponse decodes a fixture of a per-file metadata message, including the
// XPC file transfer object carrying the expected length, and parses it.
func TestFileMetadataResponse(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	err := xpc.EncodeMessage(buf, xpc.Message{
		Flags: xpc.AlwaysSetFlag | xpc.DataFlag,
		Body: map[string]interface{}{
			"DSCFilePaths": map[string]interface{}{
				"filePath":     "/System/Library/Caches/com.apple.dyld/dyld_shared_cache_arm64e",
				"fileTransfer": xpc.FileTransfer{MsgId: 1, TransferSize: 4096},
			},
		},
	})
	require.NoError(t, err)

	decoded, err := xpc.DecodeMessage(buf)
	require.NoError(t, err)
	info, err := parseFileInfo(decoded.Body)
	require.NoError(t, err)
	assert.Equal(t, FileInfo{
		FilePath: "/System/Library/Caches/com.apple.dyld/dyld_shared_cache_arm64e",
		Size:     4096,
	}, info)
}

func TestParseFileInfoErrors(t *testing.T) {
	tests := []struct {
		name string
		resp map[string]interface{}
	}{
		{name: "missing entry", resp: map[string]interface{}{}},
		{name: "entry not a dict", resp: map[string]interface{}{"DSCFilePaths": uint64(1)}},
		{name: "missing file path", resp: map[string]interface{}{"DSCFilePaths": map[string]interface{}{
			"fileTransfer": xpc.FileTransfer{MsgId: 1, TransferSize: 1},
		}}},
		{name: "empty file path", resp: map[string]interface{}{"DSCFilePaths": map[string]interface{}{
			"filePath":     "",
			"fileTransfer": xpc.FileTransfer{MsgId: 1, TransferSize: 1},
		}}},
		{name: "missing file transfer", resp: map[string]interface{}{"DSCFilePaths": map[string]interface{}{
			"filePath": "/a",
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFileInfo(tt.resp)
			assert.Error(t, err)
		})
	}
}

// chunkReader returns at most chunkSize bytes per Read call to simulate chunked arrival.
type chunkReader struct {
	data      []byte
	chunkSize int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if len(c.data) == 0 {
		return 0, io.EOF
	}
	n := c.chunkSize
	if n > len(c.data) {
		n = len(c.data)
	}
	if n > len(p) {
		n = len(p)
	}
	copy(p, c.data[:n])
	c.data = c.data[n:]
	return n, nil
}

func TestCopyFileChunksReassemblesChunks(t *testing.T) {
	payload := bytes.Repeat([]byte("0123456789abcdef"), 1024)
	out := bytes.NewBuffer(nil)
	err := copyFileChunks(out, &chunkReader{data: payload, chunkSize: 1000}, uint64(len(payload)))
	require.NoError(t, err)
	assert.Equal(t, payload, out.Bytes())
}

func TestCopyFileChunksShortStream(t *testing.T) {
	payload := []byte("not enough data")
	err := copyFileChunks(io.Discard, bytes.NewReader(payload), uint64(len(payload)+1))
	assert.ErrorContains(t, err, "ended stream")
}

func TestCopyFileChunksPropagatesReadError(t *testing.T) {
	readErr := errors.New("stream broken")
	r := io.MultiReader(bytes.NewReader([]byte("abc")), errReader{err: readErr})
	err := copyFileChunks(io.Discard, r, 10)
	assert.ErrorIs(t, err, readErr)
}

type errReader struct{ err error }

func (e errReader) Read([]byte) (int, error) { return 0, e.err }

func TestFileStreamId(t *testing.T) {
	assert.Equal(t, uint32(2), fileStreamId(0))
	assert.Equal(t, uint32(4), fileStreamId(1))
	assert.Equal(t, uint32(6), fileStreamId(2))
}

func TestCachePath(t *testing.T) {
	base := filepath.Join("cache", "17.5_21F79")
	tests := []struct {
		name       string
		devicePath string
		want       string
		wantErr    bool
	}{
		{
			name:       "device path keeps layout",
			devicePath: "/System/Library/Caches/com.apple.dyld/dyld_shared_cache_arm64e",
			want:       filepath.Join(base, "System", "Library", "Caches", "com.apple.dyld", "dyld_shared_cache_arm64e"),
		},
		{
			name:       "relative path",
			devicePath: "some/file",
			want:       filepath.Join(base, "some", "file"),
		},
		{
			name:       "traversal is neutralized",
			devicePath: "/../../etc/passwd",
			want:       filepath.Join(base, "etc", "passwd"),
		},
		{
			name:       "backslash traversal is neutralized",
			devicePath: `..\..\etc\passwd`,
			want:       filepath.Join(base, "etc", "passwd"),
		},
		{name: "empty path", devicePath: "", wantErr: true},
		{name: "root only", devicePath: "/", wantErr: true},
		{name: "dots only", devicePath: "/../..", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CachePath(base, tt.devicePath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsCached(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "cachefile")
	assert.False(t, IsCached(dest, 4), "missing file must not count as cached")

	require.NoError(t, os.WriteFile(dest, []byte("data"), 0o644))
	assert.True(t, IsCached(dest, 4))
	assert.False(t, IsCached(dest, 5), "size mismatch must not count as cached")
	assert.False(t, IsCached(dir, 4), "directory must not count as cached")
}

func TestWriteFileAtomically(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")

	err := writeFileAtomically(dest, func(w io.Writer) error {
		_, err := w.Write([]byte("content"))
		return err
	})
	require.NoError(t, err)
	content, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, []byte("content"), content)
	_, err = os.Stat(dest + ".download")
	assert.True(t, os.IsNotExist(err), "temp file must be gone after success")
}

func TestWriteFileAtomicallyFailureLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")
	writeErr := errors.New("transfer interrupted")

	err := writeFileAtomically(dest, func(w io.Writer) error {
		_, _ = w.Write([]byte("partial"))
		return writeErr
	})
	assert.ErrorIs(t, err, writeErr)
	_, statErr := os.Stat(dest)
	assert.True(t, os.IsNotExist(statErr), "failed download must not leave a destination file")
	_, statErr = os.Stat(dest + ".download")
	assert.True(t, os.IsNotExist(statErr), "failed download must not leave a temp file")
}

func TestProgressWriterReportsRunningTotal(t *testing.T) {
	var reported []uint64
	w := &progressWriter{w: io.Discard, progress: func(written uint64) { reported = append(reported, written) }}
	_, err := w.Write([]byte("abc"))
	require.NoError(t, err)
	_, err = w.Write([]byte("de"))
	require.NoError(t, err)
	assert.Equal(t, []uint64{3, 5}, reported)
}
