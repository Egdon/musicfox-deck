package ansix

import (
	"image/color"
	"math"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

// fadeSteps quantizes fades so that frames of a slow fade repeat exactly and
// the renderer can skip unchanged lines.
const fadeSteps = 32

// Fade blends the colors of s towards target by k: 0 keeps s, 1 paints
// everything in target. Only SGR colors change, so widths are kept. Text in
// the terminal's default foreground fades from defaultFg; the terminal's
// default background is left alone.
func Fade(s string, k float64, target, defaultFg color.Color) string {
	k = math.Round(min(max(k, 0), 1)*fadeSteps) / fadeSteps
	if k == 0 || target == nil || s == "" {
		return s
	}
	tc, ok := colorful.MakeColor(target)
	if !ok {
		return s
	}
	memo := map[color.Color]color.Color{}
	blend := func(c color.Color) color.Color {
		if c == nil {
			return nil
		}
		if b, ok := memo[c]; ok {
			return b
		}
		cc, ok := colorful.MakeColor(c)
		if !ok {
			return c
		}
		r, g, b := cc.BlendLab(tc, k).Clamped().RGB255()
		out := color.RGBA{r, g, b, 0xFF}
		memo[c] = out
		return out
	}

	var b strings.Builder
	b.Grow(len(s) + len(s)/4)
	var state byte
	p := ansi.NewParser()
	var pen, out uv.Style
	for len(s) > 0 {
		seq, w, n, next := ansi.DecodeSequenceWc(s, state, p)
		state, s = next, s[n:]
		if w == 0 {
			if strings.HasPrefix(seq, "\x1b[") && strings.HasSuffix(seq, "m") {
				uv.ReadStyle(p.Params(), &pen)
				continue
			}
			if seq == "\n" && !out.IsZero() { // each line stands on its own, for callers that cut lines
				b.WriteString(reset)
				out = uv.Style{}
			}
			b.WriteString(seq)
			continue
		}
		want := pen
		fg := pen.Fg
		if fg == nil {
			fg = defaultFg
		}
		want.Fg = blend(fg)
		want.Bg = blend(pen.Bg)
		want.UnderlineColor = blend(pen.UnderlineColor)
		if d := uv.StyleDiff(&out, &want); d != "" {
			b.WriteString(d)
			out = want
		}
		b.WriteString(seq)
	}
	if !out.IsZero() {
		b.WriteString(reset)
	}
	return b.String()
}
