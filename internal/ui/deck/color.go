// Package deck renders a cassette in 3D with a perspective ray caster and fits
// the image into block-glyph cells, so that it can move freely in a terminal.
package deck

import (
	"image/color"
	"math"
)

// rgb is a working color in 0–255 units.
type rgb struct{ r, g, b float64 }

func rgbOf(c color.RGBA) rgb { return rgb{float64(c.R), float64(c.G), float64(c.B)} }

func (c rgb) rgba() color.RGBA {
	cl := func(v float64) uint8 { return uint8(math.Round(min(max(v, 0), 255))) }
	return color.RGBA{cl(c.r), cl(c.g), cl(c.b), 0xFF}
}

func (c rgb) mix(o rgb, t float64) rgb {
	return rgb{c.r + (o.r-c.r)*t, c.g + (o.g-c.g)*t, c.b + (o.b-c.b)*t}
}

func (c rgb) scale(k float64) rgb { return rgb{c.r * k, c.g * k, c.b * k} }

func (c rgb) add(o rgb) rgb { return rgb{c.r + o.r, c.g + o.g, c.b + o.b} }

func (c rgb) dist2(o rgb) float64 {
	dr, dg, db := c.r-o.r, c.g-o.g, c.b-o.b
	return dr*dr + dg*dg + db*db
}

// Palette holds the color roles the cassette is drawn with.
type Palette struct {
	Bg, Surface, Text, Hero, Secondary, Rule, Accent color.RGBA
}

// materials are the surface colors derived from a palette.
type materials struct {
	bg, hero, shell, label, ink, stripe, glass, tape, hub, hole, clamp, screw rgb
}

func materialsFor(p Palette) materials {
	bg := rgbOf(p.Bg)
	rule := rgbOf(p.Rule)
	return materials{
		bg:     bg,
		hero:   rgbOf(p.Hero),
		shell:  rgbOf(p.Surface).mix(rgbOf(p.Text), 0.10),
		label:  bg.mix(rgbOf(p.Surface), 0.6),
		ink:    rule.mix(rgbOf(p.Secondary), 0.35),
		stripe: rgbOf(p.Accent),
		glass:  bg.mix(rule, 0.12).scale(0.85),
		tape:   rgbOf(p.Accent).mix(bg, 0.78),
		hub:    rgbOf(p.Text).mix(bg, 0.28),
		hole:   bg.scale(0.45),
		clamp:  rgbOf(p.Accent),
		screw:  rule.mix(rgbOf(p.Text), 0.25),
	}
}
