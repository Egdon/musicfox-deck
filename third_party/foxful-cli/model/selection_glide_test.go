package model

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

func TestOctantAndSextantGlyphs(t *testing.T) {
	seen := map[string]bool{}
	missing := 0
	for m := range 256 {
		g := octantGlyph(m)
		if g == "" {
			missing++
			continue
		}
		if seen[g] {
			t.Fatalf("octant %#x repeats %q", m, g)
		}
		seen[g] = true
	}
	if missing != 6 || octantGlyph(0xFE) != "\U0001CDE5" {
		t.Fatalf("%d octants missing, the last one %q", missing, octantGlyph(0xFE))
	}
	if sextantGlyph(62) != "\U0001FB3B" || sextantGlyph(1) != "\U0001FB00" {
		t.Fatalf("sextants run U+1FB00–U+1FB3B: %q %q", sextantGlyph(1), sextantGlyph(62))
	}
}

func TestLeftBarGlyphs(t *testing.T) {
	for _, c := range []struct {
		steps, n int
		top      bool
		want     string
		inverse  bool
	}{
		{4, 4, true, "▌", false},
		{4, 2, true, "▘", false},
		{4, 2, false, "▖", false},
		{4, 1, true, octantGlyph(0xFE), true},
		{4, 1, false, octantGlyph(0xBF), true},
		{4, 3, true, octantGlyph(0x15), false},
		{3, 1, true, sextantGlyph(1), false},
		{3, 2, false, sextantGlyph(20), false},
		{2, 1, true, "▘", false},
		{2, 1, false, "▖", false},
	} {
		if g, inv := leftBar(c.steps, c.n, c.top); g != c.want || inv != c.inverse {
			t.Errorf("leftBar(%d, %d, %v) = %q %v, want %q %v", c.steps, c.n, c.top, g, inv, c.want, c.inverse)
		}
	}
}

func newGlideMain(t *testing.T, level motion.Level, steps int) (*Main, func(time.Time) []string, int) {
	t.Helper()
	items := make([]MenuItem, 12)
	for i := range items {
		items[i] = MenuItem{Title: "Track", Subtitle: "Artist", Meta: "3:07"}
	}
	options := DefaultOptions()
	options.DualColumn = false
	options.ListStyle = ListTracklist
	options.Motion.Level = level
	options.Motion.GlideSteps = func() int { return steps }
	options.MainMenu = &mockMenu{key: "root", items: items}
	options.MainMenuTitle = &MenuItem{Title: "Root"}
	app := NewApp(options)
	app.windowWidth, app.windowHeight = 80, 30
	app.motion.scheduler = motion.NewScheduler(16*time.Millisecond, level, func(any) {})
	t.Cleanup(app.motion.scheduler.Stop)
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: 80, Height: 30}, app)
	frame := func(at time.Time) []string {
		app.motion.frameSeq++
		app.motion.frameNow = at
		return strings.Split(main.View(app), "\n")
	}
	x := main.trackColumns(main.getCurPageMenus()).start + 1
	return main, frame, x
}

// markerColumn returns the glyphs in column x of the rows from start on.
func markerColumn(lines []string, x, start, n int) []string {
	out := make([]string, n)
	for i := range n {
		cell := ansi.Strip(ansiCut(lines[start+i], x, x+1))
		out[i] = strings.TrimSpace(cell)
	}
	return out
}

func ansiCut(s string, l, r int) string { return ansi.Cut(s, l, r) }

func TestSelectionGlidesBetweenRows(t *testing.T) {
	main, frame, x := newGlideMain(t, motion.Full, 4)
	marker := string([]rune(style.CurrentStyleSet().MenuSelectedArrow)[0])
	t0 := time.Now()
	first := main.menuListStartRow
	still := markerColumn(frame(t0), x, first, 6)
	if still[0] != marker {
		t.Fatalf("the theme marker marks the selection: %q", still)
	}

	main.selectedIndex = 4
	start := t0.Add(16 * time.Millisecond)
	moving := markerColumn(frame(start.Add(30*time.Millisecond)), x, first, 6)
	if moving[4] == marker {
		t.Fatalf("the theme marker waits for the bar: %q", moving)
	}
	bars := 0
	for _, g := range moving[:4] {
		if g != "" {
			bars++
		}
	}
	if bars == 0 || bars > 2 {
		t.Fatalf("the bar is on its way, over one or two cells: %q", moving)
	}

	settled := markerColumn(frame(start.Add(time.Second)), x, first, 6)
	if settled[4] != marker || strings.Join(settled[:4], "") != "" {
		t.Fatalf("after the glide only the theme marker shows: %q", settled)
	}
}

func TestSelectionWithoutGlide(t *testing.T) {
	for _, c := range []struct {
		level motion.Level
		steps int
	}{{motion.Subtle, 4}, {motion.Full, 1}, {motion.Off, 4}} {
		main, frame, x := newGlideMain(t, c.level, c.steps)
		marker := string([]rune(style.CurrentStyleSet().MenuSelectedArrow)[0])
		t0 := time.Now()
		frame(t0)
		main.selectedIndex = 3
		got := markerColumn(frame(t0.Add(20*time.Millisecond)), x, main.menuListStartRow, 5)
		if got[3] != marker || strings.Join(got[:3], "") != "" {
			t.Fatalf("level %v steps %d: the marker moves at once: %q", c.level, c.steps, got)
		}
	}
}

// A held key moves the selection faster than a glide: it keeps up instead.
func TestSelectionKeepsUpWithAHeldKey(t *testing.T) {
	main, frame, x := newGlideMain(t, motion.Full, 4)
	marker := string([]rune(style.CurrentStyleSet().MenuSelectedArrow)[0])
	t0 := time.Now()
	frame(t0)
	main.selectedIndex = 1
	frame(t0.Add(300 * time.Millisecond)) // the first press glides
	main.selectedIndex = 2
	got := markerColumn(frame(t0.Add(330*time.Millisecond)), x, main.menuListStartRow, 4)
	if got[2] != marker || strings.Join(got[:2], "") != "" {
		t.Fatalf("a repeated move shows the marker on the new row at once: %q", got)
	}
}
