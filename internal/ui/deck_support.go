package ui

import (
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/ansix"
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
// draw show bg. Each line is exactly f.W cells wide.
func deckLines(f *deck.Frame, bg color.Color) []string {
	lines := make([]string, f.H)
	var bgR, bgG, bgB uint32
	if bg != nil {
		bgR, bgG, bgB, _ = bg.RGBA()
		bgR, bgG, bgB = bgR>>8, bgG>>8, bgB>>8
	}
	b := make([]byte, 0, f.W*40)
	for y := range f.H {
		b = b[:0]
		var curBg, curFg sgrColor
		for x := range f.W {
			c := f.Cells[y*f.W+x]
			switch {
			case c.Ch == "":
				if bg != nil {
					b = curBg.set(b, "48", bgR, bgG, bgB)
				}
				b = append(b, ' ')
			default:
				b = curBg.set(b, "48", uint32(c.Bg.R), uint32(c.Bg.G), uint32(c.Bg.B))
				if c.Ch != " " {
					b = curFg.set(b, "38", uint32(c.Fg.R), uint32(c.Fg.G), uint32(c.Fg.B))
				}
				b = append(b, c.Ch...)
			}
		}
		b = append(b, "\x1b[m"...)
		lines[y] = string(b)
	}
	return lines
}

// sgrColor is the color last set on a line for one SGR parameter.
type sgrColor struct {
	r, g, b uint32
	ok      bool
}

// set appends the sequence setting code to the color unless it is in effect.
func (s *sgrColor) set(b []byte, code string, r, g, bl uint32) []byte {
	if s.ok && s.r == r && s.g == g && s.b == bl {
		return b
	}
	*s = sgrColor{r, g, bl, true}
	return appendSGR(b, code, r, g, bl)
}

// appendSGR appends the sequence setting a true color: code 38 is the
// foreground and 48 the background.
func appendSGR(b []byte, code string, r, g, bl uint32) []byte {
	b = append(b, "\x1b["...)
	b = append(b, code...)
	b = append(b, ";2;"...)
	b = strconv.AppendUint(b, uint64(r), 10)
	b = append(b, ';')
	b = strconv.AppendUint(b, uint64(g), 10)
	b = append(b, ';')
	b = strconv.AppendUint(b, uint64(bl), 10)
	return append(b, 'm')
}

// fitRow pads a line sw cells wide to w cells as ansix.Fit does, without
// measuring it again.
func fitRow(s string, sw, w int, pad string) string {
	if sw >= w || pad == "" || !strings.Contains(s, "\x1b") {
		return ansix.Fit(s, w, pad)
	}
	return s + "\x1b[m" + pad + strings.Repeat(" ", w-sw) + "\x1b[m"
}

// deckPacer redraws a slowly moving cassette at most deckPace apart: the
// interface draws faster while it animates other things or takes keys, and
// a floating cassette looks the same at that pace. The float moves it by a
// fraction of a sub-pixel in that time, so its pose, lift and tape packs are
// taken anew only every deckPosePace; in between only the reels turn, which
// the renderer redraws by the hubs alone. A cassette moving briskly, or
// drawn into a different view, is redrawn in full every time.
type deckPacer struct {
	last   *deck.Frame
	at     time.Time
	view   deck.View
	pose   deck.Pose
	lift   float64
	radii  [2]float64
	poseAt time.Time
}

const (
	deckPace     = 33 * time.Millisecond
	deckPosePace = 60 * time.Millisecond // every other frame at deckPace
)

func (c *deckPacer) frame(now time.Time, r *deck.Renderer, v deck.View, pal deck.Palette, pose deck.Pose, st deck.State, brisk bool) *deck.Frame {
	fresh := c.last == nil || brisk || v != c.view || now.Before(c.at)
	if !fresh && now.Sub(c.at) < deckPace {
		return c.last
	}
	if fresh || now.Sub(c.poseAt) >= deckPosePace {
		c.pose, c.lift, c.radii, c.poseAt = pose, st.Lift, st.Radii, now
	}
	st.Lift, st.Radii = c.lift, c.radii
	c.last, c.at, c.view = r.Render(v, pal, c.pose, st), now, v
	return c.last
}
