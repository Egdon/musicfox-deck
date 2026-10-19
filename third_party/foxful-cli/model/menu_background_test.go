package model

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// renderMainToScreen renders the main view into a screen buffer for cell inspection.
func renderMainToScreen(content string, w, h int) uv.ScreenBuffer {
	screen := uv.NewScreenBuffer(w, h)
	screen.Method = ansi.GraphemeWidth
	uv.NewStyledString(content).Draw(screen, screen.Bounds())
	return screen
}

// TestMainViewPaintsAppBackgroundOnEveryCell verifies that when a theme sets a
// solid AppBackground, every rendered cell carries a non-transparent background.
// Transparent cells would let content drawn beneath the TUI (e.g. a cover image
// placed with the Kitty graphics protocol at a negative z-index) bleed through,
// producing the "background holes" seen with light themes.
//
// Wide-glyph continuation cells (width 0, following a width-2 rune) are exempt:
// the terminal paints them with the preceding glyph's background, so they are
// never visible holes.
func TestMainViewPaintsAppBackgroundOnEveryCell(t *testing.T) {
	th := style.DefaultLightTheme()
	th.AppBackground = style.Highlight{Bg: lipgloss.Color("#FAF0C8")}
	style.SetStyleSet(style.NewStyleSet(th))

	options := DefaultOptions()
	options.WhetherDisplayTitle = true
	options.DualColumn = true
	options.MainMenu = &mockMenu{key: "main", items: []MenuItem{
		{Title: "每日推荐歌曲"}, {Title: "每日推荐歌单"},
		{Title: "我的歌单"}, {Title: "我的收藏"},
		{Title: "私人FM"}, {Title: "专辑列表"},
		{Title: "搜索"}, {Title: "排行榜"},
		{Title: "精选歌单"}, {Title: "热门歌手"},
	}}
	options.MainMenuTitle = &MenuItem{Title: "网易云音乐", Subtitle: "[anhoder]"}

	const w, h = 100, 30
	app := NewApp(options)
	app.windowWidth = w
	app.windowHeight = h
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: w, Height: h}, app)
	screen := renderMainToScreen(main.View(app), w, h)

	isTransparent := func(c color.Color) bool {
		if c == nil {
			return true
		}
		_, ok := c.(lipgloss.NoColor)
		return ok
	}

	for y := range screen.Lines {
		for x := range screen.Lines[y] {
			cell := screen.CellAt(x, y)
			if cell == nil {
				continue
			}
			// Wide-glyph continuation cell: painted by the preceding glyph.
			if cell.Width == 0 {
				continue
			}
			if isTransparent(cell.Style.Bg) {
				t.Fatalf("cell (%d,%d) content=%q has transparent background; "+
					"cover image would bleed through", x, y, cell.Content)
			}
		}
	}
}

// TestMainViewPaintsAppBackgroundOnPartialDualColumnPage covers the last page
// of an odd-sized dual-column menu. Its empty right column and remaining page
// rows must stay opaque so the terminal or Kitty cover cannot bleed through.
func TestMainViewPaintsAppBackgroundOnPartialDualColumnPage(t *testing.T) {
	previousStyles := style.CurrentStyleSet()
	t.Cleanup(func() { style.SetStyleSet(previousStyles) })

	theme := style.DefaultLightTheme()
	theme.AppBackground = style.Highlight{Bg: lipgloss.Color("#FAF0C8")}
	style.SetStyleSet(style.NewStyleSet(theme))

	items := make([]MenuItem, 11)
	for i := range items {
		items[i] = MenuItem{Title: "item"}
	}
	options := DefaultOptions()
	options.WhetherDisplayTitle = false
	options.DualColumn = true
	options.MainMenu = &mockMenu{key: "main", items: items}
	options.MainMenuTitle = &MenuItem{Title: "Main"}

	const w, h = 100, 30
	app := NewApp(options)
	app.windowWidth = w
	app.windowHeight = h
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: w, Height: h}, app)
	main.menuCurPage = 2
	main.selectedIndex = 10

	screen := renderMainToScreen(main.View(app), w, h)
	for y := main.menuListStartRow; y < main.menuBottomRow; y++ {
		for x := range w {
			cell := screen.CellAt(x, y)
			if cell == nil || cell.Width == 0 {
				continue
			}
			if isCellTransparent(cell.Style.Bg) {
				t.Fatalf("partial-page cell (%d,%d) content=%q has transparent background", x, y, cell.Content)
			}
		}
	}
}

// isCellTransparent reports whether a screen cell's background is unset or the
// NoColor sentinel — i.e. a hole through which content drawn beneath the TUI
// (e.g. a cover image) would bleed.
func isCellTransparent(c color.Color) bool {
	if c == nil {
		return true
	}
	_, ok := c.(lipgloss.NoColor)
	return ok
}

// TestStatusBarAndBackButtonPaintAppBackground verifies that with a top status
// bar and inside a submenu (back button shown), every cell of the status-bar
// row and the menu-title row carries a non-transparent background. This guards
// the two regressions where the status-bar filler between the breadcrumb and
// the time, and the space between the back button and the menu title, stayed
// transparent and let the cover image bleed through.
func TestStatusBarAndBackButtonPaintAppBackground(t *testing.T) {
	th := style.DefaultLightTheme()
	th.AppBackground = style.Highlight{Bg: lipgloss.Color("#FAF0C8")}
	style.SetStyleSet(style.NewStyleSet(th))
	t.Cleanup(func() { style.SetStyleSet(style.NewStyleSet(style.DefaultDarkTheme())) })

	options := DefaultOptions()
	options.WhetherDisplayTitle = true
	options.DualColumn = true
	options.StatusBar = &DefaultStatusBar{}
	options.StatusBarPosition = StatusBarTop
	options.MainMenu = &mockMenu{key: "main", items: []MenuItem{
		{Title: "每日推荐歌曲"}, {Title: "每日推荐歌单"},
		{Title: "我的歌单"}, {Title: "我的收藏"},
	}}
	options.MainMenuTitle = &MenuItem{Title: "网易云音乐", Subtitle: "[anhoder]"}

	const w, h = 100, 30
	app := NewApp(options)
	app.windowWidth = w
	app.windowHeight = h
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: w, Height: h}, app)

	// Enter a submenu so the back button (and its trailing separator) render.
	main.menuStack.Push(&menuStackItem{menuTitle: &MenuItem{Title: "网易云音乐"}})
	main.menuTitle = &MenuItem{Title: "每日推荐歌曲"}

	screen := renderMainToScreen(main.View(app), w, h)

	// Row 0 is the status bar; menuTitleY() is the row carrying the back button.
	rows := []int{0, main.menuTitleY()}
	for _, y := range rows {
		if y < 0 || y >= len(screen.Lines) {
			t.Fatalf("row %d out of range (screen has %d lines)", y, len(screen.Lines))
		}
		for x := range screen.Lines[y] {
			cell := screen.CellAt(x, y)
			if cell == nil || cell.Width == 0 {
				continue
			}
			if isCellTransparent(cell.Style.Bg) {
				t.Fatalf("cell (%d,%d) content=%q has transparent background; "+
					"cover image would bleed through", x, y, cell.Content)
			}
		}
	}

	// Sanity: the back button must actually be present on the title row, else
	// the test would vacuously pass without exercising the separator.
	titleRow := main.menuTitleY()
	foundBack := false
	for x := range screen.Lines[titleRow] {
		if cell := screen.CellAt(x, titleRow); cell != nil && cell.Content == "←" {
			foundBack = true
			break
		}
	}
	if !foundBack {
		t.Fatalf("back button not found on title row %d; test would not exercise the separator", titleRow)
	}
}
