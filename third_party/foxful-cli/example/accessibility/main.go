package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

type displayOption struct {
	name       string
	accessible bool
	theme      style.Theme
}

var displayOptions = []displayOption{
	{name: "Default Mode", accessible: false, theme: style.DefaultTheme()},
	{name: "Accessible Mode", accessible: true, theme: style.DefaultTheme()},
	{name: "High Contrast Dark", accessible: true, theme: style.HighContrastDarkTheme()},
	{name: "High Contrast Light", accessible: true, theme: style.HighContrastLightTheme()},
}

type accessibilityMenu struct {
	model.DefaultMenu
	items   []model.MenuItem
	current int
}

func newAccessibilityMenu() *accessibilityMenu {
	return &accessibilityMenu{items: []model.MenuItem{
		{Title: "Default Mode", Subtitle: "standard theme"},
		{Title: "Accessible Mode", Subtitle: "reverse, bold, underline"},
		{Title: "High Contrast Dark", Subtitle: "white on black"},
		{Title: "High Contrast Light", Subtitle: "black on white"},
	}}
}

func (m *accessibilityMenu) GetMenuKey() string          { return "accessibility_modes" }
func (m *accessibilityMenu) MenuViews() []model.MenuItem { return m.items }
func (m *accessibilityMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}
func (m *accessibilityMenu) IsSearchable() bool { return false }
func (m *accessibilityMenu) HelpHints() []model.HelpHint {
	return []model.HelpHint{
		{Key: "up/down", Desc: "choose display mode"},
		{Key: "enter", Desc: "apply mode"},
		{Key: "q", Desc: "quit"},
	}
}

func (m *accessibilityMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	if index >= 0 && index < len(displayOptions) {
		m.current = index
		applyDisplayOption(app, displayOptions[index])
	}
	return nil, app.RerenderCmd(true)
}

func (m *accessibilityMenu) currentOption() displayOption {
	return displayOptions[m.current]
}

type accessibilityComponent struct {
	menu *accessibilityMenu
}

func (c *accessibilityComponent) Update(_ tea.Msg, _ *model.App) {}

func (c *accessibilityComponent) View(app *model.App, _ *model.Main) (string, int) {
	styles := app.StyleSet()
	option := c.menu.currentOption()
	mode := "Off"
	if option.accessible {
		mode = "On: reverse, bold, underline"
	}

	content := strings.Join([]string{
		styles.MenuTitle.Render("Accessible mode: ") + mode,
		styles.MenuTitle.Render("Theme: ") + option.name,
		"",
		styles.Border.Render(strings.Join([]string{
			styles.MenuTitle.Render("Interactive state comparison"),
			"",
			styles.SelectedItem.Render("  > Focused menu item"),
			styles.MenuItemHover.Render("    Hoverable menu item"),
			"",
			styles.Button.Render(" Save ") + "  " + styles.ButtonBlurred.Render(" Cancel "),
			"",
			styles.Popup.ActionFocused.Render(" Confirm ") + "  " + styles.Popup.ActionHover.Render(" More options "),
		}, "\n")),
		"",
		styles.Info.Render("Choose a display mode above to update this demonstration."),
	}, "\n")
	return content, 12
}

func applyDisplayOption(app *model.App, option displayOption) {
	style.SetAccessibleMode(option.accessible)
	styleSet := style.NewStyleSet(option.theme)
	style.SetStyleSet(styleSet)
	app.SetStyleSet(styleSet)
}

func main() {
	style.SetAccessibleMode(false)
	style.SetStyleSet(style.NewStyleSet(style.DefaultTheme()))

	menu := newAccessibilityMenu()
	options := model.DefaultOptions()
	options.AppName = "Accessibility Demo"
	options.MainMenuTitle = &model.MenuItem{Title: "Choose a display mode"}
	options.MainMenu = menu
	options.Components = []model.Component{&accessibilityComponent{menu: menu}}
	options.BottomHeight = 16

	app := model.NewApp(options)
	applyDisplayOption(app, displayOptions[0])
	fmt.Println(app.Run())
}
