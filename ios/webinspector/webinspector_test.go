package webinspector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseApplication(t *testing.T) {
	app, err := parseApplication(map[string]any{
		"WIRApplicationIdentifierKey":       "PID:123",
		"WIRApplicationBundleIdentifierKey": "com.apple.mobilesafari",
		"WIRApplicationNameKey":             "Safari",
		"WIRAutomationAvailabilityKey":      "WIRAutomationAvailabilityAvailable",
		"WIRIsApplicationActiveKey":         true,
		"WIRIsApplicationProxyKey":          false,
		"WIRIsApplicationReadyKey":          true,
	})
	if err != nil {
		t.Fatalf("parse application: %v", err)
	}
	if app.PID != 123 {
		t.Fatalf("expected pid 123, got %d", app.PID)
	}
	if app.BundleID != "com.apple.mobilesafari" {
		t.Fatalf("unexpected bundle id: %s", app.BundleID)
	}
}

func TestParseWebPage(t *testing.T) {
	page, err := parsePage("1", map[string]any{
		"WIRPageIdentifierKey": uint64(1),
		"WIRTypeKey":           "WIRTypeWebPage",
		"WIRTitleKey":          "Example",
		"WIRURLKey":            "https://example.test/",
	})
	if err != nil {
		t.Fatalf("parse page: %v", err)
	}
	if page.ID != 1 || page.Type != WIRTypeWebPage || page.Title != "Example" {
		t.Fatalf("unexpected page: %#v", page)
	}
}

func TestDecodeDispatchMessage(t *testing.T) {
	decoded, ok := decodeDispatchMessage(map[string]any{
		"method": "Target.dispatchMessageFromTarget",
		"params": map[string]any{
			"message": `{"id":7,"result":{"ok":true}}`,
		},
	})
	if !ok {
		t.Fatal("expected dispatch message to decode")
	}
	if id, _ := numericInt(decoded["id"]); id != 7 {
		t.Fatalf("expected id 7, got %d", id)
	}
}

func TestParseEvaluateResult(t *testing.T) {
	value, err := parseEvaluateResult(map[string]any{
		"result": map[string]any{
			"result": map[string]any{
				"type":  "string",
				"value": "hello",
			},
		},
	})
	if err != nil {
		t.Fatalf("parse evaluate result: %v", err)
	}
	if value != "hello" {
		t.Fatalf("expected hello, got %#v", value)
	}
}

func TestParseEvaluateResultObjectPreview(t *testing.T) {
	value, err := parseEvaluateResult(map[string]any{
		"result": map[string]any{
			"result": map[string]any{
				"type":      "object",
				"className": "Object",
				"preview": map[string]any{
					"properties": []any{
						map[string]any{"name": "answer", "value": "42", "type": "number"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("parse evaluate result: %v", err)
	}
	text, _ := value.(string)
	if !strings.Contains(text, "answer: 42") {
		t.Fatalf("expected object preview, got %#v", value)
	}
}

func TestAutomationSessionDisabledErrorIsActionable(t *testing.T) {
	client := &Client{state: AutomationNotAvailable}
	_, err := client.AutomationSession(context.Background(), Application{})
	if !errors.Is(err, ErrRemoteAutomationDisabled) {
		t.Fatalf("expected remote automation disabled error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Settings > Safari > Advanced > Remote Automation") {
		t.Fatalf("error should include enablement instructions: %v", err)
	}
}

func TestAutomationSessionTimeoutErrorIsActionable(t *testing.T) {
	client := &Client{
		apps:     make(map[string]Application),
		pages:    make(map[string]map[string]Page),
		errs:     make(chan error),
		disabled: make(chan error),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	_, err := client.waitForAutomationPage(ctx, "missing-session")
	if !errors.Is(err, ErrRemoteAutomationDisabled) {
		t.Fatalf("expected remote automation disabled error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Settings > Safari > Advanced > Remote Automation") {
		t.Fatalf("error should include enablement instructions: %v", err)
	}
}

func TestNormalizeStartupErrorMapsReadEOFToWebInspectorDisabled(t *testing.T) {
	err := normalizeStartupError(errors.New("Read: failed to read message length: EOF"))
	if !errors.Is(err, ErrWebInspectorDisabled) {
		t.Fatalf("expected Web Inspector disabled error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Settings > Safari > Advanced > Web Inspector") {
		t.Fatalf("error should include enablement instructions: %v", err)
	}
}

func TestNormalizeStartupErrorMapsShortWriteToWebInspectorDisabled(t *testing.T) {
	err := normalizeStartupError(errors.New("Write: only 0 bytes were written instead of 372"))
	if !errors.Is(err, ErrWebInspectorDisabled) {
		t.Fatalf("expected Web Inspector disabled error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Settings > Safari > Advanced > Web Inspector") {
		t.Fatalf("error should include enablement instructions: %v", err)
	}
}
