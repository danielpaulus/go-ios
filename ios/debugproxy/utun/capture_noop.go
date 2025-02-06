//go:build !darwin

package utun

import (
	"context"
	"errors"

	"github.com/danielpaulus/go-ios/ios"
)

func Live(ctx context.Context, iface string, provider ios.RsdPortProvider, dumpDir string) error {
	return errors.New("capturing traffic on the utun interface is only supported on MacOS")
}
