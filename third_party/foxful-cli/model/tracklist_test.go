package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

func newTracklistMain(t *testing.T, w, h int) (*App, *Main) {
	t.Helper()
	prev := style.CurrentStyleSet()
	t.Cleanup(func() { style.SetStyleSet(prev) })
	style.SetStyleSet(style.NewStyleSet(warmNightTestTheme()))
	songs := make([]MenuItem, len(goldenSongs))
	for i, s := range goldenSongs {
		s.Meta = "3:21"
		songs[i] = s
	}
	options := DefaultOptions()
	options.MainMenu = &mockMenu{key: "songs", items: songs}
	options.MainMenuTitle = &MenuItem{Title: "飙升榜", Subtitle: "[更新27首]"}
	options.ListStyle = ListTracklist
	options.DynamicRowCount = true
	app := NewApp(options)
	app.windowWidth, app.windowHeight = w, h
	main := NewMain(app, options)
	app.main = main
	main.Update(tea.WindowSizeMsg{Width: w, Height: h}, app)
	return app, main
}

func TestTracklistRowsAreExact(t *testing.T) {
	for _, w := range []int{40, 80, 120, 200} {
		app, main := newTracklistMain(t, w, 30)
		main.selectedIndex = 2
		main.hoveredMenuItemIdx = 4
		for i, line := range strings.Split(main.View(app), "\n") {
			if got := ansix.Width(line); got != w {
				t.Fatalf("width %d: line %d is %d cells: %q", w, i, got, ansi.Strip(line))
			}
		}
		rows := main.tracklistView(len(goldenSongs))
		cols := main.trackColumns(main.getCurPageMenus())
		for i, row := range rows {
			if got := ansix.Width(row); got != main.ContentWidth() {
				t.Fatalf("width %d: row %d is %d cells", w, i, got)
			}
		}
		if cols.start != trackMarginLeft || cols.end != w-trackMarginRight {
			t.Fatalf("width %d: row spans %d..%d", w, cols.start, cols.end)
		}
	}
}

func TestTracklistIsOneClickableColumn(t *testing.T) {
	_, main := newTracklistMain(t, 120, 30)
	if main.isDualColumn {
		t.Fatal("the tracklist must be a single column")
	}
	y := main.menuListStartRow + 3
	for _, x := range []int{trackMarginLeft, 40, 117} {
		if got := main.menuItemAt(x, y); got != 3 {
			t.Fatalf("click at %d hits %d, want 3", x, got)
		}
	}
	if got := main.menuItemAt(119, y); got != -1 {
		t.Fatalf("click in the right margin hits %d", got)
	}
}

// While a menu loads, the header already names it and shows the loading
// tips once, on its right.
func TestTracklistHeaderWhileEntering(t *testing.T) {
	app, main := newTracklistMain(t, 100, 30)
	app.page = main
	main.selectedIndex = 3
	main.enterMenuWithLoading(&mockMenu{key: "sub", items: []MenuItem{{Title: "Inner"}}}, &MenuItem{Title: "排行榜"})
	header := ""
	for _, line := range strings.Split(ansi.Strip(main.View(app)), "\n") {
		if strings.Contains(line, " ── ") && !strings.Contains(line, "foxful-cli") {
			header = line
			break
		}
	}
	if !strings.Contains(header, "排行榜") || strings.Contains(header, "飙升榜") || strings.Count(header, main.loadingTips) != 1 {
		t.Fatalf("header while entering: %q (tips %q)", header, main.loadingTips)
	}
}
