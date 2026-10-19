package ui

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
)

const PageTypeNowPlaying model.PageType = "now_playing"

const (
	npPowerOnDuration  = 300 * time.Millisecond
	npPowerOffDuration = 250 * time.Millisecond
	npSweepDuration    = 450 * time.Millisecond
	npEjectDuration    = 260 * time.Millisecond
	npInsertDuration   = 340 * time.Millisecond
	npEjectTimeout     = 1500 * time.Millisecond // put the tape back if no track change follows
	npTransitionFrame  = 33 * time.Millisecond
	npReelFrame        = 50 * time.Millisecond // at most 12.3° per step on an empty reel
	npPlayOSDDuration  = 2 * time.Second
	npOSDFadeDuration  = 400 * time.Millisecond
	npEventOSDDuration = 1500 * time.Millisecond
	npBlueScreenDelay  = 800 * time.Millisecond // stopped with an empty queue
	npUpcomingCount    = 12
)

// npRenderMsg redraws the page; App.Rerender sends it to the current page.
type npRenderMsg struct{}

// npFrameMsg drives the animation loop; gen discards ticks of an older loop.
type npFrameMsg struct{ gen int }

// npDeferredMsg runs a blocking action after the frame announcing it is drawn.
type npDeferredMsg struct {
	gen int
	run func()
}

type npQueueKey struct {
	songID    int64
	index     int
	mode      types.Mode
	updatedAt time.Time
}

// NowPlayingPage is the full-screen now-playing view: a 3D cassette deck with
// a VCR-style on-screen display.
type NowPlayingPage struct {
	netease *Netease
	handler *EventHandler

	gen    int
	loopOn bool // a npFrameMsg is in flight

	effects    configs.NowPlayingEffects
	lyricsMode bool
	keys       npKeys

	poweredOnAt  time.Time  // start of the warm-up; zero once it is over
	exitAt       time.Time  // start of the power-off; zero while not leaving
	exitTo       model.Page // shown after the power-off; nil means Main
	drivePending bool       // Main holds deferred work that needs a tick

	sweepAt  time.Time
	ejectAt  time.Time // the cassette is being taken out for a track change
	insertAt time.Time // the next cassette slides in

	osdText  string
	osdUntil time.Time

	lastSongID   int64
	lastState    types.State
	switching    bool // the playlist moved on but the engine still plays the old song
	stoppedSince time.Time

	reels     npReelMotion
	lastFrame time.Time
	deck      npDeck

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

// enter resets the page each time it is shown.
func (p *NowPlayingPage) enter(now time.Time) {
	p.gen++
	p.loopOn = false
	p.effects = configs.AppConfig.Main.NowPlaying.EffectsLevel()
	p.keys = npBoundKeys()
	p.poweredOnAt = time.Time{}
	if p.effects == configs.NowPlayingEffectsFull {
		p.poweredOnAt = now
	}
	p.exitAt, p.exitTo, p.drivePending = time.Time{}, nil, false
	p.sweepAt, p.ejectAt, p.insertAt = time.Time{}, time.Time{}, time.Time{}
	p.lastFrame = now
	p.lastSongID = p.currentSong().Id
	p.lastState = p.netease.player.State()
	p.osdText, p.osdUntil = "", time.Time{}
	p.reels.speed, p.reels.vel = 0, 0
	if p.lastState == types.Playing {
		p.showOSD(osdStateText(types.Playing), now, npPlayOSDDuration)
		if p.effects != configs.NowPlayingEffectsOff {
			p.reels.speed = 1 // already running: no spin-up on entry
		}
	}
	p.stoppedSince = time.Time{}
	p.queueKey = npQueueKey{index: -2}
}

func (p *NowPlayingPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	now := time.Now()
	switch m := msg.(type) {
	case npFrameMsg:
		if m.gen != p.gen {
			return p, nil
		}
		p.loopOn = false
		p.advance(now)
		if !p.exitAt.IsZero() && now.Sub(p.exitAt) >= npPowerOffDuration {
			return p.leave(a)
		}
		if !p.poweredOnAt.IsZero() && now.Sub(p.poweredOnAt) >= npPowerOnDuration {
			p.poweredOnAt = time.Time{}
			// The asynchronous cover renderer may have drawn once more after
			// the page was entered.
			p.clearCover()
		}
	case npDeferredMsg:
		if m.gen == p.gen {
			m.run()
		}
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
		return page, tea.Batch(cmd, p.handleMotion(m, a))
	}
	return p.settle(a, now)
}

// settle polls the player, runs Main's deferred work once no popup is open and
// keeps the animation loop going.
func (p *NowPlayingPage) settle(a *model.App, now time.Time) (model.Page, tea.Cmd) {
	p.poll(now)
	if p.drivePending && !a.HasPopup() {
		p.drivePending = false
		// Operations such as liking a song are deferred until Main handles a
		// tick, which never happens while another page is shown.
		main := p.netease.MustMain()
		if page, cmd := main.Update(main.Msg(), a); page != nil && page != model.Page(main) {
			return page, cmd
		}
	}
	return p, p.ensureLoop(now)
}

func (p *NowPlayingPage) frameInterval(now time.Time) time.Duration {
	switch {
	case !p.poweredOnAt.IsZero() || !p.exitAt.IsZero() || now.Sub(p.sweepAt) < npSweepDuration ||
		!p.ejectAt.IsZero() || !p.insertAt.IsZero():
		return npTransitionFrame
	case p.effects != configs.NowPlayingEffectsOff && (p.reels.moving() || p.netease.player.State() == types.Playing):
		return npReelFrame
	case now.Before(p.osdUntil) || p.drivePending:
		return npReelFrame
	case !p.stoppedSince.IsZero() && now.Sub(p.stoppedSince) < npBlueScreenDelay:
		return npReelFrame
	}
	return 0
}

// ensureLoop schedules the next animation frame unless one is in flight or
// nothing moves; the loop restarts with the next message that needs it.
func (p *NowPlayingPage) ensureLoop(now time.Time) tea.Cmd {
	if p.loopOn {
		return nil
	}
	interval := p.frameInterval(now)
	if interval <= 0 {
		return nil
	}
	p.loopOn = true
	gen := p.gen
	return tea.Tick(interval, func(time.Time) tea.Msg { return npFrameMsg{gen: gen} })
}

// advance turns the reels. Both turn at the same tape speed, so the emptier
// reel spins faster.
func (p *NowPlayingPage) advance(now time.Time) {
	dt := min(now.Sub(p.lastFrame).Seconds(), 0.1) // no jump after a stall
	p.lastFrame = now
	if p.effects == configs.NowPlayingEffectsOff || dt <= 0 {
		return
	}
	playing := p.netease.player.State() == types.Playing
	p.reels.advance(dt, playing, now, npReelRadii(p.reels.progress(now, p.progress())))
}

// poll picks up player changes, including those made outside the page
// (automatic advance, MPRIS, global hotkeys).
func (p *NowPlayingPage) poll(now time.Time) {
	player := p.netease.player
	song := p.currentSong()
	state := player.State()
	if song.Id != p.lastSongID {
		if p.lastSongID != 0 {
			p.cueInsert(now)
			if now.Sub(p.sweepAt) > time.Second {
				p.startSweep(now)
			}
		}
		p.lastSongID = song.Id
		// A new cassette: its reels start at rest in their own position.
		p.reels.speed, p.reels.vel = 0, 0
		p.reels.angles = [2]float64{math.Mod(float64(song.Id)*0.37, 2*math.Pi), math.Mod(float64(song.Id)*0.71, 2*math.Pi)}
	}
	switch {
	case !p.insertAt.IsZero() && now.Sub(p.insertAt) >= npInsertDuration:
		p.ejectAt, p.insertAt = time.Time{}, time.Time{}
	case !p.ejectAt.IsZero() && p.insertAt.IsZero() && now.Sub(p.ejectAt) >= npEjectTimeout:
		p.insertAt = now // the track did not change: put the same tape back
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
	if p.switching || p.effects == configs.NowPlayingEffectsOff || left <= 0 || left >= npOSDFadeDuration {
		return 0
	}
	return 1 - float64(left)/float64(npOSDFadeDuration)
}

// startEject takes the cassette out for a track change.
func (p *NowPlayingPage) startEject(now time.Time) {
	if p.effects == configs.NowPlayingEffectsOff {
		return
	}
	if p.ejectAt.IsZero() || !p.insertAt.IsZero() {
		p.ejectAt, p.insertAt = now, time.Time{}
	}
}

// cueInsert slides the next cassette in once the current one is out; a track
// change the page did not start ejects first.
func (p *NowPlayingPage) cueInsert(now time.Time) {
	if p.effects == configs.NowPlayingEffectsOff {
		return
	}
	if p.ejectAt.IsZero() || !p.insertAt.IsZero() {
		p.ejectAt = now
	}
	p.insertAt = p.ejectAt.Add(npEjectDuration)
	if p.insertAt.Before(now) {
		p.insertAt = now
	}
}

// deckPose places the cassette during a swap; shown is false while the slot
// is empty.
func (p *NowPlayingPage) deckPose(now time.Time) (pose npPose, shown bool) {
	if !p.insertAt.IsZero() && !now.Before(p.insertAt) {
		t := min(float64(now.Sub(p.insertAt))/float64(npInsertDuration), 1)
		e := 1 - math.Pow(1-t, 3) // ease out
		return npPose{DY: -20 * (1 - e), Tilt: 0.35 * (1 - e), Fade: 0.8 * (1 - e)}, true
	}
	if !p.ejectAt.IsZero() {
		t := float64(now.Sub(p.ejectAt)) / float64(npEjectDuration)
		if t >= 1 {
			return npPose{}, false
		}
		e := t * t // ease in
		return npPose{DY: -16 * e, Tilt: 0.44 * e, Fade: e}, true
	}
	return npPose{}, true
}

// startSweep runs the tracking line of a track change.
func (p *NowPlayingPage) startSweep(now time.Time) {
	if p.effects != configs.NowPlayingEffectsFull {
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
	if d := p.currentSong().Duration; d > 0 && p.effects != configs.NowPlayingEffectsOff {
		dir := 1.0
		if delta < 0 {
			dir = -1
		}
		p.reels.wind(dir, now, from, float64(to)/float64(d))
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

// switchTrack shows the track change first and switches on the next message:
// PlaySong resolves the song URL synchronously, freezing the UI on the change
// instead of on a stale frame.
func (p *NowPlayingPage) switchTrack(now time.Time, osd string, run func()) tea.Cmd {
	p.showOSD(osd, now, npEventOSDDuration)
	p.startSweep(now)
	p.startEject(now)
	delay := npTransitionFrame
	if p.effects != configs.NowPlayingEffectsOff {
		delay = npEjectDuration // let the tape come out before the UI blocks
	}
	gen := p.gen
	return tea.Batch(p.ensureLoop(now), tea.Tick(delay, func(time.Time) tea.Msg {
		return npDeferredMsg{gen: gen, run: run}
	}))
}

// jumpTo plays the k-th song of the up-next list. Stepping the playlist keeps
// the shuffle order, so the song played is the one shown.
func (p *NowPlayingPage) jumpTo(k int, now time.Time) tea.Cmd {
	player := p.netease.player
	if player.Mode() == types.PmSingleLoop {
		return nil
	}
	return p.switchTrack(now, "▶▶ NEXT", func() {
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

func (p *NowPlayingPage) startExit(a *model.App, now time.Time, to model.Page) (model.Page, tea.Cmd) {
	p.exitTo = to
	if p.effects != configs.NowPlayingEffectsFull {
		return p.leave(a)
	}
	p.exitAt = now
	return p, p.ensureLoop(now)
}

func (p *NowPlayingPage) leave(a *model.App) (model.Page, tea.Cmd) {
	p.gen++
	p.loopOn = false
	p.exitAt = time.Time{}
	pointer := a.SetMousePointer("default")
	main := p.netease.MustMain()
	if to := p.exitTo; to != nil && to != model.Page(main) {
		p.exitTo = nil
		return to, tea.Batch(pointer, func() tea.Msg { return to.Msg() })
	}
	// The rerender tick also lets Main run work deferred by operations that
	// sent us back to it.
	return main, tea.Batch(pointer, a.RerenderCmd(true))
}

// ---- input ----

func (p *NowPlayingPage) handleKey(k tea.KeyPressMsg, a *model.App, now time.Time) (model.Page, tea.Cmd) {
	if !p.exitAt.IsZero() {
		return p, nil
	}
	p.poweredOnAt = time.Time{} // any key skips the warm-up
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
		return p, p.switchTrack(now, "◀◀ PREV", func() { player.PreviousSong(true) })
	case keybindings.OpNext:
		return p, p.switchTrack(now, "▶▶ NEXT", func() { player.NextSong(true) })
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
	if mouse.Button != tea.MouseLeft || !p.exitAt.IsZero() {
		return p.settle(a, now)
	}
	hit, ok := p.hitAt(mouse.X, mouse.Y)
	if !ok {
		return p.settle(a, now)
	}
	p.poweredOnAt = time.Time{}
	player := p.netease.player
	switch hit.kind {
	case npHitProgress:
		if d := p.currentSong().Duration; d > 0 && hit.w > 1 {
			target := time.Duration(float64(mouse.X-hit.x) / float64(hit.w-1) * float64(d))
			p.seek(target-player.PassedTime(), now)
		}
	case npHitPrev:
		return p, p.switchTrack(now, "◀◀ PREV", func() { player.PreviousSong(true) })
	case npHitNext:
		return p, p.switchTrack(now, "▶▶ NEXT", func() { player.NextSong(true) })
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

// handleMotion shows a pointer over clickable regions. It is re-sent on every
// move because notifications may change the pointer in between.
func (p *NowPlayingPage) handleMotion(m tea.MouseMotionMsg, a *model.App) tea.Cmd {
	mouse := m.Mouse()
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
	now := time.Now()
	p.scr.reset(w, h)
	p.hits = npRender(&p.scr, p.snapshot(now, w, h))
	p.applyEffects(now)
	return p.scr.render()
}

// palette follows the active theme; unset roles fall back to Warm Night.
func (p *NowPlayingPage) palette() npPalette {
	if gen := style.StyleGeneration(); !p.palOK || gen != p.palGen {
		p.pal, p.palGen, p.palOK = npPaletteFromTheme(style.CurrentStyleSet().Theme()), gen, true
	}
	return p.pal
}

func (p *NowPlayingPage) applyEffects(now time.Time) {
	p.scr.fillBackground()
	if t := now.Sub(p.sweepAt); t >= 0 && t < npSweepDuration {
		p.scr.sweep(float64(t) / float64(npSweepDuration))
	}
	if !p.poweredOnAt.IsZero() {
		if t := now.Sub(p.poweredOnAt); t < npPowerOnDuration {
			p.scr.powerOn(float64(t) / float64(npPowerOnDuration))
		}
	}
	if !p.exitAt.IsZero() {
		p.scr.powerOff(min(float64(now.Sub(p.exitAt))/float64(npPowerOffDuration), 1))
	}
}

func (p *NowPlayingPage) snapshot(now time.Time, w, h int) *npSnapshot {
	player := p.netease.player
	song := p.currentSong()
	state := player.State()
	lines, index := p.lyrics()
	pal := p.palette()
	blue := song.Id == 0 && !p.stoppedSince.IsZero() && now.Sub(p.stoppedSince) >= npBlueScreenDelay
	var deck *npDeckFrame
	if pose, shown := p.deckPose(now); shown && !blue && !p.lyricsMode && w >= 100 && h >= 30 {
		deck = p.deck.frame(pal, npOctantSupport(), npDeckState{
			Progress: p.reels.progress(now, p.progress()),
			Angles:   p.reels.angles,
			Blur:     p.reels.blur(),
			Pose:     pose,
		})
	}
	return &npSnapshot{
		Now:        now,
		Pal:        pal,
		Song:       song,
		State:      state,
		BlueScreen: blue,
		Scanlines:  p.effects == configs.NowPlayingEffectsFull,
		Passed:     player.PassedTime(),
		Volume:     player.Volume(),
		Mode:       player.Mode(),
		Liked:      song.Id > 0 && likelist.IsLikeSong(song.Id),
		Quality:    configs.AppConfig.Player.SongLevel,
		QueuePos:   p.queuePos,
		QueueLen:   p.queueLen,
		TapeMins:   p.tapeMins,
		Upcoming:   p.upcoming,
		Lyrics:     lines,
		LyricIndex: index,
		LyricsMode: p.lyricsMode,
		OSD:        p.visibleOSD(now, state),
		OSDFade:    p.osdFade(now),
		Keys:       p.keys,
		Angles:     p.reels.angles,
		Deck:       deck,
	}
}

func (p *NowPlayingPage) lyrics() ([]npLyricLine, int) {
	if p.netease.lyricService == nil {
		return nil, -1
	}
	st := p.netease.lyricService.State()
	switch {
	case st.YRCEnabled && len(st.YRCLines) > 0:
		lines := make([]npLyricLine, len(st.YRCLines))
		for i, l := range st.YRCLines {
			var b strings.Builder
			for _, w := range l.Words {
				b.WriteString(w.Word)
			}
			lines[i].Text = b.String()
			if st.ShowTranslation {
				lines[i].Trans = l.TranslatedLyric
			}
		}
		return lines, st.YRCLineIndex
	case len(st.Fragments) > 0:
		lines := make([]npLyricLine, len(st.Fragments))
		for i, f := range st.Fragments {
			lines[i].Text = f.Content
			if st.ShowTranslation {
				lines[i].Trans = st.TranslatedFragments[f.StartTimeMs]
			}
		}
		return lines, st.CurrentIndex
	}
	return nil, -1
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
