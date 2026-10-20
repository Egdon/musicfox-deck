package ansix

import (
	"image/color"
	"math"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

// Restyle passes the style of every cell in the columns [l, r) of each line
// of s through f. Text and widths are kept, and each line carries its own
// styles.
func Restyle(s string, l, r int, f func(uv.Style) uv.Style) string {
	if s == "" || r <= l {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + len(s)/4)
	var state byte
	p := ansi.NewParser()
	var pen, out uv.Style
	x := 0
	for len(s) > 0 {
		seq, w, n, next := ansi.DecodeSequenceWc(s, state, p)
		state, s = next, s[n:]
		if w == 0 {
			if strings.HasPrefix(seq, "\x1b[") && strings.HasSuffix(seq, "m") {
				uv.ReadStyle(p.Params(), &pen)
				continue
			}
			if seq == "\n" {
				if !out.IsZero() {
					b.WriteString(reset)
					out = uv.Style{}
				}
				x = 0
			}
			b.WriteString(seq)
			continue
		}
		want := pen
		if x >= l && x < r {
			want = f(pen)
		}
		if d := uv.StyleDiff(&out, &want); d != "" {
			b.WriteString(d)
			out = want
		}
		b.WriteString(seq)
		x += w
	}
	if !out.IsZero() {
		b.WriteString(reset)
	}
	return b.String()
}

// Mix blends a towards b by k in Lab space, quantized like Fade. A nil
// color stays as it is.
func Mix(a, b color.Color, k float64) color.Color {
	k = math.Round(min(max(k, 0), 1)*fadeSteps) / fadeSteps
	if a == nil || b == nil || k == 0 {
		return a
	}
	ca, ok1 := colorful.MakeColor(a)
	cb, ok2 := colorful.MakeColor(b)
	if !ok1 || !ok2 {
		return a
	}
	r, g, bl := ca.BlendLab(cb, k).Clamped().RGB255()
	return color.RGBA{r, g, bl, 0xFF}
}

// SameColor reports whether two colors are the same, nil only matching nil.
func SameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
