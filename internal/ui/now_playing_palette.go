package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

// npPalette holds the color roles of the now-playing page. Text is near
// neutral; Accent marks what is happening now and Accent2 switches that are on.
// Ghost, for unlit segments and lamps, is derived from Bg and Rule.
type npPalette struct {
	Bg, Surface, Text, Hero, Secondary, Rule, Ghost, Accent, Accent2 color.RGBA
}

// npWarmNight is the Warm Night theme palette (utils/filex/embed/themes/warm_night.toml),
// used for any role the active theme leaves unset.
var npWarmNight = npPalette{
	Bg:        color.RGBA{0x15, 0x14, 0x12, 0xFF},
	Surface:   color.RGBA{0x1D, 0x1C, 0x19, 0xFF},
	Text:      color.RGBA{0xCB, 0xC4, 0xB6, 0xFF},
	Hero:      color.RGBA{0xDD, 0xD6, 0xC9, 0xFF},
	Secondary: color.RGBA{0x9E, 0x97, 0x8A, 0xFF},
	Rule:      color.RGBA{0x54, 0x4F, 0x47, 0xFF},
	Ghost:     color.RGBA{0x2E, 0x2C, 0x27, 0xFF}, // npGhostOf(Bg, Rule)
	Accent:    color.RGBA{0xD2, 0x71, 0x4F, 0xFF},
	Accent2:   color.RGBA{0x93, 0xAC, 0x82, 0xFF},
}

// The idle "no signal" screen keeps a dim VCR blue that stays calm at night.
var (
	npNoSignal     = color.RGBA{0x1C, 0x24, 0x33, 0xFF}
	npNoSignalScan = color.RGBA{0x19, 0x20, 0x2E, 0xFF}
)

// npSolid converts a theme color to RGBA; nil and lipgloss.NoColor (whose
// RGBA() reports opaque black) count as unset.
func npSolid(c color.Color) (color.RGBA, bool) {
	if c == nil {
		return color.RGBA{}, false
	}
	if _, ok := c.(lipgloss.NoColor); ok {
		return color.RGBA{}, false
	}
	r, g, b, _ := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 0xFF}, true
}

// npPaletteFromTheme maps the theme roles onto the page palette.
func npPaletteFromTheme(t style.Theme) npPalette {
	p := npWarmNight
	pick := func(dst *color.RGBA, candidates ...color.Color) {
		for _, c := range candidates {
			if rgba, ok := npSolid(c); ok {
				*dst = rgba
				return
			}
		}
	}
	pick(&p.Bg, t.AppBackground.Bg, t.Background)
	pick(&p.Surface, t.Surface)
	pick(&p.Text, t.Foreground)
	pick(&p.Hero, t.Primary)
	pick(&p.Secondary, t.Secondary)
	pick(&p.Rule, t.Muted)
	pick(&p.Accent, t.Accent)
	pick(&p.Accent2, t.Success)
	p.Ghost = npGhostOf(p.Bg, p.Rule)
	return p
}

// npGhostOf derives the unlit color, barely above the background, from the
// background and the rule color.
func npGhostOf(bg, rule color.RGBA) color.RGBA {
	mix := func(a, b uint8) uint8 { return uint8(float64(a) + (float64(b)-float64(a))*0.4 + 0.5) }
	return color.RGBA{mix(bg.R, rule.R), mix(bg.G, rule.G), mix(bg.B, rule.B), 0xFF}
}
