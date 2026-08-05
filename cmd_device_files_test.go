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
