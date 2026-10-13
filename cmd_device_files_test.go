package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullToLocalFileRemovesFileOnError(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "manifest.json")
	pullErr := errors.New("device error: File paths cannot contain '..'.")

	size, err := pullToLocalFile(func(remotePath string, writer io.Writer) error {
		return pullErr
	}, "shared-data/manifest.json", localPath)

	require.ErrorIs(t, err, pullErr)
	assert.Zero(t, size)
	_, statErr := os.Stat(localPath)
	assert.True(t, os.IsNotExist(statErr), "failed pull must not leave a local file behind, stat err: %v", statErr)
}

func TestPullToLocalFileRemovesPartialFileOnError(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "partial.bin")
	pullErr := errors.New("connection reset")

	size, err := pullToLocalFile(func(remotePath string, writer io.Writer) error {
		_, _ = writer.Write([]byte("partial data"))
		return pullErr
	}, "big.bin", localPath)

	require.ErrorIs(t, err, pullErr)
	assert.Zero(t, size)
	_, statErr := os.Stat(localPath)
	assert.True(t, os.IsNotExist(statErr), "failed pull must not leave a partial file behind, stat err: %v", statErr)
}

// TestPullToLocalFileRemovesFileWhenFinalizeFails proves the cleanup covers
// failures that surface after the pull itself succeeds (Stat/Close), where a
// full-disk write error would otherwise be reported. The pull callback closes
// the file, so the subsequent Stat fails and the wrapper must still remove the
// (potentially partial) file and return the error.
func TestPullToLocalFileRemovesFileWhenFinalizeFails(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "finalize.bin")

	size, err := pullToLocalFile(func(remotePath string, writer io.Writer) error {
		f, ok := writer.(*os.File)
		require.True(t, ok, "expected the wrapper to pass an *os.File")
		_, _ = f.Write([]byte("data that must not survive a finalize failure"))
		return f.Close() // pull "succeeds" but leaves the descriptor closed
	}, "big.bin", localPath)

	// The pull reported success, but finalizing (Stat on the now-closed
	// descriptor) fails, so the wrapper must surface the error and clean up.
	require.Error(t, err)
	assert.Zero(t, size)
	_, statErr := os.Stat(localPath)
	assert.True(t, os.IsNotExist(statErr), "a post-pull finalize failure must not leave a file behind, stat err: %v", statErr)
}

func TestPullToLocalFileWritesFileOnSuccess(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "manifest.json")
	content := []byte(`{"hello":"world"}`)

	size, err := pullToLocalFile(func(remotePath string, writer io.Writer) error {
		_, err := writer.Write(content)
		return err
	}, "shared-data/manifest.json", localPath)

	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
	written, err := os.ReadFile(localPath)
	require.NoError(t, err)
	assert.Equal(t, content, written)
}
