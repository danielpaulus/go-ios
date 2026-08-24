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
	// streamStartTimeout bounds the media-stream negotiation. A wedged
	// mediastream daemon never answers at all, so failing fast with an
	// actionable error beats hanging. It is deliberately longer than the
	// deadline the device applies to its own side (display's
	// negotiationTimeoutSeconds): the device gives up first and tells us, rather
	// than us abandoning a negotiation it goes on to complete.
	streamStartTimeout = 15 * time.Second

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

	// indigo is created on the first button press and reused afterwards, so a
	// script full of button presses does not redial the service each time.
	indigo *IndigoConnection

	// Stream state, all guarded by mutex and populated by ensureStream.
	displayService *display.Service
	receiver       *display.Receiver
	streamAnswer   display.StreamAnswer
	drainDone      chan struct{}
	// streamLost is set by the drain goroutine when the host stops receiving the
	// stream. Reports would still be accepted and silently dropped from that
	// point on, so the next gesture re-negotiates instead.
	streamLost atomic.Bool

	// contactDown tracks a contact held by the streaming Touch* methods, so a
	// session that is closed mid-gesture can lift it instead of leaving the
	// device believing a finger is still on the screen.
	contactDown bool
	lastContact Point

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

// Stroke drags a single contact through every point in order and releases at the
// last one, which is how any path that is not a straight line - an arc, a
// circle, a hand-drawn shape - is drawn. Drag is the two-point case.
//
// duration is spread evenly across the path, so it sets the speed the device's
// gesture recogniser reads from the report timestamps. A path of one point is a
// tap.
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

// stroke sends the contact samples for a path. Call with the mutex held and the
// stream already ensured.
//
// Once a contact is down the device believes a finger is on the screen until it
// is lifted, so the release has to happen even if a sample or the context fails
// part way through - otherwise every later gesture, in this session or the next,
// lands on a device that still thinks it is being touched.
func (s *Session) stroke(ctx context.Context, points []Point, duration time.Duration) error {
	if len(points) == 0 {
		return fmt.Errorf("a stroke needs at least one point")
	}

	contactDown := false
	last := points[len(points)-1]
	defer func() {
		if !contactDown {
			return
		}
		if err := s.hid.SendTouchscreen(TouchRelease, last.X, last.Y, SurfaceMainTouchscreen); err != nil {
			golog.Warn("failed to lift the contact after a gesture, the device may still consider the screen touched",
				"module", logModule, "error", err)
		}
	}()

	var interval time.Duration
	if duration > 0 && len(points) > 1 {
		interval = duration / time.Duration(len(points)-1)
	}
	for i, p := range points {
		if err := s.hid.SendTouchscreen(TouchContact, p.X, p.Y, SurfaceMainTouchscreen); err != nil {
			return fmt.Errorf("contact sample %d/%d: %w", i+1, len(points), err)
		}
		contactDown = true
		if err := sleepCtx(ctx, interval); err != nil {
			return err
		}
	}
	return nil
}

// MoveDigitizer posts a single pointer sample on the trackpad-style gesture
// surface. It moves the pointer a mirroring host draws without putting a contact
// on the screen, so it does not by itself produce a touch.
//
// Unlike Tap and Drag, which take normalised Points, the coordinates here are
// the gesture surface's own digitizer units.
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
		if err := sleepCtx(ctx, s.typingInterval); err != nil {
			return fmt.Errorf("Type: %w", err)
		}
	}
	return nil
}

// TouchDown puts a contact on the screen at p and leaves it there.
//
// TouchDown, TouchMove and TouchUp are the streaming counterpart to Tap and
// Drag: they exist for live input, where each pointer sample arrives separately
// (a user dragging a finger in a remote-control UI) rather than as a gesture
// known up front. The device's model is absolute - every contact report says
// "in contact at this position" - so a stream is TouchDown, any number of
// TouchMove, then TouchUp.
//
// A contact left down is released by Close, but a caller that drops a stream
// should send TouchUp itself: until then the device believes a finger is down
// and interprets everything else in that light.
func (s *Session) TouchDown(ctx context.Context, p Point) error {
	return s.sendContact(ctx, p, "TouchDown")
}

// TouchMove moves a contact that is already down. It is equivalent to
// TouchDown at the new position - the device tracks position, not transitions -
// but the separate name keeps caller code honest about intent.
func (s *Session) TouchMove(ctx context.Context, p Point) error {
	return s.sendContact(ctx, p, "TouchMove")
}

// TouchUp lifts the contact at p. It is a no-op when nothing is down, so an
// input stream that sends a stray release cannot desynchronise the device.
func (s *Session) TouchUp(ctx context.Context, p Point) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}
	if !s.contactDown {
		return nil
	}
	if err := s.hid.SendTouchscreen(TouchRelease, p.X, p.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("TouchUp: %w", err)
	}
	s.contactDown = false
	return nil
}

func (s *Session) sendContact(ctx context.Context, p Point, op string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.beginGatedGesture(ctx); err != nil {
		return err
	}
	if err := s.hid.SendTouchscreen(TouchContact, p.X, p.Y, SurfaceMainTouchscreen); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.contactDown = true
	s.lastContact = p
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

// Close stops the media stream if one is running and closes every connection.
// It is idempotent.
//
// Close waits for any gesture in flight, including a media-stream negotiation,
// so it can block for as long as that negotiation is allowed to take.
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

// StreamActive reports whether a media stream is currently holding the touch
// auth gate open. It turns false again if the host stops receiving the stream.
// Exposed for diagnostics and tests.
func (s *Session) StreamActive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.displayService != nil && !s.streamLost.Load()
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
// unless one is already running. Call with the mutex held.
//
// A stream the host has stopped receiving is torn down and re-negotiated: the
// device would keep accepting reports and silently dropping them, which is the
// failure this package exists to prevent.
//
// On any failure it tears down whatever was already set up, so a failed gesture
// leaves no half-open stream behind and the next attempt starts clean.
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
		if err := receiver.Drain(); err != nil {
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
		DisplayID:    s.displayID,
	})
	s.streamAnswer = answer
	if err != nil {
		s.teardownStream()
		return fmt.Errorf("failed to start the media stream touch input requires: %w", err)
	}

	// Logged at info: it happens once per session, and it is the only positive
	// evidence that touch reports will actually reach the app rather than being
	// silently discarded.
	golog.Info("media stream started, touch surfaces authenticated", "module", logModule,
		"receiver", receiver.IP(), "port", receiver.Port())

	// Give backboardd a moment to re-match the surfaces against the new stream.
	// This is deliberately not tied to ctx: the stream is already negotiated, and
	// discarding it here would mean re-negotiating on the next gesture, which is
	// the churn that wedges the device's mediastream daemon.
	time.Sleep(surfaceSettleDelay)
	return nil
}

// teardownStream reverses ensureStream. Call with the mutex held. Safe to call
// at any point of a partially completed setup.
func (s *Session) teardownStream() {
	if s.displayService != nil {
		if s.streamAnswer.ClientSessionID != uuid.Nil {
			stopCtx, cancel := context.WithTimeout(context.Background(), streamStopTimeout)
			// The receiver is still draining here, so the device is not
			// throttled while it works through the stop. An error is worth
			// surfacing: a stream left running is what wedges the device's
			// mediastream daemon, though the device dropping the channel
			// mid-stop is normal and means the stop took effect.
			if err := s.displayService.StopMediaStream(stopCtx, s.streamAnswer.ClientSessionID); err != nil {
				golog.Warn("stopping the media stream failed; if this repeats, the device may need a reboot",
					"module", logModule, "error", err)
			}
			cancel()
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

// interpolate returns the coordinate at step of steps between from and to.
// The arithmetic is 64-bit: the product reaches 65535*steps, which overflows a
// 32-bit int on the platforms go-ios builds for.
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
