package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteVerifiedUIArtifact(t *testing.T) {
	payload := "trusted artifact"
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))
	target := filepath.Join(t.TempDir(), "artifact.zip")

	digest, err := writeVerifiedUIArtifact(strings.NewReader(payload), target, expected)
	if err != nil {
		t.Fatalf("writeVerifiedUIArtifact failed: %v", err)
	}
	if digest != expected {
		t.Fatalf("digest = %q, want %q", digest, expected)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read verified artifact: %v", err)
	}
	if string(content) != payload {
		t.Fatalf("content = %q, want %q", content, payload)
	}
}

func TestWriteVerifiedUIArtifactRejectsMismatch(t *testing.T) {
	target := filepath.Join(t.TempDir(), "artifact.ipa")
	if err := os.WriteFile(target, []byte("previous verified artifact"), 0o600); err != nil {
		t.Fatalf("seed existing artifact: %v", err)
	}
	_, err := writeVerifiedUIArtifact(strings.NewReader("malicious"), target, strings.Repeat("0", sha256.Size*2))
	if err == nil {
		t.Fatal("expected digest mismatch")
	}
	content, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read existing artifact: %v", readErr)
	}
	if string(content) != "previous verified artifact" {
		t.Fatalf("digest mismatch replaced existing artifact with %q", content)
	}
}
