package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/ui/deck"
)

const (
	deckPanelMinWidth  = 110 // narrower windows keep the classic bottom area
	deckPanelMinHeight = 24
	deckPanelFooter    = 3 // song info (2 rows) and progress bar
	deckPanelLyrics    = 5
	deckPanelQueue     = 30
)

// deckJumpMsg plays the k-th song of the panel's up-next list.
type deckJumpMsg struct{ k int }

// DeckPanel is the column at the right of the main screen: a small cassette
// that plays along, the track, its lyrics and what plays next.
type DeckPanel struct {
	netease *Netease
	handler *EventHandler

	renderer deck.Renderer
	motion   deck.Motion
	reels    deck.Reels
	stepped  time.Time
	songID   int64

	pal    npPalette
	palGen uint64
	palOK  bool

	queue   []structs.Song
	queueAt time.Time

	// Where things were drawn in the last frame, in panel cells.
	deckBox  image.Rectangle
	queueTop int
	queueLen int
}

// NewDeckPanel returns the main screen's side panel.
func NewDeckPanel(n *Netease, h *EventHandler) *DeckPanel {
	p := &DeckPanel{netease: n, handler: h}
	p.motion.FloatAmount = 0.5
	return p
}

func (p *DeckPanel) Layout(_ *model.App, w, h int) model.SidePanelLayout {
	if !configs.AppConfig.Theme.DeckLayout() || w < deckPanelMinWidth || h < deckPanelMinHeight {
		return model.SidePanelLayout{}
	}
	return model.SidePanelLayout{Width: min(max(int(float64(w)*0.34+0.5), 36), 56), FooterRows: deckPanelFooter}
}

func (p *DeckPanel) Update(msg tea.Msg, a *model.App) {
	if m, ok := msg.(deckJumpMsg); ok {
		p.jump(m.k)
		return
	}
	p.step(a, time.Now())
}

func (p *DeckPanel) currentSong() structs.Song {
	song, _ := p.netease.player.playlistManager.GetCurrentSong()
	return song
}

// step moves the cassette on to now and asks for frames while it moves.
func (p *DeckPanel) step(a *model.App, now time.Time) {
	player := p.netease.player
	song := p.currentSong()
	playing := player.State() == types.Playing
	level := a.MotionLevel()
	if song.Id != p.songID {
		switch {
		case p.songID == 0:
		case level == motion.Full:
			p.motion.Flip(1, now)
		default:
			p.motion.FlipNow()
		}
		p.songID = song.Id
		p.reels.Reset(song.Id)
	}
	dt := 0.0
	if !p.stepped.IsZero() {
		dt = min(max(now.Sub(p.stepped).Seconds(), 0), 0.1)
	}
	p.stepped = now
	p.motion.FloatAmount = 0
	if level == motion.Full {
		p.motion.FloatAmount = 0.5
	}
	p.motion.Step(now, playing)
	if level != motion.Off {
		sideB := p.motion.SideB(now)
		dir := 1.0
		if sideB {
			dir = -1
		}
		p.reels.Advance(dt, playing, now, deck.PackRadii(p.progress(), sideB), dir)
	}
	main := p.netease.MustMain()
	if main == nil || !main.PanelVisible() {
		return
	}
	switch {
	case p.motion.Flipping(now):
		a.Animate(100 * time.Millisecond)
	case playing || p.motion.Busy(now) || p.reels.Moving():
		a.AnimateSlow(200 * time.Millisecond)
	}
}

func (p *DeckPanel) progress() float64 {
	d := p.currentSong().Duration
	if d <= 0 {
		return 0
	}
	return min(max(float64(p.netease.player.PassedTime())/float64(d), 0), 1)
}

func (p *DeckPanel) palette() npPalette {
	if gen := style.StyleGeneration(); !p.palOK || gen != p.palGen {
		p.pal, p.palGen, p.palOK = npPaletteFromTheme(style.CurrentStyleSet().Theme()), gen, true
	}
	return p.pal
}

// upcoming returns what plays next, refreshed every second: some playlist
// edits do not mark the playlist as changed.
func (p *DeckPanel) upcoming(now time.Time) []structs.Song {
	if now.Sub(p.queueAt) > time.Second {
		p.queue = p.netease.player.playlistManager.Upcoming(deckPanelQueue)
		p.queueAt = now
	}
	return p.queue
}

// panelText builds a styled line from colored runs.
type panelText struct {
	b  strings.Builder
	bg color.Color
}

func (t *panelText) add(s string, fg color.Color, bold bool) {
	if s == "" {
		return
	}
	r, g, b, _ := fg.RGBA()
	fmt.Fprintf(&t.b, "\x1b[38;2;%d;%d;%dm", r>>8, g>>8, b>>8)
	if bold {
		t.b.WriteString("\x1b[1m")
	}
	t.b.WriteString(s)
	if bold {
		t.b.WriteString("\x1b[22m")
	}
}

func (t *panelText) String() string {
	r, g, b, _ := t.bg.RGBA()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r>>8, g>>8, b>>8) + t.b.String() + "\x1b[m"
}

func (p *DeckPanel) View(a *model.App, _ *model.Main, w, h int) string {
	now := a.FrameTime()
	player := p.netease.player
	song := p.currentSong()
	s := &deckPanelSnapshot{Pal: p.palette(), Song: song, Idle: song.Id == 0}
	if total := len(player.playlistManager.GetPlaylist()); total > 0 {
		s.Pos, s.Total = player.CurSongIndex()+1, total
	}
	cw, ch := deckPanelCassette(w, h, deckAspect(a))
	pose, lift := p.motion.At(now)
	sideB := p.motion.SideB(now)
	s.Deck = p.renderer.Render(deck.View{W: cw, H: ch, Tier: deckTier(a), Aspect: deckAspect(a)}, deckPaletteOf(s.Pal), pose,
		deck.State{Radii: deck.PackRadii(p.progress(), sideB), Angles: p.reels.Angles, Blur: p.reels.Blur(), Lift: lift})
	if !s.Idle {
		s.Lyrics, s.Cur, s.Sung = lyricsNow(p.netease.lyricService, player.PassedTime())
		s.Queue = p.upcoming(now)
	}
	lines, layout := renderDeckPanel(s, w, h)
	p.deckBox, p.queueTop, p.queueLen = layout.deck, layout.queueTop, layout.queueLen
	return strings.Join(lines, "\n")
}

// deckPanelCassette sizes the cassette to the panel's width, leaving room for
// the text below it.
func deckPanelCassette(w, h int, aspect float64) (cw, ch int) {
	if aspect <= 0 {
		aspect = 2
	}
	cw = w - 2
	return cw, min(int(float64(cw)*60/88/aspect+0.5), max((h-12)/2, 4))
}

// deckPanelSnapshot is what the panel shows in one frame.
type deckPanelSnapshot struct {
	Pal        npPalette
	Song       structs.Song
	Idle       bool
	Pos, Total int
	Deck       *deck.Frame
	Lyrics     []lyricLine
	Cur        int
	Sung       float64
	Queue      []structs.Song
}

// deckPanelLayout is where things went, in panel cells.
type deckPanelLayout struct {
	deck               image.Rectangle
	queueTop, queueLen int
}

// renderDeckPanel draws the panel: exactly h lines of exactly w cells.
func renderDeckPanel(s *deckPanelSnapshot, w, h int) ([]string, deckPanelLayout) {
	var out deckPanelLayout
	pal := s.Pal
	bgSGR := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", pal.Bg.R, pal.Bg.G, pal.Bg.B)
	lines := make([]string, 0, h)
	line := func(build func(t *panelText)) {
		t := &panelText{bg: pal.Bg}
		build(t)
		lines = append(lines, ansix.Fit(t.String(), w, bgSGR))
	}
	blank := func() { line(func(*panelText) {}) }
	text := func(str string, max int) string { return ansix.Truncate(ansix.Normalize(str), max, "…") }

	// Header: section mark and queue position.
	line(func(t *panelText) {
		label := npSpaced("NOW PLAYING")
		t.add("■ ", pal.Accent, false)
		t.add(label, pal.Secondary, false)
		if s.Total > 0 && !s.Idle {
			pos := fmt.Sprintf("%02d / %02d", s.Pos, s.Total)
			t.add(strings.Repeat(" ", max(w-2-npWidth(label)-len(pos)-1, 1)), pal.Secondary, false)
			t.add(pos, pal.Secondary, false)
		}
	})
	blank()

	if s.Deck != nil {
		out.deck = image.Rect(1, len(lines), 1+s.Deck.W, len(lines)+s.Deck.H)
		for _, l := range deckLines(s.Deck, pal.Bg) {
			lines = append(lines, ansix.Fit(bgSGR+" \x1b[m"+l, w, bgSGR))
		}
	}
	blank()

	if s.Idle {
		line(func(t *panelText) { t.add(" 未在播放", pal.Secondary, false) })
		blank()
		for _, hint := range [][2]string{{"Enter", "播放选中的歌"}, {"Space", "播放 / 暂停"}, {"i", "打开播放页"}} {
			line(func(t *panelText) {
				t.add(" "+hint[0], pal.Text, false)
				t.add(strings.Repeat(" ", max(8-len(hint[0]), 1))+hint[1], pal.Secondary, false)
			})
		}
		return fitPanelRows(lines, h, w, bgSGR), out
	}

	line(func(t *panelText) { t.add(" "+text(s.Song.Name, w-2), pal.Hero, true) })
	line(func(t *panelText) { t.add(" "+text(npArtistAlbum(s.Song), w-2), pal.Secondary, false) })
	blank()

	// Lyrics around the line being sung.
	lyricRows := min(deckPanelLyrics, max(h-len(lines)-4, 0))
	if lyricRows > 0 && len(s.Lyrics) == 0 {
		line(func(t *panelText) { t.add("   暂无歌词", pal.Rule, false) })
	}
	first := max(s.Cur-1, 0)
	for i := first; i < first+lyricRows && i < len(s.Lyrics); i++ {
		l := s.Lyrics[i]
		line(func(t *panelText) {
			str := text(l.Text, w-4)
			switch {
			case i == s.Cur:
				t.add(" ▸ ", pal.Accent, false)
				runes := []rune(str)
				for j, c := range karaokeColors(str, s.Sung, pal.Hero, pal.Secondary) {
					t.add(string(runes[j]), c, true)
				}
			case i == s.Cur+1:
				t.add("   "+str, pal.Secondary, false)
			default:
				t.add("   "+str, pal.Rule, false)
			}
		})
	}

	// Up next fills the rest.
	if rows := h - len(lines) - 2; rows > 0 && len(s.Queue) > 0 {
		blank()
		line(func(t *panelText) { t.add(" "+npSpaced("UP NEXT"), pal.Secondary, false) })
		out.queueTop = len(lines)
		for k, song := range s.Queue[:min(rows, len(s.Queue))] {
			line(func(t *panelText) {
				t.add(fmt.Sprintf(" %02d  ", k+1), pal.Rule, false)
				name := text(song.Name, w-6)
				t.add(name, pal.Text, false)
				if rest := w - 6 - npWidth(name) - 3; rest > 2 {
					t.add(" · "+text(artistsString(song), rest), pal.Secondary, false)
				}
			})
			out.queueLen++
		}
	}
	return fitPanelRows(lines, h, w, bgSGR), out
}

// fitPanelRows pads or cuts the panel to h rows.
func fitPanelRows(lines []string, h, w int, bgSGR string) []string {
	if len(lines) > h {
		return lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, ansix.Fit("", w, bgSGR))
	}
	return lines
}

func (p *DeckPanel) HandleMouse(a *model.App, _ *model.Main, msg tea.MouseMsg, x, y int) model.SidePanelMouseResult {
	now := time.Now()
	mouse := msg.Mouse()
	onDeck := image.Pt(x, y).In(p.deckBox)
	queueRow := -1
	if y >= p.queueTop && y < p.queueTop+p.queueLen {
		queueRow = y - p.queueTop
	}
	pointer := ""
	if onDeck || queueRow >= 0 {
		pointer = "pointer"
	}
	switch msg.(type) {
	case tea.MouseMotionMsg:
		if onDeck && a.MotionLevel() == motion.Full {
			c := p.deckBox.Min.Add(p.deckBox.Max).Div(2)
			p.motion.Point(float64(x-c.X)/float64(max(p.deckBox.Dx()/2, 1)), float64(y-c.Y)/float64(max(p.deckBox.Dy()/2, 1)), now)
			a.Animate(100 * time.Millisecond)
		} else {
			p.motion.Leave()
		}
	case tea.MouseClickMsg:
		if mouse.Button != tea.MouseLeft {
			break
		}
		switch {
		case onDeck:
			page, cmd := p.netease.ToNowPlayingPage(p.handler)
			return model.SidePanelMouseResult{Handled: true, Page: page, Cmd: cmd, Pointer: pointer}
		case queueRow >= 0:
			k := queueRow
			return model.SidePanelMouseResult{Handled: true, Pointer: pointer, Cmd: func() tea.Msg { return deckJumpMsg{k: k} }}
		}
	case tea.MouseWheelMsg:
		if mouse.Button == tea.MouseWheelUp || mouse.Button == tea.MouseWheelDown {
			p.handler.wheelVolume(mouse)
			p.handler.osd.show(a, osdVolumeText(p.netease.player.Volume()))
		}
	}
	return model.SidePanelMouseResult{Handled: true, Pointer: pointer}
}

func (p *DeckPanel) MouseLeave(*model.App, *model.Main) { p.motion.Leave() }

// jump plays the k-th song of the up-next list, stepping the playlist so a
// shuffled order is kept.
func (p *DeckPanel) jump(k int) {
	player := p.netease.player
	if player.Mode() == types.PmSingleLoop {
		return
	}
	for range k {
		if _, err := player.playlistManager.NextSong(true); err != nil {
			return
		}
	}
	player.NextSong(true)
}

// MenuHeader adds units to the item count in a tracklist's section header.
func MenuHeader(_ *model.App, menu model.Menu, h model.MenuHeader) model.MenuHeader {
	if _, ok := menu.(*SearchResultMenu); ok {
		return h
	}
	switch menu.(type) {
	case SongsMenu:
		h.Meta += " 首"
	case PlaylistsMenu:
		h.Meta += " 个歌单"
	case AlbumsMenu:
		h.Meta += " 张专辑"
	case ArtistsMenu:
		h.Meta += " 位歌手"
	}
	return h
}
