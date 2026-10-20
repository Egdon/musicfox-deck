package model

import (
	"strings"
	"testing"
	"time"

	"github.com/anhoder/foxful-cli/motion"
	"github.com/charmbracelet/x/ansi"
)

func newOverlayApp(t *testing.T, level motion.Level) *App {
	t.Helper()
	options := DefaultOptions()
	options.Motion.Level = level
	a := NewApp(options)
	a.windowWidth, a.windowHeight = 60, 16
	a.motion.scheduler = motion.NewScheduler(16*time.Millisecond, level, func(any) {})
	t.Cleanup(a.motion.scheduler.Stop)
	return a
}

func overlayBase(a *App) string {
	line := strings.Repeat(".", a.windowWidth)
	return strings.TrimSuffix(strings.Repeat(line+"\n", a.windowHeight), "\n")
}

// overlayAt composites the overlays as they look d after start.
func overlayAt(a *App, start time.Time, d time.Duration, modals bool) []string {
	a.motion.frameNow = start.Add(d)
	out := overlayBase(a)
	if modals {
		out = a.compositeModals(out)
	} else {
		out = a.compositeNotifications(out)
	}
	return strings.Split(out, "\n")
}

// find returns the row and column where text shows, or -1, -1.
func find(lines []string, text string) (int, int) {
	for y, l := range lines {
		if x := strings.Index(ansi.Strip(l), text); x >= 0 {
			return y, ansi.StringWidth(ansi.Strip(l)[:x])
		}
	}
	return -1, -1
}

func checkOverlayFrame(t *testing.T, a *App, lines []string, when string) {
	t.Helper()
	if len(lines) != a.windowHeight {
		t.Fatalf("%s: %d lines, want %d", when, len(lines), a.windowHeight)
	}
	for i, l := range lines {
		if w := ansi.StringWidth(l); w != a.windowWidth {
			t.Fatalf("%s: line %d is %d cells: %q", when, i, w, ansi.Strip(l))
		}
	}
}

func TestPopupRisesInAndFadesOut(t *testing.T) {
	a := newOverlayApp(t, motion.Full)
	p, err := NewPopup(PopupSpec{Title: "Saved", Content: "All good"})
	if err != nil {
		t.Fatal(err)
	}
	a.ShowPopup(p)
	start := a.overlay.modalStart[p]

	first := overlayAt(a, start, 5*time.Millisecond, true)
	settled := overlayAt(a, start, time.Second, true)
	checkOverlayFrame(t, a, first, "entering")
	checkOverlayFrame(t, a, settled, "settled")
	y0, _ := find(first, "All good")
	y1, _ := find(settled, "All good")
	if y0 != y1+modalRise || first[y0] == settled[y1] {
		t.Fatalf("an entering popup shows a row lower and faded: rows %d and %d", y0, y1)
	}
	if again := overlayAt(a, start, 2*time.Second, true); strings.Join(again, "\n") != strings.Join(settled, "\n") {
		t.Fatal("a settled popup must not change between frames")
	}

	a.DismissPopup()
	if a.HasPopup() {
		t.Fatal("a dismissed popup is off the stack at once")
	}
	gone := a.overlay.ghosts[0].start
	fading := overlayAt(a, gone, modalOut[motion.Full]/2, true)
	checkOverlayFrame(t, a, fading, "leaving")
	if y, _ := find(fading, "All good"); y != y1 || fading[y] == settled[y1] {
		t.Fatalf("a closing popup fades where it was, at row %d", y)
	}
	if y, _ := find(overlayAt(a, gone, modalOut[motion.Full]+time.Millisecond, true), "All good"); y >= 0 || len(a.overlay.ghosts) != 0 {
		t.Fatalf("the ghost is gone once faded: row %d, %d ghosts", y, len(a.overlay.ghosts))
	}
}

func TestToastSlidesFromItsEdge(t *testing.T) {
	a := newOverlayApp(t, motion.Full)
	a.handleShowNotification(NotificationSpec{Level: NotificationInfo, Title: "Saved", Message: "All good"})
	n := a.notifications[0]

	first := overlayAt(a, n.createdAt, 5*time.Millisecond, false)
	settled := overlayAt(a, n.createdAt, time.Second, false)
	checkOverlayFrame(t, a, first, "entering")
	_, x0 := find(first, "All good")
	y1, x1 := find(settled, "All good")
	if x0 != x1+toastSlide {
		t.Fatalf("an entering toast starts %d columns right of its place: %d, %d", toastSlide, x0, x1)
	}

	a.removeNotification(n.id)
	gone := a.overlay.ghosts[0].start
	leaving := overlayAt(a, gone, toastOut[motion.Full]*3/4, false)
	checkOverlayFrame(t, a, leaving, "leaving")
	if y, x := find(leaving, "All"); y != y1 || x <= x1 {
		t.Fatalf("a removed toast slides out to the right: at %d,%d from %d,%d", y, x, y1, x1)
	}
	if y, _ := find(overlayAt(a, gone, toastOut[motion.Full]+time.Millisecond, false), "All"); y >= 0 {
		t.Fatal("the toast's ghost is gone once it has slid out")
	}
}

func TestOverlaysStillWithoutMotion(t *testing.T) {
	for _, level := range []motion.Level{motion.Off, motion.Subtle} {
		a := newOverlayApp(t, level)
		p, _ := NewPopup(PopupSpec{Title: "Saved", Content: "All good"})
		a.ShowPopup(p)
		start := a.overlay.modalStart[p]
		y0, _ := find(overlayAt(a, start, 5*time.Millisecond, true), "All good")
		y1, _ := find(overlayAt(a, start, time.Second, true), "All good")
		if y0 != y1 {
			t.Fatalf("level %v: only full motion moves popups, rows %d and %d", level, y0, y1)
		}
		a.DismissPopup()
		if ghosts := len(a.overlay.ghosts); (level == motion.Off) != (ghosts == 0) {
			t.Fatalf("level %v: %d ghosts", level, ghosts)
		}
	}
}

// A popup closed under another keeps its place in the stack while it fades.
func TestModalGhostKeepsItsLayer(t *testing.T) {
	a := newOverlayApp(t, motion.Full)
	under, _ := NewPopup(PopupSpec{Title: "Under", Content: "bottom popup text"})
	over, _ := NewPopup(PopupSpec{Title: "Over", Content: "top popup"})
	a.ShowPopup(under)
	a.ShowPopup(over)
	overlayAt(a, time.Now(), time.Second, true)
	a.DismissPopup() // over
	fresh, _ := NewPopup(PopupSpec{Title: "Fresh", Content: "newest popup"})
	a.ShowPopup(fresh)
	layers := a.ghostLayers(false, time.Now())
	if len(layers) != 1 || layers[0].z != a.overlay.z-1 || a.overlay.modalZ[fresh] <= layers[0].z || a.overlay.modalZ[under] >= layers[0].z {
		t.Fatalf("ghost z %v between the live popups %d and %d", layers, a.overlay.modalZ[under], a.overlay.modalZ[fresh])
	}
}
