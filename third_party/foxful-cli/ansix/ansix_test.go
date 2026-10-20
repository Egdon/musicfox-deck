package ansix

import (
	"image/color"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

var corpus = []string{
	"",
	"hello",
	"王子系变装",
	"\x1b[31mred\x1b[0m plain",
	"\x1b[38;2;210;113;79m▎\x1b[0m  01 遗憾总是贯穿人生始终 \x1b[2m边缘,粘了咕叽\x1b[0m",
	"❤️ heart",
	"👨‍👩‍👧 family",
	"café au lait",
	" 10:31",
	"a\tb",
	"\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\ after",
	"\x1b[48;2;21;20;18m\x1b[38;5;214m▀▄█ 青山テルマ — GRAY SMOKE\x1b[m",
}

func TestFitAndShiftAreExact(t *testing.T) {
	for _, s := range corpus {
		for w := range 34 {
			if got := Width(Fit(s, w, "\x1b[48;2;1;2;3m")); got != w {
				t.Errorf("Fit(%q, %d) is %d wide", s, w, got)
			}
			for _, dx := range []int{-40, -5, -1, 1, 3, 40} {
				if got := Width(Shift(s, dx, w, "")); got != w {
					t.Errorf("Shift(%q, %d, %d) is %d wide", s, dx, w, got)
				}
			}
		}
	}
}

func TestCutSpansExactCells(t *testing.T) {
	for _, s := range corpus {
		n := Width(s)
		if got, want := ansi.Strip(Cut(s, 0, n)), ansi.Strip(s); got != want {
			t.Errorf("Cut(%q, 0, %d) lost text: %q", s, n, got)
		}
		for l := range n + 1 {
			for r := l; r <= n; r++ {
				if got := Width(Cut(s, l, r)); got != r-l {
					t.Fatalf("Cut(%q, %d, %d) is %d wide", s, l, r, got)
				}
			}
		}
	}
}

func TestCutKeepsStyles(t *testing.T) {
	if got, want := Cut("\x1b[31mred\x1b[0m plain", 1, 3), "\x1b[31med\x1b[m"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := Cut("王子系变装", 1, 4), " 子"; got != want {
		t.Errorf("wide left edge: got %q, want %q", got, want)
	}
	if got, want := Cut("王子系变装", 1, 5), " 子 "; got != want {
		t.Errorf("wide edges: got %q, want %q", got, want)
	}
}

func TestShiftMovesContent(t *testing.T) {
	if got, want := Shift("abcdef", 2, 6, ""), "  abcd"; got != want {
		t.Errorf("right: got %q, want %q", got, want)
	}
	if got, want := Shift("abcdef", -2, 6, ""), "cdef  "; got != want {
		t.Errorf("left: got %q, want %q", got, want)
	}
}

func TestTruncate(t *testing.T) {
	for _, s := range corpus {
		for w := range 20 {
			if got := Width(Truncate(s, w, "…")); got > w {
				t.Errorf("Truncate(%q, %d) is %d wide", s, w, got)
			}
		}
	}
	if got := Truncate("遗憾总是贯穿人生始终", 7, "…"); got != "遗憾总…" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeAgreesOnWidth(t *testing.T) {
	for _, s := range corpus {
		n := Normalize(s)
		if ansi.StringWidth(n) != ansi.StringWidthWc(n) {
			t.Errorf("Normalize(%q) = %q: grapheme width %d, wcwidth %d", s, n, ansi.StringWidth(n), ansi.StringWidthWc(n))
		}
	}
	if got := Normalize("a\tb"); got != "a b" {
		t.Errorf("tab: got %q", got)
	}
}

func TestFadeKeepsTextAndWidth(t *testing.T) {
	target, fg := color.RGBA{21, 20, 18, 255}, color.RGBA{203, 196, 182, 255}
	for _, s := range corpus {
		if Fade(s, 0, target, fg) != s {
			t.Errorf("Fade(%q, 0) changed the string", s)
		}
		for _, k := range []float64{0.1, 0.5, 1} {
			f := Fade(s, k, target, fg)
			if ansi.Strip(f) != ansi.Strip(s) || Width(f) != Width(s) {
				t.Errorf("Fade(%q, %v) = %q changed text or width", s, k, f)
			}
		}
	}
}

func TestFadeToTargetPaintsTarget(t *testing.T) {
	target, fg := color.RGBA{21, 20, 18, 255}, color.RGBA{203, 196, 182, 255}
	s := "plain \x1b[38;2;210;113;79;48;2;36;34;31mstyled\x1b[0m \x1b[31mred\x1b[m"
	f := Fade(s, 1, target, fg)
	var state byte
	p := ansi.NewParser()
	var pen uv.Style
	for rest := f; len(rest) > 0; {
		seq, w, n, next := ansi.DecodeSequenceWc(rest, state, p)
		state, rest = next, rest[n:]
		if w == 0 {
			if strings.HasPrefix(seq, "\x1b[") && strings.HasSuffix(seq, "m") {
				uv.ReadStyle(p.Params(), &pen)
			}
			continue
		}
		if seq == " " && pen.Bg == nil {
			continue
		}
		if !sameRGB(pen.Fg, target) || (pen.Bg != nil && !sameRGB(pen.Bg, target)) {
			t.Fatalf("%q drawn in fg %v bg %v, want the target", seq, pen.Fg, pen.Bg)
		}
	}
}

func sameRGB(a, b color.Color) bool {
	if a == nil || b == nil {
		return false
	}
	r1, g1, b1, _ := a.RGBA()
	r2, g2, b2, _ := b.RGBA()
	return r1>>8 == r2>>8 && g1>>8 == g2>>8 && b1>>8 == b2>>8
}
