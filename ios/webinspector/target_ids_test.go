package webinspector

import "testing"

// Page keys are numbered per application, so Safari and an app WebView can
// both have page "1". CDP target ids must tell them apart.
func newTwoAppClient() *Client {
	return &Client{
		apps: map[string]Application{
			"PID:10": {ID: "PID:10", BundleID: "com.apple.mobilesafari", PID: 10},
			"PID:20": {ID: "PID:20", BundleID: "com.example.app", PID: 20},
		},
		pages: map[string]map[string]Page{
			"PID:10": {"1": {ID: 1, Key: "1", Type: WIRTypeWebPage, Title: "Safari page"}},
			"PID:20": {"1": {ID: 1, Key: "1", Type: WIRTypeWebPage, Title: "App page"}},
		},
	}
}

func TestCDPTargetIDIsUniqueAcrossApplications(t *testing.T) {
	c := newTwoAppClient()
	safari := cdpTargetID(c.apps["PID:10"], c.pages["PID:10"]["1"])
	app := cdpTargetID(c.apps["PID:20"], c.pages["PID:20"]["1"])
	if safari == app {
		t.Fatalf("both pages got target id %q", safari)
	}
}

func TestFindPageByTarget(t *testing.T) {
	c := newTwoAppClient()
	for _, want := range []string{"PID:10", "PID:20"} {
		target := cdpTargetID(c.apps[want], c.pages[want]["1"])
		app, page, ok := c.FindPageByTarget(target)
		if !ok || app.ID != want || page.Key != "1" {
			t.Fatalf("FindPageByTarget(%q) = %v, %v, %v; want app %s", target, app.ID, page.Key, ok, want)
		}
	}
	if _, _, ok := c.FindPageByTarget("1"); !ok {
		t.Fatalf("a bare page key must still resolve")
	}
	for _, target := range []string{"30.1", "x.1", "10.9"} {
		if _, _, ok := c.FindPageByTarget(target); ok {
			t.Fatalf("FindPageByTarget(%q) resolved, want not found", target)
		}
	}
}
