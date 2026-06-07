package webinspector

import "testing"

func TestLocalCDPResponse(t *testing.T) {
	handled, response, extra := localCDPResponse(map[string]any{"id": 4, "method": "CSS.takeComputedStyleUpdates"}, "target-1", "session-1", Page{})
	if !handled {
		t.Fatal("expected CSS.takeComputedStyleUpdates to be handled locally")
	}
	result := response["result"].(map[string]any)
	nodeIDs := result["nodeIds"].([]int)
	if len(nodeIDs) != 0 {
		t.Fatalf("unexpected node id result: %#v", response)
	}
	if len(extra) != 0 {
		t.Fatalf("unexpected extra events: %#v", extra)
	}
}

func TestLocalCDPResponseDoesNotSwallowStatefulCommands(t *testing.T) {
	statefulMethods := []string{
		"Debugger.setBlackboxPatterns",
		"Emulation.setAutoDarkModeOverride",
		"Page.getNavigationHistory",
		"Runtime.compileScript",
		"Runtime.getIsolateId",
	}
	for _, method := range statefulMethods {
		handled, _, _ := localCDPResponse(map[string]any{"id": 5, "method": method}, "target-1", "session-1", Page{})
		if handled {
			t.Fatalf("%s should be routed through the stateful CDP session", method)
		}
	}
}

func TestTranslateCDPCommand(t *testing.T) {
	message := translateCDPCommand(map[string]any{
		"id":     1,
		"method": "Network.setCacheDisabled",
		"params": map[string]any{"cacheDisabled": true},
	})
	if message["method"] != "Network.setResourceCachingDisabled" {
		t.Fatalf("unexpected method: %#v", message["method"])
	}
	params := message["params"].(map[string]any)
	if params["disabled"] != true {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestTranslateDebuggerBreakpointCondition(t *testing.T) {
	message := translateCDPCommand(map[string]any{
		"id":     2,
		"method": "Debugger.setBreakpointByUrl",
		"params": map[string]any{"condition": "x > 1"},
	})
	params := message["params"].(map[string]any)
	options := params["options"].(map[string]any)
	if options["condition"] != "x > 1" {
		t.Fatalf("unexpected options: %#v", options)
	}
	if _, ok := params["condition"]; ok {
		t.Fatalf("condition should be moved into options: %#v", params)
	}
}

func TestTranslateEmulationAutoDarkMode(t *testing.T) {
	message := translateCDPCommand(map[string]any{
		"id":     3,
		"method": "Emulation.setAutoDarkModeOverride",
		"params": map[string]any{"enabled": true},
	})
	if message["method"] != "Page.setForcedAppearance" {
		t.Fatalf("unexpected method: %#v", message["method"])
	}
	params := message["params"].(map[string]any)
	if params["appearance"] != "Dark" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestRuntimeGetIsolateIDUsesDefaultExecutionContext(t *testing.T) {
	session := &cdpPageSession{defaultExecutionID: 42}
	response := session.runtimeGetIsolateID(7)
	result := response["result"].(map[string]any)
	if result["id"] != 42 {
		t.Fatalf("unexpected isolate id result: %#v", response)
	}
}

func TestNormalizeConsoleEvent(t *testing.T) {
	normalized, drop := normalizeCDPEvent(map[string]any{
		"method": "Console.messageAdded",
		"params": map[string]any{
			"message": map[string]any{
				"source": "console-api",
				"level":  "debug",
				"text":   "hello",
			},
		},
	})
	if drop {
		t.Fatal("expected console event to be emitted")
	}
	if normalized["method"] != "Log.entryAdded" {
		t.Fatalf("unexpected normalized method: %#v", normalized)
	}
}

func TestNormalizeDebuggerPaused(t *testing.T) {
	normalized, drop := normalizeCDPEvent(map[string]any{
		"method": "Debugger.paused",
		"params": map[string]any{
			"reason": "Listener",
			"data":   map[string]any{"breakpointId": "bp-1"},
		},
	})
	if drop {
		t.Fatal("expected debugger paused event to be emitted")
	}
	params := normalized["params"].(map[string]any)
	if params["reason"] != "EventListener" {
		t.Fatalf("unexpected reason: %#v", params["reason"])
	}
	breakpoints := params["hitBreakpoints"].([]string)
	if breakpoints[0] != "bp-1" {
		t.Fatalf("unexpected hit breakpoints: %#v", breakpoints)
	}
}
