package imagemounter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validateBaseDirAndLookForImage must not panic when the directory doesn't
// exist; it should create it and return an empty path.
func TestValidateBaseDirAndLookForImageNonExistentPath(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "does", "not", "exist")

	assert.NotPanics(t, func() {
		path, err := validateBaseDirAndLookForImage(nonExistent, "image.dmg")
		require.NoError(t, err)
		assert.Empty(t, path)
	})

	// The directory should have been created.
	info, err := os.Stat(nonExistent)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// downloadFile must return an error when the server responds with a non-200
// status code instead of writing the error body to disk.
func TestDownloadFileNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "out.dmg")
	err := downloadFile(dest, server.URL)
	assert.Error(t, err)
}
