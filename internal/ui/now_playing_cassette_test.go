package ui

import (
	"math"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

func npTestDeckFrame(octants bool, pose npPose) *npDeckFrame {
	var d npDeck
	f := d.frame(npWarmNight, octants, npDeckState{Progress: 0.3, Angles: [2]float64{0.4, 1.9}, Pose: pose})
	cp := *f
	return &cp
}

func TestOctantTable(t *testing.T) {
	codes := 0
	for m, g := range npOctantGlyphs {
		mask := uint8(m)
		if npMissingOctants[mask] {
			if g != "" {
				t.Errorf("mask %#02x has no character but got %q", mask, g)
			}
			continue
		}
		if g == "" {
			t.Fatalf("mask %#02x has no glyph", mask)
		}
		if _, ok := npExistingBlocks[mask]; !ok {
			codes++
		}
		if gw, wc := ansi.GraphemeWidth.StringWidth(g), ansi.WcWidth.StringWidth(g); gw != 1 || wc != 1 {
			t.Errorf("glyph %q for mask %#02x has width grapheme=%d wc=%d, want 1", g, mask, gw, wc)
		}
	}
	if codes != 0x1CDE5-0x1CD00+1 {
		t.Errorf("octant codepoints = %d, want %d", codes, 0x1CDE5-0x1CD00+1)
	}
	// Spot checks against Unicode's "BLOCK OCTANT-n" names.
	for mask, want := range map[uint8]rune{0x04: 0x1CD00, 0x17: 0x1CD0F, 0xFE: 0x1CDE5} {
		if got := npOctantGlyphs[mask]; got != string(want) {
			t.Errorf("mask %#02x = %U, want %U", mask, []rune(got)[0], want)
		}
	}
	for _, g := range npQuadrantGlyphs {
		if gw, wc := ansi.GraphemeWidth.StringWidth(g), ansi.WcWidth.StringWidth(g); gw != 1 || wc != 1 {
			t.Errorf("quadrant %q has width grapheme=%d wc=%d, want 1", g, gw, wc)
		}
	}
}

// A cell made of exactly two colors is reproduced exactly, including the six
// patterns that only exist as complements.
func TestFitCellReproducesTwoColorCells(t *testing.T) {
	fg, bg := npRGB{200, 190, 170}, npRGB{20, 20, 18}
	for m := 1; m < 255; m++ {
		var px [8]npRGB
		for i := range px {
			px[i] = bg
			if m&(1<<i) != 0 {
				px[i] = fg
			}
		}
		c := npFitCell(&px, true)
		var got uint8
		for i, g := range npOctantGlyphs {
			if g == c.Ch {
				got = uint8(i)
			}
		}
		want := uint8(m)
		if c.Fg != fg.rgba() {
			want = ^want
		}
		if got != want {
			t.Fatalf("mask %#02x fitted as %q (mask %#02x)", m, c.Ch, got)
		}
	}
}

func TestDeckLabelIsFlatAtRest(t *testing.T) {
	f := npTestDeckFrame(true, npPose{})
	if !f.Label {
		t.Fatal("label hidden at rest")
	}
	for row := npLabelRow; row <= npLabelRow+1; row++ {
		for col := npLabelLeft; col < npLabelRight; col++ {
			c := f.Cells[row*npDeckW+col]
			if c.Ch != " " || c.Bg != f.LabelBg {
				t.Fatalf("label cell (%d,%d) = %+v, want flat %v", col, row, c, f.LabelBg)
			}
		}
	}
}

func TestDeckMovingHidesLabelAndIsDeterministic(t *testing.T) {
	pose := npPose{DY: -12, Tilt: 0.35, Fade: 0.5}
	a, b := npTestDeckFrame(true, pose), npTestDeckFrame(true, pose)
	if a.Label {
		t.Fatal("label shown while moving")
	}
	if a.Cells != b.Cells {
		t.Fatal("rendering is not deterministic")
	}
}

// The cached static layer plus the per-frame window must equal a deck that
// was never cached.
func TestDeckCacheMatchesFreshRender(t *testing.T) {
	var cached npDeck
	cached.frame(npWarmNight, true, npDeckState{Progress: 0.8})
	got := *cached.frame(npWarmNight, true, npDeckState{Progress: 0.3, Angles: [2]float64{0.4, 1.9}})
	if want := npTestDeckFrame(true, npPose{}); got.Cells != want.Cells {
		t.Fatal("cached frame differs from a fresh render")
	}
}

func TestReelSpring(t *testing.T) {
	var m npReelMotion
	now := time.Now()
	radii := npReelRadii(0)
	step := func(seconds float64, playing bool) {
		for i := 0; i < int(seconds/0.05+0.5); i++ {
			m.advance(0.05, playing, now, radii)
		}
	}
	step(0.3, true)
	if m.speed < 0.85 {
		t.Errorf("speed after 300 ms spin-up = %.2f, want >= 0.85", m.speed)
	}
	step(1, true)
	step(0.9, false)
	if m.speed > 0.05 {
		t.Errorf("speed after 900 ms spin-down = %.2f, want <= 0.05", m.speed)
	}
	step(2, false)
	if m.moving() {
		t.Error("reels never come to rest")
	}
	// At playing speed a step never turns far enough to look reversed.
	var p npReelMotion
	p.speed = 1
	p.advance(0.05, true, now, radii)
	if p.turn >= 15*math.Pi/180 {
		t.Errorf("turn per frame = %.1f°, want < 15°", p.turn*180/math.Pi)
	}
}

func BenchmarkDeckSpinFrame(b *testing.B) {
	var d npDeck
	st := npDeckState{Progress: 0.4}
	d.frame(npWarmNight, true, st)
	for i := 0; b.Loop(); i++ {
		st.Angles = [2]float64{float64(i) * 0.1, float64(i) * 0.05}
		d.frame(npWarmNight, true, st)
	}
}

func BenchmarkDeckMovingFrame(b *testing.B) {
	var d npDeck
	st := npDeckState{Progress: 0.4, Pose: npPose{DY: -8, Tilt: 0.2, Fade: 0.3}}
	for b.Loop() {
		d.frame(npWarmNight, true, st)
	}
}

// The Warm Night theme file and the built-in fallback palette must not drift.
func TestWarmNightPaletteMatchesTheme(t *testing.T) {
	configs.LoadThemeRegistry(t.TempDir())
	reg := configs.CurrentThemeRegistry()
	reg.SelectTheme("Warm Night", true)
	ss := reg.CurrentStyleSet(true)
	if ss == nil || reg.CurrentName(true) != "Warm Night" {
		t.Fatal("Warm Night theme missing")
	}
	if got := npPaletteFromTheme(ss.Theme()); got != npWarmNight {
		t.Fatalf("palette from theme = %+v\nwant %+v", got, npWarmNight)
	}
	if npGhostOf(npWarmNight.Bg, npWarmNight.Rule) != npWarmNight.Ghost {
		t.Fatal("fallback ghost is not derived from Bg and Rule")
	}
}
