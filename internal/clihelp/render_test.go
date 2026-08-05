package clihelp

import (
	"strings"
	"testing"
)

func TestRender_Global(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("load help catalog: %v", err)
	}
	out, err := c.Render("test-version", "")
	if err != nil {
		t.Fatalf("render global: %v", err)
	}
	if !strings.Contains(out, "go-ios test-version") {
		t.Fatalf("missing header: %q", out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Fatalf("missing commands section: %q", out)
	}
}

func TestRender_Command(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("load help catalog: %v", err)
	}
	out, err := c.Render("test-version", "apps")
	if err != nil {
		t.Fatalf("render command: %v", err)
	}
	if !strings.Contains(out, "Command: apps") {
		t.Fatalf("missing command title: %q", out)
	}
	if !strings.Contains(out, "ios apps") {
		t.Fatalf("missing command usage: %q", out)
	}
}

func TestRender_CommandExamples(t *testing.T) {
    c, err := Load()
    if err != nil {
        t.Fatalf("load help catalog: %v", err)
    }
    out, err := c.Render("test-version", "pcap")
    if err != nil {
        t.Fatalf("render pcap: %v", err)
    }
    if !strings.Contains(out, "Examples:") {
        t.Fatalf("missing examples section: %q", out)
    }
    if !strings.Contains(out, "capture.pcap") {
        t.Fatalf("missing pcap example: %q", out)
    }
}
