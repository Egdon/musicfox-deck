package model

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/charmbracelet/x/ansi"
)

func newTransitionMain(t *testing.T, level motion.Level) (*App, *Main, func(time.Time) []string) {
	t.Helper()
	options := DefaultOptions()
	options.DualColumn = false
	options.Motion.Level = level
	options.MainMenu = &mockMenu{key: "root", items: []MenuItem{{Title: "Alpha"}, {Title: "Beta"}}}
	options.MainMenuTitle = &MenuItem{Title: "Root"}
	app := NewApp(options)
	app.windowWidth, app.windowHeight = 80, 24
	app.motion.scheduler = motion.NewScheduler(16*time.Millisecond, level, func(any) {})
	t.Cleanup(app.motion.scheduler.Stop)
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: 80, Height: 24}, app)
	// frame renders Main as the App would in its next frame, at a given time.
	frame := func(at time.Time) []string {
		app.motion.frameSeq++
		app.motion.frameNow = at
		return strings.Split(ansi.Strip(main.View(app)), "\n")
	}
	return app, main, frame
}

// column returns where text starts in the frame, or -1.
func column(lines []string, text string) int {
	for _, l := range lines {
		if i := strings.Index(l, text); i >= 0 {
			return ansi.StringWidth(l[:i])
		}
	}
	return -1
}

func TestMenuEnterSlidesInFromTheRight(t *testing.T) {
	_, main, frame := newTransitionMain(t, motion.Full)
	t0 := time.Now()
	home := column(frame(t0), "Alpha")

	main.EnterMenu(&mockMenu{key: "sub", items: []MenuItem{{Title: "Inner"}}}, &MenuItem{Title: "Sub"})
	start := t0.Add(16 * time.Millisecond)
	if got := frame(start); column(got, "Alpha") != home || column(got, "Inner") >= 0 {
		t.Fatal("a transition starts on the old rows")
	}
	if got := column(frame(start.Add(menuOut*3/5)), "Alpha"); got != home-menuShiftOut {
		t.Fatalf("the old rows drift left as they fade: column %d, from %d", got, home)
	}
	settled := column(frame(start.Add(time.Second)), "Inner")
	main.menuTr.start = start // rewind to look inside the transition
	if got := column(frame(start.Add(menuOut+time.Millisecond)), "Inner"); got != settled+menuShiftIn {
		t.Fatalf("the new rows come from the right: column %d, settled %d", got, settled)
	}
	if got := column(frame(start.Add(time.Second)), "Inner"); got != settled {
		t.Fatalf("after the transition the rows rest at %d, got %d", settled, got)
	}

	main.BackMenu()
	back := start.Add(2 * time.Second)
	frame(back)
	if got := column(frame(back.Add(menuOut+time.Millisecond)), "Alpha"); got != home-menuShiftIn {
		t.Fatalf("going back, rows come from the left: column %d, home %d", got, home)
	}
}

func TestMenuTransitionNeedsMainOnScreen(t *testing.T) {
	app, main, frame := newTransitionMain(t, motion.Full)
	t0 := time.Now()
	frame(t0)
	main.EnterMenu(&mockMenu{key: "sub", items: []MenuItem{{Title: "Inner"}}}, &MenuItem{Title: "Sub"})
	app.motion.frameSeq++ // a frame of another page
	if got := frame(t0.Add(40 * time.Millisecond)); column(got, "Inner") < 0 || !main.menuTr.start.IsZero() {
		t.Fatal("coming back to Main shows the new menu at once")
	}

	app.snapshotting = true
	main.BackMenu()
	frame(t0.Add(80 * time.Millisecond))
	app.snapshotting = false
	if !main.menuTr.start.IsZero() {
		t.Fatal("an off-screen snapshot starts no transition")
	}
}

func TestMenuTransitionLevels(t *testing.T) {
	for _, level := range []motion.Level{motion.Off, motion.Subtle} {
		_, main, frame := newTransitionMain(t, level)
		t0 := time.Now()
		frame(t0)
		main.EnterMenu(&mockMenu{key: "sub", items: []MenuItem{{Title: "Inner"}}}, &MenuItem{Title: "Sub"})
		start := t0.Add(16 * time.Millisecond)
		first := frame(start)
		if level == motion.Off && !main.menuTr.start.IsZero() {
			t.Fatal("no transition with motion off")
		}
		if got, settled := column(first, "Inner"), column(frame(start.Add(time.Second)), "Inner"); got != settled {
			t.Fatalf("level %v: rows do not move, column %d vs %d", level, got, settled)
		}
	}
}

func TestMenuDirection(t *testing.T) {
	root := menuKey{menu: "a", depth: 1, page: 2}
	for _, c := range []struct {
		to   menuKey
		want int
	}{
		{menuKey{menu: "b", depth: 2, page: 1}, 1},
		{menuKey{menu: "c", depth: 0, page: 1}, -1},
		{menuKey{menu: "a", depth: 1, page: 3}, 1},
		{menuKey{menu: "a", depth: 1, page: 1}, -1},
		{menuKey{menu: "a", depth: 1, page: 2, tab: -1}, -1},
		{menuKey{menu: "z", depth: 1, page: 1}, 1},
	} {
		if got := menuDirection(root, c.to); got != c.want {
			t.Errorf("%+v → %+v: %d, want %d", root, c.to, got, c.want)
		}
	}
}
