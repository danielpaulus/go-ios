package main

import "testing"

func TestUsedDiskBytes(t *testing.T) {
	if got := usedDiskBytes(100, 40); got != 60 {
		t.Fatalf("usedDiskBytes(100, 40) = %d, want 60", got)
	}
}

func TestDiskspaceByteCount(t *testing.T) {
	if got := diskspaceByteCount(1500); got != "1.5kB" {
		t.Fatalf("diskspaceByteCount(1500) = %q, want %q", got, "1.5kB")
	}
}
