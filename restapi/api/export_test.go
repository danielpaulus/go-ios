package api

import (
	"context"
	"io"
	"net/http"
)

// WriteMJPEGFramesForTest exposes the unexported multipart framing core so the
// external api_test package can exercise it device-free with an injected frame
// channel.
func WriteMJPEGFramesForTest(ctx context.Context, w io.Writer, flusher http.Flusher, frames <-chan []byte) error {
	return writeMJPEGFrames(ctx, w, flusher, frames)
}
