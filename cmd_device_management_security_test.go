package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFileWithPermissionsTightensExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-key.pem")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0o666))
	require.NoError(t, writeFileWithPermissions(path, []byte("new"), 0o600))

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, []byte("new"), content)
}
