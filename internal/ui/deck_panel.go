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
	"github.com/go-musicfox/go-musicfox/utils/likelist"
)

const (
	deckPanelMinWidth  = 110 // narrower windows dock the panel under the menu
	deckPanelMinHeight = 24
	deckPanelFooter    = 3 // song info (2 rows) and progress bar
	deckPanelLyrics    = 5
	deckPanelQueue     = 30

	deckDockRows        = 6 // a rule, then five rows of cassette and track
	deckDockTallRows    = 7 // with a row of air under the rule, in windows deckDockTall rows high
	deckDockTall        = 28
	deckDockFooter      = 1  // the progress bar: the dock shows the track itself
	deckDockMinMenu     = 5  // menu rows a dock must leave
	deckDockMinCassette = 48 // narrower docks leave the cassette out

	// The dock's cassette is a few rows high, where the rest pose's
	// perspective only blurs it: it is turned this far towards the viewer
	// and fills this much more of its box, as it barely floats there.
	deckDockTurn = 0.5
	deckDockGrow = 0.1
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
	docked   bool
	queueTop int
	queueLen int
}

// NewDeckPanel returns the main screen's side panel.
func NewDeckPanel(n *Netease, h *EventHandler) *DeckPanel {
	p := &DeckPanel{netease: n, handler: h}
	p.motion.FloatAmount = 0.5
	n.deckPanel = p
	return p
}

// deckPlacement is how the panel drew its cassette, in window cells: the
// cells it may cover, the box it was fitted to and how far it was turned
// towards the viewer.
type deckPlacement struct {
	Cells, Fit deck.Rect
	Turn       float64
}

// CassettePlacement returns how the cassette was drawn in the last frame,
// and whether the panel showed.
func (p *DeckPanel) CassettePlacement() (deckPlacement, bool) {
	main := p.netease.MustMain()
	if main == nil {
		return deckPlacement{}, false
	}
	x, y, _, _, ok := main.PanelRect()
	if !ok || p.deckBox.Empty() {
		return deckPlacement{}, false
	}
	b := p.deckBox
	cells := deck.Rect{X: float64(x + b.Min.X), Y: float64(y + b.Min.Y), W: float64(b.Dx()), H: float64(b.Dy())}
	if !p.docked {
		return deckPlacement{Cells: cells, Fit: cells}, true
	}
	return deckPlacement{Cells: cells, Fit: deckGrow(cells, deckDockGrow), Turn: deckDockTurn}, true
}

// deckTurned turns a pose k of the way from the rest pose's angles towards
// facing the viewer.
func deckTurned(pose deck.Pose, k float64) deck.Pose {
	rest := deck.RestPose()
	pose.Yaw -= rest.Yaw * k
	pose.Pitch -= rest.Pitch * k
	return pose
}

// deckGrow grows a box by a fraction of its size around its center.
func deckGrow(r deck.Rect, k float64) deck.Rect {
	return deck.Rect{X: r.X - r.W*k/2, Y: r.Y - r.H*k/2, W: r.W * (1 + k), H: r.H * (1 + k)}
}

// adopt takes over the cassette's state from the now-playing page, so the
// panel carries on where the page left it.
func (p *DeckPanel) adopt(m deck.Motion, r deck.Reels) {
	m.FloatAmount = p.motion.FloatAmount
	p.motion, p.reels = m, r
}

// Layout puts the panel beside the menu in wide windows and docks it under
// the menu in narrow ones, keeping at least five rows of the list: the
// rows above it are the status bar, two blank rows, the header and its rule.
func (p *DeckPanel) Layout(_ *model.App, w, h int) model.SidePanelLayout {
	switch {
	case !configs.AppConfig.Theme.DeckLayout():
		return model.SidePanelLayout{}
	case w >= deckPanelMinWidth && h >= deckPanelMinHeight:
		return model.SidePanelLayout{Width: min(max(int(float64(w)*0.34+0.5), 36), 56), FooterRows: deckPanelFooter}
	}
	rows := deckDockRows
	if h >= deckDockTall {
		rows = deckDockTallRows
	}
	if h-5-rows-deckDockFooter >= deckDockMinMenu { // 5: the status bar and the menu's header
		return model.SidePanelLayout{Height: rows, FooterRows: deckDockFooter}
	}
	return model.SidePanelLayout{}
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
	s := &deckPanelSnapshot{Pal: p.palette(), Song: song, Idle: song.Id == 0,
		Liked: song.Id > 0 && likelist.IsLikeSong(song.Id), Mode: player.Mode(), Volume: player.Volume()}
	if total := len(player.playlistManager.GetPlaylist()); total > 0 {
		s.Pos, s.Total = player.CurSongIndex()+1, total
	}
	docked := h <= deckDockTallRows // a column is always taller
	cw, ch := deckPanelCassette(w, h, deckAspect(a))
	if docked {
		cw, ch = deckDockCassette(w, h, deckAspect(a))
	}
	if cw > 0 && ch > 0 {
		pose, lift := p.motion.At(now)
		sideB := p.motion.SideB(now)
		v := deck.View{W: cw, H: ch, Tier: deckTier(a), Aspect: deckAspect(a)}
		if docked {
			pose, v.Fit = deckTurned(pose, deckDockTurn), deckGrow(deck.Rect{W: float64(cw), H: float64(ch)}, deckDockGrow)
		}
		s.Deck = p.renderer.Render(v, deckPaletteOf(s.Pal), pose,
			deck.State{Radii: deck.PackRadii(p.progress(), sideB), Angles: p.reels.Angles, Blur: p.reels.Blur(), Lift: lift})
	}
	p.docked = docked
	if !s.Idle {
		s.Lyrics, s.Cur, s.Sung = lyricsNow(p.netease.lyricService, player.PassedTime())
		s.Queue = p.upcoming(now)
	}
	render := renderDeckPanel
	if docked {
		render = renderDeckDock
	}
	lines, layout := render(s, w, h)
	p.deckBox, p.queueTop, p.queueLen = layout.deck, layout.queueTop, layout.queueLen
	return strings.Join(lines, "\n")
}

// deckDockCassette sizes the dock's cassette to its height; narrow docks
// have none.
func deckDockCassette(w, h int, aspect float64) (cw, ch int) {
	if w < deckDockMinCassette {
		return 0, 0
	}
	if aspect <= 0 {
		aspect = 2
	}
	ch = h - 1
	return min(int(float64(ch)*aspect*deck.FitAspect+0.5), w*2/5), ch
}

// deckPanelCassette sizes the cassette to the panel's width, leaving room for
// the text below it.
func deckPanelCassette(w, h int, aspect float64) (cw, ch int) {
	if aspect <= 0 {
		aspect = 2
	}
	cw = w - 2
	return cw, min(int(float64(cw)/deck.FitAspect/aspect+0.5), max((h-8)/2, 4))
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
	Liked      bool
	Mode       types.Mode
	Volume     int
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

// renderDeckDock draws the panel as a strip under the menu: a rule, then
// the cassette with the track beside it, the player's state flush right and
// the line being sung under them. Exactly h lines of w cells.
func renderDeckDock(s *deckPanelSnapshot, w, h int) ([]string, deckPanelLayout) {
	var out deckPanelLayout
	pal := s.Pal
	bgSGR := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", pal.Bg.R, pal.Bg.G, pal.Bg.B)
	text := func(str string, max int) string { return ansix.Truncate(ansix.Normalize(str), max, "…") }

	rule := &panelText{bg: pal.Bg}
	rule.add("  "+strings.Repeat("─", max(w-4, 0)), pal.Rule, false)
	lines := []string{ansix.Fit(rule.String(), w, bgSGR)}

	var deckRows []string
	x0 := 2
	if s.Deck != nil {
		deckRows = deckLines(s.Deck, pal.Bg)
		out.deck = image.Rect(2, 1, 2+s.Deck.W, 1+s.Deck.H)
		x0 = 2 + s.Deck.W + 3
	}

	// The text beside the cassette: each row draws its left part within the
	// columns it is given and may have a part flush right.
	type span struct {
		s  string
		fg color.Color
	}
	type row struct {
		left  func(t *panelText, w int)
		right []span
	}
	plain := func(str string, fg color.Color, bold bool) func(t *panelText, w int) {
		return func(t *panelText, w int) { t.add(text(str, w), fg, bold) }
	}
	var info []row
	if s.Idle {
		info = append(info, row{left: plain("未在播放", pal.Secondary, false)}, row{})
		for _, hint := range [][2]string{{"Enter", "播放选中的歌"}, {"Space", "播放 / 暂停"}, {"i", "打开播放页"}} {
			info = append(info, row{left: func(t *panelText, w int) {
				t.add(hint[0], pal.Text, false)
				t.add(text(strings.Repeat(" ", max(8-len(hint[0]), 1))+hint[1], max(w-len(hint[0]), 0)), pal.Secondary, false)
			}})
		}
	} else {
		var state []span
		if s.Liked {
			state = append(state, span{"♥  ", pal.Accent})
		}
		if s.Total > 0 {
			state = append(state, span{fmt.Sprintf("%02d / %02d", s.Pos, s.Total), pal.Secondary})
		}
		info = append(info,
			row{left: plain(s.Song.Name, pal.Hero, true), right: state},
			row{left: plain(npArtistAlbum(s.Song), pal.Secondary, false),
				right: []span{{s.Mode.String() + "  ·  ", pal.Rule}, {fmt.Sprintf("VOL %d", s.Volume), pal.Secondary}}},
			row{})
		switch cur := s.Cur; {
		case len(s.Lyrics) == 0:
			info = append(info, row{left: plain("暂无歌词", pal.Rule, false)})
		case cur < 0: // before the first line
			for _, l := range s.Lyrics[:min(2, len(s.Lyrics))] {
				info = append(info, row{left: plain("  "+l.Text, pal.Secondary, false)})
			}
		default:
			info = append(info, row{left: func(t *panelText, w int) {
				t.add("▸ ", pal.Accent, false)
				str := text(s.Lyrics[cur].Text, max(w-2, 0))
				runes := []rune(str)
				for j, c := range karaokeColors(str, s.Sung, pal.Hero, pal.Secondary) {
					t.add(string(runes[j]), c, true)
				}
			}})
			// The translation of the line, or else the line after it.
			if tr := s.Lyrics[cur].Trans; tr != "" {
				info = append(info, row{left: plain("  "+tr, pal.Text, false)})
			} else if cur+1 < len(s.Lyrics) {
				info = append(info, row{left: plain("  "+s.Lyrics[cur+1].Text, pal.Secondary, false)})
			}
		}
	}

	iw := w - x0 - 2
	if h > deckDockRows { // the taller dock: a row of air under the rule
		info = append([]row{{}}, info...)
	}
	for r := range h - 1 {
		var b strings.Builder
		b.WriteString(bgSGR + "  \x1b[m")
		if r < len(deckRows) {
			b.WriteString(deckRows[r])
		}
		line := ansix.Fit(b.String(), x0, bgSGR)
		if r < len(info) && info[r].left != nil {
			right := &panelText{bg: pal.Bg}
			rw := 0
			for _, sp := range info[r].right {
				right.add(sp.s, sp.fg, false)
				rw += npWidth(sp.s)
			}
			if rw > 0 && iw-rw-3 < 24 { // too narrow for both parts
				rw = 0
			}
			left := &panelText{bg: pal.Bg}
			info[r].left(left, iw-rw-min(rw, 3))
			line += left.String()
			if rw > 0 {
				line = ansix.Fit(line, w-2-rw, bgSGR) + right.String()
			}
		}
		lines = append(lines, ansix.Fit(line, w, bgSGR))
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
	player.async(func() {
		for range k {
			if _, err := player.playlistManager.NextSong(true); err != nil {
				return
			}
		}
		player.NextSong(true)
	})
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
