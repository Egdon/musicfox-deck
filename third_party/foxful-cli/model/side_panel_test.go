package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

type testPanel struct {
	minWidth int
	dock     int // rows of the dock below minWidth; 0 hides the panel there
	lastX    int
	lastY    int
	left     bool
}

func (p *testPanel) Layout(_ *App, w, _ int) SidePanelLayout {
	if w < p.minWidth {
		return SidePanelLayout{Height: p.dock, FooterRows: 2}
	}
	return SidePanelLayout{Width: 30, FooterRows: 2}
}

func (p *testPanel) Update(tea.Msg, *App) {}

func (p *testPanel) View(_ *App, _ *Main, w, h int) string {
	lines := make([]string, h)
	for i := range lines {
		lines[i] = ansix.Fit("PANEL 王", w, "")
	}
	return strings.Join(lines, "\n")
}

func (p *testPanel) HandleMouse(_ *App, _ *Main, msg tea.MouseMsg, x, y int) SidePanelMouseResult {
	p.lastX, p.lastY = x, y
	return SidePanelMouseResult{Handled: true, Pointer: "pointer"}
}

func (p *testPanel) MouseLeave(*App, *Main) { p.left = true }

type footerComponent struct{}

func (footerComponent) Update(tea.Msg, *App) {}

func (footerComponent) View(*App, *Main) (string, int) { return "FOOTER\nPROGRESS", 2 }

func newPanelTestMain(t *testing.T, w, h int, panel *testPanel) (*App, *Main) {
	t.Helper()
	prev := style.CurrentStyleSet()
	t.Cleanup(func() { style.SetStyleSet(prev) })
	style.SetStyleSet(style.NewStyleSet(warmNightTestTheme()))
	options := DefaultOptions()
	options.MainMenu = &mockMenu{key: "songs", items: goldenSongs}
	options.MainMenuTitle = &MenuItem{Title: "飙升榜"}
	options.SidePanel = panel
	options.Components = []Component{footerComponent{}}
	options.StatusBar = goldenStatusBar{}
	options.StatusBarPosition = StatusBarTop
	app := NewApp(options)
	app.windowWidth, app.windowHeight = w, h
	main := NewMain(app, options)
	app.main = main
	main.Update(tea.WindowSizeMsg{Width: w, Height: h}, app)
	return app, main
}

func TestSidePanelLayout(t *testing.T) {
	panel := &testPanel{minWidth: 100}
	app, main := newPanelTestMain(t, 120, 30, panel)
	if main.ContentWidth() != 90 || !main.PanelVisible() {
		t.Fatalf("content width %d, panel visible %v", main.ContentWidth(), main.PanelVisible())
	}
	x, y, w, h, ok := main.PanelRect()
	if !ok || x != 90 || w != 30 || y < 1 || y+h != 30-2 {
		t.Fatalf("panel rect %d,%d %dx%d", x, y, w, h)
	}
	lines := strings.Split(main.View(app), "\n")
	if len(lines) != 30 {
		t.Fatalf("%d lines", len(lines))
	}
	for i, line := range lines {
		if got := ansix.Width(line); got != 120 {
			t.Fatalf("line %d is %d cells: %q", i, got, ansi.Strip(line))
		}
		plain := ansi.Strip(line)
		inPanel := i >= y && i < y+h
		if inPanel != strings.Contains(plain, "PANEL 王") {
			t.Fatalf("line %d panel content %v: %q", i, !inPanel, plain)
		}
	}
	if !strings.HasPrefix(ansi.Strip(lines[28]), "FOOTER") || !strings.HasPrefix(ansi.Strip(lines[29]), "PROGRESS") {
		t.Fatalf("footer rows: %q / %q", ansi.Strip(lines[28]), ansi.Strip(lines[29]))
	}
	// The list fills the rows above the footer.
	if main.menuPageSize < 20 {
		t.Fatalf("page size %d", main.menuPageSize)
	}

	// Mouse events over the panel arrive in panel coordinates.
	page, _ := main.Update(tea.MouseMotionMsg{X: 95, Y: y + 3}, app)
	if page != main || panel.lastX != 5 || panel.lastY != 3 || main.panelPointer != "pointer" {
		t.Fatalf("panel got (%d,%d) pointer %q", panel.lastX, panel.lastY, main.panelPointer)
	}
	main.Update(tea.MouseMotionMsg{X: 10, Y: y + 3}, app)
	if !panel.left || main.panelPointer != "" {
		t.Fatal("leaving the panel did not reset it")
	}
}

func TestSidePanelHiddenWhenNarrow(t *testing.T) {
	_, main := newPanelTestMain(t, 90, 30, &testPanel{minWidth: 100})
	if main.PanelVisible() || main.ContentWidth() != 90 {
		t.Fatalf("panel visible %v, content width %d", main.PanelVisible(), main.ContentWidth())
	}
}

func TestSidePanelDocksUnderTheMenu(t *testing.T) {
	panel := &testPanel{minWidth: 100, dock: 8}
	app, main := newPanelTestMain(t, 80, 30, panel)
	if !main.PanelVisible() || main.ContentWidth() != 80 {
		t.Fatalf("panel visible %v, content width %d", main.PanelVisible(), main.ContentWidth())
	}
	x, y, w, h, ok := main.PanelRect()
	if !ok || x != 0 || w != 80 || h != 8 || y != 30-2-8 {
		t.Fatalf("dock rect %d,%d %dx%d", x, y, w, h)
	}
	lines := strings.Split(main.View(app), "\n")
	if len(lines) != 30 {
		t.Fatalf("%d lines", len(lines))
	}
	for i, line := range lines {
		if got := ansix.Width(line); got != 80 {
			t.Fatalf("line %d is %d cells: %q", i, got, ansi.Strip(line))
		}
		if inDock := i >= y && i < y+h; inDock != strings.Contains(ansi.Strip(line), "PANEL 王") {
			t.Fatalf("line %d: dock content %v: %q", i, !inDock, ansi.Strip(line))
		}
	}
	if rows := main.menuBottomRow - main.menuListStartRow; main.menuListStartRow+rows >= y {
		t.Fatalf("the list (rows %d..%d) runs into the dock at %d", main.menuListStartRow, main.menuBottomRow, y)
	}
	main.Update(tea.MouseMotionMsg{X: 7, Y: y + 2}, app)
	if panel.lastX != 7 || panel.lastY != 2 {
		t.Fatalf("dock got (%d,%d)", panel.lastX, panel.lastY)
	}
}

// The width memo measures as ansi.StringWidth does and forgets rows unused
// for two frames.
func TestWidthMemo(t *testing.T) {
	var m widthMemo
	rows := []string{"", "plain", "\x1b[31m红色\x1b[m", "\x1b[1m👩‍👩‍👧\x1b[22m family"}
	for range 3 {
		for _, r := range rows {
			if got, want := m.width(r), ansi.StringWidth(r); got != want {
				t.Fatalf("%q: %d, want %d", r, got, want)
			}
		}
		m.next()
	}
	m.next()
	m.next()
	if len(m.cur)+len(m.prev) != 0 {
		t.Fatal("rows unused for two frames are still remembered")
	}
}

// The menu block is stacked again only when its sections change.
func TestStackMenuFollowsTheSections(t *testing.T) {
	var m Main
	for _, sections := range [][]string{{"ab", "c"}, {"ab", "c"}, {"ab", "\x1b[31mx\x1b[m\nyy"}, {"ab", "c"}} {
		if got, want := m.stackMenu(sections), layout.JoinVertical(lipgloss.Left, sections...); got != want {
			t.Fatalf("%q: %q, want %q", sections, got, want)
		}
	}
}
