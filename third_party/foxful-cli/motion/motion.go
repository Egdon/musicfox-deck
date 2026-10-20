// Package motion times animations: easing curves, tweens, critically damped
// springs and a scheduler that keeps frames coming only while something
// moves.
package motion

import (
	"math"
	"sync"
	"time"
)

// Level is how much the interface animates.
type Level uint8

const (
	Full   Level = iota // every transition
	Subtle              // fades only, no movement
	Off                 // nothing moves
)

func (l Level) String() string {
	switch l {
	case Subtle:
		return "subtle"
	case Off:
		return "off"
	}
	return "full"
}

// ParseLevel reads a level name; unknown names count as Full.
func ParseLevel(s string) Level {
	switch s {
	case "subtle":
		return Subtle
	case "off":
		return Off
	}
	return Full
}

// Ease maps progress from 0..1 to 0..1. None of the curves overshoot.
type Ease func(float64) float64

func Linear(t float64) float64   { return t }
func InQuad(t float64) float64   { return t * t }
func InCubic(t float64) float64  { return t * t * t }
func OutCubic(t float64) float64 { return 1 - math.Pow(1-t, 3) }
func OutQuart(t float64) float64 { return 1 - math.Pow(1-t, 4) }

func InOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// Tween runs from Start+Delay for Dur.
type Tween struct {
	Start      time.Time
	Delay, Dur time.Duration
	Ease       Ease
}

// At returns the eased progress at now: 0 before the tween starts, 1 after
// it ends and for a zero tween.
func (tw Tween) At(now time.Time) float64 {
	if tw.Start.IsZero() || tw.Dur <= 0 {
		return 1
	}
	t := float64(now.Sub(tw.Start)-tw.Delay) / float64(tw.Dur)
	t = min(max(t, 0), 1)
	if tw.Ease != nil {
		return tw.Ease(t)
	}
	return t
}

// End returns when the tween is over.
func (tw Tween) End() time.Time { return tw.Start.Add(tw.Delay + tw.Dur) }

// Done reports whether the tween is over at now.
func (tw Tween) Done(now time.Time) bool { return tw.Start.IsZero() || !now.Before(tw.End()) }

// Spring is a critically damped spring evaluated in closed form, so its
// position depends only on time and frames can come at any rate.
type Spring struct {
	From, To float64
	V0       float64 // velocity at Start
	Start    time.Time
	Omega    float64 // stiffness; it settles in about 6.6/Omega seconds
}

// At returns position and velocity at now.
func (s Spring) At(now time.Time) (x, v float64) {
	t := max(now.Sub(s.Start).Seconds(), 0)
	c1 := s.From - s.To
	c2 := s.V0 + s.Omega*c1
	e := math.Exp(-s.Omega * t)
	return s.To + (c1+c2*t)*e, (c2 - s.Omega*(c1+c2*t)) * e
}

// Retarget moves the target, keeping position and velocity continuous.
func (s *Spring) Retarget(now time.Time, to float64) {
	x, v := s.At(now)
	s.From, s.V0, s.To, s.Start = x, v, to, now
}

// Jump puts the spring at rest at x.
func (s *Spring) Jump(x float64) { s.From, s.To, s.V0 = x, x, 0 }

// Settled reports whether the spring is within eps of its target and nearly
// still.
func (s Spring) Settled(now time.Time, eps float64) bool {
	x, v := s.At(now)
	return math.Abs(x-s.To) < eps && math.Abs(v) < eps*s.Omega
}

// SettleTime is how long the spring takes to come within 1% of its target
// from rest.
func (s Spring) SettleTime() time.Duration {
	if s.Omega <= 0 {
		return 0
	}
	return time.Duration(6.64 / s.Omega * float64(time.Second))
}

// StaggerDelay is the start delay of item i in a staggered entrance: items
// start step apart, and the last starts no later than spread.
func StaggerDelay(i int, step, spread time.Duration) time.Duration {
	return min(time.Duration(i)*step, spread)
}

// FrameMsg asks for a new frame.
type FrameMsg struct{ gen uint64 }

// SlowInterval paces slow continuous animations, such as something floating,
// when nothing faster runs.
const SlowInterval = 33 * time.Millisecond

// Scheduler keeps a single frame timer in flight while animations run and
// stops when they end. Animations started anywhere, even while rendering,
// extend its deadline.
type Scheduler struct {
	mu       sync.Mutex
	interval time.Duration
	send     func(any)
	level    Level
	deadline time.Time // fast frames until then
	slow     time.Time // slow frames until then
	inFlight bool
	slowTick bool // the pending frame is on the slow interval
	gen      uint64
	timer    *time.Timer
	now      func() time.Time
	after    func(time.Duration, func()) *time.Timer
}

// NewScheduler returns a scheduler that sends a FrameMsg through send every
// interval while animations run. A nil send never schedules.
func NewScheduler(interval time.Duration, level Level, send func(any)) *Scheduler {
	return &Scheduler{interval: interval, send: send, level: level, now: time.Now, after: time.AfterFunc}
}

// Level returns the configured level.
func (s *Scheduler) Level() Level {
	if s == nil {
		return Off
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.level
}

// SetLevel changes the level; Off stops the frames.
func (s *Scheduler) SetLevel(l Level) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.level = l
	if l == Off {
		s.deadline, s.slow = time.Time{}, time.Time{}
	}
}

// Animate keeps frames coming for d from now.
func (s *Scheduler) Animate(d time.Duration) {
	if s == nil || d <= 0 {
		return
	}
	s.AnimateUntil(s.now().Add(d))
}

// AnimateUntil keeps frames coming until t.
func (s *Scheduler) AnimateUntil(t time.Time) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.level == Off {
		return
	}
	if t.After(s.deadline) {
		s.deadline = t
	}
	// A pending slow frame would hold up the first fast one.
	if s.inFlight && s.slowTick && s.timer != nil {
		s.timer.Stop()
		s.gen++
		s.inFlight = false
	}
	s.arm()
}

// AnimateSlow keeps frames coming at SlowInterval for d from now.
func (s *Scheduler) AnimateSlow(d time.Duration) {
	if s == nil || d <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.level == Off {
		return
	}
	if t := s.now().Add(d); t.After(s.slow) {
		s.slow = t
	}
	s.arm()
}

// Running reports whether frames are scheduled.
func (s *Scheduler) Running() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inFlight
}

// HandleFrame accounts for a received frame and schedules the next one while
// the deadline lies ahead. It reports whether the frame was current.
func (s *Scheduler) HandleFrame(m FrameMsg) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m.gen != s.gen {
		return false
	}
	s.inFlight = false
	s.arm()
	return true
}

// Stop cancels any pending frame.
func (s *Scheduler) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deadline, s.slow = time.Time{}, time.Time{}
	s.gen++
	s.inFlight = false
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}

// arm schedules the next frame unless one is pending or the deadline passed.
// The frame after the deadline still comes, to draw the final state.
func (s *Scheduler) arm() {
	if s.inFlight || s.send == nil {
		return
	}
	now := s.now()
	if !s.deadline.IsZero() && now.After(s.deadline.Add(s.interval)) {
		s.deadline = time.Time{}
	}
	if !s.slow.IsZero() && now.After(s.slow.Add(SlowInterval)) {
		s.slow = time.Time{}
	}
	interval := s.interval
	switch {
	case !s.deadline.IsZero():
	case !s.slow.IsZero():
		interval = max(SlowInterval, s.interval)
	default:
		return
	}
	s.inFlight, s.slowTick = true, interval != s.interval
	gen, send := s.gen, s.send
	s.timer = s.after(interval, func() { send(FrameMsg{gen: gen}) })
}
