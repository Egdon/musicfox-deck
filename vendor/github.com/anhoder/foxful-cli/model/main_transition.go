package model

import (
	"math"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"
)

// Menu transitions: entering a menu slides its rows in from the right,
// going back slides them in from the left, and turning a page or a tab
// follows the turn. At full motion the old rows first fade out as they
// drift the other way, and the new ones come in one after another.

var (
	menuOut        = 50 * time.Millisecond // the old rows fade out, at Full
	menuIn         = map[motion.Level]time.Duration{motion.Full: 140 * time.Millisecond, motion.Subtle: 120 * time.Millisecond}
	menuStagger    = 6 * time.Millisecond // between the new rows, at Full
	menuStaggerMax = 50 * time.Millisecond
)

const (
	menuShiftIn  = 3 // columns the new rows travel
	menuShiftOut = 2 // columns the old rows travel
)

// menuKey tells menus apart for transitions.
type menuKey struct {
	menu  string
	depth int
	page  int
	tab   int
}

type menuTransition struct {
	key   menuKey // the menu last shown
	known bool
	seq   uint64   // the frame Main was last shown in
	size  [2]int   // the window then
	last  []string // its menu block, before any animation

	start time.Time // zero when no transition runs
	dir   int       // 1: the new rows come from the right, -1: from the left
	from  []string
	level motion.Level
}

func (m *Main) menuKeyNow() menuKey {
	k := menuKey{depth: m.menuStack.Len(), page: m.menuCurPage, tab: m.activeTab}
	if m.menu != nil {
		k.menu = m.menu.GetMenuKey()
	}
	return k
}

// menuDirection is where the rows of the menu shown next come from.
func menuDirection(from, to menuKey) int {
	switch {
	case to.tab != from.tab:
		return sign(to.tab - from.tab)
	case to.depth != from.depth:
		return sign(to.depth - from.depth)
	case to.menu == from.menu && to.page != from.page:
		return sign(to.page - from.page)
	}
	return 1
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}

func menuTransitionLength(level motion.Level) time.Duration {
	if level == motion.Full {
		return menuOut + menuStaggerMax + menuIn[level]
	}
	return menuIn[level]
}

// animateMenu starts a transition when the menu changed since the last
// frame, and returns the menu block as the transition shows it. Off-screen
// snapshots, frames after Main was away or the window changed, and menus
// still loading start none.
func (m *Main) animateMenu(a *App, sections []string) []string {
	if a.snapshotting {
		return sections
	}
	tr := &m.menuTr
	lines := strings.Split(strings.Join(sections, "\n"), "\n")
	key, now, seq := m.menuKeyNow(), a.FrameTime(), a.FrameSeq()
	size := [2]int{a.WindowWidth(), a.WindowHeight()}
	steady := tr.seq+1 == seq && tr.size == size // Main was shown last frame, in this window
	if !steady {
		tr.start = time.Time{}
	}
	if m.pendingEnterMenu == nil {
		level := a.overlayLevel()
		if tr.known && key != tr.key && level != motion.Off && steady {
			tr.start, tr.dir, tr.from, tr.level = now, menuDirection(tr.key, key), tr.last, level
			a.Animate(menuTransitionLength(level))
		}
		tr.key, tr.known = key, true
	}
	tr.seq, tr.size, tr.last = seq, size, lines
	if tr.start.IsZero() {
		return sections
	}
	t := now.Sub(tr.start)
	if t >= menuTransitionLength(tr.level) {
		tr.start, tr.from = time.Time{}, nil
		return sections
	}
	return []string{strings.Join(m.menuFrame(a, lines, t), "\n")}
}

// menuFrame draws the menu block t into a transition.
func (m *Main) menuFrame(a *App, lines []string, t time.Duration) []string {
	tr := &m.menuTr
	w := m.ContentWidth()
	bg, fg := a.SurfaceColors()
	pad := backgroundSGR(style.CurrentStyleSet().AppBackground.GetBackground())
	full := tr.level == motion.Full
	out := make([]string, len(lines))
	if full && t < menuOut {
		e := motion.OutCubic(float64(t) / float64(menuOut))
		shift := -tr.dir * int(math.Round(menuShiftOut*e))
		for i := range out {
			old := ""
			if i < len(tr.from) {
				old = tr.from[i]
			}
			out[i] = ansix.Fade(ansix.Shift(old, shift, w, pad), e, bg, fg)
		}
		return out
	}
	if full {
		t -= menuOut
	}
	for i, l := range lines {
		var delay time.Duration
		if full {
			delay = min(time.Duration(i)*menuStagger, menuStaggerMax)
		}
		e := motion.OutCubic(min(max(float64(t-delay)/float64(menuIn[tr.level]), 0), 1))
		shift := 0
		if full {
			shift = tr.dir * int(math.Round(menuShiftIn*(1-e)))
		}
		out[i] = ansix.Fade(ansix.Shift(l, shift, w, pad), 1-e, bg, fg)
	}
	return out
}
