//go:build integration
// +build integration

package accessibility_test

import (
	"context"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
)

type noopCallbacks struct{}

func (noopCallbacks) HostAppStateChanged(accessibility.Notification)              {}
func (noopCallbacks) HostInspectorNotificationReceived(accessibility.Notification) {}

func TestMove(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	conn, err := accessibility.New(device, ctx, noopCallbacks{})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.TurnOff()

	conn.SwitchToDevice()
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
			conn.Move(direction)
			element, err := conn.AwaitElementChanged(ctx)
			if err != nil {
				t.Logf("Move %v failed (expected on some devices): %v", direction, err)
				continue
			}

			t.Logf("Move %v succeeded: %+v", direction, element)
		}
	})
}
