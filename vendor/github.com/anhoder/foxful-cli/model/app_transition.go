package model

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/motion"
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

type pageTransition struct {
	active bool
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
	const out = 0.45
	if t < out {
		return ansix.Fade(tx.from, motion.InQuad(t/out), bg, fg)
	}
	return ansix.Fade(frame, 1-motion.OutCubic((t-out)/(1-out)), bg, fg)
}

// Snapshot renders a page without showing it, for pages that animate
// towards another one.
func (a *App) Snapshot(p Page) string {
	if p == nil {
		return ""
	}
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
