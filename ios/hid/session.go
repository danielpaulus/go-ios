package hid

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/display"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/google/uuid"
)

const logModule = "go-ios/hid"

const (
	// streamStartTimeout bounds the media-stream negotiation. A wedged
	// mediastream daemon never answers at all, so failing fast with an
	// actionable error beats hanging.
	streamStartTimeout = 10 * time.Second

	// streamStopTimeout bounds the best-effort stop during Close.
	streamStopTimeout = 5 * time.Second

	// surfaceSettleDelay gives backboardd a moment to re-match the HID surfaces
	// against the newly authenticated stream. Without it the first report of a
	// gesture can still land while the surface is externalAccessory and be
	// dropped.
	surfaceSettleDelay = 300 * time.Millisecond

	// defaultTypingInterval paces keyboard reports; typing with no delay drops
	// characters on the device side.
	defaultTypingInterval = 20 * time.Millisecond
)

// Point is a position on the touchscreen in normalised coordinates: 0 is the
// left/top edge and 65535 the right/bottom edge, independent of the display's
// pixel size or the current orientation.
type Point struct {
	X uint16
	Y uint16
}

// Session delivers HID input to a device and owns the media stream that touch
// input requires.
//
// Open one session and reuse it for as many gestures as needed. Touch reports
// only reach the app while a media stream holds dtuhidd's auth gate open, and
// negotiating a stream per gesture both adds seconds of latency and has been
// observed to wedge the device's mediastream daemon until it is rebooted.
//
// The stream is started lazily, on the first gesture that needs it: pressing
// hardware buttons never starts one, because the button surface is authenticated
// out of the box.
//
// A Session is safe for concurrent use; gestures are serialised.
type Session struct {
	device ios.DeviceEntry

	mutex sync.Mutex
	hid   *UniversalConnection

	// Stream state, all guarded by mutex and populated by ensureStream.
	displayService *display.Service
	receiver       *display.Receiver
	streamAnswer   display.StreamAnswer
	drainCancel    context.CancelFunc
	drainDone      chan struct{}

	// keyboardServiceID is the surface registered on first use by Type.
	keyboardServiceID uint64
	keyboardCreated   bool

	closed bool

	// displayID selects which display the stream mirrors.
	displayID int
	// typingInterval paces keyboard reports.
	typingInterval time.Duration
}

// SessionOption configures a Session at open time.
type SessionOption func(*Session)

// WithDisplayID selects the display whose stream holds the auth gate open.
// Defaults to the main display.
func WithDisplayID(id int) SessionOption {
	return func(s *Session) { s.displayID = id }
}

// WithTypingInterval sets the delay between keyboard reports sent by Type.
func WithTypingInterval(d time.Duration) SessionOption {
	return func(s *Session) { s.typingInterval = d }
}

// NewSession connects to the HID service on the device. It does not start a
// media stream: that happens on the first gesture that needs one.
//
// Requires iOS 17+ with the Developer Disk Image mounted, and - because touch
// input needs the device to stream back to the host - a kernel tunnel. On a
// userspace tunnel the session opens but the first touch fails with
// display.ErrUserspaceTunnelUnsupported.
func NewSession(device ios.DeviceEntry, opts ...SessionOption) (*Session, error) {
	s := &Session{
		device:         device,
		displayID:      display.DefaultDisplayID,
		typingInterval: defaultTypingInterval,
	}
	for _, opt := range opts {
		opt(s)
	}

	conn, err := NewUniversal(device)
	if err != nil {
		return nil, fmt.Errorf("NewSession: %w", err)
	}
	s.hid = conn
	return s, nil
}

// ListServices enumerates the HID surfaces registered on the device.
func (s *Session) ListServices() (map[string]interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return nil, err
	}
	return s.hid.ListConnectedServices()
}

// Tap taps once at p: one contact report followed by a release at the same
// position.
func (s *Session) Tap(ctx context.Context, p Point) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.hid.SendTouchscreen(TouchContact, p.X, p.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("Tap: %w", err)
	}
	if err := s.hid.SendTouchscreen(TouchRelease, p.X, p.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("Tap: %w", err)
	}
	return nil
}

// Drag holds a contact down while moving it from one point to another, then
// releases at the end position.
//
// steps is how many intermediate contact samples to send and duration how long
// the drag takes; together they set the velocity the device's gesture recogniser
// reads from the report timestamps. A short duration over a modest distance
// reads as a flick - which is how VoiceOver navigation is driven - while a long
// one reads as a slow drag. steps below 1 is treated as 1.
func (s *Session) Drag(ctx context.Context, from, to Point, steps int, duration time.Duration) error {
	if steps < 1 {
		steps = 1
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}

	var interval time.Duration
	if duration > 0 {
		interval = duration / time.Duration(steps)
	}
	for i := 1; i <= steps; i++ {
		x := interpolate(from.X, to.X, i, steps)
		y := interpolate(from.Y, to.Y, i, steps)
		if err := s.hid.SendTouchscreen(TouchContact, x, y, SurfaceMainTouchscreen); err != nil {
			return fmt.Errorf("Drag: contact sample %d/%d: %w", i, steps, err)
		}
		if interval > 0 && i < steps {
			if err := sleepCtx(ctx, interval); err != nil {
				return fmt.Errorf("Drag: %w", err)
			}
		}
	}
	if err := s.hid.SendTouchscreen(TouchRelease, to.X, to.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("Drag: release: %w", err)
	}
	return nil
}

// Move posts a single pointer sample on the gesture surface. It moves the
// pointer a mirroring host draws without putting a contact on the screen, so it
// does not by itself produce a touch.
func (s *Session) Move(ctx context.Context, x, y int32) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.hid.SendDigitizer(x, y, SurfaceTouchscreenGesture); err != nil {
		return fmt.Errorf("Move: %w", err)
	}
	return nil
}

// Type sends text through a virtual keyboard, registering the keyboard surface
// on first use. Characters without a HID mapping are skipped.
func (s *Session) Type(ctx context.Context, text string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if !s.keyboardCreated {
		id, err := s.hid.CreateKeyboardService(SurfaceKeyboardDefault, "go-ios Keyboard", "go-ios", 0x05AC, 0x0267)
		if err != nil {
			return fmt.Errorf("Type: %w", err)
		}
		s.keyboardServiceID = id
		s.keyboardCreated = true
	}

	for _, ch := range text {
		key, ok := KeyForRune(ch)
		if !ok {
			golog.Warn("skipping character without a HID mapping", "module", logModule, "char", string(ch))
			continue
		}
		usages := []uint8{key.Usage}
		if key.Shift {
			usages = append(usages, KeyLeftShift)
		}
		if err := s.hid.SendKeyboard(s.keyboardServiceID, usages...); err != nil {
			return fmt.Errorf("Type: failed to press %q: %w", string(ch), err)
		}
		// Release everything before the next character, otherwise repeated
		// characters collapse into one keypress.
		if err := s.hid.SendKeyboard(s.keyboardServiceID); err != nil {
			return fmt.Errorf("Type: failed to release %q: %w", string(ch), err)
		}
		if err := sleepCtx(ctx, s.typingInterval); err != nil {
			return fmt.Errorf("Type: %w", err)
		}
	}
	return nil
}

// PressButton presses and releases a hardware button, identified in HID terms:
// usage page 0x0C (Consumer) covers the media buttons, 0x09 (Button) the generic
// ones.
//
// Buttons need no media stream, so this never starts one.
func (s *Session) PressButton(ctx context.Context, usagePage, usageCode uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}

	indigo, err := NewIndigo(s.device)
	if err != nil {
		return fmt.Errorf("PressButton: %w", err)
	}
	defer indigo.Close()

	if err := indigo.SendButton(usagePage, usageCode, ButtonDown); err != nil {
		return fmt.Errorf("PressButton: %w", err)
	}
	if err := indigo.SendButton(usagePage, usageCode, ButtonUp); err != nil {
		return fmt.Errorf("PressButton: %w", err)
	}
	return nil
}

// Close stops the media stream if one is running and closes every connection.
// It is idempotent.
func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	s.teardownStream()

	if s.hid != nil {
		if err := s.hid.Close(); err != nil {
			return fmt.Errorf("Session.Close: %w", err)
		}
		s.hid = nil
	}
	return nil
}

// StreamActive reports whether the media stream that gates touch input is
// currently running. Exposed for diagnostics and tests.
func (s *Session) StreamActive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.displayService != nil
}

// beginGatedGesture is the common preamble for gestures that need the auth gate
// open. Call with the mutex held.
func (s *Session) beginGatedGesture(ctx context.Context) error {
	if err := s.checkOpen(); err != nil {
		return err
	}
	return s.ensureStream(ctx)
}

func (s *Session) checkOpen() error {
	if s.closed || s.hid == nil {
		return fmt.Errorf("session is closed")
	}
	return nil
}

// ensureStream starts the media stream that authenticates the touch surfaces,
// unless it is already running. Call with the mutex held.
//
// On any failure it tears down whatever was already set up, so a failed gesture
// leaves no half-open stream behind and the next attempt starts clean.
func (s *Session) ensureStream(ctx context.Context) error {
	if s.displayService != nil {
		return nil
	}

	receiver, err := display.OpenReceiver(s.device)
	if err != nil {
		return fmt.Errorf("failed to prepare the media stream touch input requires: %w", err)
	}
	s.receiver = receiver

	svc, err := display.New(s.device)
	if err != nil {
		s.teardownStream()
		return fmt.Errorf("failed to prepare the media stream touch input requires: %w", err)
	}
	s.displayService = svc

	startCtx, cancel := context.WithTimeout(ctx, streamStartTimeout)
	defer cancel()

	answer, err := svc.StartVideoStream(startCtx, display.VideoStreamRequest{
		ReceiverIP:   receiver.IP(),
		ReceiverPort: receiver.Port(),
		SenderIP:     s.device.Address,
		DisplayID:    s.displayID,
	})
	if err != nil {
		s.teardownStream()
		return fmt.Errorf("failed to start the media stream touch input requires: %w", err)
	}
	s.streamAnswer = answer

	drainCtx, drainCancel := context.WithCancel(context.Background())
	s.drainCancel = drainCancel
	s.drainDone = make(chan struct{})
	go func(done chan struct{}) {
		defer close(done)
		receiver.Drain(drainCtx)
	}(s.drainDone)

	golog.Debug("media stream started, touch surfaces authenticated", "module", logModule,
		"receiver", receiver.IP(), "port", receiver.Port())

	if err := sleepCtx(ctx, surfaceSettleDelay); err != nil {
		s.teardownStream()
		return err
	}
	return nil
}

// teardownStream reverses ensureStream. Call with the mutex held. Safe to call
// at any point of a partially completed setup.
func (s *Session) teardownStream() {
	if s.drainCancel != nil {
		s.drainCancel()
		s.drainCancel = nil
	}

	if s.displayService != nil {
		if s.streamAnswer.ClientSessionID != uuid.Nil {
			stopCtx, cancel := context.WithTimeout(context.Background(), streamStopTimeout)
			// The device routinely drops the channel while processing the stop,
			// which surfaces as an error even though the stop took effect.
			if err := s.displayService.StopMediaStream(stopCtx, s.streamAnswer.ClientSessionID); err != nil {
				golog.Debug("stopping the media stream reported an error, which is expected when the device closes the channel first",
					"module", logModule, "error", err)
			}
			cancel()
		}
		if err := s.displayService.Close(); err != nil {
			golog.Debug("closing the display service failed", "module", logModule, "error", err)
		}
		s.displayService = nil
	}

	if s.receiver != nil {
		if err := s.receiver.Close(); err != nil {
			golog.Debug("closing the RTP receiver failed", "module", logModule, "error", err)
		}
		s.receiver = nil
	}

	// The drain goroutine exits once the socket is closed; wait so Close does
	// not leave it running.
	if s.drainDone != nil {
		<-s.drainDone
		s.drainDone = nil
	}

	s.streamAnswer = display.StreamAnswer{}
}

func interpolate(from, to uint16, step, steps int) uint16 {
	delta := (int(to) - int(from)) * step / steps
	return uint16(int(from) + delta)
}

// sleepCtx sleeps for d unless ctx is cancelled first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
