package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

type markerComponent struct{}

func (markerComponent) Update(tea.Msg, *App) {}

func (markerComponent) View(*App, *Main) (string, int) {
	return "component-marker", 1
}

func TestMainViewKeepsComponentRowOnPartialLastPage(t *testing.T) {
	items := make([]MenuItem, 11)
	for i := range items {
		items[i] = MenuItem{Title: "item"}
	}

	options := DefaultOptions()
	options.WhetherDisplayTitle = false
	options.DualColumn = false
	options.MainMenu = &mockMenu{key: "main", items: items}
	options.MainMenuTitle = &MenuItem{Title: "Main"}
	options.Components = []Component{markerComponent{}}

	app := NewApp(options)
	app.windowWidth = 80
	app.windowHeight = 30
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: 80, Height: 30}, app)

	firstPageMarkerRow := rowContaining(t, main.View(app), "component-marker")

	main.menuCurPage = 2
	main.selectedIndex = 10
	lastPageMarkerRow := rowContaining(t, main.View(app), "component-marker")

	if lastPageMarkerRow != firstPageMarkerRow {
		t.Fatalf("component row changed on partial last page: first page row = %d, last page row = %d", firstPageMarkerRow, lastPageMarkerRow)
	}
}

func rowContaining(t *testing.T, view, marker string) int {
	t.Helper()
	for row, line := range strings.Split(ansi.Strip(view), "\n") {
		if strings.Contains(line, marker) {
			return row
		}
	}
	t.Fatalf("view does not contain %q", marker)
	return -1
}

func TestMainMenuWhitespaceDoesNotTriggerHoverOrLeftClick(t *testing.T) {
	app, main := newMainForMenuMouseTest(t)
	menuRow := main.menuListStartRow
	blankX := app.WindowWidth() - 1

	main.mouseMotionHandle(tea.Mouse{X: blankX, Y: menuRow}, app)
	if main.hoveredMenuItemIdx != -1 {
		t.Fatalf("hovered menu item = %d, want none for right-side whitespace", main.hoveredMenuItemIdx)
	}

	main.selectedIndex = 1
	main.mouseClickHandle(tea.Mouse{X: blankX, Y: menuRow, Button: tea.MouseLeft}, app)
	if main.selectedIndex != 1 {
		t.Fatalf("left click changed selection to %d, want 1", main.selectedIndex)
	}

	main.mouseMotionHandle(tea.Mouse{X: main.menuStartColumn, Y: menuRow}, app)
	if main.hoveredMenuItemIdx != 0 {
		t.Fatalf("hovered menu item = %d, want 0 for visible menu text", main.hoveredMenuItemIdx)
	}
}

func TestMainMenuWhitespaceCanOpenCustomContextMenu(t *testing.T) {
	app, main := newMainForMenuMouseTest(t)
	main.mouseClickHandle(tea.Mouse{X: app.WindowWidth() - 1, Y: main.menuListStartRow, Button: tea.MouseRight}, app)

	if !app.HasPopup() {
		t.Fatal("right click on whitespace did not open the menu-provided context menu")
	}
	contextMenu, ok := app.modalStack[len(app.modalStack)-1].(*ContextMenu)
	if !ok {
		t.Fatalf("top modal = %T, want *ContextMenu", app.modalStack[len(app.modalStack)-1])
	}
	if contextMenu.itemIndex != -1 {
		t.Fatalf("context menu item index = %d, want -1 for whitespace", contextMenu.itemIndex)
	}
}

func newMainForMenuMouseTest(t *testing.T) (*App, *Main) {
	t.Helper()
	options := DefaultOptions()
	options.DualColumn = false
	options.MainMenu = &testMenu{items: []MenuItem{
		{Title: "Alpha", Subtitle: "first item"},
		{Title: "Beta", Subtitle: "second item"},
	}}

	app := NewApp(options)
	app.windowWidth = 80
	app.windowHeight = 30
	main := NewMain(app, options)
	app.main = main
	_, _ = main.Update(tea.WindowSizeMsg{Width: app.windowWidth, Height: app.windowHeight}, app)
	return app, main
}

type blankRightClickController struct {
	calls int
}

func (c *blankRightClickController) MouseMsgHandle(msg tea.MouseMsg, _ *App) (bool, Page, tea.Cmd) {
	if msg.Mouse().Button != tea.MouseRight {
		return false, nil, nil
	}
	c.calls++
	return true, nil, nil
}

func TestMainMouseControllerReceivesBlankRightClick(t *testing.T) {
	app, main := newMainForMenuMouseTest(t)
	controller := &blankRightClickController{}
	main.mouseCtrls = []MouseController{controller}

	main.mouseMsgHandle(tea.MouseClickMsg(tea.Mouse{X: 0, Y: 0, Button: tea.MouseRight}), app)
	if controller.calls != 1 {
		t.Fatalf("blank right-click controller calls = %d, want 1", controller.calls)
	}
	if app.HasPopup() {
		t.Fatal("blank right click must be handled by the controller, not open a built-in menu")
	}
}

type actionLoadingMenu struct {
	DefaultMenu
	actionCalls         int
	submenuCalls        int
	loadingDuringAction string
	actionCmd           tea.Cmd
}

func (m *actionLoadingMenu) GetMenuKey() string {
	return "action_loading"
}

func (m *actionLoadingMenu) MenuViews() []MenuItem {
	return []MenuItem{{Title: "Run action"}}
}

func (m *actionLoadingMenu) Action(app *App, _ int) (Page, tea.Cmd) {
	m.actionCalls++
	m.loadingDuringAction = app.Main().loadingTips
	return nil, m.actionCmd
}

func (m *actionLoadingMenu) SubMenu(_ *App, _ int) Menu {
	m.submenuCalls++
	return nil
}

func TestMenuActionShowsLoadingUntilInvocationCompletes(t *testing.T) {
	tests := []struct {
		name    string
		trigger func(*Main, *App) (Page, tea.Cmd)
	}{
		{
			name: "keyboard",
			trigger: func(main *Main, app *App) (Page, tea.Cmd) {
				return main.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}), app)
			},
		},
		{
			name: "mouse double click",
			trigger: func(main *Main, app *App) (Page, tea.Cmd) {
				click := tea.MouseClickMsg(tea.Mouse{
					X:      main.menuStartColumn,
					Y:      main.menuListStartRow,
					Button: tea.MouseLeft,
				})
				_, _ = main.Update(click, app)
				return main.Update(click, app)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdExecuted := false
			menu := &actionLoadingMenu{
				actionCmd: func() tea.Msg {
					cmdExecuted = true
					return nil
				},
			}
			options := DefaultOptions()
			options.DualColumn = false
			options.LoadingText = "[Action loading]"
			options.MainMenu = menu
			options.MainMenuTitle = &MenuItem{Title: "Actions"}

			app := NewApp(options)
			app.windowWidth = 80
			app.windowHeight = 30
			main := NewMain(app, options)
			app.main = main
			app.setPage(main)
			_, _ = main.Update(tea.WindowSizeMsg{Width: app.windowWidth, Height: app.windowHeight}, app)
			_ = main.View(app)

			page, triggerCmd := tt.trigger(main, app)
			if page != main {
				t.Fatalf("trigger page = %T, want current main page", page)
			}
			if triggerCmd == nil {
				t.Fatal("trigger did not schedule the deferred action")
			}
			if menu.actionCalls != 0 {
				t.Fatalf("Action called before loading rendered: calls = %d", menu.actionCalls)
			}
			if view := ansi.Strip(main.View(app)); !strings.Contains(view, options.LoadingText) {
				t.Fatalf("view before Action = %q, want loading text %q", view, options.LoadingText)
			}

			actionPage, actionCmd := main.Update(tickMainMsg{}, app)
			if actionPage != nil {
				t.Fatalf("Action page = %T, want nil", actionPage)
			}
			if menu.actionCalls != 1 {
				t.Fatalf("Action calls = %d, want 1", menu.actionCalls)
			}
			if menu.loadingDuringAction != options.LoadingText {
				t.Fatalf("loading text during Action = %q, want %q", menu.loadingDuringAction, options.LoadingText)
			}
			if main.loadingTips != "" {
				t.Fatalf("loading text after Action = %q, want cleared", main.loadingTips)
			}
			if view := ansi.Strip(main.View(app)); strings.Contains(view, options.LoadingText) {
				t.Fatalf("view after Action still contains loading text %q", options.LoadingText)
			}
			if menu.submenuCalls != 0 {
				t.Fatalf("submenu calls = %d, want 0 when Action returns a command", menu.submenuCalls)
			}
			if actionCmd == nil {
				t.Fatal("Action command was not returned")
			}
			_ = actionCmd()
			if !cmdExecuted {
				t.Fatal("returned Action command did not execute")
			}
		})
	}
}
