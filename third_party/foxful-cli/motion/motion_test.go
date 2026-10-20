package motion

import (
	"math"
	"testing"
	"time"
)

func TestEasesStayInRange(t *testing.T) {
	for name, e := range map[string]Ease{"linear": Linear, "inQuad": InQuad, "inCubic": InCubic,
		"outCubic": OutCubic, "outQuart": OutQuart, "inOutCubic": InOutCubic} {
		if e(0) != 0 || math.Abs(e(1)-1) > 1e-12 {
			t.Errorf("%s endpoints: %v %v", name, e(0), e(1))
		}
		prev := 0.0
		for i := 1; i <= 100; i++ {
			v := e(float64(i) / 100)
			if v < prev || v > 1+1e-12 {
				t.Fatalf("%s is not monotonic within 0..1 at %d: %v", name, i, v)
			}
			prev = v
		}
	}
}

func TestTween(t *testing.T) {
	start := time.Unix(0, 0)
	tw := Tween{Start: start, Delay: 10 * time.Millisecond, Dur: 100 * time.Millisecond, Ease: Linear}
	if tw.At(start) != 0 || tw.At(start.Add(60*time.Millisecond)) != 0.5 || tw.At(start.Add(time.Second)) != 1 {
		t.Fatalf("progress: %v %v %v", tw.At(start), tw.At(start.Add(60*time.Millisecond)), tw.At(start.Add(time.Second)))
	}
	if !tw.Done(start.Add(110*time.Millisecond)) || tw.Done(start.Add(109*time.Millisecond)) {
		t.Fatal("done at the wrong time")
	}
	if (Tween{}).At(start) != 1 {
		t.Fatal("a zero tween is finished")
	}
}

// The closed form must match numeric integration of x” = -2ωx' - ω²(x-to).
func TestSpringMatchesIntegration(t *testing.T) {
	start := time.Unix(0, 0)
	s := Spring{From: 3, To: 10, V0: -4, Start: start, Omega: 9}
	x, v := s.From, s.V0
	const dt = 1e-5
	for i := 1; i <= 50000; i++ {
		a := -2*s.Omega*v - s.Omega*s.Omega*(x-s.To)
		v += a * dt
		x += v * dt
		if i%5000 == 0 {
			cx, cv := s.At(start.Add(time.Duration(float64(i) * dt * float64(time.Second))))
			if math.Abs(cx-x) > 1e-3 || math.Abs(cv-v) > 1e-2 {
				t.Fatalf("t=%v: closed form (%v, %v), integrated (%v, %v)", float64(i)*dt, cx, cv, x, v)
			}
		}
	}
}

func TestSpringRetargetIsContinuous(t *testing.T) {
	start := time.Unix(0, 0)
	s := Spring{From: 0, To: 1, Start: start, Omega: 12}
	mid := start.Add(80 * time.Millisecond)
	x0, v0 := s.At(mid)
	s.Retarget(mid, -2)
	x1, v1 := s.At(mid)
	if math.Abs(x0-x1) > 1e-12 || math.Abs(v0-v1) > 1e-9 {
		t.Fatalf("jump at retarget: (%v,%v) -> (%v,%v)", x0, v0, x1, v1)
	}
	if !s.Settled(mid.Add(s.SettleTime()+200*time.Millisecond), 0.03) {
		t.Fatal("not settled after the settle time")
	}
	// Starting at rest, it never overshoots.
	r := Spring{From: 0, To: 1, Start: start, Omega: 30}
	for i := range 200 {
		if x, _ := r.At(start.Add(time.Duration(i) * time.Millisecond)); x > 1 {
			t.Fatalf("overshoot %v at %d ms", x, i)
		}
	}
}

type fakeTimers struct {
	armed []func()
	now   time.Time
}

func (f *fakeTimers) after(_ time.Duration, fn func()) *time.Timer {
	f.armed = append(f.armed, fn)
	return time.NewTimer(time.Hour)
}

func newTestScheduler(level Level) (*Scheduler, *fakeTimers, *[]FrameMsg) {
	f := &fakeTimers{now: time.Unix(100, 0)}
	var sent []FrameMsg
	s := NewScheduler(16*time.Millisecond, level, func(m any) { sent = append(sent, m.(FrameMsg)) })
	s.after = f.after
	s.now = func() time.Time { return f.now }
	return s, f, &sent
}

func TestSchedulerKeepsOneFrameInFlight(t *testing.T) {
	s, f, sent := newTestScheduler(Full)
	s.Animate(100 * time.Millisecond)
	s.Animate(200 * time.Millisecond)
	if len(f.armed) != 1 {
		t.Fatalf("%d timers armed, want 1", len(f.armed))
	}
	frames := 0
	for len(f.armed) > 0 && frames < 100 {
		fire := f.armed[0]
		f.armed = f.armed[1:]
		fire()
		f.now = f.now.Add(16 * time.Millisecond)
		s.HandleFrame((*sent)[len(*sent)-1])
		frames++
	}
	if frames < 12 || frames > 15 {
		t.Fatalf("%d frames for a 200 ms animation", frames)
	}
	if s.Running() {
		t.Fatal("still running after the deadline")
	}
}

func TestSchedulerOffAndStale(t *testing.T) {
	s, f, _ := newTestScheduler(Off)
	s.Animate(time.Second)
	if len(f.armed) != 0 {
		t.Fatal("Off must not schedule frames")
	}
	s2, f2, sent := newTestScheduler(Full)
	s2.Animate(time.Second)
	f2.armed[0]()
	s2.Stop()
	if s2.HandleFrame((*sent)[0]) {
		t.Fatal("a frame from before Stop must be stale")
	}
}

func TestSchedulerSlowFrames(t *testing.T) {
	s, f, sent := newTestScheduler(Full)
	var intervals []time.Duration
	s.after = func(d time.Duration, fn func()) *time.Timer {
		intervals = append(intervals, d)
		f.armed = append(f.armed, fn)
		return time.NewTimer(time.Hour)
	}
	s.AnimateSlow(time.Second)
	s.Animate(50 * time.Millisecond)
	for len(f.armed) > 0 && len(intervals) < 200 {
		fire := f.armed[0]
		f.armed = f.armed[1:]
		fire()
		f.now = f.now.Add(intervals[len(intervals)-1])
		s.HandleFrame((*sent)[len(*sent)-1])
	}
	fast, slow := 0, 0
	for _, d := range intervals {
		if d == SlowInterval {
			slow++
		} else {
			fast++
		}
	}
	// A second of slow frames, less the fast start.
	want := int((time.Second - 50*time.Millisecond) / SlowInterval)
	if fast < 3 || slow < want-3 || slow > want+3 {
		t.Fatalf("%d fast and %d slow frames, want about %d slow", fast, slow, want)
	}
}
