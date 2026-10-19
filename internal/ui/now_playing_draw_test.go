package ui

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func npTestSnapshot() *npSnapshot {
	song := func(id int64, name, artist string) structs.Song {
		return structs.Song{
			Id:       id,
			Name:     name,
			Duration: 4*time.Minute + time.Second,
			Artists:  []structs.Artist{{Name: artist}},
			Album:    structs.Album{Name: "GRAY SMOKE"},
		}
	}
	return &npSnapshot{
		Now:      time.Date(2026, 10, 18, 21, 43, 0, 0, time.Local),
		Pal:      npWarmNight,
		Song:     song(1, "ON & ON", "青山テルマ"),
		State:    types.Playing,
		Passed:   76 * time.Second,
		Volume:   35,
		Mode:     types.PmListLoop,
		Liked:    true,
		Quality:  service.Exhigh,
		QueuePos: 7,
		QueueLen: 20,
		TapeMins: 74,
		Upcoming: []structs.Song{song(2, "海边的卡带 ❤️ 一首很长很长很长很长很长的歌名", "某某乐队"), song(3, "Night Drive", "Tape Unit")},
		Lyrics: []npLyricLine{
			{Text: "ON & ON & ON 思い出すと身体しびれる FOREPLAY", Trans: "回响着 每每想起身体麻木到动弹不得"},
			{Text: "ON & ON & ON 唇重ねながら見つめる BABY", Trans: "回响着 双唇交叠着感受你的专属体温"},
			{Text: "I WANNA HAVE FUN TONIGHT"},
		},
		LyricIndex: 1,
		OSD:        "▶ PLAY",
		Keys:       npKeys{Prev: "[", Toggle: "SPC", Next: "]", Like: ",", Mode: "p", VolDown: "-", VolUp: "="},
		Angles:     [2]float64{1, 2},
	}
}

// Every frame must be exactly w×h cells under both width methods, with every
// cell painted, whatever the layout and effect.
func TestNowPlayingFrameFillsScreen(t *testing.T) {
	sizes := [][2]int{{140, 45}, {120, 40}, {100, 30}, {99, 30}, {80, 24}, {60, 20}, {59, 20}, {50, 15}, {30, 6}, {20, 2}, {12, 1}}
	type variant struct {
		name string
		snap func(*npSnapshot)
		fx   func(*npScreen)
	}
	variants := []variant{
		{name: "no deck"},
		{name: "deck octants", snap: func(s *npSnapshot) { s.Deck = npTestDeckFrame(true, npPose{}) }},
		{name: "deck quadrants", snap: func(s *npSnapshot) { s.Deck = npTestDeckFrame(false, npPose{}) }},
		{name: "deck ejecting", snap: func(s *npSnapshot) { s.Deck = npTestDeckFrame(true, npPose{DY: -10, Tilt: 0.3, Fade: 0.4}) }},
		{name: "lyrics", snap: func(s *npSnapshot) { s.LyricsMode = true }},
		{name: "blue", snap: func(s *npSnapshot) { s.BlueScreen, s.State, s.Scanlines = true, types.Stopped, true }},
		{name: "empty", snap: func(s *npSnapshot) { *s = npSnapshot{Now: s.Now, Pal: s.Pal, Keys: s.Keys} }},
		{name: "shuffle", snap: func(s *npSnapshot) { s.Mode, s.Upcoming = types.PmInfRandom, nil }},
		{name: "osd fading", snap: func(s *npSnapshot) { s.OSDFade = 0.5 }},
		{name: "sweep", fx: func(scr *npScreen) { scr.sweep(0.5) }},
		{name: "power-on", fx: func(scr *npScreen) { scr.powerOn(0.5) }},
		{name: "power-off", fx: func(scr *npScreen) { scr.powerOff(0.9) }},
	}
	for _, size := range sizes {
		w, h := size[0], size[1]
		for _, v := range variants {
			name := v.name
			s := npTestSnapshot()
			if v.snap != nil {
				v.snap(s)
			}
			var scr npScreen
			scr.reset(w, h)
			npRender(&scr, s)
			scr.fillBackground()
			if v.fx != nil {
				v.fx(&scr)
			}

			out := scr.render()
			lines := strings.Split(out, "\n")
			if len(lines) != h {
				t.Fatalf("%dx%d %s: %d lines, want %d", w, h, name, len(lines), h)
			}
			for y, line := range lines {
				if gw, wc := ansi.GraphemeWidth.StringWidth(line), ansi.WcWidth.StringWidth(line); gw != w || wc != w {
					t.Fatalf("%dx%d %s: line %d width grapheme=%d wc=%d, want %d: %q", w, h, name, y, gw, wc, w, line)
				}
			}
			screen := customPageStyledScreen(out, w, h)
			for y := range h {
				for x := range w {
					if c := screen.CellAt(x, y); c != nil && c.Width > 0 && c.Style.Bg == nil {
						t.Fatalf("%dx%d %s: cell (%d,%d) %q has no background", w, h, name, x, y, c.Content)
					}
				}
			}
		}
	}
}

func TestNowPlayingSevenSegmentDigits(t *testing.T) {
	lit, ghost := npWarmNight.Text, npWarmNight.Ghost
	litCount := map[int]int{1: 7, 7: 10, 8: 20}
	for digit, want := range litCount {
		p := newNPPixels(4, 8)
		npDrawDigit(p, 0, digit, lit, ghost)
		n := 0
		for _, c := range p.px {
			if npSameColor(c, lit) {
				n++
			}
		}
		if n != want {
			t.Errorf("digit %d lights %d pixels, want %d", digit, n, want)
		}
	}
	// The inner pixels are never part of a segment.
	p := newNPPixels(4, 8)
	npDrawDigit(p, 0, 8, lit, ghost)
	for _, xy := range [][2]int{{1, 1}, {2, 2}, {1, 4}, {2, 5}} {
		if p.at(xy[0], xy[1]) != nil {
			t.Errorf("inner pixel %v of 8 must stay empty", xy)
		}
	}
}

func TestNowPlayingReelRadii(t *testing.T) {
	centerGap := npReelCenters[1][0] - npReelCenters[0][0]
	prev := npReelRadii(0)
	if prev[0] != npReelMax || prev[1] != npReelMin {
		t.Fatalf("radii at start = %v, want [%v %v]", prev, npReelMax, npReelMin)
	}
	for i := 1; i <= 20; i++ {
		r := npReelRadii(float64(i) / 20)
		if r[0] >= prev[0] || r[1] <= prev[1] {
			t.Fatalf("radii not monotonic at %d/20: %v after %v", i, r, prev)
		}
		if r[0]+r[1] >= centerGap-1 {
			t.Fatalf("reels overlap at %d/20: %v", i, r)
		}
		if area := r[0]*r[0] + r[1]*r[1]; math.Abs(area-(npReelMin*npReelMin+npReelMax*npReelMax)) > 1e-9 {
			t.Fatalf("tape area not conserved at %d/20: %v", i, area)
		}
		prev = r
	}
}

func TestNowPlayingLegendFitsAndKeepsBack(t *testing.T) {
	for _, w := range []int{140, 100, 80, 60, 40, 24} {
		var scr npScreen
		scr.reset(w, 3)
		hits := npLegend(&scr, 2, npTestSnapshot(), false)
		hasBack := false
		end := 0
		for _, h := range hits {
			if h.x < end {
				t.Fatalf("width %d: overlapping hit %+v", w, h)
			}
			if h.x+h.w > w {
				t.Fatalf("width %d: hit %+v exceeds the screen", w, h)
			}
			end = h.x + h.w
			hasBack = hasBack || h.kind == npHitBack
		}
		if !hasBack {
			t.Fatalf("width %d: legend lost the BACK button", w)
		}
	}
}

func TestOSDCanUpdate(t *testing.T) {
	now := time.Now()
	created := now.Add(-time.Second)
	expire := now.Add(500 * time.Millisecond)
	tests := []struct {
		name    string
		created time.Time
		expire  time.Time
		foreign time.Time
		want    bool
	}{
		{"no osd yet", time.Time{}, time.Time{}, time.Time{}, false},
		{"expired", created, now, time.Time{}, false},
		{"no foreign notification", created, expire, time.Unix(0, 0), true},
		{"foreign long before", created, expire, created.Add(-2 * time.Second), true},
		{"foreign right before", created, expire, created.Add(-100 * time.Millisecond), false},
		{"foreign right after", created, expire, created.Add(100 * time.Millisecond), false},
		{"foreign long after", created, expire, created.Add(800 * time.Millisecond), true},
	}
	for _, tt := range tests {
		if got := osdCanUpdate(now, tt.created, tt.expire, tt.foreign); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestOSDTexts(t *testing.T) {
	if got := osdVolumeText(35); got != "VOL ▪▪▪▪▫▫▫▫▫▫ 35" {
		t.Errorf("volume text = %q", got)
	}
	if got := osdSeekText(-5*time.Second, 71*time.Second); got != "◀◀ -5s  01:11" {
		t.Errorf("seek text = %q", got)
	}
}
