package main
//taken from
// https://github.com/calebdoxsey/tutorials/blob/master/talks/2018-01-30--extending-gopherjs/example-08/client/io.go
import (
	"context"
	"io"
	"time"
)

// A ChannelReader reads bytes from a channel and buffers them
type ChannelReader struct {
	c        <-chan []byte
	buf      []byte
	deadline time.Time
}

// NewChannelReader creates a new ChannelReader
func NewChannelReader(c <-chan []byte) *ChannelReader {
	return &ChannelReader{
		c: c,
	}
}

// Read reads from the channel. It should not be called by multiple goroutines
func (r *ChannelReader) Read(b []byte) (sz int, err error) {
	if len(b) == 0 {
		return 0, io.ErrShortBuffer
	}

	for {
		if len(r.buf) > 0 {
			if len(r.buf) <= len(b) {
				sz = len(r.buf)
				copy(b, r.buf)
				r.buf = nil
			} else {
				copy(b, r.buf)
				r.buf = r.buf[len(b):]
				sz = len(b)
			}
			return sz, nil
		}

		if r.deadline.IsZero() {
			r.buf = <-r.c
		} else {
			timer := time.NewTimer(r.deadline.Sub(time.Now()))
			defer timer.Stop()

			select {
			case r.buf = <-r.c:
			case <-timer.C:
				return 0, context.DeadlineExceeded
			}
		}
	}
}

// SetDeadline sets the deadline to read to the channel
func (r *ChannelReader) SetDeadline(deadline time.Time) {
	r.deadline = deadline
}