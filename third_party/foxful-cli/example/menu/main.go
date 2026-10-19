package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

var (
	mainMenu      = NewMainMenu()
	secondaryMenu = NewSecondaryMenu()
)

type MainMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewMainMenu() *MainMenu {
	m := &MainMenu{}
	m.menus = []model.MenuItem{
		{Title: "Title 1", Subtitle: "subtitle"},
		{Title: "Title 2", Subtitle: "very long long long long long long long long long long long subtitle"},
		{Title: "Title 3"},
		{Title: "Title 4"},
		{Title: "Title 5"},
		{Title: "Title 6"},
		{Title: "Title 7"},
		{Title: "Title 8"},
		{Title: "Title 9"},
		{Title: "Title 10"},
		{Title: "Title 11"},
		{Title: "Title 12"},
		{Title: "Title 13"},
	}

	return m
}

func (m *MainMenu) IsSearchable() bool {
	return true
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menus) {
		return nil
	}

	return secondaryMenu
}

func (m *MainMenu) ContextMenuItems(_ *model.App, index int) []model.ContextMenuItem {
	if index >= len(m.menus) {
		return nil
	}
	items := []model.ContextMenuItem{
		{ID: "play", Label: "󰐊  Play"},
		{ID: "queue", Label: "󰆴  Add to Queue"},
		{ID: "play_next", Label: "󰒭  Play Next"},
		{Separator: true},
		{ID: "album", Label: "󰀥  Go to Album"},
		{ID: "artist", Label: "󰠃  Go to Artist"},
		{ID: "details", Label: "󰋼  View Track Details"},
		{Separator: true},
		{ID: "delete", Label: "󰅖  Delete"},
	}
	// Keep this label deliberately long so the configured maximum width
	// demonstrates ellipsis truncation. Disable it for the first item as a demo.
	favorite := model.ContextMenuItem{ID: "favorite", Label: "󰓡  Add to a Very Long Favorites Collection"}
	if index == 0 {
		favorite.Disabled = true
	}
	items = append(items, favorite)
	return items
}

func (m *MainMenu) ContextMenuAction(app *model.App, index int, item model.ContextMenuItem) (model.Page, tea.Cmd) {
	// Show a popup confirming the action
	menuItem := m.menus[index]
	popup, _ := model.NewPopup(model.PopupSpec{
		Title:   "Context Action",
		Content: fmt.Sprintf("Action '%s' on '%s'", item.Label, menuItem.Title),
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK", IsCancel: true},
		},
	})
	app.ShowPopup(popup)
	return nil, app.RerenderCmd(true)
}

type SecondaryMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewSecondaryMenu() *SecondaryMenu {
	m := &SecondaryMenu{}
	m.menus = []model.MenuItem{
		{Title: "Submenu 1"},
		{Title: "Submenu 2"},
	}

	return m
}

func (m *SecondaryMenu) GetMenuKey() string {
	return "secondary_menu"
}

func (m *SecondaryMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SecondaryMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *SecondaryMenu) BeforeEnterMenuHook() model.Hook {
	return func(_ *model.Main) (bool, model.Page) {
		// mock request
		time.Sleep(time.Millisecond * 200)
		return true, nil
	}
}

func main() {
	opts := model.DefaultOptions()
	opts.StatusBar = &model.DefaultStatusBar{}
	opts.ContextMenuOptions = model.ContextMenuOptions{MaxWidth: 24, MaxHeight: 7}
	// opts.DynamicRowCount = true
	app := model.NewApp(opts)
	app.With(model.WithMainMenu(mainMenu, nil))

	fmt.Println(app.Run())
}
