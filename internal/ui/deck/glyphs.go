package deck

import (
	"image/color"
	"math/bits"
)

// Tier is how finely a terminal cell is split into sub-pixels: each cell is
// drawn as one character with two colors over a cols×rows grid.
type Tier uint8

const (
	Half     Tier = iota + 1 // 1×2: ▀ ▄
	Quadrant                 // 2×2: ▘ ▝ ▖ ▗ …
	Sextant                  // 2×3: U+1FB00–U+1FB3B
	Octant                   // 2×4: U+1CD00–U+1CDE5
)

// Grid returns the sub-pixel columns and rows of one cell.
func (t Tier) Grid() (cols, rows int) {
	switch t {
	case Octant:
		return 2, 4
	case Sextant:
		return 2, 3
	case Quadrant:
		return 2, 2
	default:
		return 1, 2
	}
}

func (t Tier) String() string {
	switch t {
	case Octant:
		return "octant"
	case Sextant:
		return "sextant"
	case Quadrant:
		return "quadrant"
	case Half:
		return "half"
	}
	return ""
}

// ParseTier reads a tier name as used in the configuration.
func ParseTier(s string) (Tier, bool) {
	for _, t := range []Tier{Half, Quadrant, Sextant, Octant} {
		if t.String() == s {
			return t, true
		}
	}
	return 0, false
}

// Sub-pixels are numbered row-major: bit 0 is the top-left one.

// octantExisting are the 2×4 patterns Unicode encodes outside the octant
// block; the octant range U+1CD00–U+1CDE5 skips them.
var octantExisting = map[uint8]string{
	0x00: " ", 0xFF: "█",
	0x0F: "▀", 0xF0: "▄", 0x55: "▌", 0xAA: "▐",
	0x05: "▘", 0x0A: "▝", 0x50: "▖", 0xA0: "▗",
	0xA5: "▚", 0x5A: "▞", 0xF5: "▙", 0x5F: "▛", 0xAF: "▜", 0xFA: "▟",
	0xC0: "▂", 0xFC: "▆", 0x03: "\U0001FB82", 0x3F: "\U0001FB85",
}

// octantMissing have no character at all; they are drawn as their complement
// with foreground and background swapped.
var octantMissing = map[uint8]bool{0x01: true, 0x02: true, 0x14: true, 0x28: true, 0x40: true, 0x80: true}

var octantGlyphs = func() (t [256]string) {
	next := rune(0x1CD00)
	for m := range 256 {
		mask := uint8(m)
		if g, ok := octantExisting[mask]; ok {
			t[m] = g
			continue
		}
		if octantMissing[mask] {
			continue
		}
		t[m] = string(next)
		next++
	}
	return t
}()

// sextantGlyphs covers 2×3 patterns: the sextant block U+1FB00–U+1FB3B skips
// the empty, full and half-block patterns, which exist elsewhere.
var sextantGlyphs = func() (t [64]string) {
	existing := map[int]string{0: " ", 63: "█", 21: "▌", 42: "▐"}
	next := rune(0x1FB00)
	for m := range 64 {
		if g, ok := existing[m]; ok {
			t[m] = g
			continue
		}
		t[m] = string(next)
		next++
	}
	return t
}()

var quadrantGlyphs = [16]string{" ", "▘", "▝", "▀", "▖", "▌", "▞", "▛", "▗", "▚", "▐", "▜", "▄", "▙", "▟", "█"}

var halfGlyphs = [4]string{" ", "▀", "▄", "█"}

func glyphFor(t Tier, mask uint8) string {
	switch t {
	case Octant:
		return octantGlyphs[mask]
	case Sextant:
		return sextantGlyphs[mask&63]
	case Quadrant:
		return quadrantGlyphs[mask&15]
	default:
		return halfGlyphs[mask&3]
	}
}

// Cell is one fitted character cell. An empty Ch leaves the screen as it is.
type Cell struct {
	Ch     string
	Fg, Bg color.RGBA
}

// minSplitGain is the per-sub-pixel squared error a split must remove to be
// worth a glyph; below it the cell is drawn flat, which keeps gradients calm.
const minSplitGain = 18.0

// fitCell picks the character and two colors that best reproduce a cell's
// sub-pixels, given in row-major order.
func fitCell(px []rgb, t Tier) Cell {
	n := len(px)
	full := uint8(1<<n - 1)
	a, b, mask, gain := split2(px)
	if mask == 0 || mask == full || gain < minSplitGain*float64(n) {
		return Cell{Ch: " ", Bg: mean(px).rgba()}
	}
	fg, bg := b, a
	if bits.OnesCount8(mask) > n/2 {
		mask, fg, bg = ^mask&full, a, b
	}
	g := glyphFor(t, mask)
	if g == "" {
		mask, fg, bg = ^mask&full, bg, fg
		g = glyphFor(t, mask)
	}
	return Cell{Ch: g, Fg: fg.rgba(), Bg: bg.rgba()}
}

func mean(cs []rgb) rgb {
	var m rgb
	for _, c := range cs {
		m = m.add(c)
	}
	return m.scale(1 / float64(len(cs)))
}

// split2 partitions colors into two groups with two rounds of 2-means and
// returns the group means, the membership bits of the second group and how
// much squared error the split removes compared with a single color.
func split2(cs []rgb) (a, b rgb, mask uint8, gain float64) {
	m := mean(cs)
	a = cs[0]
	for _, c := range cs {
		if c.dist2(m) > a.dist2(m) {
			a = c
		}
	}
	b = a
	for _, c := range cs {
		if c.dist2(a) > b.dist2(a) {
			b = c
		}
	}
	if a.dist2(b) == 0 {
		return m, m, 0, 0
	}
	for range 2 {
		mask = 0
		var sa, sb rgb
		var na, nb int
		for i, c := range cs {
			if c.dist2(b) < c.dist2(a) {
				mask |= 1 << i
				sb, nb = sb.add(c), nb+1
			} else {
				sa, na = sa.add(c), na+1
			}
		}
		if na > 0 {
			a = sa.scale(1 / float64(na))
		}
		if nb > 0 {
			b = sb.scale(1 / float64(nb))
		}
	}
	var err1, err2 float64
	for i, c := range cs {
		err1 += c.dist2(m)
		if mask&(1<<i) != 0 {
			err2 += c.dist2(b)
		} else {
			err2 += c.dist2(a)
		}
	}
	return a, b, mask, err1 - err2
}
