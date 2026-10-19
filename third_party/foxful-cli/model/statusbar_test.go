package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

type clickableStatusBarTestComponent struct {
	clicks int
}

func (c *clickableStatusBarTestComponent) View(*App, *Main) string {
	return "musicfox · [1/2] · 无损"
}

func (c *clickableStatusBarTestComponent) HandleMouse(mouse tea.Mouse, x, _ int) (bool, tea.Cmd) {
	if mouse.Button != tea.MouseLeft || !c.IsMouseOver(x, 0) {
		return false, nil
	}
	return true, func() tea.Msg {
		c.clicks++
		return nil
	}
}

func (c *clickableStatusBarTestComponent) IsMouseOver(x, _ int) bool {
	return x >= 0 && x < len("musicfox")
}

type passiveStatusBarTestComponent struct{}

func (passiveStatusBarTestComponent) View(*App, *Main) string { return "passive" }

func TestDefaultStatusBarKeepsLongBreadcrumbTitleOnOneLine(t *testing.T) {
	a := &App{windowWidth: 80, windowHeight: 24}
	m := &Main{
		menuTitle: &MenuItem{Title: "Startup Animation Gallery"},
		menuStack: &util.Stack{},
	}

	segments := computeBreadcrumbSegments(m)
	if got, want := segments[0].DisplayTitle, "Startup Animation Gallery"; got != want {
		t.Fatalf("breadcrumb title = %q, want %q", got, want)
	}

	bar := (&DefaultStatusBar{}).View(a, m)
	if got := lipgloss.Height(bar); got != 1 {
		t.Fatalf("status bar height = %d, want 1", got)
	}
}

// TestStatusBarTopDoesNotLeaveBottomGap verifies that when StatusBarPosition=StatusBarTop,
// the total rendered view height equals windowHeight without off-by-one overflow that
// produces an extra blank line at the bottom.
// Regression test for the layout bug where targetHeight did not account for the status bar
// already being included in body, causing lipgloss.Height(targetHeight) to fill to h instead of h-1.
func TestStatusBarTopDoesNotLeaveBottomGap(t *testing.T) {
	windowH := 24
	a := &App{windowWidth: 80, windowHeight: windowH}
	m := NewMain(a, &Options{
		StatusBar:         &DefaultStatusBar{},
		StatusBarPosition: StatusBarTop,
		MainMenu:          &testMenu{items: []MenuItem{{Title: "Item"}}},
		MainMenuTitle:     &MenuItem{Title: "Test Menu"},
	})
	m.menuTitle = &MenuItem{Title: "Test Menu"}

	rendered := m.View(a)
	actualH := lipgloss.Height(rendered)

	if actualH != windowH {
		t.Errorf("rendered view height = %d, want %d (windowHeight) — status bar at top leaves bottom gap",
			actualH, windowH)
	}
}

func TestDefaultStatusBarKeepsPrimaryBreadcrumbLabelWithTransparentAppBackground(t *testing.T) {
	theme := style.DefaultDarkTheme()
	theme.Primary = lipgloss.Color("#E040FB")
	theme.AppBackground = style.Highlight{Bg: lipgloss.NoColor{}}
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })

	bar := (&DefaultStatusBar{}).View(
		&App{windowWidth: 80, windowHeight: 24},
		&Main{menuTitle: &MenuItem{Title: "Menu"}, menuStack: &util.Stack{}},
	)
	markerX := strings.Index(ansi.Strip(bar), "»")
	if markerX < 0 {
		t.Fatal("status bar is missing breadcrumb marker")
	}

	screen := uv.NewScreenBuffer(lipgloss.Width(bar), lipgloss.Height(bar))
	screen.Method = ansi.GraphemeWidth
	uv.NewStyledString(bar).Draw(screen, screen.Bounds())
	cell := screen.CellAt(markerX, 0)
	if cell == nil {
		t.Fatal("breadcrumb marker cell is missing")
	}
	if cell.Style.Bg == nil {
		t.Fatal("breadcrumb marker background is missing")
	}
	gotR, gotG, gotB, gotA := cell.Style.Bg.RGBA()
	wantR, wantG, wantB, wantA := theme.Primary.RGBA()
	if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
		t.Fatalf("breadcrumb marker background = %v, want primary %v", cell.Style.Bg, theme.Primary)
	}
}

func TestDefaultStatusBarDispatchesToInjectedComponents(t *testing.T) {
	clickable := &clickableStatusBarTestComponent{}
	bar := &DefaultStatusBar{
		Components: []StatusBarComponent{clickable, passiveStatusBarTestComponent{}},
	}
	options := DefaultOptions()
	options.StatusBar = bar
	options.StatusBarPosition = StatusBarTop
	options.MainMenu = &testMenu{items: []MenuItem{{Title: "Item"}}}
	options.MainMenuTitle = &MenuItem{Title: "Menu"}

	app := NewApp(options)
	app.windowWidth = 80
	app.windowHeight = 24
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: app.windowWidth, Height: app.windowHeight}, app)
	_ = main.View(app)

	if got, want := len(bar.componentBounds), 2; got != want {
		t.Fatalf("component bounds = %d, want %d", got, want)
	}
	y := main.statusBarRowY(app)
	clickableBounds := bar.componentBounds[0]
	if !main.isOverClickableElement(clickableBounds.start, y, app) {
		t.Fatal("clickable component prefix is not recognized as clickable")
	}
	_, cmd := main.mouseClickHandle(tea.Mouse{X: clickableBounds.start, Y: y, Button: tea.MouseLeft}, app)
	if cmd == nil {
		t.Fatal("clickable component did not return a command")
	}
	_ = cmd()
	if clickable.clicks != 1 {
		t.Fatalf("clickable component count = %d, want 1", clickable.clicks)
	}

	adjacentX := clickableBounds.start + len("musicfox")
	if main.isOverClickableElement(adjacentX, y, app) {
		t.Fatalf("adjacent component text at x=%d is unexpectedly clickable", adjacentX)
	}
	_, _ = main.mouseClickHandle(tea.Mouse{X: adjacentX, Y: y, Button: tea.MouseLeft}, app)
	if clickable.clicks != 1 {
		t.Fatalf("adjacent component click count = %d, want 1", clickable.clicks)
	}

	passiveBounds := bar.componentBounds[1]
	if main.isOverClickableElement(passiveBounds.start, y, app) {
		t.Fatal("passive component is unexpectedly clickable")
	}
}
