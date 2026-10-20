package ui

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/ui/deck"
)

// deckPaletteOf picks the cassette's colors from the page palette.
func deckPaletteOf(p npPalette) deck.Palette {
	return deck.Palette{
		Bg: p.Bg, Surface: p.Surface, Text: p.Text, Hero: p.Hero,
		Secondary: p.Secondary, Rule: p.Rule, Accent: p.Accent,
	}
}

// deckTier is the block glyph tier to draw the cassette with: the configured
// one, else what the terminal is known to draw itself.
func deckTier(a *model.App) deck.Tier {
	if t, ok := deck.ParseTier(configs.AppConfig.Main.BlockGlyphs); ok {
		return t
	}
	version := ""
	if a != nil {
		version = a.TerminalVersion()
	}
	return deck.DetectTier(deck.Terminal{Getenv: os.Getenv, Version: version})
}

// GlideSteps returns how many steps of a row the tracklist's selection
// marker can glide by, from the block glyphs the terminal draws.
func GlideSteps(n *Netease) func() int {
	return func() int {
		switch deckTier(n.App) {
		case deck.Octant:
			return 4
		case deck.Sextant:
			return 3
		case deck.Quadrant:
			return 2
		}
		return 0
	}
}

// deckAspect is the terminal's cell height over width, 0 when unknown. Some
// terminals and multiplexers report a made-up cell size; only plausible
// shapes count.
func deckAspect(a *model.App) float64 {
	if a == nil {
		return 0
	}
	if r := a.CellAspect(); r >= 1.4 && r <= 3.2 {
		return r
	}
	return 0
}

// deckLines turns a rendered cassette into styled lines; cells it does not
// draw show bg.
func deckLines(f *deck.Frame, bg color.Color) []string {
	lines := make([]string, f.H)
	var b strings.Builder
	var curBg, curFg string
	sgr := func(code string, c color.Color, cur *string) {
		r, g, bl, _ := c.RGBA()
		s := fmt.Sprintf("\x1b[%s;2;%d;%d;%dm", code, r>>8, g>>8, bl>>8)
		if s != *cur {
			b.WriteString(s)
			*cur = s
		}
	}
	for y := range f.H {
		b.Reset()
		curBg, curFg = "", ""
		for x := range f.W {
			c := f.Cells[y*f.W+x]
			switch {
			case c.Ch == "":
				if bg != nil {
					sgr("48", bg, &curBg)
				}
				b.WriteByte(' ')
			default:
				sgr("48", c.Bg, &curBg)
				if c.Ch != " " {
					sgr("38", c.Fg, &curFg)
				}
				b.WriteString(c.Ch)
			}
		}
		b.WriteString("\x1b[m")
		lines[y] = b.String()
	}
	return lines
}

// deckPacer redraws a slowly moving cassette at most deckPace apart: the
// interface draws faster while it animates other things or takes keys, and
// a floating cassette looks the same at that pace. A cassette moving
// briskly, or drawn into a different view, is redrawn every time.
type deckPacer struct {
	last *deck.Frame
	at   time.Time
	view deck.View
}

const deckPace = 33 * time.Millisecond

func (c *deckPacer) frame(now time.Time, v deck.View, brisk bool, render func() *deck.Frame) *deck.Frame {
	if c.last == nil || brisk || v != c.view || now.Sub(c.at) >= deckPace || now.Before(c.at) {
		c.last, c.at, c.view = render(), now, v
	}
	return c.last
}
