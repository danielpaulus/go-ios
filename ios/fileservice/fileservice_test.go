//go:build !fast

package fileservice_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/fileservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListDirectory tests listing files in an app's Documents directory
func TestListDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	device, err := ios.GetDevice("")
	require.NoError(t, err, "Failed to get device")

	if !device.SupportsRsd() {
		t.Skip("Device does not support RSD (iOS 17+)")
	}

	// Test with a known app - you'll need to replace this with an actual app bundle ID
	bundleID := "com.apple.Preferences" // System Preferences app
	conn, err := fileservice.New(device, fileservice.DomainAppDataContainer, bundleID)
	if err != nil {
		t.Skipf("Failed to create file service connection (app may not exist): %v", err)
	}
	defer conn.Close()

	// List root directory
	files, err := conn.ListDirectory(".")
	require.NoError(t, err, "Failed to list directory")

	t.Logf("Found %d files in root directory", len(files))
	for _, file := range files {
		t.Logf("  - %s", file)
	}
}

// TestListSystemCrashLogs tests listing system crash logs
func TestListSystemCrashLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	device, err := ios.GetDevice("")
	require.NoError(t, err, "Failed to get device")

	if !device.SupportsRsd() {
		t.Skip("Device does not support RSD (iOS 17+)")
	}

	conn, err := fileservice.New(device, fileservice.DomainSystemCrashLogs, "")
	if err != nil {
		t.Skipf("Failed to create file service connection: %v", err)
	}
	defer conn.Close()

	// List root directory of crash logs
	files, err := conn.ListDirectory(".")
	if err != nil {
		t.Logf("Failed to list crash logs (may be empty or access denied): %v", err)
		return
	}

	t.Logf("Found %d crash log files", len(files))
	for i, file := range files {
		if i < 10 { // Only print first 10
			t.Logf("  - %s", file)
		}
	}
}

// TestPullFile tests downloading a file from the device
func TestPullFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	device, err := ios.GetDevice("")
	require.NoError(t, err, "Failed to get device")

	if !device.SupportsRsd() {
		t.Skip("Device does not support RSD (iOS 17+)")
	}

	// This test requires a known file to exist on the device
	// You'll need to adjust this based on your test device
	t.Skip("Skipping PullFile test - requires known file on device")

	bundleID := "com.apple.Preferences"
	conn, err := fileservice.New(device, fileservice.DomainAppDataContainer, bundleID)
	require.NoError(t, err, "Failed to create file service connection")
	defer conn.Close()

	// Try to pull a file (adjust path as needed)
	var buf bytes.Buffer
	err = conn.PullFile("some/known/file.txt", &buf)
	require.NoError(t, err, "Failed to pull file")

	assert.Greater(t, buf.Len(), 0, "File data should not be empty")
	t.Logf("Downloaded file: %d bytes", buf.Len())
}

// TestPushFile tests uploading a file to the device
func TestPushFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	device, err := ios.GetDevice("")
	require.NoError(t, err, "Failed to get device")

	if !device.SupportsRsd() {
		t.Skip("Device does not support RSD (iOS 17+)")
	}

	// This test requires write access to an app's container
	t.Skip("Skipping PushFile test - requires app with file sharing enabled")

	bundleID := "com.example.testapp" // Replace with actual app
	conn, err := fileservice.New(device, fileservice.DomainAppDataContainer, bundleID)
	require.NoError(t, err, "Failed to create file service connection")
	defer conn.Close()

	// Create a test file
	testData := []byte("Hello from go-ios fileservice!")
	testFileName := "test_upload.txt"

	// Push the file (0o644 = rw-r--r--)
	err = conn.PushFile(testFileName, bytes.NewReader(testData), int64(len(testData)), 0o644, 501, 501)
	require.NoError(t, err, "Failed to push file")

	t.Logf("Successfully uploaded file: %s", testFileName)

	// Verify the file exists
	files, err := conn.ListDirectory(".")
	require.NoError(t, err, "Failed to list directory")

	found := false
	for _, file := range files {
		if file == testFileName {
			found = true
			break
		}
	}
	assert.True(t, found, "Uploaded file should appear in directory listing")
}

// TestDomainTypes tests creating connections with different domain types
func TestDomainTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	device, err := ios.GetDevice("")
	require.NoError(t, err, "Failed to get device")

	if !device.SupportsRsd() {
		t.Skip("Device does not support RSD (iOS 17+)")
	}

	domains := []struct {
		domain     fileservice.Domain
		identifier string
		name       string
	}{
		{fileservice.DomainAppDataContainer, "com.apple.Preferences", "App Data Container"},
		{fileservice.DomainTemporary, "", "Temporary"},
		{fileservice.DomainSystemCrashLogs, "", "System Crash Logs"},
	}

	for _, tc := range domains {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := fileservice.New(device, tc.domain, tc.identifier)
			if err != nil {
				t.Logf("Failed to create connection for %s: %v", tc.name, err)
				return
			}
			defer conn.Close()

			// Try to list root directory
			files, err := conn.ListDirectory(".")
			if err != nil {
				t.Logf("Failed to list directory for %s: %v", tc.name, err)
				return
			}

			t.Logf("%s: Found %d files", tc.name, len(files))
		})
	}
}

// Example usage documentation
func ExampleConnection_ListDirectory() {
	device, _ := ios.GetDevice("")

	// Create a connection to an app's Documents directory
	conn, _ := fileservice.New(device, fileservice.DomainAppDataContainer, "com.example.myapp")
	defer conn.Close()

	// List files in the root directory
	files, _ := conn.ListDirectory(".")
	for _, file := range files {
		fmt.Println(file)
	}
}

// Example of pulling a file
func ExampleConnection_PullFile() {
	device, _ := ios.GetDevice("")

	conn, _ := fileservice.New(device, fileservice.DomainAppDataContainer, "com.example.myapp")
	defer conn.Close()

	// Create output file for streaming
	outputFile, _ := os.Create(filepath.Join(".", "myfile.txt"))
	defer outputFile.Close()

	// Download file (streaming)
	_ = conn.PullFile("Documents/myfile.txt", outputFile)
}

// Example of pushing a file
func ExampleConnection_PushFile() {
	device, _ := ios.GetDevice("")

	conn, _ := fileservice.New(device, fileservice.DomainAppDataContainer, "com.example.myapp")
	defer conn.Close()

	// Open local file for streaming
	file, _ := os.Open("local_file.txt")
	defer file.Close()

	// Get file info for size and permissions
	fileInfo, _ := file.Stat()

	// Upload to device (streaming, preserves permissions)
	_ = conn.PushFile("Documents/uploaded.txt", file, fileInfo.Size(), int64(fileInfo.Mode().Perm()), 501, 501)
}
