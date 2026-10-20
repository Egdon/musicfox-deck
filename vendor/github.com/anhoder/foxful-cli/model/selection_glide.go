package model

import (
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"
	uv "github.com/charmbracelet/ultraviolet"
)

// Selection glide: when the selection moves within a page of the tracklist,
// its band fades across from the old row to the new one and, at full
// motion, a bar slides between them in steps of a row before the theme's
// marker comes back.

const (
	glideOmega = 30.0 // the bar's spring: it arrives in about 0.15 s
	glideEps   = 0.02 // rows within which the bar has arrived
)

var (
	bandFade = 150 * time.Millisecond
	// Moves closer together than this come from a held key or the wheel:
	// the selection keeps up with them instead of animating each.
	glideRapid = 120 * time.Millisecond
)

type selectionGlide struct {
	key   menuKey // the page the selection was on
	known bool
	row   int // its selected row, last frame
	seq   uint64

	bar   motion.Spring // the bar's row; a zero Start when it is still
	steps int           // steps of a row the bar moves by
	band  time.Time     // when the band left from; zero when it is still
	from  int
	moved time.Time // when the selection last moved
}

// glideSteps is how many steps of a row the bar can move by.
func (m *Main) glideSteps() int {
	if f := m.options.Motion.GlideSteps; f != nil {
		return min(f(), 4)
	}
	return 0
}

// glideRows animates the selection over the rows of the page shown, the
// first of which is item first.
func (m *Main) glideRows(rows []string, cols trackColumns, first int) []string {
	a := m.app
	if a == nil || a.snapshotting {
		return rows
	}
	g := &m.glide
	now, seq, key := a.FrameTime(), a.FrameSeq(), m.menuKeyNow()
	row, level := m.selectedIndex-first, a.overlayLevel()
	inPage := func(r int) bool { return r >= 0 && r < len(rows) }
	switch {
	case !g.known || g.key != key || g.seq+1 != seq || level == motion.Off:
		g.bar, g.band = motion.Spring{}, time.Time{}
	case row != g.row && inPage(row) && inPage(g.row):
		rapid := now.Sub(g.moved) < glideRapid
		g.moved = now
		if rapid {
			g.bar, g.band = motion.Spring{}, time.Time{}
			break
		}
		g.band, g.from = now, g.row
		wait := bandFade
		if steps := m.glideSteps(); level == motion.Full && steps > 1 {
			if g.bar.Start.IsZero() {
				g.bar = motion.Spring{From: float64(g.row), To: float64(g.row), Omega: glideOmega}
			}
			g.bar.Retarget(now, float64(row))
			g.steps = steps
			wait = max(wait, g.bar.SettleTime())
		}
		a.Animate(wait)
	}
	g.key, g.known, g.row, g.seq = key, true, row, seq

	// How much of the band each row has, 0 to 1.
	bandOf := func(r int) float64 {
		if r == row {
			return 1
		}
		return 0
	}
	if !g.band.IsZero() {
		t := now.Sub(g.band)
		if t >= bandFade || !inPage(row) || !inPage(g.from) || g.from == row {
			g.band = time.Time{}
		} else {
			e := motion.OutCubic(float64(t) / float64(bandFade))
			bandOf = func(r int) float64 {
				switch r {
				case row:
					return e
				case g.from:
					return 1 - e
				}
				return 0
			}
			rows[g.from] = m.withBand(rows[g.from], cols, 1-e, false)
			rows[row] = m.withBand(rows[row], cols, e, true)
		}
	}
	if !g.bar.Start.IsZero() {
		if g.bar.Settled(now, glideEps) || !inPage(row) {
			g.bar = motion.Spring{}
		} else {
			y, _ := g.bar.At(now)
			m.drawGlideBar(rows, cols, row, y, g.steps, bandOf)
		}
	}
	return rows
}

// glideColors are the colors the glide paints with.
func glideColors() (appBg, band, bar color.Color) {
	ss := style.CurrentStyleSet()
	appBg, band = ss.AppBackground.GetBackground(), ss.SelectedItem.GetBackground()
	bar = ss.MenuArrow.GetForeground()
	if isNoColor(bar) {
		bar = ss.SelectedItem.GetForeground()
	}
	if isNoColor(appBg) {
		appBg = nil
	}
	if isNoColor(band) {
		band = nil
	}
	return appBg, band, bar
}

// withBand gives a row k of the selection band across its span: a selected
// row fading in, or an unselected one fading out.
func (m *Main) withBand(row string, cols trackColumns, k float64, selected bool) string {
	appBg, band, _ := glideColors()
	if band == nil {
		return row
	}
	bg := band
	if appBg != nil {
		bg = ansix.Mix(appBg, band, k)
	}
	return ansix.Restyle(row, cols.start, cols.end, func(st uv.Style) uv.Style {
		switch {
		case selected && ansix.SameColor(st.Bg, band):
			st.Bg = bg
		case selected && ansix.SameColor(st.Fg, band): // the band's rounded ends
			st.Fg = bg
		case !selected && (st.Bg == nil || ansix.SameColor(st.Bg, appBg)):
			st.Bg = bg
		}
		return st
	})
}

// drawGlideBar draws the bar at row y of the page, in the marker's column,
// over the cells it covers, and hides the theme's marker on the selected
// row meanwhile.
func (m *Main) drawGlideBar(rows []string, cols trackColumns, selected int, y float64, steps int, bandOf func(int) float64) {
	appBg, band, bar := glideColors()
	x, w := cols.start+1, m.ContentWidth()
	bgOf := func(r int) color.Color {
		if band == nil || appBg == nil {
			return appBg
		}
		return ansix.Mix(appBg, band, bandOf(r))
	}
	rows[selected] = overlayCell(rows[selected], x, w, " ", uv.Style{Bg: bgOf(selected)})

	pos := int(math.Round(y * float64(steps)))
	top := floorDiv(pos, steps)
	part := pos - top*steps // sub-rows of the bar in the row under top
	cover := func(r, n int, fromTop bool) {
		if r < 0 || r >= len(rows) || n <= 0 {
			return
		}
		glyph, inverse := leftBar(steps, n, fromTop)
		st := uv.Style{Fg: bar, Bg: bgOf(r)}
		if inverse {
			st.Fg, st.Bg = st.Bg, st.Fg
		}
		rows[r] = overlayCell(rows[r], x, w, glyph, st)
	}
	cover(top, steps-part, false)
	cover(top+1, part, true)
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}

// overlayCell puts a glyph with a style in the cell at column x of a line w
// cells wide.
func overlayCell(line string, x, w int, glyph string, st uv.Style) string {
	if x < 0 || x >= w {
		return line
	}
	var b strings.Builder
	b.WriteString(ansix.Cut(line, 0, x))
	b.WriteString(uv.StyleDiff(&uv.Style{}, &st))
	b.WriteString(glyph)
	b.WriteString("\x1b[m")
	b.WriteString(ansix.Cut(line, x+1, w))
	return b.String()
}

// leftBar returns the glyph that fills the left half of a cell for n of its
// steps sub-rows, from its top or from its bottom. inverse means the glyph
// draws the rest of the cell, so foreground and background swap.
func leftBar(steps, n int, fromTop bool) (glyph string, inverse bool) {
	if n >= steps {
		return "▌", false
	}
	var mask int
	for i := range n {
		row := i
		if !fromTop {
			row = steps - 1 - i
		}
		mask |= 1 << (2 * row)
	}
	switch steps {
	case 4:
		if g := octantGlyph(mask); g != "" {
			return g, false
		}
		return octantGlyph(^mask & 0xFF), true
	case 3:
		return sextantGlyph(mask), false
	default: // quadrants: half a row
		if fromTop {
			return "▘", false
		}
		return "▖", false
	}
}

// octantGlyph returns the character for a 2×4 pattern, bit 0 its top-left
// sub-cell, or "" for the six patterns Unicode has none for. The octant
// block U+1CD00–U+1CDE5 skips the patterns encoded elsewhere.
func octantGlyph(mask int) string {
	existing := map[int]string{
		0x00: " ", 0xFF: "█", 0x0F: "▀", 0xF0: "▄", 0x55: "▌", 0xAA: "▐",
		0x05: "▘", 0x0A: "▝", 0x50: "▖", 0xA0: "▗", 0xA5: "▚", 0x5A: "▞",
		0xF5: "▙", 0x5F: "▛", 0xAF: "▜", 0xFA: "▟", 0xC0: "▂", 0xFC: "▆",
		0x03: "\U0001FB82", 0x3F: "\U0001FB85",
	}
	missing := map[int]bool{0x01: true, 0x02: true, 0x14: true, 0x28: true, 0x40: true, 0x80: true}
	if g, ok := existing[mask]; ok {
		return g
	}
	if missing[mask] {
		return ""
	}
	next := rune(0x1CD00)
	for m := range mask {
		if _, ok := existing[m]; !ok && !missing[m] {
			next++
		}
	}
	return string(next)
}

// sextantGlyph returns the character for a 2×3 pattern, bit 0 its top-left
// sub-cell. The sextant block U+1FB00–U+1FB3B skips the empty, full and
// half-block patterns.
func sextantGlyph(mask int) string {
	existing := map[int]string{0: " ", 63: "█", 21: "▌", 42: "▐"}
	if g, ok := existing[mask]; ok {
		return g
	}
	next := rune(0x1FB00)
	for m := range mask {
		if _, ok := existing[m]; !ok {
			next++
		}
	}
	return string(next)
}
