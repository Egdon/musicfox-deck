package deck

import (
	"image/color"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

var testPal = Palette{
	Bg:        color.RGBA{0x15, 0x14, 0x12, 0xFF},
	Surface:   color.RGBA{0x1D, 0x1C, 0x19, 0xFF},
	Text:      color.RGBA{0xCB, 0xC4, 0xB6, 0xFF},
	Hero:      color.RGBA{0xDD, 0xD6, 0xC9, 0xFF},
	Secondary: color.RGBA{0x9E, 0x97, 0x8A, 0xFF},
	Rule:      color.RGBA{0x54, 0x4F, 0x47, 0xFF},
	Accent:    color.RGBA{0xD2, 0x71, 0x4F, 0xFF},
}

func TestOctantTable(t *testing.T) {
	var generated []rune
	for m, g := range octantGlyphs {
		if octantMissing[uint8(m)] {
			if g != "" {
				t.Errorf("mask %#x has no character but got %q", m, g)
			}
			continue
		}
		if g == "" {
			t.Fatalf("mask %#x has no glyph", m)
		}
		if r := []rune(g)[0]; r >= 0x1CD00 && r <= 0x1CDE5 {
			generated = append(generated, r)
		}
	}
	if len(generated) != 0x1CDE6-0x1CD00 {
		t.Fatalf("generated %d octants, want %d", len(generated), 0x1CDE6-0x1CD00)
	}
	for i, r := range generated {
		if r != rune(0x1CD00+i) {
			t.Fatalf("octant %d is %U, want %U", i, r, 0x1CD00+i)
		}
	}
	for mask, want := range map[uint8]rune{0x04: 0x1CD00, 0x17: 0x1CD0F, 0xFE: 0x1CDE5} {
		if got := []rune(octantGlyphs[mask])[0]; got != want {
			t.Errorf("mask %#x is %U, want %U", mask, got, want)
		}
	}
}

func TestSextantTable(t *testing.T) {
	if sextantGlyphs[21] != "▌" || sextantGlyphs[42] != "▐" || sextantGlyphs[0] != " " || sextantGlyphs[63] != "█" {
		t.Fatalf("existing blocks are wrong: %q %q %q %q", sextantGlyphs[21], sextantGlyphs[42], sextantGlyphs[0], sextantGlyphs[63])
	}
	if sextantGlyphs[1] != "\U0001FB00" || sextantGlyphs[62] != "\U0001FB3B" {
		t.Fatalf("range ends: %U %U", []rune(sextantGlyphs[1])[0], []rune(sextantGlyphs[62])[0])
	}
}

// Both width methods the renderer uses must count every glyph as one cell, or
// rows would shift.
func TestGlyphsAreOneCellWide(t *testing.T) {
	var all []string
	all = append(all, octantGlyphs[:]...)
	all = append(all, sextantGlyphs[:]...)
	all = append(all, quadrantGlyphs[:]...)
	all = append(all, halfGlyphs[:]...)
	for _, g := range all {
		if g == "" {
			continue
		}
		if w := ansi.StringWidthWc(g); w != 1 {
			t.Errorf("%q (%U) wcwidth %d", g, []rune(g)[0], w)
		}
		if w := ansi.StringWidth(g); w != 1 {
			t.Errorf("%q (%U) grapheme width %d", g, []rune(g)[0], w)
		}
	}
}

func TestFitCellReproducesTwoColorCells(t *testing.T) {
	fg, bg := rgb{210, 113, 79}, rgb{21, 20, 18}
	for _, tier := range []Tier{Half, Quadrant, Sextant, Octant} {
		cols, rows := tier.Grid()
		n := cols * rows
		for mask := 1; mask < 1<<n-1; mask++ {
			px := make([]rgb, n)
			for i := range px {
				px[i] = bg
				if mask>>i&1 != 0 {
					px[i] = fg
				}
			}
			c := fitCell(px, tier)
			// Draw the glyph back and compare.
			got := make([]rgb, n)
			gm := glyphMask(t, tier, c.Ch)
			for i := range got {
				got[i] = rgbOf(c.Bg)
				if gm>>i&1 != 0 {
					got[i] = rgbOf(c.Fg)
				}
			}
			for i := range px {
				if got[i].dist2(px[i]) > 3 {
					t.Fatalf("%s mask %#x: sub-pixel %d is %v, want %v (glyph %q)", tier, mask, i, got[i], px[i], c.Ch)
				}
			}
		}
	}
}

func glyphMask(t *testing.T, tier Tier, g string) int {
	cols, rows := tier.Grid()
	for m := range 1 << (cols * rows) {
		if glyphFor(tier, uint8(m)) == g {
			return m
		}
	}
	t.Fatalf("%s has no glyph %q", tier, g)
	return 0
}

// sdShell is the exact signed distance to the shell, for checking the
// analytic intersection by ray marching.
func sdShell(p vec3) float64 {
	d2 := roundRectSD(p.x+faceW/2, p.y+faceH/2, faceW, faceH, radius)
	dz := math.Abs(p.z) - depth/2
	return math.Hypot(max(d2, 0), max(dz, 0)) + min(max(d2, dz), 0)
}

func TestIntersectShellMatchesMarching(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	checked := 0
	for range 20000 {
		o := vec3{rng.Float64()*200 - 100, rng.Float64()*160 - 80, -60 - rng.Float64()*100}
		target := vec3{rng.Float64()*90 - 45, rng.Float64()*60 - 30, rng.Float64()*12 - 6}
		d := target.sub(o).norm()
		// March.
		tm, hit := 0.0, false
		for range 400 {
			dist := sdShell(o.at(d, tm))
			if dist < 1e-4 {
				hit = true
				break
			}
			tm += dist
			if tm > 400 {
				break
			}
		}
		h, ok := intersectShell(o, d)
		if hit != ok {
			// Grazing rays can go either way within the march tolerance.
			if near := sdShell(o.at(d, tm)); near > 1e-2 && hit {
				t.Fatalf("hit mismatch: march %v analytic %v (o=%v d=%v)", hit, ok, o, d)
			}
			continue
		}
		if ok && math.Abs(h.t-tm) > 1e-2 {
			t.Fatalf("t = %v, marched %v (o=%v d=%v surf=%v)", h.t, tm, o, d, h.surf)
		}
		checked++
	}
	if checked < 15000 {
		t.Fatalf("only %d rays checked", checked)
	}
}

// At rest the A face shows the palette's material colors exactly.
func TestRestFaceIsUnshaded(t *testing.T) {
	m := materialsFor(testPal)
	pose := RestPose()
	rot := rotation(pose.Yaw, pose.Pitch, pose.Roll)
	sc := &scene{m: &m, rot: rot, inv: 1, o: rot.applyT(vec3{0, 0, camDist}.mul(-1)), lObj: rot.applyT(keyLight), detail: 1}
	// Aim at the label paper between the letter and the window.
	target := rot.apply(vec3{30 - faceW/2, 14.5 - faceH/2, -depth / 2}).add(vec3{0, 0, camDist})
	c, mt, ok := sc.sample(target)
	if !ok || mt != matLabel {
		t.Fatalf("hit %v material %v, want the label", ok, mt)
	}
	if c.dist2(m.label) > 1 {
		t.Fatalf("label at rest is %v, want %v", c, m.label)
	}
}

func TestRenderIsDeterministicAndStaysInView(t *testing.T) {
	st := State{Radii: PackRadii(0.4, false), Angles: [2]float64{1, 2}, Lift: restLift}
	v := View{W: 50, H: 18, Tier: Octant}
	var a, b Renderer
	fa := a.Render(v, testPal, RestPose(), st)
	fb := b.Render(v, testPal, RestPose(), st)
	drawn := 0
	for i := range fa.Cells {
		if fa.Cells[i] != fb.Cells[i] {
			t.Fatalf("cell %d differs between renders", i)
		}
		if fa.Cells[i].Ch != "" {
			drawn++
		}
	}
	if drawn < len(fa.Cells)/3 {
		t.Fatalf("only %d of %d cells drawn", drawn, len(fa.Cells))
	}
	for x := range v.W {
		if fa.Cells[x].Ch != "" || fa.Cells[(v.H-1)*v.W+x].Ch != "" {
			t.Fatalf("the cassette touches the top or bottom row at column %d", x)
		}
	}
	// The fit box confines the cassette.
	box := Rect{X: 30, Y: 4, W: 18, H: 8}
	f := a.Render(View{W: 50, H: 18, Tier: Octant, Fit: box}, testPal, RestPose(), st)
	for y := range v.H {
		for x := range v.W {
			if f.Cells[y*v.W+x].Ch != "" && (x < 28 || y < 2) {
				t.Fatalf("cell (%d,%d) drawn outside the fit box", x, y)
			}
		}
	}
}

// Cells only the shadow falls on say how dark it is; the cassette's own
// cells do not.
// Turning the reels redraws only the hubs, which must give the frame a full
// render would.
func TestTurningReelsDrawsLikeAFullRender(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 5))
	sizes := [][2]int{{40, 14}, {64, 22}, {20, 6}, {96, 32}}
	tiers := []Tier{Octant, Sextant, Quadrant, Half}
	for i := range 64 {
		sz := sizes[i%len(sizes)]
		v := View{W: sz[0], H: sz[1], Tier: tiers[i/len(sizes)%len(tiers)]}
		pose := RestPose()
		pose.Yaw += (rng.Float64()*2 - 1) * math.Pi
		pose.Pitch += (rng.Float64()*2 - 1) * 0.3
		pose.Roll = (rng.Float64()*2 - 1) * 0.1
		st := State{Radii: ReelRadii(rng.Float64()), Lift: 2}
		var turned Renderer
		for range 6 {
			st.Angles = [2]float64{rng.Float64() * 2 * math.Pi, rng.Float64() * 2 * math.Pi}
			st.Blur = []float64{0, 0.3, 1}[rng.IntN(3)]
			got := turned.Render(v, testPal, pose, st)
			var fresh Renderer
			if want := fresh.Render(v, testPal, pose, st); !slices.Equal(got.Cells, want.Cells) {
				t.Fatalf("case %d: turning the reels drew another frame than a full render", i)
			}
		}
	}
}

func TestShadowCellsAreMarked(t *testing.T) {
	st := State{Radii: PackRadii(0.4, false), Lift: restLift}
	v := View{W: 50, H: 18, Tier: Octant}
	var r Renderer
	f := r.Render(v, testPal, RestPose(), st)
	x0, y0, x1, y1 := Bounds(v, RestPose())
	shadows := 0
	for i, c := range f.Cells {
		x, y := float64(i%v.W)+0.5, float64(i/v.W)+0.5
		if x > x0+1 && x < x1-1 && y > y0+1 && y < y1-1 && c.Shadow != 0 {
			t.Fatalf("cell (%v,%v) on the cassette is marked as shadow", x, y)
		}
		if c.Shadow < 0 || c.Shadow > shadowDepth {
			t.Fatalf("cell (%v,%v): shadow %v", x, y, c.Shadow)
		}
		if c.Shadow > 0 {
			shadows++
		}
	}
	if shadows == 0 {
		t.Fatal("no cell is marked as shadow")
	}
}

func TestPackRadiiConserveArea(t *testing.T) {
	for _, p := range []float64{0, 0.25, 0.5, 1} {
		r := ReelRadii(p)
		if area := r[0]*r[0] + r[1]*r[1]; math.Abs(area-(hubR*hubR+packR*packR)) > 1e-9 {
			t.Errorf("progress %v: area %v", p, area)
		}
	}
	if a, b := PackRadii(0.2, false), PackRadii(0.2, true); a[0] != b[1] || a[1] != b[0] {
		t.Errorf("the B side does not swap supply and take-up: %v %v", a, b)
	}
}

func TestReelsSpinUpAndDown(t *testing.T) {
	var m Reels
	now := time.Unix(0, 0)
	radii := ReelRadii(0.5)
	for range 6 { // 0.3 s
		m.Advance(0.05, true, now, radii, 1)
	}
	if m.speed < 0.8 {
		t.Fatalf("speed after 0.3 s is %v", m.speed)
	}
	for range 40 {
		m.Advance(0.05, false, now, radii, 1)
	}
	if m.Moving() {
		t.Fatalf("still moving after 2 s: %v %v", m.speed, m.vel)
	}
}

func TestMotionFloatsAndSettles(t *testing.T) {
	m := Motion{FloatAmount: 1}
	now := time.Unix(1_800_000_000, 0)
	for i := range 60 {
		m.Step(now.Add(time.Duration(i)*50*time.Millisecond), true)
	}
	if m.amp < 0.8 {
		t.Fatalf("float amplitude after 3 s is %v", m.amp)
	}
	now = now.Add(3 * time.Second)
	for i := range 100 {
		m.Step(now.Add(time.Duration(i)*50*time.Millisecond), false)
	}
	if m.Busy(now.Add(5 * time.Second)) {
		t.Fatalf("still moving 5 s after pausing: amp %v", m.amp)
	}
	if p, _ := m.At(now.Add(5 * time.Second)); math.Abs(p.Yaw-restYaw) > 1e-6 || math.Abs(p.Pitch-restPitch) > 1e-6 {
		t.Fatalf("settled pose %+v", p)
	}
}

func TestMotionFlipTurnsOver(t *testing.T) {
	var m Motion
	now := time.Unix(0, 0)
	m.Step(now, false)
	m.Flip(1, now)
	if !m.Flipping(now.Add(FlipDuration / 2)) {
		t.Fatal("not flipping halfway")
	}
	end := now.Add(FlipDuration)
	m.Step(end, false)
	if !m.SideB(end) || m.Busy(end) {
		t.Fatalf("after one flip: side B %v busy %v", m.SideB(end), m.Busy(end))
	}
	m.Flip(-1, end)
	m.Step(end.Add(FlipDuration), false)
	if m.SideB(end.Add(FlipDuration)) {
		t.Fatal("two flips should show the A side again")
	}
}

func TestFlipEaseTurnsEvenly(t *testing.T) {
	prev, peak := 0.0, 0.0
	const n = 1000
	for i := 1; i <= n; i++ {
		x := flipEase(float64(i) / n)
		if x < prev {
			t.Fatalf("the flip turns back at %v", float64(i)/n)
		}
		peak = max(peak, (x-prev)*n)
		prev = x
	}
	if prev != 1 || peak > 1.3 {
		t.Fatalf("ends at %v, peak pace %v", prev, peak)
	}
	if flipEase(0.2) < 0.1 {
		t.Fatalf("a fifth in, the flip has turned only %v", flipEase(0.2))
	}
}

// A second flip during one under way neither jumps nor stacks up: the same
// way it carries on, the other way it turns back from where it is.
func TestFlipDuringFlipIsContinuous(t *testing.T) {
	now := time.Unix(0, 0)
	for _, dir := range []int{1, -1} {
		var m Motion
		m.Step(now, false)
		m.Flip(1, now)
		mid := now.Add(FlipDuration * 2 / 5)
		before, liftBefore := m.At(mid)
		m.Flip(dir, mid)
		after, liftAfter := m.At(mid)
		if math.Abs(after.Yaw-before.Yaw) > 1e-9 || math.Abs(liftAfter-liftBefore) > 1e-9 || math.Abs(after.Scale-before.Scale) > 1e-9 {
			t.Fatalf("dir %d: the pose jumps from %+v to %+v", dir, before, after)
		}
		end := mid.Add(2 * FlipDuration)
		m.Step(end, false)
		if want := dir == 1; m.SideB(end) != want || m.Busy(end) {
			t.Fatalf("dir %d: ends on side B %v, want %v", dir, m.SideB(end), want)
		}
		if p, lift := m.At(end); math.Abs(p.Scale-1) > 1e-9 || math.Abs(lift-restLift) > 1e-9 {
			t.Fatalf("dir %d: does not settle flat: scale %v lift %v", dir, p.Scale, lift)
		}
	}
}

func TestDetectTier(t *testing.T) {
	env := func(kv ...string) func(string) string {
		m := map[string]string{}
		for i := 0; i+1 < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) string { return m[k] }
	}
	cases := []struct {
		name string
		term Terminal
		want Tier
	}{
		{"ghostty xtversion", Terminal{Version: "ghostty 1.3.1"}, Octant},
		{"old ghostty", Terminal{Version: "ghostty 1.0.1"}, Sextant},
		{"kitty", Terminal{Version: "kitty(0.40.1)"}, Octant},
		{"old kitty", Terminal{Version: "kitty(0.35.2)"}, Sextant},
		{"foot", Terminal{Version: "foot(1.21.0)"}, Octant},
		{"tmux answers", Terminal{Version: "tmux 3.5a"}, Quadrant},
		{"wezterm", Terminal{Version: "WezTerm 20240203-110809-5046fc22"}, Sextant},
		{"old iterm2", Terminal{Version: "iTerm2 3.5.4"}, Quadrant},
		{"ghostty with generic TERM", Terminal{Getenv: env("TERM", "xterm-256color", "TERM_PROGRAM", "ghostty", "TERM_PROGRAM_VERSION", "1.3.1")}, Octant},
		{"alacritty launched from ghostty", Terminal{Getenv: env("TERM", "alacritty", "TERM_PROGRAM", "ghostty")}, Sextant},
		{"inside tmux", Terminal{Getenv: env("TERM", "tmux-256color", "TMUX", "/tmp/tmux", "TERM_PROGRAM", "tmux")}, Quadrant},
		{"vte", Terminal{Getenv: env("TERM", "xterm-256color", "VTE_VERSION", "8002")}, Octant},
		{"old vte", Terminal{Getenv: env("TERM", "xterm-256color", "VTE_VERSION", "6800")}, Sextant},
		{"unknown", Terminal{Getenv: env("TERM", "xterm-256color")}, Quadrant},
		{"unknown xtversion falls back to env", Terminal{Version: "XTerm(390)", Getenv: env("TERM", "xterm-kitty")}, Octant},
	}
	for _, c := range cases {
		if got := DetectTier(c.term); got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
	}
}

func benchmarkRender(b *testing.B, w, h int, tier Tier) {
	var r Renderer
	var m Motion
	m.FloatAmount = 1
	now := time.Unix(1_800_000_000, 0)
	m.Step(now, true)
	b.ResetTimer()
	for i := range b.N {
		t := now.Add(time.Duration(i) * 33 * time.Millisecond)
		m.Step(t, true)
		pose, lift := m.At(t)
		r.Render(View{W: w, H: h, Tier: tier}, testPal, pose, State{Radii: ReelRadii(0.4), Angles: [2]float64{float64(i) * 0.1, float64(i) * 0.2}, Lift: lift})
	}
}

// benchmarkTurnReels draws a cassette that holds still while its reels turn.
func benchmarkTurnReels(b *testing.B, w, h int, tier Tier) {
	var r Renderer
	st := State{Radii: ReelRadii(0.4), Lift: RestLift}
	b.ResetTimer()
	for i := range b.N {
		st.Angles = [2]float64{float64(i) * 0.1, float64(i) * 0.2}
		r.Render(View{W: w, H: h, Tier: tier}, testPal, RestPose(), st)
	}
}

func BenchmarkTurnReelsPageOctant(b *testing.B)  { benchmarkTurnReels(b, 64, 22, Octant) }
func BenchmarkTurnReelsLargeOctant(b *testing.B) { benchmarkTurnReels(b, 96, 32, Octant) }

func BenchmarkRenderPageOctant(b *testing.B)  { benchmarkRender(b, 64, 22, Octant) }
func BenchmarkRenderPageSextant(b *testing.B) { benchmarkRender(b, 64, 22, Sextant) }
func BenchmarkRenderPanelOctant(b *testing.B) { benchmarkRender(b, 40, 14, Octant) }
func BenchmarkRenderLargeOctant(b *testing.B) { benchmarkRender(b, 96, 32, Octant) }
