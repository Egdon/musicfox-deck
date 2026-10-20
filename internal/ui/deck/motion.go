package deck

import (
	"math"
	"time"
)

// spring advances a critically damped spring by dt towards target.
func spring(x, v, target, omega, dt float64) (float64, float64) {
	d := x - target
	e := math.Exp(-omega * dt)
	return target + (d+(v+omega*d)*dt)*e, (v - omega*(v+omega*d)*dt) * e
}

func deg(d float64) float64 { return d * math.Pi / 180 }

// ---- reels ----

const windDuration = 400 * time.Millisecond

// Reels turns the reels: a spring eases the speed up when playback starts
// and down when it stops, and a seek winds the tape fast for a moment.
type Reels struct {
	Angles    [2]float64
	speed     float64 // 1 is playing speed, negative rewinds
	vel       float64
	turn      float64 // largest turn of the last step, radians
	windStart time.Time
	windDir   float64
	windFrom  float64
	windTo    float64
}

// Reset loads a new cassette: its reels rest in their own position.
func (m *Reels) Reset(seed int64) {
	m.speed, m.vel = 0, 0
	m.Angles = [2]float64{math.Mod(float64(seed)*0.37, 2*math.Pi), math.Mod(float64(seed)*0.71, 2*math.Pi)}
}

// Run sets the reels turning at playing speed without a spin-up.
func (m *Reels) Run() { m.speed, m.vel = 1, 0 }

// Stop halts the reels at once.
func (m *Reels) Stop() { m.speed, m.vel = 0, 0 }

func (m *Reels) winding(now time.Time) bool { return now.Sub(m.windStart) < windDuration }

// Advance turns the reels by dt seconds. Both carry the tape at the same
// speed, so the emptier reel spins faster. dir is 1 while the A side plays
// and -1 for the B side, which turns the other way in object space.
func (m *Reels) Advance(dt float64, playing bool, now time.Time, radii [2]float64, dir float64) {
	target, omega := 0.0, 7.0 // spin down in about 0.8 s
	if playing {
		target = 1
	}
	if m.winding(now) {
		target = 9 * m.windDir
	}
	if math.Abs(target) > math.Abs(m.speed) {
		omega = 13 // spin up in about 0.3 s
	}
	m.speed, m.vel = spring(m.speed, m.vel, target, omega, dt)
	if math.Abs(m.speed) < 0.002 && math.Abs(m.vel) < 0.002 && target == 0 {
		m.speed, m.vel = 0, 0
	}
	m.turn = 0
	for i := range m.Angles {
		d := m.speed * tapeSpeed / max(radii[i], hubR) * dt
		m.Angles[i] = math.Mod(m.Angles[i]+dir*d, 2*math.Pi)
		m.turn = max(m.turn, math.Abs(d))
	}
}

// Moving reports whether the reels still turn.
func (m *Reels) Moving() bool { return m.speed != 0 || m.vel != 0 }

// Blur blends the hub teeth into a ring once a step turns more than 15°,
// beyond which the eye can no longer follow them.
func (m *Reels) Blur() float64 {
	return min(max((m.turn-deg(15))/deg(30), 0), 1)
}

// Wind starts a seek: the reels spin fast in dir while the packs travel from
// one progress to the other.
func (m *Reels) Wind(dir float64, now time.Time, from, to float64) {
	m.windStart, m.windDir, m.windFrom, m.windTo = now, dir, from, to
}

// Progress is the progress the tape packs show: while winding they travel
// from where the seek started to where it lands.
func (m *Reels) Progress(now time.Time, real float64) float64 {
	if !m.winding(now) {
		return real
	}
	t := float64(now.Sub(m.windStart)) / float64(windDuration)
	return m.windFrom + (m.windTo-m.windFrom)*(1-math.Pow(1-t, 3))
}

// PackRadii returns the radius of each reel in object space. On the B side
// the tape runs from the other reel.
func PackRadii(progress float64, sideB bool) [2]float64 {
	r := ReelRadii(progress)
	if sideB {
		r[0], r[1] = r[1], r[0]
	}
	return r
}

// ---- pose ----

// The float: incommensurate periods so that it never visibly repeats.
var (
	floatYaw, floatPitch, floatRoll = deg(6), deg(3), deg(1.2)
	floatBob                        = 0.6 // face units
	leanYaw, leanPitch              = deg(10), deg(6)
)

const (
	leanIdle     = 2500 * time.Millisecond // the cassette straightens up after this
	FlipDuration = 650 * time.Millisecond
	flipLift     = 3.0 // face units
	restLift     = 2.0
)

// RestLift is how high the cassette rests above the backdrop, in face units.
const RestLift = restLift

// Motion moves the cassette as a whole: it floats while music plays, leans
// towards the pointer, settles when paused and flips over to change tracks.
type Motion struct {
	// FloatAmount scales the float: 1 on the page, less in a panel, 0
	// keeps the cassette still.
	FloatAmount float64

	amp, ampV float64 // float amplitude, eased in and out
	lean      [2]float64
	leanV     [2]float64
	leanTo    [2]float64 // yaw and pitch towards the pointer
	pointerAt time.Time
	base      float64 // rest yaw; the B side faces the viewer after an odd number of flips
	flipAt    time.Time
	flipBy    float64 // yaw the flip under way turns by: ±π, or less to turn back halfway
	last      time.Time
}

// Step advances the motion to now.
func (m *Motion) Step(now time.Time, playing bool) {
	dt := 0.0
	if !m.last.IsZero() {
		dt = min(max(now.Sub(m.last).Seconds(), 0), 0.1)
	}
	m.last = now
	target := 0.0
	if playing {
		target = m.FloatAmount
	}
	m.amp, m.ampV = spring(m.amp, m.ampV, target, 2.2, dt)
	if target == 0 && math.Abs(m.amp) < 1e-3 && math.Abs(m.ampV) < 1e-3 {
		m.amp, m.ampV = 0, 0
	}
	if !m.pointerAt.IsZero() && now.Sub(m.pointerAt) > leanIdle {
		m.leanTo = [2]float64{}
	}
	for i := range m.lean {
		m.lean[i], m.leanV[i] = spring(m.lean[i], m.leanV[i], m.leanTo[i], 6, dt)
		if m.leanTo[i] == 0 && math.Abs(m.lean[i]) < 1e-4 && math.Abs(m.leanV[i]) < 1e-4 {
			m.lean[i], m.leanV[i] = 0, 0
		}
	}
	if !m.flipAt.IsZero() && now.Sub(m.flipAt) >= FlipDuration {
		m.base = normAngle(m.base + m.flipBy)
		m.flipAt = time.Time{}
	}
}

// Point leans the cassette towards a pointer at (nx, ny), each from -1 to 1
// across the area around it.
func (m *Motion) Point(nx, ny float64, now time.Time) {
	nx, ny = min(max(nx, -1), 1), min(max(ny, -1), 1)
	m.leanTo = [2]float64{-nx * leanYaw, ny * leanPitch}
	m.pointerAt = now
}

// Leave lets the cassette straighten up.
func (m *Motion) Leave() { m.leanTo = [2]float64{} }

// Flip starts turning the cassette over, to the right for dir > 0. A flip
// already under way the same way carries on, and one the other way turns
// back from where it is: starting over would make the cassette jump.
func (m *Motion) Flip(dir int, now time.Time) {
	by := math.Copysign(math.Pi, float64(dir))
	if m.Flipping(now) {
		if math.Signbit(by) == math.Signbit(m.flipBy) {
			return
		}
		turned := m.flipBy * m.flipProgress(now)
		m.base += turned
		by = -turned
	} else if !m.flipAt.IsZero() {
		m.base = normAngle(m.base + m.flipBy)
	}
	m.flipAt, m.flipBy = now, by
}

// FlipNow turns the cassette over without animation.
func (m *Motion) FlipNow() {
	m.flipAt = time.Time{}
	m.base = normAngle(m.base + math.Pi)
}

// Flipping reports whether a flip is under way.
func (m *Motion) Flipping(now time.Time) bool {
	return !m.flipAt.IsZero() && now.Sub(m.flipAt) < FlipDuration
}

func (m *Motion) flipProgress(now time.Time) float64 {
	if m.flipAt.IsZero() {
		return 0
	}
	return flipEase(min(float64(now.Sub(m.flipAt))/float64(FlipDuration), 1))
}

// flipEase speeds the flip up over its first fifth, turns at an even pace,
// and brakes over its last fifth. An ease-in-out would sit nearly still at
// first and then whip through the middle, which reads as a twitch.
func flipEase(t float64) float64 {
	const a = 0.2
	v := 1 / (1 - a) // the even pace
	switch {
	case t <= 0:
		return 0
	case t < a:
		return v * t * t / (2 * a)
	case t <= 1-a:
		return v * (t - a/2)
	case t < 1:
		return 1 - v*(1-t)*(1-t)/(2*a)
	}
	return 1
}

// At returns the pose at now and how high the cassette floats above the
// backdrop.
func (m *Motion) At(now time.Time) (Pose, float64) {
	turn := m.base + m.flipBy*m.flipProgress(now)
	arc := math.Abs(math.Sin(turn)) // how far from lying flat; base rests on a multiple of π
	t := float64(now.UnixNano()) / 1e9
	wave := func(period, phase float64) float64 { return math.Sin(2*math.Pi*t/period + phase) }
	bob := m.amp * floatBob * wave(7.3, 0)
	p := Pose{
		Yaw:   restYaw + turn + m.amp*floatYaw*wave(9, 0) + m.lean[0],
		Pitch: restPitch + m.amp*floatPitch*wave(12.7, 1.3) + m.lean[1] + deg(5)*arc,
		Roll:  m.amp * floatRoll * wave(15.1, 0.4),
		Y:     bob - flipLift*arc,
		Scale: 1 - 0.05*arc,
	}
	return p, restLift - bob + flipLift*arc
}

// SideB reports whether the B side faces the viewer at now.
func (m *Motion) SideB(now time.Time) bool {
	p, _ := m.At(now)
	return math.Cos(p.Yaw) < 0
}

// Busy reports whether the cassette is still moving.
func (m *Motion) Busy(now time.Time) bool {
	return m.amp != 0 || m.ampV != 0 || m.lean != [2]float64{} || m.leanTo != [2]float64{} || m.Flipping(now)
}

// Brisk reports whether the cassette moves fast enough to be drawn on every
// frame: flipping over, or leaning towards or away from the pointer. Its
// float alone is slow.
func (m *Motion) Brisk(now time.Time) bool {
	return m.Flipping(now) || m.lean != m.leanTo || m.leanV != [2]float64{}
}

func normAngle(a float64) float64 {
	a = math.Mod(a, 2*math.Pi)
	if a > math.Pi {
		a -= 2 * math.Pi
	} else if a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}
