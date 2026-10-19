package main

import (
	"fmt"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/robotn/gohook"
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
	var app = model.NewApp(model.DefaultOptions())
	var m = map[string]model.GlobalKeyHandler{
		"ctrl+shift+c": func(event hook.Event) model.Page {
			app.MustMain().EnterMenu(nil, nil)
			return nil
		},
	}
	app.With(model.WithGlobalKeyHandlers(m), model.WithMainMenu(mainMenu, nil))

	fmt.Println(app.Run())
}
