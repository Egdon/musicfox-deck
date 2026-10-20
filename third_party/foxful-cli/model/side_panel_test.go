package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

type testPanel struct {
	minWidth int
	lastX    int
	lastY    int
	left     bool
}

func (p *testPanel) Layout(_ *App, w, _ int) SidePanelLayout {
	if w < p.minWidth {
		return SidePanelLayout{}
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
