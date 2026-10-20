package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/charmbracelet/x/ansi"
)

type framePage struct {
	text    string
	handles bool
	got     *PageTransition
}

func (p *framePage) IgnoreQuitKeyMsg(tea.KeyMsg) bool     { return false }
func (p *framePage) Type() PageType                       { return PageType("frame") }
func (p *framePage) Update(tea.Msg, *App) (Page, tea.Cmd) { return p, nil }
func (p *framePage) Msg() tea.Msg                         { return nil }
func (p *framePage) View(*App) string                     { return "\x1b[38;2;200;200;200m" + p.text + "\x1b[m" }
func (p *framePage) HandlePageTransition(_ *App, tx PageTransition) bool {
	p.got = &tx
	return p.handles
}

func newTransitionApp(level motion.Level) *App {
	options := DefaultOptions()
	options.Motion.Level = level
	a := NewApp(options)
	a.windowWidth, a.windowHeight = 20, 1
	return a
}

func TestPageCrossfade(t *testing.T) {
	a := newTransitionApp(motion.Full)
	from, to := &framePage{text: "FROM"}, &framePage{text: "TO"}
	a.page = from
	a.View()
	a.page = to
	if got := ansi.Strip(a.View().Content); !strings.Contains(got, "FROM") {
		t.Fatalf("right after the switch the old page shows: %q", got)
	}
	if to.got == nil || to.got.FromFrame == "" || to.got.From != Page(from) {
		t.Fatalf("the new page was not offered the transition: %+v", to.got)
	}
	time.Sleep(pageFade[motion.Full] + 20*time.Millisecond)
	if got := ansi.Strip(a.View().Content); !strings.Contains(got, "TO") {
		t.Fatalf("after the crossfade the new page shows: %q", got)
	}
}

func TestPageTransitionHandledOrOff(t *testing.T) {
	a := newTransitionApp(motion.Full)
	from, to := &framePage{text: "FROM"}, &framePage{text: "TO", handles: true}
	a.page = from
	a.View()
	a.page = to
	if got := ansi.Strip(a.View().Content); !strings.Contains(got, "TO") {
		t.Fatalf("a page that animates itself shows at once: %q", got)
	}

	off := newTransitionApp(motion.Off)
	off.page = from
	off.View()
	off.page = &framePage{text: "TO"}
	if got := ansi.Strip(off.View().Content); !strings.Contains(got, "TO") {
		t.Fatalf("with motion off the switch is instant: %q", got)
	}
}

// After the startup page the first page comes in row by row: the top rows
// settle while the lower ones still slide in.
func TestStartupHandoffStaggersRows(t *testing.T) {
	a := newTransitionApp(motion.Full)
	a.windowWidth, a.windowHeight = 16, 21
	a.transition = pageTransition{active: true, rows: true, from: "startup"}
	rows := make([]string, a.windowHeight)
	for i := range rows {
		rows[i] = fmt.Sprintf("row%-13d", i)
	}
	frame := strings.Join(rows, "\n")
	bg, fg := a.SurfaceColors()
	if got := ansi.Strip(a.rowsIn(frame, handoffOut/2, bg, fg)); got != "startup" {
		t.Fatalf("first the startup frame fades out: %q", got)
	}
	mid := strings.Split(ansi.Strip(a.rowsIn(frame, handoffOut+handoffSpread, bg, fg)), "\n")
	if !strings.HasPrefix(mid[0], "row0") || !strings.HasPrefix(mid[20], strings.Repeat(" ", handoffShift)+"row20") {
		t.Fatalf("the top row is in place while the last still slides: %q / %q", mid[0], mid[20])
	}
	end := a.rowsIn(frame, handoffOut+handoffSpread+handoffIn, bg, fg)
	if ansi.Strip(end) != frame {
		t.Fatal("after the entrance the page shows as it is")
	}
}

func TestStartupViewHook(t *testing.T) {
	a := newTransitionApp(motion.Full)
	a.windowWidth, a.windowHeight = 10, 2
	var got time.Duration
	s := NewStartup(&StartupOptions{LoadingDuration: time.Second, View: func(_ *App, w, h int, elapsed, total time.Duration) string {
		got = total
		return strings.Repeat("x", w) + "\n" + strings.Repeat("y", w)
	}}, nil)
	s.startedAt = time.Now()
	if v := ansi.Strip(s.View(a)); !strings.Contains(v, "xxxxxxxxxx") || got != time.Second {
		t.Fatalf("the hook draws the startup page: %q", v)
	}
}
