package model

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/motion"
)

// Overlay motion: modals and toasts fade in as they appear and fade out once
// they are gone. One that is gone leaves a ghost, its last frame, which is
// only drawn: input, bounds and HasPopup see the live stack alone.

// Durations per motion level.
var (
	modalIn  = map[motion.Level]time.Duration{motion.Full: 180 * time.Millisecond, motion.Subtle: 150 * time.Millisecond}
	modalOut = map[motion.Level]time.Duration{motion.Full: 120 * time.Millisecond, motion.Subtle: 120 * time.Millisecond}
	toastIn  = map[motion.Level]time.Duration{motion.Full: 220 * time.Millisecond, motion.Subtle: 150 * time.Millisecond}
	toastOut = map[motion.Level]time.Duration{motion.Full: 160 * time.Millisecond, motion.Subtle: 120 * time.Millisecond}
)

const (
	modalRise  = 1 // rows a modal rises while it appears, at Full
	toastSlide = 4 // columns a toast slides from its edge, at Full
)

// overlayFrame is a modal or a toast as drawn, before any animation.
type overlayFrame struct {
	content string
	x, y    int
	z       uint64 // stacking order among modals
}

type overlayGhost struct {
	overlayFrame
	start time.Time
	level motion.Level
	edge  int // a toast leaves towards it: 1 right, -1 left, 0 in place
	toast bool
}

type overlayMotion struct {
	z          uint64
	modalZ     map[Modal]uint64
	modalStart map[Modal]time.Time
	modalLast  map[Modal]overlayFrame
	toastLast  map[NotificationID]overlayFrame
	ghosts     []overlayGhost
}

// overlayLevel is how much overlays animate: not at all without a frame
// scheduler, which would leave an animation stuck on its first frame.
func (a *App) overlayLevel() motion.Level {
	if a.motion.scheduler == nil {
		return motion.Off
	}
	return a.MotionLevel()
}

// modalShown starts a modal's entrance.
func (a *App) modalShown(m Modal) {
	o := &a.overlay
	if o.modalZ == nil {
		o.modalZ, o.modalStart, o.modalLast = map[Modal]uint64{}, map[Modal]time.Time{}, map[Modal]overlayFrame{}
	}
	o.z++
	o.modalZ[m], o.modalStart[m] = o.z, time.Now()
	a.Animate(modalIn[a.overlayLevel()])
}

// modalGone leaves a ghost of a modal that was taken off the stack.
func (a *App) modalGone(m Modal) {
	o := &a.overlay
	last, drawn := o.modalLast[m]
	delete(o.modalZ, m)
	delete(o.modalStart, m)
	delete(o.modalLast, m)
	if level := a.overlayLevel(); drawn && level != motion.Off {
		o.ghosts = append(o.ghosts, overlayGhost{overlayFrame: last, start: time.Now(), level: level})
		a.Animate(modalOut[level])
	}
}

// toastGone leaves a ghost of a notification that was removed.
func (a *App) toastGone(id NotificationID) {
	o := &a.overlay
	last, drawn := o.toastLast[id]
	delete(o.toastLast, id)
	if level := a.overlayLevel(); drawn && level != motion.Off {
		edge := toastEdge(a.options.NotificationOptions.Anchor)
		o.ghosts = append(o.ghosts, overlayGhost{overlayFrame: last, start: time.Now(), level: level, edge: edge, toast: true})
		a.Animate(toastOut[level])
	}
}

// hasGhosts reports whether closed modals or toasts are still fading.
func (a *App) hasGhosts(toast bool) bool {
	for _, g := range a.overlay.ghosts {
		if g.toast == toast {
			return true
		}
	}
	return false
}

// toastEdge is the screen edge toasts at an anchor slide from.
func toastEdge(anchor PopupAnchor) int {
	switch anchor {
	case AnchorTopRight, AnchorBottomRight:
		return 1
	case AnchorTopLeft, AnchorBottomLeft:
		return -1
	}
	return 0
}

// overlayLayer is a layer and its stacking order.
type overlayLayer struct {
	layer *layout.Layer
	z     uint64
}

// modalLayer remembers how a modal is drawn and returns its layer as it
// looks at now: rising into place while it fades in.
func (a *App) modalLayer(m Modal, content string, x, y int, now time.Time) overlayLayer {
	o := &a.overlay
	z := o.modalZ[m]
	if o.modalLast != nil {
		o.modalLast[m] = overlayFrame{content: content, x: x, y: y, z: z}
	}
	if start, ok := o.modalStart[m]; ok {
		level := a.overlayLevel()
		if t := progress(now, start, modalIn[level]); t < 1 && level != motion.Off {
			e := motion.OutCubic(t)
			if level == motion.Full {
				y += int(math.Round(modalRise * (1 - e)))
			}
			content = a.fadeOverlay(content, 1-e)
		} else {
			delete(o.modalStart, m)
		}
	}
	return a.clippedLayer(content, x, y, z)
}

// toastLayer remembers how a toast is drawn and returns its layer as it
// looks at now: sliding in from its edge while it fades in.
func (a *App) toastLayer(n *Notification, content string, x, y int, now time.Time) overlayLayer {
	o := &a.overlay
	if o.toastLast == nil {
		o.toastLast = map[NotificationID]overlayFrame{}
	}
	o.toastLast[n.id] = overlayFrame{content: content, x: x, y: y}
	level := a.overlayLevel()
	if t := progress(now, n.createdAt, toastIn[level]); t < 1 && level != motion.Off {
		e := motion.OutCubic(t)
		if level == motion.Full {
			x += toastEdge(a.options.NotificationOptions.Anchor) * int(math.Round(toastSlide*(1-e)))
		}
		content = a.fadeOverlay(content, 1-e)
	}
	return a.clippedLayer(content, x, y, 0)
}

// ghostLayers returns the layers of closed modals, or of removed toasts, as
// they fade out at now; toasts also slide towards their edge. Ghosts that
// are done are dropped.
func (a *App) ghostLayers(toast bool, now time.Time) []overlayLayer {
	o := &a.overlay
	var out []overlayLayer
	kept := o.ghosts[:0]
	for _, g := range o.ghosts {
		dur := modalOut[g.level]
		if g.toast {
			dur = toastOut[g.level]
		}
		t := progress(now, g.start, dur)
		switch {
		case g.toast != toast:
		case t >= 1:
			continue
		default:
			e := motion.OutCubic(t)
			x := g.x
			if g.toast && g.level == motion.Full {
				x += g.edge * int(math.Round(toastSlide*e))
			}
			out = append(out, a.clippedLayer(a.fadeOverlay(g.content, e), x, g.y, g.z))
		}
		kept = append(kept, g)
	}
	o.ghosts = kept
	return out
}

// fadeOverlay fades an overlay towards the background by k.
func (a *App) fadeOverlay(content string, k float64) string {
	bg, fg := a.SurfaceColors()
	return ansix.Fade(content, k, bg, fg)
}

// clippedLayer places content at (x, y), cut to the window: the compositor
// would otherwise grow its canvas around a layer that sticks out.
func (a *App) clippedLayer(content string, x, y int, z uint64) overlayLayer {
	w, h := a.WindowWidth(), a.WindowHeight()
	lines := strings.Split(content, "\n")
	if y < 0 {
		lines = lines[min(-y, len(lines)):]
		y = 0
	}
	lines = lines[:max(min(len(lines), h-y), 0)]
	if len(lines) == 0 || x >= w {
		return overlayLayer{z: z}
	}
	left, right := max(-x, 0), w-x
	for i, l := range lines {
		if left > 0 || ansix.Width(l) > right {
			lines[i] = ansix.Cut(l, left, right)
		}
	}
	return overlayLayer{layer: layout.NewLayer(strings.Join(lines, "\n")).X(max(x, 0)).Y(y), z: z}
}

// composeOverlays draws layers over base in their stacking order.
func composeOverlays(base string, ls []overlayLayer) string {
	sort.SliceStable(ls, func(i, j int) bool { return ls[i].z < ls[j].z })
	layers := []*layout.Layer{layout.NewLayer(base)}
	for _, l := range ls {
		if l.layer != nil {
			layers = append(layers, l.layer)
		}
	}
	return layout.NewCompositor(layers...).Render()
}

// progress is how far into d now is since start, from 0 to 1.
func progress(now, start time.Time, d time.Duration) float64 {
	if d <= 0 {
		return 1
	}
	return min(max(float64(now.Sub(start))/float64(d), 0), 1)
}
