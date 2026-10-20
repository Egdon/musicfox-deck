package ui

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/ui/deck"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
)

const PageTypeNowPlaying model.PageType = "now_playing"

const (
	npMorphDuration    = 320 * time.Millisecond // the panel's cassette growing onto the stage
	npSweepDuration    = 450 * time.Millisecond
	npFlipTimeout      = 1500 * time.Millisecond // flip back if no track change follows
	npPlayOSDDuration  = 2 * time.Second
	npOSDFadeDuration  = 400 * time.Millisecond
	npEventOSDDuration = 1500 * time.Millisecond
	npBlueScreenDelay  = 800 * time.Millisecond // stopped with an empty queue
	npUpcomingCount    = 12
)

// npRenderMsg redraws the page; App.Rerender sends it to the current page.
type npRenderMsg struct{}

type npQueueKey struct {
	songID    int64
	index     int
	mode      types.Mode
	updatedAt time.Time
}

// NowPlayingPage is the full-screen now-playing view: a 3D cassette on a
// stage, the track and its lyrics, with a VCR-style on-screen display.
type NowPlayingPage struct {
	netease *Netease
	handler *EventHandler

	level      motion.Level
	lyricsMode bool
	keys       npKeys

	// The cassette morphs between the main screen's panel and the stage.
	morphIn   time.Time
	morphOut  time.Time
	fromFrame string    // the main screen as it was left
	toFrame   string    // the main screen to return to
	panelBox  deck.Rect // the panel's cassette, in window cells
	stage     deck.Rect // where the cassette rests in the current layout

	exitTo       model.Page // shown after leaving; nil means Main
	drivePending bool       // Main holds deferred work that needs a tick

	sweepAt time.Time
	flipAt  time.Time // a flip the page started for a track change
	flipDir int

	osdText  string
	osdUntil time.Time

	lastSongID   int64
	lastState    types.State
	switching    bool // the playlist moved on but the engine still plays the old song
	stoppedSince time.Time

	renderer  deck.Renderer // the cassette on the stage
	morph     deck.Renderer // the cassette while it morphs, over the whole window
	motion    deck.Motion
	reels     deck.Reels
	lastFrame time.Time

	pal    npPalette
	palGen uint64
	palOK  bool

	queueKey npQueueKey
	upcoming []structs.Song
	queuePos int
	queueLen int
	tapeMins int

	scr          npScreen
	hits         []npHit
	pointerOnHit bool
}

func NewNowPlayingPage(netease *Netease, handler *EventHandler) *NowPlayingPage {
	return &NowPlayingPage{netease: netease, handler: handler}
}

func (p *NowPlayingPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return false
}

func (p *NowPlayingPage) Type() model.PageType {
	return PageTypeNowPlaying
}

func (p *NowPlayingPage) Msg() tea.Msg {
	return npRenderMsg{}
}

// enter resets the page each time it is shown. The cassette carries on from
// the panel's, so the morph starts from exactly what was on screen.
func (p *NowPlayingPage) enter(now time.Time) {
	p.level = p.netease.MotionLevel()
	p.keys = npBoundKeys()
	p.morphIn, p.morphOut, p.fromFrame, p.toFrame = time.Time{}, time.Time{}, "", ""
	p.exitTo, p.drivePending = nil, false
	p.sweepAt, p.flipAt = time.Time{}, time.Time{}
	p.lastFrame = now
	p.lastSongID = p.currentSong().Id
	p.lastState = p.netease.player.State()
	p.osdText, p.osdUntil = "", time.Time{}
	if panel := p.netease.deckPanel; panel != nil {
		p.motion, p.reels = panel.motion, panel.reels
	} else {
		p.motion, p.reels = deck.Motion{}, deck.Reels{}
		p.reels.Reset(p.lastSongID)
	}
	p.motion.FloatAmount = 0
	if p.level == motion.Full {
		p.motion.FloatAmount = 1
	}
	if p.lastState == types.Playing {
		p.showOSD(osdStateText(types.Playing), now, npPlayOSDDuration)
		if p.level != motion.Off && !p.reels.Moving() {
			p.reels.Run()
		}
	}
	p.stoppedSince = time.Time{}
	p.queueKey = npQueueKey{index: -2}
}

// HandlePageTransition grows the main screen's panel cassette onto the stage
// instead of the App's crossfade.
func (p *NowPlayingPage) HandlePageTransition(a *model.App, tx model.PageTransition) bool {
	panel := p.netease.deckPanel
	if p.level != motion.Full || panel == nil || tx.From != model.Page(p.netease.MustMain()) {
		return false
	}
	box, ok := panel.CassetteBox()
	if !ok {
		return false
	}
	p.morphIn, p.fromFrame, p.panelBox = tx.Start, tx.FromFrame, box
	a.Animate(npMorphDuration + 50*time.Millisecond)
	return true
}

func (p *NowPlayingPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	now := time.Now()
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		// Keep Main's geometry current so that returning to it does not
		// flash a frame laid out for the old size.
		_, _ = p.netease.MustMain().Update(msg, a)
	case tea.KeyPressMsg:
		return p.handleKey(m, a, now)
	case tea.MouseClickMsg:
		return p.handleClick(m, a, now)
	case tea.MouseWheelMsg:
		p.handleWheel(m, now)
	case tea.MouseMotionMsg:
		page, cmd := p.settle(a, now)
		return page, tea.Batch(cmd, p.handleMotion(m, a, now))
	}
	return p.settle(a, now)
}

// settle moves everything on to now, runs Main's deferred work once no popup
// is open and keeps frames coming while anything moves.
func (p *NowPlayingPage) settle(a *model.App, now time.Time) (model.Page, tea.Cmd) {
	p.advance(now)
	p.poll(now)
	if !p.morphOut.IsZero() && now.Sub(p.morphOut) >= npMorphDuration {
		return p.leave(a)
	}
	if p.drivePending && !a.HasPopup() {
		p.drivePending = false
		// Operations such as liking a song are deferred until Main handles a
		// tick, which never happens while another page is shown.
		main := p.netease.MustMain()
		if page, cmd := main.Update(main.Msg(), a); page != nil && page != model.Page(main) {
			return page, cmd
		}
	}
	p.requestFrames(a, now)
	return p, nil
}

// requestFrames asks the App for frames: fast ones for morphs, flips and the
// sweep, calm ones for floating and turning reels.
func (p *NowPlayingPage) requestFrames(a *model.App, now time.Time) {
	playing := p.netease.player.State() == types.Playing
	switch {
	case !p.morphIn.IsZero() && now.Sub(p.morphIn) < npMorphDuration,
		!p.morphOut.IsZero(),
		p.motion.Flipping(now),
		now.Sub(p.sweepAt) < npSweepDuration:
		a.Animate(50 * time.Millisecond)
	case p.level != motion.Off && (playing || p.motion.Busy(now) || p.reels.Moving()),
		now.Before(p.osdUntil.Add(npOSDFadeDuration)),
		p.drivePending,
		!p.stoppedSince.IsZero() && now.Sub(p.stoppedSince) < npBlueScreenDelay:
		a.AnimateSlow(100 * time.Millisecond)
	}
}

// advance moves the cassette and turns the reels. Both reels carry the tape
// at the same speed, so the emptier one spins faster.
func (p *NowPlayingPage) advance(now time.Time) {
	dt := min(max(now.Sub(p.lastFrame).Seconds(), 0), 0.1) // no jump after a stall
	p.lastFrame = now
	playing := p.netease.player.State() == types.Playing
	p.motion.Step(now, playing)
	if p.level == motion.Off || dt <= 0 {
		return
	}
	sideB := p.motion.SideB(now)
	dir := 1.0
	if sideB {
		dir = -1
	}
	p.reels.Advance(dt, playing, now, deck.PackRadii(p.reels.Progress(now, p.progress()), sideB), dir)
}

// poll picks up player changes, including those made outside the page
// (automatic advance, MPRIS, global hotkeys).
func (p *NowPlayingPage) poll(now time.Time) {
	player := p.netease.player
	song := p.currentSong()
	state := player.State()
	if song.Id != p.lastSongID {
		if p.lastSongID != 0 {
			switch {
			case !p.flipAt.IsZero(): // the page flipped for this change already
				p.flipAt = time.Time{}
			case p.level == motion.Full:
				p.motion.Flip(1, now)
			default:
				p.motion.FlipNow()
			}
			if now.Sub(p.sweepAt) > time.Second {
				p.startSweep(now)
			}
		}
		p.lastSongID = song.Id
		p.reels.Reset(song.Id)
	}
	if !p.flipAt.IsZero() && now.Sub(p.flipAt) >= npFlipTimeout {
		p.motion.Flip(-p.flipDir, now) // the track did not change: turn back
		p.flipAt = time.Time{}
	}
	if state != p.lastState {
		switch {
		case state == types.Playing && !now.Before(p.osdUntil):
			p.showOSD(osdStateText(state), now, npPlayOSDDuration)
		case state != types.Playing:
			p.osdUntil = time.Time{} // the state OSD takes over
		}
		p.lastState = state
	}
	playing := player.CurMusic().Id
	p.switching = song.Id != 0 && playing != 0 && playing != song.Id
	if state == types.Stopped && !p.switching {
		if p.stoppedSince.IsZero() {
			p.stoppedSince = now
		}
	} else {
		p.stoppedSince = time.Time{}
	}
	p.refreshQueue(song.Id)
}

func (p *NowPlayingPage) refreshQueue(songID int64) {
	player := p.netease.player
	key := npQueueKey{songID: songID, index: player.CurSongIndex(), mode: player.Mode(), updatedAt: player.playlistUpdateAt}
	if key == p.queueKey {
		return
	}
	p.queueKey = key
	p.upcoming = player.playlistManager.Upcoming(npUpcomingCount)
	playlist := player.playlistManager.GetPlaylist()
	p.queueLen = len(playlist)
	p.queuePos = 0
	if key.index >= 0 && key.index < len(playlist) {
		p.queuePos = key.index + 1
	}
	var total time.Duration
	for _, s := range playlist {
		total += s.Duration
	}
	p.tapeMins = int(math.Ceil(total.Minutes()))
}

func (p *NowPlayingPage) currentSong() structs.Song {
	song, _ := p.netease.player.playlistManager.GetCurrentSong()
	return song
}

func (p *NowPlayingPage) progress() float64 {
	d := p.currentSong().Duration
	if d <= 0 {
		return 0
	}
	return min(max(float64(p.netease.player.PassedTime())/float64(d), 0), 1)
}

func (p *NowPlayingPage) showOSD(text string, now time.Time, d time.Duration) {
	p.osdText, p.osdUntil = text, now.Add(d)
}

// visibleOSD returns the OSD text: the latest event while it lasts, otherwise
// the persistent state (PAUSE, STOP); PLAY fades out like on a VCR.
func (p *NowPlayingPage) visibleOSD(now time.Time, state types.State) string {
	switch {
	case p.switching:
		return "TRACKING"
	case now.Before(p.osdUntil):
		return p.osdText
	case state != types.Playing:
		return osdStateText(state)
	}
	return ""
}

// osdFade fades a timed OSD out over its last moments, as a VCR does.
func (p *NowPlayingPage) osdFade(now time.Time) float64 {
	left := p.osdUntil.Sub(now)
	if p.switching || p.level == motion.Off || left <= 0 || left >= npOSDFadeDuration {
		return 0
	}
	return 1 - float64(left)/float64(npOSDFadeDuration)
}

// startSweep runs the tracking line of a track change.
func (p *NowPlayingPage) startSweep(now time.Time) {
	if p.level != motion.Full {
		return
	}
	p.sweepAt = now
}

func (p *NowPlayingPage) clearCover() {
	if p.netease.coverRenderer != nil {
		p.netease.coverRenderer.ClearDisplayed()
	}
}

// ---- actions ----

// toggle pauses or resumes the current song. Unlike playOrToggle on Main it
// never starts the song selected in Main's menu.
func (p *NowPlayingPage) toggle() {
	player := p.netease.player
	switch player.State() {
	case types.Playing:
		player.Pause()
	case types.Paused, types.Interrupted:
		player.Resume()
	default:
		if p.currentSong().Id != 0 {
			player.StartPlay()
		}
	}
}

func (p *NowPlayingPage) seek(delta time.Duration, now time.Time) {
	player := p.netease.player
	if s := player.State(); s != types.Playing && s != types.Paused {
		return
	}
	from := p.progress()
	to := min(max(player.PassedTime()+delta, 0), p.currentSong().Duration)
	player.Seek(to)
	p.showOSD(osdSeekText(delta, to), now, npEventOSDDuration)
	if d := p.currentSong().Duration; d > 0 && p.level != motion.Off {
		dir := 1.0
		if delta < 0 {
			dir = -1
		}
		p.reels.Wind(dir, now, from, float64(to)/float64(d))
	}
}

func (p *NowPlayingPage) volume(up bool, now time.Time) {
	player := p.netease.player
	if up {
		player.UpVolume()
	} else {
		player.DownVolume()
	}
	p.showOSD(osdVolumeText(player.Volume()), now, npEventOSDDuration)
}

func (p *NowPlayingPage) switchMode(now time.Time) {
	player := p.netease.player
	player.SwitchMode()
	p.showOSD(osdModeText(player.Mode()), now, npEventOSDDuration)
}

// switchTrack turns the cassette over while the player switches in the
// background.
func (p *NowPlayingPage) switchTrack(now time.Time, osd string, dir int, run func()) tea.Cmd {
	p.showOSD(osd, now, npEventOSDDuration)
	p.startSweep(now)
	if p.level == motion.Full {
		p.motion.Flip(dir, now)
		p.flipAt, p.flipDir = now, dir
	}
	p.netease.player.async(run)
	return nil
}

// jumpTo plays the k-th song of the up-next list. Stepping the playlist keeps
// the shuffle order, so the song played is the one shown.
func (p *NowPlayingPage) jumpTo(k int, now time.Time) tea.Cmd {
	player := p.netease.player
	if player.Mode() == types.PmSingleLoop {
		return nil
	}
	return p.switchTrack(now, "▶▶ NEXT", 1, func() {
		for range k {
			if _, err := player.playlistManager.NextSong(true); err != nil {
				return
			}
		}
		player.NextSong(true)
	})
}

// forward runs an operation of Main's event handler that works on the playing
// song, then lets Main run the work it deferred.
func (p *NowPlayingPage) forward(op keybindings.OperateType, a *model.App, now time.Time) (model.Page, tea.Cmd) {
	_, page, _ := p.handler.handle(op)
	if page != nil && page != model.Page(p.netease.MustMain()) {
		return page, func() tea.Msg { return page.Msg() }
	}
	p.drivePending = true
	return p.settle(a, now)
}

// startExit leaves the page. Back to the main screen the cassette shrinks
// into the panel; anywhere else the App crossfades.
func (p *NowPlayingPage) startExit(a *model.App, now time.Time, to model.Page) (model.Page, tea.Cmd) {
	p.exitTo = to
	main := p.netease.MustMain()
	panel := p.netease.deckPanel
	if p.level != motion.Full || panel == nil || (to != nil && to != model.Page(main)) || !p.morphOut.IsZero() {
		return p.leave(a)
	}
	frame := a.Snapshot(main) // also lays the panel out for this window
	box, ok := panel.CassetteBox()
	if !ok {
		return p.leave(a)
	}
	p.morphOut, p.toFrame, p.panelBox = now, frame, box
	a.Animate(npMorphDuration + 50*time.Millisecond)
	return p, nil
}

func (p *NowPlayingPage) leave(a *model.App) (model.Page, tea.Cmd) {
	morphed := !p.morphOut.IsZero()
	p.morphOut = time.Time{}
	if panel := p.netease.deckPanel; panel != nil {
		panel.adopt(p.motion, p.reels)
	}
	pointer := a.SetMousePointer("default")
	main := p.netease.MustMain()
	if to := p.exitTo; to != nil && to != model.Page(main) {
		p.exitTo = nil
		return to, tea.Batch(pointer, func() tea.Msg { return to.Msg() })
	}
	if morphed {
		a.SuppressPageTransition() // the cassette is already in the panel
	}
	// The tick also lets Main run work deferred by operations that sent us
	// back to it.
	return main, tea.Batch(pointer, func() tea.Msg { return main.Msg() })
}

// ---- input ----

func (p *NowPlayingPage) handleKey(k tea.KeyPressMsg, a *model.App, now time.Time) (model.Page, tea.Cmd) {
	if !p.morphOut.IsZero() {
		return p, nil
	}
	switch k.String() {
	case "esc":
		return p.startExit(a, now, nil)
	case "left":
		p.seek(-5*time.Second, now)
	case "right":
		p.seek(5*time.Second, now)
	case "up":
		p.volume(true, now)
	case "down":
		p.volume(false, now)
	case "l", "L":
		p.lyricsMode = !p.lyricsMode
	default:
		if op, ok := p.handler.keyToOperateMap[k.String()]; ok {
			return p.handleOp(op, a, now)
		}
	}
	return p.settle(a, now)
}

// handleOp maps the user's key bindings onto the page. Operations on the
// playing song run here; operations that open a menu send the user back to
// Main; operations on Main's selected item are ignored.
func (p *NowPlayingPage) handleOp(op keybindings.OperateType, a *model.App, now time.Time) (model.Page, tea.Cmd) {
	player := p.netease.player
	switch op {
	case keybindings.OpNowPlaying, keybindings.OpGoBack:
		return p.startExit(a, now, nil)
	case keybindings.OpPlayOrToggle, keybindings.OpToggle:
		p.toggle()
	case keybindings.OpPrevious:
		return p, p.switchTrack(now, "◀◀ PREV", -1, func() { player.PreviousSong(true) })
	case keybindings.OpNext:
		return p, p.switchTrack(now, "▶▶ NEXT", 1, func() { player.NextSong(true) })
	case keybindings.OpRerenderUI:
		return p, a.RerenderCmd(true)
	case keybindings.OpSeekBackward1s:
		p.seek(-time.Second, now)
	case keybindings.OpSeekBackward5s:
		p.seek(-5*time.Second, now)
	case keybindings.OpSeekForward5s:
		p.seek(5*time.Second, now)
	case keybindings.OpSeekForward10s:
		p.seek(10*time.Second, now)
	case keybindings.OpVolumeUp:
		p.volume(true, now)
	case keybindings.OpVolumeDown:
		p.volume(false, now)
	case keybindings.OpSwitchPlayMode:
		p.switchMode(now)
	case keybindings.OpHelp, keybindings.OpSharePlayingItem, keybindings.OpLogout, keybindings.OpSwitchTheme:
		p.handler.handle(op)
	case keybindings.OpCurPlaylist, keybindings.OpActionOfPlayingSong, keybindings.OpAlbumOfPlayingSong,
		keybindings.OpArtistOfPlayingSong, keybindings.OpOpenSimiSongsOfPlayingSong,
		keybindings.OpAddPlayingToUserPlaylist, keybindings.OpRemovePlayingFromUserPlaylist:
		_, page, _ := p.handler.handle(op)
		return p.startExit(a, now, page)
	case keybindings.OpLikePlayingSong, keybindings.OpDislikePlayingSong, keybindings.OpTrashPlayingSong,
		keybindings.OpDownloadPlayingSong, keybindings.OpDownloadPlayingSongLrc, keybindings.OpOpenPlayingSongInWeb,
		keybindings.OpSubscribeAlbumOfPlayingSong, keybindings.OpUnsubscribeAlbumOfPlayingSong,
		keybindings.OpSubscribeArtistOfPlayingSong, keybindings.OpUnsubscribeArtistOfPlayingSong,
		keybindings.OpClearSongCache:
		return p.forward(op, a, now)
	}
	return p.settle(a, now)
}

func (p *NowPlayingPage) hitAt(x, y int) (npHit, bool) {
	for _, h := range p.hits {
		if h.contains(x, y) {
			return h, true
		}
	}
	return npHit{}, false
}

func (p *NowPlayingPage) handleClick(m tea.MouseClickMsg, a *model.App, now time.Time) (model.Page, tea.Cmd) {
	mouse := m.Mouse()
	if mouse.Button != tea.MouseLeft || !p.morphOut.IsZero() {
		return p.settle(a, now)
	}
	hit, ok := p.hitAt(mouse.X, mouse.Y)
	if !ok {
		return p.settle(a, now)
	}
	player := p.netease.player
	switch hit.kind {
	case npHitProgress:
		if d := p.currentSong().Duration; d > 0 && hit.w > 1 {
			target := time.Duration(float64(mouse.X-hit.x) / float64(hit.w-1) * float64(d))
			p.seek(target-player.PassedTime(), now)
		}
	case npHitPrev:
		return p, p.switchTrack(now, "◀◀ PREV", -1, func() { player.PreviousSong(true) })
	case npHitNext:
		return p, p.switchTrack(now, "▶▶ NEXT", 1, func() { player.NextSong(true) })
	case npHitToggle:
		p.toggle()
	case npHitLike:
		op := keybindings.OpLikePlayingSong
		if song := p.currentSong(); song.Id > 0 && likelist.IsLikeSong(song.Id) {
			op = keybindings.OpDislikePlayingSong
		}
		return p.forward(op, a, now)
	case npHitMode:
		p.switchMode(now)
	case npHitVolDown:
		p.volume(false, now)
	case npHitVolUp:
		p.volume(true, now)
	case npHitLyrics:
		p.lyricsMode = !p.lyricsMode
	case npHitBack:
		return p.startExit(a, now, nil)
	case npHitUpNext:
		return p, p.jumpTo(hit.index, now)
	}
	return p.settle(a, now)
}

func (p *NowPlayingPage) handleWheel(m tea.MouseWheelMsg, now time.Time) {
	mouse := m.Mouse()
	if mouse.Button != tea.MouseWheelUp && mouse.Button != tea.MouseWheelDown {
		return
	}
	p.handler.wheelVolume(mouse)
	p.showOSD(osdVolumeText(p.netease.player.Volume()), now, npEventOSDDuration)
}

// handleMotion leans the cassette towards a pointer over the stage and shows
// a pointer over clickable regions. The shape is re-sent on every move
// because notifications may change it in between.
func (p *NowPlayingPage) handleMotion(m tea.MouseMotionMsg, a *model.App, now time.Time) tea.Cmd {
	mouse := m.Mouse()
	if st := p.stage; p.level == motion.Full && st.W > 0 && float64(mouse.X) >= st.X && float64(mouse.X) < st.X+st.W &&
		float64(mouse.Y) >= st.Y && float64(mouse.Y) < st.Y+st.H {
		p.motion.Point((float64(mouse.X)-st.X-st.W/2)/(st.W/2), (float64(mouse.Y)-st.Y-st.H/2)/(st.H/2), now)
		a.Animate(100 * time.Millisecond)
	} else {
		p.motion.Leave()
	}
	if _, over := p.hitAt(mouse.X, mouse.Y); over {
		p.pointerOnHit = true
		return a.SetMousePointer("pointer")
	}
	if p.pointerOnHit {
		p.pointerOnHit = false
		return a.SetMousePointer("default")
	}
	return nil
}

// ---- view ----

func (p *NowPlayingPage) View(a *model.App) string {
	w, h := a.WindowWidth(), a.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}
	now := a.FrameTime()
	p.scr.reset(w, h)
	morphing := !p.morphOut.IsZero() || (!p.morphIn.IsZero() && now.Sub(p.morphIn) < npMorphDuration)
	s := p.snapshot(a, now, w, h, !morphing)
	p.hits = npRender(&p.scr, s)
	p.scr.fillBackground()
	if t := now.Sub(p.sweepAt); t >= 0 && t < npSweepDuration {
		p.scr.sweep(float64(t) / float64(npSweepDuration))
	}
	if morphing {
		p.drawMorph(a, now, w, h)
	}
	return p.scr.render()
}

// drawMorph moves the cassette between the panel and the stage while the
// rest of the screen dips through the background: the main screen fades out
// and the page in, or the other way round.
func (p *NowPlayingPage) drawMorph(a *model.App, now time.Time, w, h int) {
	in := p.morphOut.IsZero()
	start, from, to := p.morphIn, p.panelBox, p.stage
	if !in {
		start, from, to = p.morphOut, p.stage, p.panelBox
	}
	t := min(max(float64(now.Sub(start))/float64(npMorphDuration), 0), 1)
	e := motion.OutCubic(t)
	// The main screen's own cassette is the one that moves: blank it out.
	blankPanel := func() {
		b := p.panelBox
		p.scr.fill(int(b.X), int(b.Y), int(math.Ceil(b.W)), int(math.Ceil(b.H)), p.pal.Bg)
	}
	switch {
	case in && t < 0.5:
		p.scr.paint(p.fromFrame, motion.InQuad(t/0.5))
		blankPanel()
	case in:
		p.scr.fade(1 - motion.OutCubic((t-0.5)/0.5))
	case t < 0.5:
		p.scr.fade(motion.InQuad(t / 0.5))
	default:
		p.scr.paint(p.toFrame, 1-motion.OutCubic((t-0.5)/0.5))
		blankPanel()
	}
	box := deck.Rect{
		X: from.X + (to.X-from.X)*e, Y: from.Y + (to.Y-from.Y)*e,
		W: from.W + (to.W-from.W)*e, H: from.H + (to.H-from.H)*e,
	}
	pose, lift := p.motion.At(now)
	sideB := p.motion.SideB(now)
	f := p.morph.Render(deck.View{W: w, H: h, Tier: deckTier(a), Aspect: deckAspect(a), Fit: box}, deckPaletteOf(p.palette()), pose,
		deck.State{Radii: deck.PackRadii(p.progress(), sideB), Angles: p.reels.Angles, Blur: p.reels.Blur(), Lift: lift})
	p.scr.drawDeck(f, 0, 0)
}

// palette follows the active theme; unset roles fall back to Warm Night.
func (p *NowPlayingPage) palette() npPalette {
	if gen := style.StyleGeneration(); !p.palOK || gen != p.palGen {
		p.pal, p.palGen, p.palOK = npPaletteFromTheme(style.CurrentStyleSet().Theme()), gen, true
	}
	return p.pal
}

// snapshot captures what the page draws. The cassette is rendered onto the
// stage unless it is morphing, in which case drawMorph draws it.
func (p *NowPlayingPage) snapshot(a *model.App, now time.Time, w, h int, withDeck bool) *npSnapshot {
	player := p.netease.player
	song := p.currentSong()
	state := player.State()
	pal := p.palette()
	blue := song.Id == 0 && !p.stoppedSince.IsZero() && now.Sub(p.stoppedSince) >= npBlueScreenDelay
	p.stage = deck.Rect{}
	var frame *deck.Frame
	if !blue && !p.lyricsMode && npWide(w, h) {
		x, y, sw, sh := npStage(w, h)
		p.stage = deck.Rect{X: float64(x), Y: float64(y), W: float64(sw), H: float64(sh)}
		if withDeck {
			pose, lift := p.motion.At(now)
			sideB := p.motion.SideB(now)
			frame = p.renderer.Render(deck.View{W: sw, H: sh, Tier: deckTier(a), Aspect: deckAspect(a)}, deckPaletteOf(pal), pose,
				deck.State{Radii: deck.PackRadii(p.reels.Progress(now, p.progress()), sideB), Angles: p.reels.Angles, Blur: p.reels.Blur(), Lift: lift})
		}
	}
	lines, cur, sung := lyricsNow(p.netease.lyricService, player.PassedTime())
	lyrics := make([]npLyricLine, len(lines))
	for i, l := range lines {
		lyrics[i] = npLyricLine{Text: l.Text, Trans: l.Trans}
	}
	return &npSnapshot{
		Now:        now,
		Pal:        pal,
		Song:       song,
		State:      state,
		BlueScreen: blue,
		Scanlines:  p.level == motion.Full,
		Passed:     player.PassedTime(),
		Volume:     player.Volume(),
		Mode:       player.Mode(),
		Liked:      song.Id > 0 && likelist.IsLikeSong(song.Id),
		Quality:    configs.AppConfig.Player.SongLevel,
		QueuePos:   p.queuePos,
		QueueLen:   p.queueLen,
		TapeMins:   p.tapeMins,
		Upcoming:   p.upcoming,
		Lyrics:     lyrics,
		LyricIndex: cur,
		Sung:       sung,
		LyricsMode: p.lyricsMode,
		OSD:        p.visibleOSD(now, state),
		OSDFade:    p.osdFade(now),
		Keys:       p.keys,
		Angles:     p.reels.Angles,
		Deck:       frame,
	}
}

// npBoundKeys returns display labels of the user's bindings for the legend.
func npBoundKeys() npKeys {
	label := func(op keybindings.OperateType, fallback string) string {
		for _, k := range keybindings.UserOperateToKeys()[op] {
			switch k {
			case "":
				continue
			case "space", " ":
				return "SPC"
			}
			return keybindings.FormatKeyForDisplay(k)
		}
		return fallback
	}
	return npKeys{
		Prev:    label(keybindings.OpPrevious, "["),
		Toggle:  label(keybindings.OpPlayOrToggle, "SPC"),
		Next:    label(keybindings.OpNext, "]"),
		Like:    label(keybindings.OpLikePlayingSong, ","),
		Mode:    label(keybindings.OpSwitchPlayMode, "p"),
		VolDown: label(keybindings.OpVolumeDown, "-"),
		VolUp:   label(keybindings.OpVolumeUp, "="),
	}
}
