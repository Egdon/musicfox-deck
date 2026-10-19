package ui

import (
	"image/color"
	"math"
	"math/bits"
	"os"
	"sync"
)

// A cell holds 2×4 sub-pixels numbered row-major: bit 0 is top-left, bit 1
// top-right, …, bit 7 bottom-right (Unicode's "BLOCK OCTANT-n" digit n is
// bit n-1).

// npExistingBlocks are the 2×4 patterns Unicode encodes outside the octant
// block; the octant range U+1CD00–U+1CDE5 skips them.
var npExistingBlocks = map[uint8]string{
	0x00: " ", 0xFF: "█",
	0x0F: "▀", 0xF0: "▄", 0x55: "▌", 0xAA: "▐",
	0x05: "▘", 0x0A: "▝", 0x50: "▖", 0xA0: "▗",
	0xA5: "▚", 0x5A: "▞", 0xF5: "▙", 0x5F: "▛", 0xAF: "▜", 0xFA: "▟",
	0xC0: "▂", 0xFC: "▆", 0x03: "\U0001FB82", 0x3F: "\U0001FB85",
}

// npMissingOctants have no character at all; they are drawn as their
// complement with foreground and background swapped.
var npMissingOctants = map[uint8]bool{0x01: true, 0x02: true, 0x14: true, 0x28: true, 0x40: true, 0x80: true}

// npOctantGlyphs maps every 2×4 mask to its character; the missing six are "".
var npOctantGlyphs = func() [256]string {
	var t [256]string
	next := rune(0x1CD00)
	for m := range 256 {
		mask := uint8(m)
		if g, ok := npExistingBlocks[mask]; ok {
			t[m] = g
			continue
		}
		if npMissingOctants[mask] {
			continue
		}
		t[m] = string(next)
		next++
	}
	return t
}()

// npQuadrantGlyphs maps 2×2 masks (bit 0 top-left … bit 3 bottom-right) for
// terminals that do not draw octants.
var npQuadrantGlyphs = [16]string{" ", "▘", "▝", "▀", "▖", "▌", "▞", "▛", "▗", "▚", "▐", "▜", "▄", "▙", "▟", "█"}

// npOctantSupport reports whether the terminal draws octants itself. Ghostty
// renders the whole range with its sprite font, independent of the font.
var npOctantSupport = sync.OnceValue(func() bool {
	return os.Getenv("TERM_PROGRAM") == "ghostty" || os.Getenv("TERM") == "xterm-ghostty" ||
		os.Getenv("GHOSTTY_RESOURCES_DIR") != ""
})

// npRGB is a linear-free working color in 0–255 units.
type npRGB struct{ R, G, B float64 }

func npRGBOf(c color.RGBA) npRGB { return npRGB{float64(c.R), float64(c.G), float64(c.B)} }

func (c npRGB) rgba() color.RGBA {
	cl := func(v float64) uint8 { return uint8(math.Round(min(max(v, 0), 255))) }
	return color.RGBA{cl(c.R), cl(c.G), cl(c.B), 0xFF}
}

func (c npRGB) mix(o npRGB, t float64) npRGB {
	return npRGB{c.R + (o.R-c.R)*t, c.G + (o.G-c.G)*t, c.B + (o.B-c.B)*t}
}

func (c npRGB) scale(k float64) npRGB { return npRGB{c.R * k, c.G * k, c.B * k} }

func (c npRGB) dist2(o npRGB) float64 {
	dr, dg, db := c.R-o.R, c.G-o.G, c.B-o.B
	return dr*dr + dg*dg + db*db
}

func npMean(cs []npRGB) npRGB {
	var m npRGB
	for _, c := range cs {
		m.R, m.G, m.B = m.R+c.R, m.G+c.G, m.B+c.B
	}
	n := float64(len(cs))
	return npRGB{m.R / n, m.G / n, m.B / n}
}

// npSplit2 partitions colors into two groups with two rounds of 2-means and
// returns the group means, the membership bits of the second group and how
// much squared error the split removes compared with a single color.
func npSplit2(cs []npRGB) (a, b npRGB, mask uint8, gain float64) {
	m := npMean(cs)
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
		var ga, gb []npRGB
		for i, c := range cs {
			if c.dist2(b) < c.dist2(a) {
				mask |= 1 << i
				gb = append(gb, c)
			} else {
				ga = append(ga, c)
			}
		}
		if len(ga) > 0 {
			a = npMean(ga)
		}
		if len(gb) > 0 {
			b = npMean(gb)
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

// npFitMinGain is the per-sub-pixel squared error a split must remove to be
// worth a glyph; below it the cell is drawn flat, which keeps gradients calm.
const npFitMinGain = 18.0

// npFitCell picks the character and two colors that best reproduce a cell's
// eight sub-pixels: octants when the terminal draws them, quadrants otherwise.
func npFitCell(px *[8]npRGB, octants bool) npCell {
	cs := px[:]
	full := uint8(0xFF)
	if !octants {
		q := [4]npRGB{
			px[0].mix(px[2], 0.5), px[1].mix(px[3], 0.5),
			px[4].mix(px[6], 0.5), px[5].mix(px[7], 0.5),
		}
		cs = q[:]
		full = 0x0F
	}
	a, b, mask, gain := npSplit2(cs)
	if mask == 0 || mask == full || gain < npFitMinGain*float64(len(cs)) {
		return npCell{Ch: " ", Bg: npMean(cs).rgba()}
	}
	fg, bg := b, a
	if bits.OnesCount8(mask) > len(cs)/2 {
		mask, fg, bg = ^mask&full, a, b
	}
	if !octants {
		return npCell{Ch: npQuadrantGlyphs[mask], Fg: fg.rgba(), Bg: bg.rgba()}
	}
	glyph := npOctantGlyphs[mask]
	if glyph == "" {
		mask, fg, bg = ^mask, bg, fg
		glyph = npOctantGlyphs[mask]
	}
	return npCell{Ch: glyph, Fg: fg.rgba(), Bg: bg.rgba()}
}

// npCell is one fitted character cell; an empty Ch leaves the screen as is.
type npCell struct {
	Ch     string
	Fg, Bg color.RGBA
}
