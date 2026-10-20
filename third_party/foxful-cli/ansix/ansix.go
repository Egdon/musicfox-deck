// Package ansix cuts, pads and shifts styled terminal lines by exact cell
// widths as the renderer counts them (wcwidth). The x/ansi helpers are not
// usable for this: CutWc returns the wrong span and TruncateWc decides
// whether to cut by grapheme width.
package ansix

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

const reset = "\x1b[m"

// Width returns the number of cells the renderer counts for s.
func Width(s string) int { return ansi.StringWidthWc(s) }

// Cut returns the cells [l, r) of s. Escape sequences before l are kept so
// the cut starts with the styles in effect there; a wide character split by
// either edge becomes spaces in its style.
func Cut(s string, l, r int) string {
	if r <= l {
		return ""
	}
	var b strings.Builder
	var state byte
	x, styled := 0, false
	for len(s) > 0 && x < r {
		seq, w, n, next := ansi.DecodeSequenceWc(s, state, nil)
		state, s = next, s[n:]
		if w == 0 {
			b.WriteString(seq)
			styled = styled || strings.HasPrefix(seq, "\x1b")
			continue
		}
		switch {
		case x+w <= l:
		case x < l: // a wide character across the left edge
			b.WriteString(strings.Repeat(" ", min(x+w, r)-l))
		case x+w > r: // a wide character across the right edge
			b.WriteString(strings.Repeat(" ", r-x))
		default:
			b.WriteString(seq)
		}
		x += w
	}
	if styled {
		b.WriteString(reset)
	}
	return b.String()
}

// Fit returns s cut or padded to exactly w cells. Padding is written after
// pad, typically a background color sequence.
func Fit(s string, w int, pad string) string {
	sw := Width(s)
	if sw == w {
		return s
	}
	if sw > w {
		return Cut(s, 0, w)
	}
	var b strings.Builder
	b.Grow(len(s) + len(pad) + w - sw + 2*len(reset))
	b.WriteString(s)
	if strings.Contains(s, "\x1b") {
		b.WriteString(reset)
	}
	b.WriteString(pad)
	b.WriteString(strings.Repeat(" ", w-sw))
	if pad != "" {
		b.WriteString(reset)
	}
	return b.String()
}

// Shift moves the content of a w-cell line by dx cells, right when positive,
// and fills the uncovered cells after pad. The result is w cells wide.
func Shift(s string, dx, w int, pad string) string {
	switch {
	case dx == 0:
		return Fit(s, w, pad)
	case dx >= w || -dx >= w:
		return Fit("", w, pad)
	case dx > 0:
		lead := strings.Repeat(" ", dx)
		if pad != "" {
			lead = pad + lead + reset
		}
		return Fit(lead+Cut(s, 0, w-dx), w, pad)
	default:
		return Fit(Cut(s, -dx, w), w, pad)
	}
}

// Truncate returns s cut to at most w cells, ending with tail when cut.
func Truncate(s string, w int, tail string) string {
	if Width(s) <= w {
		return s
	}
	tw := Width(tail)
	if tw > w {
		return Cut(tail, 0, w)
	}
	return Cut(s, 0, w-tw) + tail
}

// Normalize makes text safe to lay out: control characters become spaces and
// a cluster whose grapheme width and wcwidth disagree becomes its first rune,
// or "?" when that rune disagrees too. Escape sequences are kept.
func Normalize(s string) string {
	var b strings.Builder
	var state byte
	for len(s) > 0 {
		seq, w, n, next := ansi.DecodeSequence(s, state, nil)
		state, s = next, s[n:]
		switch {
		case strings.HasPrefix(seq, "\x1b"):
			b.WriteString(seq)
		case w == 0 && len(seq) == 1 && unicode.IsControl(rune(seq[0])):
			b.WriteByte(' ')
		case w != ansi.StringWidthWc(seq):
			if one := string([]rune(seq)[0]); ansi.StringWidth(one) == ansi.StringWidthWc(one) && ansi.StringWidth(one) > 0 {
				b.WriteString(one)
			} else {
				b.WriteByte('?')
			}
		default:
			b.WriteString(seq)
		}
	}
	return b.String()
}
