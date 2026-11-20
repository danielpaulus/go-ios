package ios

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestUnzipDoesNotFollowPreexistingSymlinks(t *testing.T) {
	base := t.TempDir()
	destination := filepath.Join(base, "destination")
	outside := filepath.Join(base, "outside")
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(destination, "sub")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	victim := filepath.Join(outside, "victim")
	if err := os.WriteFile(victim, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(victim, filepath.Join(destination, "file")); err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(base, "archive.zip")
	writeTestZip(t, archive, map[string]string{
		"sub/new": "inside",
		"file":    "replacement",
	})

	if _, _, err := Unzip(archive, destination); err != nil {
		t.Fatalf("Unzip: %v", err)
	}
	if contents, err := os.ReadFile(filepath.Join(outside, "victim")); err != nil || string(contents) != "unchanged" {
		t.Fatalf("outside victim changed: contents=%q err=%v", contents, err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new")); !os.IsNotExist(err) {
		t.Fatalf("archive wrote outside destination: %v", err)
	}
	if contents, err := os.ReadFile(filepath.Join(destination, "sub", "new")); err != nil || string(contents) != "inside" {
		t.Fatalf("staged descendant = %q, %v", contents, err)
	}
	if contents, err := os.ReadFile(filepath.Join(destination, "file")); err != nil || string(contents) != "replacement" {
		t.Fatalf("staged file = %q, %v", contents, err)
	}
}

func TestUnzipRejectsTraversal(t *testing.T) {
	base := t.TempDir()
	archive := filepath.Join(base, "archive.zip")
	writeTestZip(t, archive, map[string]string{"../outside": "bad"})

	if _, _, err := Unzip(archive, filepath.Join(base, "destination")); err == nil {
		t.Fatal("Unzip accepted a traversal entry")
	}
	if _, err := os.Stat(filepath.Join(base, "outside")); !os.IsNotExist(err) {
		t.Fatalf("traversal entry was created: %v", err)
	}
}

func writeTestZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, contents := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
