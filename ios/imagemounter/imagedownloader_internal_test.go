package imagemounter

import (
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// offlineServer returns a test server that fails the test on any request and
// wires it up as the download source for both the iOS 17+ DDI and the iOS <17
// signed images, so a test asserts that image resolution never goes online.
func offlineServer(t *testing.T, versions ...string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected network request for %s, the image should have been resolved offline", r.URL)
		http.Error(w, "offline", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	origDevicebox := devicebox
	devicebox = server.URL + "/"
	t.Cleanup(func() { devicebox = origDevicebox })

	for _, version := range versions {
		version := version
		origUrl := versionMap[version]
		versionMap[version] = server.URL
		t.Cleanup(func() { versionMap[version] = origUrl })
	}
}

func writeDDI(t *testing.T, baseDir string, name string) string {
	t.Helper()
	restoreDir := filepath.Join(baseDir, name, "Restore")
	require.NoError(t, os.MkdirAll(restoreDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(restoreDir, buildManifestFile), []byte("manifest"), 0o644))
	return restoreDir
}

func writeSignedImage(t *testing.T, baseDir string, versionDir string) string {
	t.Helper()
	dir := filepath.Join(baseDir, versionDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, imageFile), []byte("dmg"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, signatureFile), []byte("sig"), 0o644))
	return filepath.Join(dir, imageFile)
}

// A personalized DDI (iOS 17+) that is already present in the basedir must be
// used without any network access, even when it was stored under a different
// name than the currently pinned one (e.g. downloaded by an older go-ios).
func TestDownload17PlusOfflineWithPresentDDI(t *testing.T) {
	offlineServer(t)
	baseDir := t.TempDir()
	restoreDir := writeDDI(t, baseDir, "ddi-15F31d")

	got, err := Download17Plus(baseDir, ios.IOS17())
	require.NoError(t, err)
	assert.Equal(t, restoreDir, got)
}

// When several personalized DDIs are present, the pinned one wins.
func TestDownload17PlusPrefersPinnedDDI(t *testing.T) {
	offlineServer(t)
	baseDir := t.TempDir()
	writeDDI(t, baseDir, "ddi-15F31d")
	pinnedRestore := writeDDI(t, baseDir, xcode15_4_ddi)

	got, err := Download17Plus(baseDir, ios.IOS17())
	require.NoError(t, err)
	assert.Equal(t, pinnedRestore, got)
}

// The pinned-DDI preference must resolve to the directory named exactly like
// the pinned DDI, even when a decoy directory whose name merely *contains* the
// pinned string sorts before it. A substring match on the whole path (or a
// plain lexical first-wins) would incorrectly pick the decoy.
func TestDownload17PlusPrefersExactPinnedDDIName(t *testing.T) {
	offlineServer(t)
	baseDir := t.TempDir()
	// Decoy: structurally valid, name contains the pinned string, sorts first.
	writeDDI(t, baseDir, "aaaa-"+xcode15_4_ddi+"-copy")
	// The real pinned image (exact name, sorts after the decoy).
	pinnedRestore := writeDDI(t, baseDir, xcode15_4_ddi)

	got, err := Download17Plus(baseDir, ios.IOS17())
	require.NoError(t, err)
	assert.Equal(t, pinnedRestore, got)
}

// The pinned-DDI preference must key off the directory that holds 'Restore',
// not the whole path: an unrelated ancestor whose name contains the pinned
// string must not make a differently-named image win over an exact match.
func TestDownload17PlusIgnoresPinnedStringInAncestor(t *testing.T) {
	offlineServer(t)
	baseDir := t.TempDir()
	// A valid image under an ancestor path that contains the pinned string but
	// whose own directory is not the pinned DDI.
	decoy := writeDDI(t, filepath.Join(baseDir, xcode15_4_ddi+"-backup"), "someimage")
	got, err := Download17Plus(baseDir, ios.IOS17())
	require.NoError(t, err)
	// With only the decoy present it is still used (structural detection), but
	// the ancestor's pinned substring must not be treated as an exact/pinned
	// match that would outrank a real pinned image.
	assert.Equal(t, decoy, got)
	// Now add the real pinned image; it must win over the ancestor decoy.
	pinnedRestore := writeDDI(t, baseDir, xcode15_4_ddi)
	got, err = Download17Plus(baseDir, ios.IOS17())
	require.NoError(t, err)
	assert.Equal(t, pinnedRestore, got)
}

// A pre-downloaded DDI zip archive in the basedir must be extracted and used
// without any network access.
func TestDownload17PlusOfflineWithPresentArchive(t *testing.T) {
	offlineServer(t)
	baseDir := t.TempDir()

	zipFile, err := os.Create(filepath.Join(baseDir, xcode15_4_ddi+".zip"))
	require.NoError(t, err)
	zipWriter := zip.NewWriter(zipFile)
	manifest, err := zipWriter.Create("Restore/" + buildManifestFile)
	require.NoError(t, err)
	_, err = manifest.Write([]byte("manifest"))
	require.NoError(t, err)
	require.NoError(t, zipWriter.Close())
	require.NoError(t, zipFile.Close())

	got, err := Download17Plus(baseDir, ios.IOS17())
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(got, buildManifestFile))
}

// An iOS <17 image that go-ios itself downloaded earlier lives in the short
// '<version>' directory ('12.3'), while the catalog name is '12.3 (16F148)'.
// Resolving it again must find the present image instead of going online.
func TestDownloadImageForVersionOfflineShortDirLayout(t *testing.T) {
	offlineServer(t, "12.3 (16F148)")
	baseDir := t.TempDir()
	imagePath := writeSignedImage(t, baseDir, "12.3")

	got, err := downloadImageForVersion(baseDir, "12.3", "")
	require.NoError(t, err)
	assert.Equal(t, imagePath, got)
}

// Manually downloaded images mirroring the upstream repository layout use the
// full '<version> (<build>)' directory name and must be found offline as well.
func TestDownloadImageForVersionOfflineFullNameLayout(t *testing.T) {
	offlineServer(t, "12.3 (16F148)")
	baseDir := t.TempDir()
	imagePath := writeSignedImage(t, baseDir, "12.3 (16F148)")

	got, err := downloadImageForVersion(baseDir, "12.3", "")
	require.NoError(t, err)
	assert.Equal(t, imagePath, got)
}

// safeJoin must reject path elements that would escape the base directory.
func TestSafeJoinRejectsEscapingElements(t *testing.T) {
	baseDir := t.TempDir()

	joined, err := safeJoin(baseDir, "12.3", imageFile)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(baseDir, "12.3", imageFile), joined)

	for _, malicious := range []string{"../../etc", "..", "12.3/../../etc", "/etc/passwd"} {
		_, err := safeJoin(baseDir, malicious)
		assert.Error(t, err, "element %q must be rejected", malicious)
	}
}

// A malicious product version string must be rejected instead of being used to
// build filesystem paths.
func TestDownloadImageForVersionRejectsMaliciousVersion(t *testing.T) {
	offlineServer(t)
	baseDir := t.TempDir()

	_, err := downloadImageForVersion(baseDir, "../../etc", "")
	assert.Error(t, err)
	assert.NoDirExists(t, filepath.Join(baseDir, "..", "etc"))
}

// A present DDI archive with a zip-slip entry must not write outside the
// basedir; extraction fails and, without a reachable download source, the
// resolution errors out.
func TestDownload17PlusRejectsZipSlipArchive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "offline", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	origDevicebox := devicebox
	devicebox = server.URL + "/"
	t.Cleanup(func() { devicebox = origDevicebox })

	parentDir := t.TempDir()
	baseDir := filepath.Join(parentDir, "devimages")
	require.NoError(t, os.MkdirAll(baseDir, 0o755))

	zipFile, err := os.Create(filepath.Join(baseDir, xcode15_4_ddi+".zip"))
	require.NoError(t, err)
	zipWriter := zip.NewWriter(zipFile)
	evil, err := zipWriter.Create("../evil.txt")
	require.NoError(t, err)
	_, err = evil.Write([]byte("evil"))
	require.NoError(t, err)
	require.NoError(t, zipWriter.Close())
	require.NoError(t, zipFile.Close())

	_, err = Download17Plus(baseDir, ios.IOS17())
	assert.Error(t, err)
	assert.NoFileExists(t, filepath.Join(baseDir, "evil.txt"))
	assert.NoFileExists(t, filepath.Join(parentDir, "evil.txt"))
	assert.NoFileExists(t, filepath.Join(baseDir, xcode15_4_ddi, "evil.txt"))
}

// On a real cache miss the image and its signature are downloaded into the
// short '<version>' directory, which the offline lookup finds on the next run.
func TestDownloadImageForVersionDownloadsOnCacheMiss(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("fake-content"))
		assert.NoError(t, err)
	}))
	defer server.Close()
	origUrl := versionMap["12.3 (16F148)"]
	versionMap["12.3 (16F148)"] = server.URL
	t.Cleanup(func() { versionMap["12.3 (16F148)"] = origUrl })

	baseDir := t.TempDir()
	got, err := downloadImageForVersion(baseDir, "12.3", "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(baseDir, "12.3", imageFile), filepath.Clean(got))
	assert.FileExists(t, filepath.Join(baseDir, "12.3", imageFile))
	assert.FileExists(t, filepath.Join(baseDir, "12.3", signatureFile))

	// The next resolution must use the downloaded image without any network access.
	offlineServer(t, "12.3 (16F148)")
	cached, err := downloadImageForVersion(baseDir, "12.3", "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(baseDir, "12.3", imageFile), filepath.Clean(cached))
}

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
