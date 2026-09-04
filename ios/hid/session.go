package hid

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/display"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/google/uuid"
)

const logModule = "go-ios/hid"

const (
	// Bounds the negotiation, since a wedged daemon never answers. Kept above the
	// deadline the device applies so it gives up first and tells us.
	streamStartTimeout = 15 * time.Second

	// streamStopTimeout bounds the best-effort stop during Close.
	streamStopTimeout = 5 * time.Second

	// The gap between the stream starting and the surfaces being re-matched, when
	// reports are still dropped. Nothing reports that state, so it is a fixed wait.
	surfaceSettleDelay = 300 * time.Millisecond

	// defaultTypingInterval paces keyboard reports; typing with no delay drops
	// characters on the device side.
	defaultTypingInterval = 20 * time.Millisecond
)

// Point is a position on the touchscreen, 0 to 65535 on each axis, independent
// of pixel size and orientation.
type Point struct {
	X uint16
	Y uint16
}

// Session delivers HID input and owns the media stream touch needs. Open one and
// reuse it: a stream per gesture is slow and has been seen to wedge the daemon.
type Session struct {
	device ios.DeviceEntry

	mutex sync.Mutex
	hid   *UniversalConnection

	// indigo is created on the first button press and reused afterwards, so a
	// script full of button presses does not redial the service each time.
	indigo *IndigoConnection

	// Stream state, populated by ensureStream and guarded by mutex, except
	// streamLost which the drain goroutine writes without it.
	displayService *display.Service
	receiver       *display.Receiver
	streamAnswer   display.StreamAnswer
	drainDone      chan struct{}
	// Set by the drain goroutine when the host stops receiving. Reports would be
	// accepted and dropped from then on, so the next gesture re-negotiates.
	streamLost atomic.Bool

	// readFrames consumes the RTP stream when a caller wants the video; nil
	// discards it.
	readFrames func(*display.Receiver) error

	// Tracks a contact held by the Touch* methods, so closing mid-gesture lifts it
	// rather than leaving the device believing a finger is down.
	contactDown bool
	lastContact Point

	// keyboardServiceID is the surface registered on first use by Type.
	keyboardServiceID uint64
	keyboardCreated   bool

	closed bool
}

type Option func(*Session)

// WithFrameReader replaces the default drain. read is handed the receiver and
// must read until it errors: an unread socket makes the device throttle.
func WithFrameReader(read func(*display.Receiver) error) Option {
	return func(s *Session) { s.readFrames = read }
}

// NewSession connects to the HID service and starts no media stream. Touch needs
// iOS 27+ and a kernel tunnel, but only the first gesture needing one fails.
func NewSession(device ios.DeviceEntry, opts ...Option) (*Session, error) {
	session := &Session{device: device}
	for _, opt := range opts {
		opt(session)
	}

	conn, err := NewUniversal(device)
	if err != nil {
		return nil, fmt.Errorf("NewSession: %w", err)
	}
	session.hid = conn
	return session, nil
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
func (s *Session) Tap(ctx context.Context, point Point) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.hid.SendTouchscreen(TouchContact, point.X, point.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("Tap: %w", err)
	}
	if err := s.hid.SendTouchscreen(TouchRelease, point.X, point.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("Tap: %w", err)
	}
	return nil
}

// Drag holds a contact down from one point to another, then releases. steps and
// duration set the velocity the device reads: short is a flick, long a slow drag.
func (s *Session) Drag(ctx context.Context, from, to Point, steps int, duration time.Duration) error {
	if steps < 1 {
		steps = 1
	}
	// steps is the number of moves, so steps+1 samples are sent: the touch-down
	// at from, which is what UIKit hit-tests, then one per step up to to.
	points := make([]Point, 0, steps+1)
	for i := 0; i <= steps; i++ {
		points = append(points, Point{
			X: interpolate(from.X, to.X, i, steps),
			Y: interpolate(from.Y, to.Y, i, steps),
		})
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.stroke(ctx, points, duration); err != nil {
		return fmt.Errorf("Drag: %w", err)
	}
	return nil
}

// Stroke drags a contact through every point and releases at the last, which is
// how a curved path is drawn. duration is spread evenly across it.
func (s *Session) Stroke(ctx context.Context, points []Point, duration time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.stroke(ctx, points, duration); err != nil {
		return fmt.Errorf("Stroke: %w", err)
	}
	return nil
}

// stroke sends the samples for a path, with the mutex held. The release happens
// even when a sample fails, or later gestures land on a device still touched.
func (s *Session) stroke(ctx context.Context, points []Point, duration time.Duration) error {
	if len(points) == 0 {
		return fmt.Errorf("a stroke needs at least one point")
	}

	contactDown := false
	// Where the contact actually is, which is not the last point when a sample
	// fails: releasing there would read as a flick to somewhere untravelled.
	at := points[0]
	defer func() {
		if !contactDown {
			return
		}
		if err := s.hid.SendTouchscreen(TouchRelease, at.X, at.Y, SurfaceMainTouchscreen); err != nil {
			golog.Warn("failed to lift the contact after a gesture, the device may still consider the screen touched",
				"module", logModule, "error", err)
		}
	}()

	var interval time.Duration
	if duration > 0 && len(points) > 1 {
		interval = duration / time.Duration(len(points)-1)
	}
	for i, point := range points {
		if err := s.hid.SendTouchscreen(TouchContact, point.X, point.Y, SurfaceMainTouchscreen); err != nil {
			return fmt.Errorf("contact sample %d/%d: %w", i+1, len(points), err)
		}
		contactDown = true
		at = point
		if err := sleepCtx(ctx, interval); err != nil {
			return err
		}
	}
	return nil
}

// MoveDigitizer moves the pointer a mirroring host draws, without putting a
// contact on the screen. Coordinates are the surface's own units, not a Point.
func (s *Session) MoveDigitizer(ctx context.Context, x, y int32) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.hid.SendDigitizer(x, y, SurfaceTouchscreenGesture); err != nil {
		return fmt.Errorf("MoveDigitizer: %w", err)
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

	// A key stays held on the device until a report without it arrives, so make
	// sure everything is released even if we stop mid-word.
	defer func() {
		if err := s.hid.SendKeyboard(s.keyboardServiceID); err != nil {
			golog.Warn("failed to release the keyboard, the device may still consider a key held",
				"module", logModule, "error", err)
		}
	}()

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
		if err := sleepCtx(ctx, defaultTypingInterval); err != nil {
			return fmt.Errorf("Type: %w", err)
		}
	}
	return nil
}

// TouchDown puts a contact on the screen and leaves it there. With TouchMove and
// TouchUp it is the streaming counterpart to Tap and Drag, for live input.
func (s *Session) TouchDown(ctx context.Context, point Point) error {
	return s.sendContact(ctx, point, "TouchDown")
}

// TouchMove moves a contact already down. The device tracks position rather than
// transitions, so this is TouchDown at the new point under a clearer name.
func (s *Session) TouchMove(ctx context.Context, point Point) error {
	return s.sendContact(ctx, point, "TouchMove")
}

// TouchUp lifts the contact, and is a no-op when nothing is down. ctx is ignored
// on purpose: honouring it would leave the device touched when it matters most.
func (s *Session) TouchUp(_ context.Context, point Point) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}
	if !s.contactDown {
		return nil
	}
	if err := s.hid.SendTouchscreen(TouchRelease, point.X, point.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("TouchUp: %w", err)
	}
	s.contactDown = false
	return nil
}

func (s *Session) sendContact(ctx context.Context, point Point, op string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.hid.SendTouchscreen(TouchContact, point.X, point.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.contactDown = true
	s.lastContact = point
	return nil
}

// PressButton presses and releases a hardware button. Usage page 0x0C is the
// media buttons, 0x09 the generic ones. Buttons need no stream, so none starts,
// and with no stream to negotiate there is nothing here for ctx to bound.
func (s *Session) PressButton(_ context.Context, usagePage, usageCode uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}

	if s.indigo == nil {
		indigo, err := NewIndigo(s.device)
		if err != nil {
			return fmt.Errorf("PressButton: %w", err)
		}
		s.indigo = indigo
	}

	if err := s.indigo.SendButton(usagePage, usageCode, ButtonDown); err != nil {
		return fmt.Errorf("PressButton: %w", err)
	}
	if err := s.indigo.SendButton(usagePage, usageCode, ButtonUp); err != nil {
		// The device considers the button held until it hears otherwise; cancel
		// the press so it does not stay down.
		if cancelErr := s.indigo.SendButton(usagePage, usageCode, ButtonCanceled); cancelErr != nil {
			golog.Warn("failed to cancel a button press, the device may still consider it held",
				"module", logModule, "error", cancelErr)
		}
		return fmt.Errorf("PressButton: %w", err)
	}
	return nil
}

// Close stops any media stream and closes every connection. It is idempotent,
// and waits for a gesture in flight, including a negotiation.
func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	// Lift a contact left down by an interrupted input stream first: once the
	// stream is gone the device would keep believing a finger is on the screen.
	if s.contactDown {
		if err := s.hid.SendTouchscreen(TouchRelease, s.lastContact.X, s.lastContact.Y, SurfaceMainTouchscreen); err != nil {
			golog.Warn("failed to lift a held contact while closing, the device may still consider the screen touched",
				"module", logModule, "error", err)
		}
		s.contactDown = false
	}

	s.teardownStream()

	if s.indigo != nil {
		if err := s.indigo.Close(); err != nil {
			golog.Debug("closing the Indigo connection failed", "module", logModule, "error", err)
		}
		s.indigo = nil
	}

	if s.hid != nil {
		if err := s.hid.Close(); err != nil {
			return fmt.Errorf("Session.Close: %w", err)
		}
		s.hid = nil
	}
	return nil
}

// StreamActive reports whether a stream is holding the touch gate open. It turns
// false again if the host stops receiving.
func (s *Session) StreamActive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.displayService != nil && !s.streamLost.Load()
}

// EnsureStream starts the media stream unless one is running. Optional, but it
// moves the negotiation out of the first gesture, whose samples would be lost.
func (s *Session) EnsureStream(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}
	return s.ensureStream(ctx)
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

// ensureStream starts the stream unless one is running, with the mutex held. One
// the host stopped receiving is re-negotiated, and any failure tears down first.
func (s *Session) ensureStream(ctx context.Context) error {
	if s.displayService != nil {
		if !s.streamLost.Load() {
			return nil
		}
		golog.Warn("the media stream stopped, re-negotiating before continuing", "module", logModule)
		s.teardownStream()
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

	// Start draining before the stream is negotiated: the device begins sending
	// as soon as it answers, and an unread socket makes it throttle its encoder.
	s.streamLost.Store(false)
	s.drainDone = make(chan struct{})
	go func(done chan struct{}) {
		defer close(done)
		read := s.readFrames
		if read == nil {
			read = func(r *display.Receiver) error { return r.Drain() }
		}
		if err := read(receiver); err != nil {
			golog.Warn("the media stream that gates touch input stopped",
				"module", logModule, "error", err)
			s.streamLost.Store(true)
		}
	}(s.drainDone)

	startCtx, cancel := context.WithTimeout(ctx, streamStartTimeout)
	defer cancel()

	// The answer carries the session id even on failure, because the device may
	// have brought the stream up anyway; keep it so teardown can stop it.
	answer, err := svc.StartVideoStream(startCtx, display.VideoStreamRequest{
		ReceiverIP:   receiver.IP(),
		ReceiverPort: receiver.Port(),
		SenderIP:     s.device.Address,
		DisplayID:    display.DefaultDisplayID,
	})
	s.streamAnswer = answer
	if err != nil {
		s.teardownStream()
		return fmt.Errorf("failed to start the media stream touch input requires: %w", err)
	}

	// At info: once a session, and the only positive evidence that reports will
	// reach the app rather than being discarded.
	golog.Info("media stream started, touch surfaces authenticated", "module", logModule,
		"receiver", receiver.IP(), "port", receiver.Port())

	// Not tied to ctx on purpose: the stream is already negotiated, and giving up
	// here would mean re-negotiating next time, which is the churn that wedges it.
	time.Sleep(surfaceSettleDelay)
	return nil
}

// teardownStream reverses ensureStream. Call with the mutex held. Safe to call
// at any point of a partially completed setup.
func (s *Session) teardownStream() {
	if s.displayService != nil {
		if s.streamAnswer.ClientSessionID != uuid.Nil {
			s.stopStream(s.displayService, s.streamAnswer.ClientSessionID)
		}
		if err := s.displayService.Close(); err != nil {
			golog.Debug("closing the display service failed", "module", logModule, "error", err)
		}
		s.displayService = nil
	}

	// Closing the socket is what unblocks Drain: a blocked read is not
	// interrupted by anything else, so this has to happen before the wait below.
	if s.receiver != nil {
		if err := s.receiver.Close(); err != nil {
			golog.Debug("closing the RTP receiver failed", "module", logModule, "error", err)
		}
		s.receiver = nil
	}
	if s.drainDone != nil {
		<-s.drainDone
		s.drainDone = nil
	}

	s.streamAnswer = display.StreamAnswer{}
	s.streamLost.Store(false)
}

// stopStream stops the stream, retrying on a new connection if the first attempt
// fails. A negotiation that timed out closed its own connection, so the service
// that started the stream cannot always be the one that stops it.
func (s *Session) stopStream(svc *display.Service, sessionID uuid.UUID) {
	stopCtx, cancel := context.WithTimeout(context.Background(), streamStopTimeout)
	defer cancel()

	// The device dropping the channel mid-stop is normal and means it worked.
	err := svc.StopMediaStream(stopCtx, sessionID)
	if err == nil {
		return
	}
	golog.Warn("stopping the media stream failed, retrying on a new connection",
		"module", logModule, "error", err)

	// A stream left running is what wedges the daemon, so this is worth a second
	// attempt even though the first one already logged.
	retry, err := display.New(s.device)
	if err != nil {
		golog.Warn("could not reconnect to stop the media stream; the device may need a reboot",
			"module", logModule, "error", err)
		return
	}
	defer func() {
		if err := retry.Close(); err != nil {
			golog.Debug("closing the retry connection failed", "module", logModule, "error", err)
		}
	}()

	retryCtx, cancelRetry := context.WithTimeout(context.Background(), streamStopTimeout)
	defer cancelRetry()
	if err := retry.StopMediaStream(retryCtx, sessionID); err != nil {
		golog.Warn("stopping the media stream failed again; the device may need a reboot",
			"module", logModule, "error", err)
	}
}

// interpolate returns the coordinate at step of steps. The arithmetic is 64-bit
// because the product reaches 65535*steps.
func interpolate(from, to uint16, step, steps int) uint16 {
	delta := (int64(to) - int64(from)) * int64(step) / int64(steps)
	return uint16(int64(from) + delta)
}

// sleepCtx sleeps for d unless ctx is cancelled first. A non-positive d does not
// sleep and does not report the context: callers check ctx where it matters.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
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
