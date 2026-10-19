package ui

import (
	"image/color"
	"math"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// The deck is the cell rectangle reserved for the cassette. The cassette is
// shaded in sub-pixels (2×4 per cell) and fitted back into characters.
const (
	npDeckW = 42 // cells
	npDeckH = 15
	npSubW  = npDeckW * 2
	npSubH  = npDeckH * 4

	npFaceW = 76.0 // cassette face in sub-pixels
	npFaceH = 48.0
	// The face origin stays on the cell grid so the label text lines up.
	npFaceX = 2.0
	npFaceY = 4.0

	npShellDepth  = 9.0
	npShellRadius = 3.0
	npWindowR     = 8.5
	npReelDepth   = 3.0 // the reels sit this far behind the window

	npToothR  = 1.6 // spindle hole; six teeth reach in to this radius
	npHoleR   = 3.2
	npReelHub = 6.5 // hub ring; also the pack radius of an empty reel
	npReelMin = npReelHub
	npReelMax = 15.0

	// npTapeSpeed is the tape speed at the pack surface in sub-pixels per
	// second: an empty reel turns at about 41 rpm, a full one at about 18.
	npTapeSpeed = 27.9
)

// Label text cells, in deck cells: rows npLabelRow and npLabelRow+1 are flat
// label paper at rest.
const (
	npLabelRow   = 2
	npLabelLeft  = 4
	npLabelRight = 36 // exclusive
)

// npDepth is the oblique projection of depth: every unit backwards moves a
// point up and to the left, showing the lit top and left walls while the face
// stays square.
var npDepth = [2]float64{-0.2, -0.3}

var (
	npReelCenters = [2][2]float64{{22, 24}, {54, 24}}
	npScrews      = [2][2]float64{{3.6, 44.4}, {72.4, 44.4}}
	npAA          = [4][2]float64{{0.25, 0.25}, {0.75, 0.25}, {0.25, 0.75}, {0.75, 0.75}}
)

// npReelRadii returns the tape pack radii of the supply (left) and take-up
// (right) reels. Tape area is conserved, so one pack shrinks as the other grows.
func npReelRadii(progress float64) [2]float64 {
	p := min(max(progress, 0), 1)
	span := npReelMax*npReelMax - npReelMin*npReelMin
	return [2]float64{
		math.Sqrt(npReelMin*npReelMin + (1-p)*span),
		math.Sqrt(npReelMin*npReelMin + p*span),
	}
}

// npAngleNear reports whether angle a lies within tol of target modulo step.
func npAngleNear(a, target, step, tol float64) bool {
	d := math.Mod(a-target, step)
	if d < 0 {
		d += step
	}
	return d < tol || step-d < tol
}

func npRoundRectSD(u, v, w, h, r float64) float64 {
	qx := math.Abs(u-w/2) - (w/2 - r)
	qy := math.Abs(v-h/2) - (h/2 - r)
	return math.Hypot(max(qx, 0), max(qy, 0)) + min(max(qx, qy), 0) - r
}

func npShellSD(u, v float64) float64 {
	return npRoundRectSD(u, v, npFaceW, npFaceH, npShellRadius)
}

func npShellNormal(u, v float64) (float64, float64) {
	const e = 0.05
	nx := npShellSD(u+e, v) - npShellSD(u-e, v)
	ny := npShellSD(u, v+e) - npShellSD(u, v-e)
	l := math.Hypot(nx, ny)
	if l == 0 {
		return 0, -1
	}
	return nx / l, ny / l
}

// npWindowDist is the distance to the axis of the stadium-shaped window.
func npWindowDist(u, v float64) float64 {
	x := min(max(u, npReelCenters[0][0]), npReelCenters[1][0])
	return math.Hypot(u-x, v-npReelCenters[0][1])
}

// npLambert shades a surface normal against a key light from the top left;
// the front face comes out at exactly 1.
func npLambert(nx, ny, nz float64) float64 {
	const lx, ly, lz = -0.4, -0.75, 0.53
	l := math.Sqrt(nx*nx + ny*ny + nz*nz)
	d := (lx*nx + ly*ny + lz*nz) / l
	return 0.35 + 0.65*max(d, 0)/lz
}

// npPose places the cassette during the eject and insert animations; the
// zero pose is the cassette at rest.
type npPose struct {
	DX, DY float64 // offset in sub-pixels
	Tilt   float64 // leaning back, radians
	Fade   float64 // 0 visible, 1 gone
}

type npDeckMaterials struct {
	bg, hero, shell, label, stripe, glass, tape, hub, hole, clamp, screw npRGB
	labelInk, labelSub                                                   color.RGBA
}

func npDeckMaterialsFor(p npPalette) npDeckMaterials {
	bg := npRGBOf(p.Bg)
	return npDeckMaterials{
		bg:       bg,
		hero:     npRGBOf(p.Hero),
		shell:    npRGBOf(p.Surface).mix(npRGBOf(p.Text), 0.10),
		label:    bg.mix(npRGBOf(p.Surface), 0.6),
		stripe:   npRGBOf(p.Accent),
		glass:    bg.mix(npRGBOf(p.Ghost), 0.3).scale(0.85),
		tape:     npRGBOf(p.Accent).mix(bg, 0.78),
		hub:      npRGBOf(p.Text).mix(bg, 0.28),
		hole:     bg.scale(0.45),
		clamp:    npRGBOf(p.Accent),
		screw:    npRGBOf(p.Rule).mix(npRGBOf(p.Text), 0.25),
		labelInk: p.Text,
		labelSub: p.Secondary,
	}
}

// npDeckScene shades the cassette at one moment.
type npDeckScene struct {
	m      npDeckMaterials
	pose   npPose
	radii  [2]float64
	angles [2]float64
	blur   float64
}

// toFace maps a deck sub-pixel position at depth z to face coordinates.
func (sc *npDeckScene) toFace(x, y, z float64) (float64, float64) {
	p := sc.pose
	cos, sin := math.Cos(p.Tilt), math.Sin(p.Tilt)
	u := x - npFaceX - p.DX - npDepth[0]*z
	v := (y-npFaceY-p.DY-(npDepth[1]-sin)*z-npFaceH/2)/cos + npFaceH/2
	return u, v
}

// sample shades deck position (x, y). ok is false where the background shows;
// dyn is true inside the window, whose content depends on the reels.
func (sc *npDeckScene) sample(x, y float64, reels bool) (c npRGB, ok, dyn bool) {
	u, v := sc.toFace(x, y, 0)
	if sd := npShellSD(u, v); sd <= 0 {
		return sc.front(x, y, u, v, sd, reels)
	}
	for k := 1; k <= 8; k++ {
		u, v = sc.toFace(x, y, npShellDepth*float64(k)/8)
		if npShellSD(u, v) <= 0 {
			nx, ny := npShellNormal(u, v)
			return sc.m.shell.scale(npLambert(nx, ny, 0)), true, false
		}
	}
	return sc.m.bg, false, false
}

func (sc *npDeckScene) front(x, y, u, v, sd float64, reels bool) (npRGB, bool, bool) {
	m := sc.m
	if sd > -1 { // bevel
		nx, ny := npShellNormal(u, v)
		c := m.shell.scale(npLambert(nx*0.7, ny*0.7, 0.7))
		if ny < -0.3 {
			c = c.mix(m.hero, 0.06)
		}
		return c, true, false
	}
	for _, s := range npScrews {
		if dx, dy := u-s[0], v-s[1]; dx*dx+dy*dy < 1.3 {
			if dx+dy < 0 {
				return m.screw.scale(1.15), true, false
			}
			return m.screw.scale(0.8), true, false
		}
	}
	dw := npWindowDist(u, v)
	if dw <= npWindowR {
		c, dyn := sc.recess(x, y, u, v, reels)
		return c, true, dyn
	}
	if dw <= npWindowR+0.9 {
		return m.shell.scale(0.78), true, false // window rim
	}
	if npRoundRectSD(u-4, v-3, 68, 32, 1.5) <= 0 {
		if v >= 12 && v < 13 && u >= 6 && u <= 70 {
			return m.stripe, true, false
		}
		return m.label, true, false
	}
	if v >= 37 { // raised head area with two capstan holes and the head opening
		spread := (v - 37) * 4 / 11
		if u >= 20-spread && u <= 56+spread {
			for _, hx := range [2]float64{26.5, 49.5} {
				if math.Hypot(u-hx, v-44) < 1.4 {
					return m.hole, true, false
				}
			}
			if u >= 35 && u <= 41 && v >= 44.5 {
				return m.hole, true, false
			}
			if v < 38 {
				return m.shell.scale(1.18), true, false
			}
			return m.shell.scale(1.05), true, false
		}
	}
	return m.shell, true, false
}

// recess shades the window: the reels behind clear smoked plastic, the lit
// inner walls on the lower right, and the shadow the upper left edge casts.
func (sc *npDeckScene) recess(x, y, u, v float64, reels bool) (npRGB, bool) {
	ur, vr := sc.toFace(x, y, npReelDepth)
	if npWindowDist(ur, vr) > npWindowR {
		return sc.m.shell.scale(1.12), false
	}
	shade := 1.0
	for i := 1; i <= 3; i++ {
		d := 0.8 * float64(i)
		if npWindowDist(ur-0.47*d, vr-0.88*d) > npWindowR {
			shade = 0.55 + 0.15*float64(i-1)
			break
		}
	}
	c := sc.m.glass
	if reels {
		c = sc.reel(ur, vr)
	}
	gloss := 0.07 * max(0, 1-math.Abs(u*0.45+v-31)/2.5)
	return c.scale(shade).mix(sc.m.hero, gloss), true
}

func npAngleDist(a, b float64) float64 {
	d := math.Mod(math.Abs(a-b), 2*math.Pi)
	return min(d, 2*math.Pi-d)
}

// reel shades the reel plane: spindle with six teeth, hub ring with the
// leader clamp, tape pack with grooves and a sheen that stays put.
func (sc *npDeckScene) reel(u, v float64) npRGB {
	m := sc.m
	for i, c := range npReelCenters {
		dx, dy := u-c[0], v-c[1]
		r := math.Hypot(dx, dy)
		if r >= sc.radii[i] {
			continue
		}
		ang := math.Atan2(dy, dx)
		rel := ang - sc.angles[i]
		switch {
		case r < npToothR:
			return m.hole
		case r < npHoleR:
			teeth := m.hub.scale(0.92)
			sharp := m.hole
			if npAngleNear(rel, 0, math.Pi/3, 11*math.Pi/180) {
				sharp = teeth
			}
			return sharp.mix(m.hole.mix(teeth, 0.37), sc.blur)
		case r < npReelHub:
			h := m.hub
			if r > npReelHub-0.8 || r < npHoleR+0.6 {
				h = h.scale(0.87)
			}
			if r > 4.2 && r < 5.9 && npAngleNear(rel, math.Pi, 2*math.Pi, 0.32) {
				h = h.mix(m.clamp, 1-0.9*sc.blur)
			}
			return h
		default:
			t := m.tape.scale(1 + 0.05*math.Sin(r*2.6))
			if s := npAngleDist(ang, -2.2); s < 0.55 {
				t = t.mix(m.hub, 0.12*(1-s/0.55))
			}
			if r > sc.radii[i]-0.8 {
				t = t.scale(1.12)
			}
			return t
		}
	}
	if math.Abs(v-31.4) < 0.55 && u > npReelCenters[0][0] && u < npReelCenters[1][0] {
		return m.tape.scale(1.1) // tape running along the head side
	}
	return m.glass
}

// npDeckState is what changes the cassette from frame to frame.
type npDeckState struct {
	Progress float64
	Angles   [2]float64
	Blur     float64
	Pose     npPose
}

// npDeckFrame is a rendered cassette ready to be copied into the screen.
type npDeckFrame struct {
	Cells                       [npDeckH * npDeckW]npCell
	Label                       bool // the label text is shown
	LabelBg, LabelInk, LabelSub color.RGBA
}

type npDeckKey struct {
	pal     npPalette
	octants bool
}

// npDeck renders the cassette, caching everything that does not move while it
// rests: per frame only the cells showing the window are shaded again.
type npDeck struct {
	key     npDeckKey
	ready   bool
	m       npDeckMaterials
	sub     [npSubH * npSubW]npRGB
	dyn     [npSubH * npSubW]bool
	cells   [npDeckH * npDeckW]npCell
	dynCell [npDeckH * npDeckW]bool
	out     npDeckFrame
}

func (d *npDeck) frame(pal npPalette, octants bool, st npDeckState) *npDeckFrame {
	key := npDeckKey{pal, octants}
	if !d.ready || d.key != key {
		d.build(key)
	}
	sc := &npDeckScene{m: d.m, pose: st.Pose, radii: npReelRadii(st.Progress), angles: st.Angles, blur: st.Blur}
	d.out.LabelBg, d.out.LabelInk, d.out.LabelSub = d.m.label.rgba(), d.m.labelInk, d.m.labelSub
	if st.Pose != (npPose{}) {
		d.renderMoving(sc, octants)
		return &d.out
	}
	d.out.Label = true
	d.out.Cells = d.cells
	var px [8]npRGB
	for ci, dyn := range d.dynCell {
		if !dyn {
			continue
		}
		cx, cy := ci%npDeckW, ci/npDeckW
		for k := range 8 {
			x, y := cx*2+k%2, cy*4+k/2
			i := y*npSubW + x
			if !d.dyn[i] {
				px[k] = d.sub[i]
				continue
			}
			var acc npRGB
			for _, o := range npAA {
				c, _, _ := sc.sample(float64(x)+o[0], float64(y)+o[1], true)
				acc = npRGB{acc.R + c.R, acc.G + c.G, acc.B + c.B}
			}
			px[k] = acc.scale(0.25)
		}
		d.out.Cells[ci] = npFitCell(&px, octants)
	}
	return &d.out
}

// build renders the static layer: the shell, the label, the empty window and
// a soft drop shadow, fitted into cells.
func (d *npDeck) build(key npDeckKey) {
	d.key, d.ready = key, true
	d.m = npDeckMaterialsFor(key.pal)
	sc := &npDeckScene{m: d.m}
	var cover [npSubH * npSubW]float64
	for y := range npSubH {
		for x := range npSubW {
			i := y*npSubW + x
			var acc npRGB
			d.dyn[i] = false
			for _, o := range npAA {
				c, ok, dyn := sc.sample(float64(x)+o[0], float64(y)+o[1], false)
				acc = npRGB{acc.R + c.R, acc.G + c.G, acc.B + c.B}
				if ok {
					cover[i] += 0.25
				}
				d.dyn[i] = d.dyn[i] || dyn
			}
			d.sub[i] = acc.scale(0.25)
		}
	}
	shadow := npShadow(&cover)
	var bgOnly [npSubH * npSubW]bool
	for i := range d.sub {
		if cover[i] < 1 && shadow[i] > 0 {
			darker := d.m.bg.scale(1 - 0.5*shadow[i])
			k := 1 - cover[i]
			d.sub[i] = npRGB{
				d.sub[i].R + (darker.R-d.m.bg.R)*k,
				d.sub[i].G + (darker.G-d.m.bg.G)*k,
				d.sub[i].B + (darker.B-d.m.bg.B)*k,
			}
		}
		bgOnly[i] = cover[i] == 0 && shadow[i] < 0.02
	}
	var px [8]npRGB
	for ci := range d.cells {
		cx, cy := ci%npDeckW, ci/npDeckW
		empty := true
		d.dynCell[ci] = false
		for k := range 8 {
			i := (cy*4+k/2)*npSubW + cx*2 + k%2
			px[k] = d.sub[i]
			empty = empty && bgOnly[i]
			d.dynCell[ci] = d.dynCell[ci] || d.dyn[i]
		}
		if empty {
			d.cells[ci] = npCell{}
			continue
		}
		d.cells[ci] = npFitCell(&px, key.octants)
	}
}

// npShadow offsets the coverage by (2, 2) sub-pixels and blurs it twice.
func npShadow(cover *[npSubH * npSubW]float64) [npSubH * npSubW]float64 {
	var a, b [npSubH * npSubW]float64
	for y := 2; y < npSubH; y++ {
		for x := 2; x < npSubW; x++ {
			a[y*npSubW+x] = cover[(y-2)*npSubW+x-2]
		}
	}
	blur := func(src, dst *[npSubH * npSubW]float64, dx, dy int) {
		for y := range npSubH {
			for x := range npSubW {
				var sum float64
				for k := -2; k <= 2; k++ {
					xx, yy := x+k*dx, y+k*dy
					if xx >= 0 && yy >= 0 && xx < npSubW && yy < npSubH {
						sum += src[yy*npSubW+xx]
					}
				}
				dst[y*npSubW+x] = sum / 5
			}
		}
	}
	for range 2 {
		blur(&a, &b, 1, 0)
		blur(&b, &a, 0, 1)
	}
	return a
}

// renderMoving renders an animated pose from scratch at one sample per
// sub-pixel; the label text is hidden and there is no shadow.
func (d *npDeck) renderMoving(sc *npDeckScene, octants bool) {
	d.out.Label = false
	var px [8]npRGB
	for ci := range d.out.Cells {
		cx, cy := ci%npDeckW, ci/npDeckW
		covered := false
		for k := range 8 {
			x, y := float64(cx*2+k%2)+0.5, float64(cy*4+k/2)+0.5
			c, ok, _ := sc.sample(x, y, true)
			if ok {
				covered = true
				c = c.mix(d.m.bg, sc.pose.Fade)
			}
			px[k] = c
		}
		if !covered {
			d.out.Cells[ci] = npCell{}
			continue
		}
		d.out.Cells[ci] = npFitCell(&px, octants)
	}
}

// npDrawDeck copies the cassette into the screen and, at rest, prints the
// label on its flat paper cells.
func npDrawDeck(scr *npScreen, x, y int, s *npSnapshot) {
	f := s.Deck
	if f == nil {
		return
	}
	for ci, c := range f.Cells {
		if c.Ch == "" {
			continue
		}
		st := uv.Style{Bg: c.Bg}
		if c.Ch != " " {
			st.Fg = c.Fg
		}
		scr.set(x+ci%npDeckW, y+ci/npDeckW, c.Ch, 1, st)
	}
	if !f.Label {
		return
	}
	ink := uv.Style{Fg: f.LabelInk, Bg: f.LabelBg}
	sub := uv.Style{Fg: f.LabelSub, Bg: f.LabelBg}
	left, right := x+npLabelLeft, x+npLabelRight
	row := y + npLabelRow

	tag := npTapeTag(s.TapeMins)
	n := scr.text(left, row, "A", uv.Style{Fg: f.LabelInk, Bg: f.LabelBg, Attrs: uv.AttrBold}, 1)
	scr.text(left+n+2, row, s.Song.Name, ink, right-left-n-2-npWidth(tag)-2)
	scr.text(right-npWidth(tag), row, tag, sub, npWidth(tag))

	tape, _ := npTapeType(s.Quality)
	scr.text(left+n+2, row+1, artistsString(s.Song), sub, right-left-n-2-npWidth(tape)-2)
	scr.text(right-npWidth(tape), row+1, tape, sub, npWidth(tape))
}

// ---- reel motion ----

// npSpring advances a critically damped spring by dt towards target.
func npSpring(x, v, target, omega, dt float64) (float64, float64) {
	d := x - target
	e := math.Exp(-omega * dt)
	return target + (d+(v+omega*d)*dt)*e, (v - omega*(v+omega*d)*dt) * e
}

const npWindDuration = 400 * time.Millisecond

// npReelMotion turns the reels: a spring eases the speed up when playback
// starts and down when it stops, and a seek winds the tape fast for a moment.
type npReelMotion struct {
	angles    [2]float64
	speed     float64 // 1 is playing speed, negative rewinds
	vel       float64
	turn      float64 // largest turn of the last step, radians
	windStart time.Time
	windDir   float64
	windFrom  float64
	windTo    float64
}

func (m *npReelMotion) winding(now time.Time) bool {
	return now.Sub(m.windStart) < npWindDuration
}

func (m *npReelMotion) advance(dt float64, playing bool, now time.Time, radii [2]float64) {
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
	m.speed, m.vel = npSpring(m.speed, m.vel, target, omega, dt)
	if math.Abs(m.speed) < 0.002 && math.Abs(m.vel) < 0.002 && target == 0 {
		m.speed, m.vel = 0, 0
	}
	m.turn = 0
	for i := range m.angles {
		d := m.speed * npTapeSpeed / radii[i] * dt
		m.angles[i] = math.Mod(m.angles[i]+d, 2*math.Pi)
		m.turn = max(m.turn, math.Abs(d))
	}
}

func (m *npReelMotion) moving() bool {
	return m.speed != 0 || m.vel != 0
}

// blur blends the hub teeth into a ring once a step turns more than 15°,
// beyond which the eye can no longer follow them.
func (m *npReelMotion) blur() float64 {
	return min(max((m.turn-15*math.Pi/180)/(30*math.Pi/180), 0), 1)
}

func (m *npReelMotion) wind(dir float64, now time.Time, from, to float64) {
	m.windStart, m.windDir, m.windFrom, m.windTo = now, dir, from, to
}

// progress is the progress the tape packs show: while winding they travel
// from where the seek started to where it lands.
func (m *npReelMotion) progress(now time.Time, real float64) float64 {
	if !m.winding(now) {
		return real
	}
	t := float64(now.Sub(m.windStart)) / float64(npWindDuration)
	return m.windFrom + (m.windTo-m.windFrom)*(1-math.Pow(1-t, 3))
}
