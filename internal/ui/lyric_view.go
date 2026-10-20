package ui

import (
	"image/color"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/lyric"
)

// lyricLine is one line of lyrics as the views draw it.
type lyricLine struct {
	Text, Trans string
	start       int64 // ms
}

// lyricsNow returns the lyrics, the line being sung at pos and how much of it
// has been sung, from 0 to 1. Without word timing a current line counts as
// sung in full. cur is -1 before the first line.
func lyricsNow(svc *lyric.Service, pos time.Duration) (lines []lyricLine, cur int, sung float64) {
	if svc == nil {
		return nil, -1, 0
	}
	st := svc.State()
	t := pos.Milliseconds() + st.OffsetMs
	cur = -1
	switch {
	case st.YRCEnabled && len(st.YRCLines) > 0:
		lines = make([]lyricLine, len(st.YRCLines))
		for i, l := range st.YRCLines {
			var b strings.Builder
			for _, w := range l.Words {
				b.WriteString(w.Word)
			}
			lines[i] = lyricLine{Text: b.String(), start: l.StartTime}
			if st.ShowTranslation {
				lines[i].Trans = l.TranslatedLyric
			}
			if l.StartTime <= t {
				cur = i
			}
		}
		if cur >= 0 {
			sung = yrcSung(st.YRCLines[cur], t)
		}
	case len(st.Fragments) > 0:
		lines = make([]lyricLine, len(st.Fragments))
		for i, f := range st.Fragments {
			lines[i] = lyricLine{Text: f.Content, start: f.StartTimeMs}
			if st.ShowTranslation {
				lines[i].Trans = st.TranslatedFragments[f.StartTimeMs]
			}
			if f.StartTimeMs <= t {
				cur = i
			}
		}
		sung = 1
	}
	return lines, cur, sung
}

// yrcSung measures the sung part of a line in display cells, so the light
// moves evenly across wide and narrow characters.
func yrcSung(l lyric.YRCLine, t int64) float64 {
	p := lyric.ProgressYRCLineAtTimeMs(l, t)
	total, done := 0.0, 0.0
	for i, w := range l.Words {
		ww := float64(runewidth.StringWidth(w.Word))
		total += ww
		switch {
		case i < p.CompletedWords:
			done += ww
		case i == p.CurrentWord:
			done += ww * p.CurrentProgress
		}
	}
	if total == 0 {
		return 1
	}
	return min(done/total, 1)
}

// karaokeColors returns the color of every rune of text with sung (0..1) of
// it lit: lit runes in lit, the rest in dim, and the rune at the edge in
// between, so the light glides rather than jumps.
func karaokeColors(text string, sung float64, lit, dim color.Color) []color.Color {
	runes := []rune(text)
	widths := make([]float64, len(runes))
	total := 0.0
	for i, r := range runes {
		widths[i] = float64(runewidth.RuneWidth(r))
		total += widths[i]
	}
	edge := sung * total
	colors := make([]color.Color, len(runes))
	x := 0.0
	for i := range runes {
		switch {
		case widths[i] == 0 || x+widths[i] <= edge:
			colors[i] = lit
		case x >= edge:
			colors[i] = dim
		default:
			colors[i] = npLerp(dim, lit, (edge-x)/widths[i])
		}
		x += widths[i]
	}
	return colors
}
