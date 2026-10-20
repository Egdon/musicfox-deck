package ui

import (
	"image/color"
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/model"
	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/ui/deck"
)

func deckPanelTestSnapshot(t *testing.T, w, h int, idle bool) *deckPanelSnapshot {
	t.Helper()
	var r deck.Renderer
	cw, ch := deckPanelCassette(w, h, 0)
	frame := r.Render(deck.View{W: cw, H: ch, Tier: deck.Octant}, deckPaletteOf(npWarmNight), deck.RestPose(),
		deck.State{Radii: deck.ReelRadii(0.3)})
	s := &deckPanelSnapshot{Pal: npWarmNight, Deck: frame, Idle: idle}
	if idle {
		return s
	}
	s.Song = structs.Song{Id: 7, Name: "一首很长很长很长很长很长很长很长的歌名", Artists: []structs.Artist{{Name: "青山テルマ"}}}
	s.Song.Album.Name = "GRAY SMOKE"
	s.Pos, s.Total = 7, 20
	s.Lyrics = []lyricLine{{Text: "回响着 每每想起身体麻木到动弹不得"}, {Text: "ON & ON 唇重ねながら見つめる BABY"}, {Text: "I wanna have fun tonight"}}
	s.Cur, s.Sung = 1, 0.4
	for i := range 30 {
		s.Queue = append(s.Queue, structs.Song{Id: int64(100 + i), Name: "Night Drive ❤️ 海边的卡带", Artists: []structs.Artist{{Name: "Tape Unit"}}})
	}
	return s
}

func TestDeckPanelLinesAreExact(t *testing.T) {
	for _, idle := range []bool{false, true} {
		for _, size := range [][2]int{{36, 24}, {40, 34}, {56, 50}} {
			w, h := size[0], size[1]
			lines, layout := renderDeckPanel(deckPanelTestSnapshot(t, w, h, idle), w, h)
			if len(lines) != h {
				t.Fatalf("%dx%d idle=%v: %d lines", w, h, idle, len(lines))
			}
			for i, l := range lines {
				if got := ansix.Width(l); got != w {
					t.Fatalf("%dx%d idle=%v: line %d is %d cells: %q", w, h, idle, i, got, ansi.Strip(l))
				}
			}
			if layout.deck.Min.Y != 2 || layout.deck.Dx() != w-2 {
				t.Fatalf("%dx%d: cassette at %v", w, h, layout.deck)
			}
			if idle != (layout.queueLen == 0) {
				t.Fatalf("%dx%d idle=%v: %d up-next rows", w, h, idle, layout.queueLen)
			}
			if !idle && !strings.Contains(ansi.Strip(lines[layout.queueTop]), "01") {
				t.Fatalf("%dx%d: the first up-next row is %q", w, h, ansi.Strip(lines[layout.queueTop]))
			}
		}
	}
}

func TestKaraokeColors(t *testing.T) {
	lit, dim := color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255}
	same := func(a, b color.Color) bool {
		r1, g1, b1, _ := a.RGBA()
		r2, g2, b2, _ := b.RGBA()
		return r1 == r2 && g1 == g2 && b1 == b2
	}
	for i, c := range karaokeColors("唱歌ab", 0, lit, dim) {
		if !same(c, dim) {
			t.Fatalf("nothing sung, rune %d lit", i)
		}
	}
	for i, c := range karaokeColors("唱歌ab", 1, lit, dim) {
		if !same(c, lit) {
			t.Fatalf("all sung, rune %d dim", i)
		}
	}
	// 6 cells; four sung light 唱 and 歌 fully and nothing else.
	cs := karaokeColors("唱歌ab", 4.0/6, lit, dim)
	if !same(cs[0], lit) || !same(cs[1], lit) || !same(cs[2], dim) {
		t.Fatalf("half sung: %v", cs)
	}
	// A third of 歌 sung leaves it in between.
	cs = karaokeColors("唱歌", 0.75, lit, dim)
	if same(cs[1], lit) || same(cs[1], dim) {
		t.Fatalf("the edge rune is not in between: %v", cs[1])
	}
}

type headerSongsMenu struct{ baseMenu }

func (*headerSongsMenu) Songs() []structs.Song { return nil }

func TestMenuHeaderUnits(t *testing.T) {
	if got := MenuHeader(nil, &headerSongsMenu{}, model.MenuHeader{Meta: "100"}).Meta; got != "100 首" {
		t.Fatalf("songs: %q", got)
	}
	if got := MenuHeader(nil, &pageLayoutTestMenu{}, model.MenuHeader{Meta: "8"}).Meta; got != "8" {
		t.Fatalf("other menus: %q", got)
	}
}
