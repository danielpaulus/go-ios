//go:build integration
// +build integration

package accessibility_test

import (
	"context"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
)

func TestMove(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := accessibility.New(device)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.TurnOff()

	conn.SwitchToDevice()
	if err != nil {
		t.Fatal(err)
	}
	conn.EnableSelectionMode()

	t.Run("Test Move directions", func(t *testing.T) {
		directions := []accessibility.MoveDirection{
			// newer ios(18+) devices sometimes doesn't bring focus to first element, so we need to move twice
			accessibility.DirectionNext,
			accessibility.DirectionNext,
			accessibility.DirectionPrevious,
		}

		for _, direction := range directions {
			t.Logf("Testing direction: %v", direction)
			element, err := conn.Move(context.Background(), direction)
			if err != nil {
				t.Logf("Move %v failed (expected on some devices): %v", direction, err)
				continue
			}

			t.Logf("Move %v succeeded: %+v", direction, element)
		}
	})
}

// to execute this unit test, you can use `go test -tags integration ./ios/accessibility/...` with a device connected
func TestResetAccessibilitySettings(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test ResetToDefaultAccessibilitySettings", func(t *testing.T) {
		// create connection for testing
		conn, err := accessibility.NewWithoutEventChangeListeners(device)
		if err != nil {
			t.Fatal(err)
		}

		err = conn.ResetToDefaultAccessibilitySettings()
		if err != nil {
			t.Fatalf("ResetToDefaultAccessibilitySettings failed: %v", err)
		}

		t.Log("Successfully reset accessibility settings to defaults")
	})
}
