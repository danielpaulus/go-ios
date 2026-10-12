package ios_test

import (
	"context"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
)

// TestBrowseRemotedTimesOutCleanly is a smoke test that does not depend on any
// device being present: with a short timeout BrowseRemoted must return promptly
// (within the timeout plus a small margin) without error, yielding whatever it
// happened to find (possibly nothing).
func TestBrowseRemotedTimesOutCleanly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	var err error
	go func() {
		_, err = ios.BrowseRemoted(ctx)
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatalf("BrowseRemoted returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("BrowseRemoted did not return within the timeout")
	}
}
