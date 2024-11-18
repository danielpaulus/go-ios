//go:build !windows

package tunnel

import "io"

func setupWindowsTUN(tunnelInfo tunnelParameters) (io.ReadWriteCloser, error) {
	panic("this should never be called, it only exists so the build system can compile the code without errors")
}
