package zipconduit

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

func TestValidateZipEntryNames(t *testing.T) {
	decomposed := "cafe\u0301.txt"

	if err := validateZipEntryNames([]string{"README.txt", "Resources/Info.plist", decomposed}); err != nil {
		t.Fatalf("validateZipEntryNames() unexpected error: %v", err)
	}

	err := validateZipEntryNames([]string{"Frameworks/Test.framework/Resources/café.txt"})
	if err == nil || !strings.Contains(err.Error(), "Frameworks/Test.framework/Resources/café.txt") {
		t.Fatalf("validateZipEntryNames() error = %v, want offending relative path", err)
	}
}

func TestSendDirectoryRejectsNFCEntryBeforeInitTransfer(t *testing.T) {
	dir := t.TempDir()
	resourceDir := filepath.Join(dir, "Frameworks", "Test.framework", "Resources")
	if err := os.MkdirAll(resourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(resourceDir, "café.txt"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	var written bytes.Buffer
	conn := Connection{
		deviceConn: &testReadWriteCloser{Buffer: &written},
		plistCodec: ios.NewPlistCodecReadWriter(bytes.NewReader(nil), &written),
	}

	err := conn.sendDirectory(dir)
	if err == nil || !strings.Contains(err.Error(), "Frameworks/Test.framework/Resources/café.txt") {
		t.Fatalf("sendDirectory() error = %v, want offending relative path", err)
	}
	if written.Len() != 0 {
		t.Fatalf("sendDirectory() wrote %d bytes before rejecting the entry", written.Len())
	}
}

func TestSendIpaFileRejectsNFCEntryBeforeInitTransfer(t *testing.T) {
	ipaPath := filepath.Join(t.TempDir(), "app.ipa")
	file, err := os.Create(ipaPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("Payload/App.app/café.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("test")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	var written bytes.Buffer
	conn := Connection{
		deviceConn: &testReadWriteCloser{Buffer: &written},
		plistCodec: ios.NewPlistCodecReadWriter(bytes.NewReader(nil), &written),
	}

	err = conn.sendIpaFile(ipaPath)
	if err == nil || !strings.Contains(err.Error(), "Payload/App.app/café.txt") {
		t.Fatalf("sendIpaFile() error = %v, want offending archive path", err)
	}
	if written.Len() != 0 {
		t.Fatalf("sendIpaFile() wrote %d bytes before rejecting the entry", written.Len())
	}
}

type testReadWriteCloser struct {
	*bytes.Buffer
}

func (c *testReadWriteCloser) Close() error { return nil }
