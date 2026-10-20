package model

import (
	"image/color"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"
	uv "github.com/charmbracelet/ultraviolet"
)

// MotionOptions set how much the interface animates.
type MotionOptions struct {
	Level         motion.Level  // Full unless set
	FrameInterval time.Duration // 16 ms when zero

	// GlideSteps returns how many steps of a row the selection marker of a
	// tracklist can glide by: 4 where the terminal draws octants, 3 with
	// sextants, 2 with quadrants. Nil, or under 2, keeps the marker still
	// while the selection band fades across.
	GlideSteps func() int
}

// appMotion is the App's animation state.
type appMotion struct {
	scheduler *motion.Scheduler
	frameNow  time.Time // sampled once per View
	frameSeq  uint64
	termBg    color.Color // reported by the terminal
	termFg    color.Color

	termVersion  string // XTVERSION reply
	cellW, cellH int    // cell size in pixels, when reported
}

func (a *App) startMotion() {
	interval := a.options.Motion.FrameInterval
	if interval <= 0 {
		interval = 16 * time.Millisecond
	}
	a.motion.scheduler = motion.NewScheduler(interval, a.options.Motion.Level, func(m any) {
		if a.program != nil {
			a.program.Send(m)
		}
	})
}

// MotionLevel returns how much the interface may animate.
func (a *App) MotionLevel() motion.Level {
	if a.motion.scheduler == nil {
		return a.options.Motion.Level
	}
	return a.motion.scheduler.Level()
}

// Animate keeps frames coming for d. Frames re-render the current page,
// which also receives them as motion.FrameMsg.
func (a *App) Animate(d time.Duration) { a.motion.scheduler.Animate(d) }

// AnimateSlow keeps frames coming for d at the slow rate, for continuous
// motion that does not need every frame.
func (a *App) AnimateSlow(d time.Duration) { a.motion.scheduler.AnimateSlow(d) }

// AnimateUntil keeps frames coming until t.
func (a *App) AnimateUntil(t time.Time) { a.motion.scheduler.AnimateUntil(t) }

// FrameTime is the time of the frame being rendered; animations sample it so
// that everything in one frame agrees.
func (a *App) FrameTime() time.Time {
	if a.motion.frameNow.IsZero() {
		return time.Now()
	}
	return a.motion.frameNow
}

// FrameSeq counts rendered frames.
func (a *App) FrameSeq() uint64 { return a.motion.frameSeq }

// SurfaceColors returns the background fades blend into and the foreground
// of text in the terminal's default color: the theme's when it sets them,
// otherwise what the terminal reported, otherwise a guess.
func (a *App) SurfaceColors() (bg, fg color.Color) {
	ss := a.StyleSet()
	bg, fg = ss.AppBackground.GetBackground(), ss.Theme().Foreground
	if isNoColor(bg) {
		bg = a.motion.termBg
	}
	if isNoColor(fg) {
		fg = a.motion.termFg
	}
	dark := style.HasDarkBackground()
	if isNoColor(bg) {
		bg = lipgloss.Color("#FFFFFF")
		if dark {
			bg = lipgloss.Color("#000000")
		}
	}
	if isNoColor(fg) {
		fg = lipgloss.Color("#202020")
		if dark {
			fg = lipgloss.Color("#D0D0D0")
		}
	}
	return bg, fg
}

// handleMotionMsg takes frame and terminal color messages. It reports
// whether the message was a frame, which the caller passes on to the page.
func (a *App) handleMotionMsg(msg tea.Msg) bool {
	switch m := msg.(type) {
	case motion.FrameMsg:
		if a.motion.scheduler != nil {
			a.motion.scheduler.HandleFrame(m)
		}
		return true
	case tea.BackgroundColorMsg:
		a.motion.termBg = m.Color
	case tea.ForegroundColorMsg:
		a.motion.termFg = m.Color
	case tea.TerminalVersionMsg:
		a.motion.termVersion = m.Name
	case uv.CellSizeEvent:
		a.motion.cellW, a.motion.cellH = m.Width, m.Height
	}
	return false
}

// TerminalVersion returns the terminal's XTVERSION reply, such as
// "ghostty 1.3.1", or "" until it answers.
func (a *App) TerminalVersion() string { return a.motion.termVersion }

// CellAspect returns the height of a terminal cell over its width, or 0 when
// the terminal has not reported its cell size.
func (a *App) CellAspect() float64 {
	if a.motion.cellW <= 0 || a.motion.cellH <= 0 {
		return 0
	}
	return float64(a.motion.cellH) / float64(a.motion.cellW)
}

func (a *App) stopMotion() {
	a.motion.scheduler.Stop()
}
