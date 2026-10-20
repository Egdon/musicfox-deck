package model

import (
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"
)

// PageTransition describes a page switch to the page that is shown next.
type PageTransition struct {
	From      Page
	FromFrame string // the outgoing page as last shown, without popups
	Start     time.Time
}

// PageTransitionHandler is implemented by pages that animate their own
// entrance. Returning true stops the App's default crossfade.
type PageTransitionHandler interface {
	HandlePageTransition(a *App, tx PageTransition) bool
}

// Page crossfade durations: the old page dips to the background, the new
// one rises from it.
var pageFade = map[motion.Level]time.Duration{motion.Full: 200 * time.Millisecond, motion.Subtle: 150 * time.Millisecond}

// After the startup page, at full motion, the first page comes in row by
// row: the startup frame fades out, then each row fades in as it slides a
// little from the right, one after another.
var (
	handoffOut     = 80 * time.Millisecond
	handoffStagger = 6 * time.Millisecond
	handoffSpread  = 100 * time.Millisecond
	handoffIn      = 180 * time.Millisecond
)

const handoffShift = 2 // columns each row slides

type pageTransition struct {
	active bool
	rows   bool // the row-by-row entrance after the startup page
	from   string
	start  time.Time
	dur    time.Duration
}

// pageLayer renders the current page and, right after a page switch, the
// transition into it.
func (a *App) pageLayer() string {
	page := a.page
	now := a.FrameTime()
	if page != a.lastPage && a.lastPage != nil {
		level := a.MotionLevel()
		switch {
		case a.suppressTransition || level == motion.Off || a.lastPageFrame == "":
		case a.pageHandlesTransition(page, now):
		default:
			a.transition = pageTransition{active: true, from: a.lastPageFrame, start: now, dur: pageFade[level]}
			if _, ok := a.lastPage.(*StartupPage); ok && level == motion.Full {
				a.transition.rows, a.transition.dur = true, handoffOut+handoffSpread+handoffIn
			}
			a.Animate(a.transition.dur)
		}
		a.suppressTransition = false
	}
	a.lastPage = page
	frame := page.View(a)
	if a.transition.active {
		frame = a.crossfade(frame, now)
	}
	a.lastPageFrame = frame
	return frame
}

func (a *App) pageHandlesTransition(page Page, now time.Time) bool {
	h, ok := page.(PageTransitionHandler)
	return ok && h.HandlePageTransition(a, PageTransition{From: a.lastPage, FromFrame: a.lastPageFrame, Start: now})
}

// crossfade dips the old frame to the background, then raises the new one.
func (a *App) crossfade(frame string, now time.Time) string {
	tx := &a.transition
	t := float64(now.Sub(tx.start)) / float64(tx.dur)
	if t >= 1 || tx.dur <= 0 {
		tx.active = false
		return frame
	}
	bg, fg := a.SurfaceColors()
	if tx.rows {
		return a.rowsIn(frame, now.Sub(tx.start), bg, fg)
	}
	const out = 0.45
	if t < out {
		return ansix.Fade(tx.from, motion.InQuad(t/out), bg, fg)
	}
	return ansix.Fade(frame, 1-motion.OutCubic((t-out)/(1-out)), bg, fg)
}

// rowsIn is the row-by-row entrance, t into it.
func (a *App) rowsIn(frame string, t time.Duration, bg, fg color.Color) string {
	if t < handoffOut {
		return ansix.Fade(a.transition.from, motion.InQuad(float64(t)/float64(handoffOut)), bg, fg)
	}
	t -= handoffOut
	w := a.WindowWidth()
	pad := backgroundSGR(style.CurrentStyleSet().AppBackground.GetBackground())
	lines := strings.Split(frame, "\n")
	for i, l := range lines {
		d := motion.StaggerDelay(i, handoffStagger, handoffSpread)
		e := motion.OutCubic(min(max(float64(t-d)/float64(handoffIn), 0), 1))
		if e < 1 {
			lines[i] = ansix.Fade(ansix.Shift(l, int(math.Round(handoffShift*(1-e))), w, pad), 1-e, bg, fg)
		}
	}
	return strings.Join(lines, "\n")
}

// Snapshot renders a page without showing it, for pages that animate
// towards another one.
func (a *App) Snapshot(p Page) string {
	if p == nil {
		return ""
	}
	a.snapshotting = true
	defer func() { a.snapshotting = false }()
	return p.View(a)
}

// LastPageFrame returns the page layer as last shown.
func (a *App) LastPageFrame() string { return a.lastPageFrame }

// SuppressPageTransition makes the next page switch instant, for a page
// that has already animated it.
func (a *App) SuppressPageTransition() { a.suppressTransition = true }

// cancelTransition drops a running crossfade, as after a resize.
func (a *App) cancelTransition(msg tea.Msg) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		a.transition.active = false
	}
}
