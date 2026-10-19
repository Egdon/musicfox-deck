package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

// AppCustomTheme defines the custom styles with full Highlight fields
// that support presets, foreground, background, and attributes.
type AppCustomTheme struct {
	BannerColor    style.Highlight
	HighlightColor style.Highlight
	AppExtra       style.Highlight
}

// AppCustomStyles is the resolved version of AppCustomTheme, where all
// Highlight fields have been resolved through the preset pipeline and
// converted to lipgloss.Style. Use style.CustomStyles[AppCustomStyles]()
// to obtain a typed instance from the global StyleSet.
type AppCustomStyles struct {
	BannerColor    lipgloss.Style
	HighlightColor lipgloss.Style
	AppExtra       lipgloss.Style
}

// ThemeController handles the 't' key by showing a popup with theme info.
// It demonstrates accessing custom styles via style.CurrentStyleSet().Custom.
type ThemeController struct{}

func mustPopup(spec model.PopupSpec) *model.Popup {
	popup, err := model.NewPopup(spec)
	if err != nil {
		panic(err)
	}
	return popup
}

func (c *ThemeController) KeyMsgHandle(msg tea.KeyMsg, a *model.App) (bool, model.Page, tea.Cmd) {
	k := msg.String()
	if k != "t" && k != "T" {
		return false, nil, nil
	}

	styles := style.CurrentStyleSet()

	// Access custom domain styles via type-safe struct conversion.
	custom := style.CustomStyles[AppCustomStyles](styles)
	bannerStyle := custom.BannerColor
	highlightStyle := custom.HighlightColor

	// Use bannerColor to style the popup title.
	popupTitle := bannerStyle.Bold(true).Render("Theme Colors (press T to switch)")
	now := time.Now().Format("15:04:05")

	body := lipgloss.JoinVertical(lipgloss.Left,
		popupTitle,
		"",
		"Custom domain colors in action:",
		"",
		bannerStyle.Render("  ▶ Banner color text"),
		highlightStyle.Render("  ▶ Highlight color text"),
		"",
		styles.MenuTitle.Render("  ▶ Primary (menu title)"),
		styles.Subtitle.Render("  ▶ Secondary (subtitle)"),
		"",
		fmt.Sprintf("  Time: %s", now),
	)

	popup := mustPopup(model.PopupSpec{
		Content: body,
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK"},
		},
	})
	a.ShowPopup(popup)

	return true, nil, a.RerenderCmd(true)
}

// MainMenu is the main menu with two items.
type MainMenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		items: []model.MenuItem{
			{Title: "Theme Info", Subtitle: "Press 't' to view theme details"},
			{Title: "Quit", Subtitle: "Press 'q' or Ctrl+C to quit"},
		},
	}
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	return m.items
}

func (m *MainMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

// Action demonstrates custom actions on menu items without navigating to a submenu.
// When the first item is activated (Enter/double-click), a popup is shown.
// Return (nil, nil) to fall through to normal SubMenu behavior.
func (m *MainMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	if index == 0 {
		// Show a popup as a custom action
		body := style.CurrentStyleSet().MenuTitle.Render("Custom Action Demo")
		body += "\n\nThis popup was triggered by a menu item action\n"
		body += fmt.Sprintf("\nSelected index: %d", index)
		body += "\n\nUse this for popups, logging, API calls, or any\narbitrary logic that shouldn't navigate to a submenu."

		popup := mustPopup(model.PopupSpec{
			Title:   "Action Fired",
			Content: body,
			Actions: []model.PopupAction{
				{ID: "got-it", Label: "Got It"},
			},
		})
		app.ShowPopup(popup)
		return nil, app.RerenderCmd(true)
	}
	// For other items (index 1 = Quit), fall through to SubMenu
	return nil, nil
}

func main() {
	// ── Step 1: Auto-adaptive default theme ──
	// DefaultTheme() auto-detects terminal background (dark/light).
	// Use DefaultDarkTheme() or DefaultLightTheme() for explicit choice.
	isDark := style.HasDarkBackground()
	mode := "dark"
	if !isDark {
		mode = "light"
	}
	fmt.Printf("Detected terminal background: %s mode\n", mode)

	// Build theme list for runtime switching. Press 'T' to cycle.
	darkTheme := style.DefaultDarkTheme()
	darkTheme.HighlightPresets = map[string]style.Highlight{
		"accent": {Fg: lipgloss.Color("#FF5F87"), Bold: style.BoolPtr(true)},
	}
	darkTheme.Custom = AppCustomTheme{
		BannerColor:    style.Highlight{Fg: lipgloss.Color("#FF5F87")},
		HighlightColor: style.Highlight{Fg: lipgloss.Color("#00FF00")},
		AppExtra:       style.Highlight{Fg: lipgloss.Color("#FFA500")},
	}

	lightTheme := style.DefaultLightTheme()
	lightTheme.Custom = AppCustomTheme{
		BannerColor:    style.Highlight{Fg: lipgloss.Color("#1565C0")},
		HighlightColor: style.Highlight{Fg: lipgloss.Color("#2E7D32")},
		AppExtra:       style.Highlight{Fg: lipgloss.Color("#E65100")},
	}

	vscodeTheme := style.VSCodeDarkTheme()
	vscodeTheme.Custom = AppCustomTheme{
		BannerColor:    style.Highlight{Fg: lipgloss.Color("#569CD6")},
		HighlightColor: style.Highlight{Fg: lipgloss.Color("#6A9955")},
		AppExtra:       style.Highlight{Fg: lipgloss.Color("#DCDCAA")},
	}

	// ── Step 6: Build the app with the custom theme ──
	opts := model.DefaultOptions()
	opts.EnableStartup = false
	opts.WhetherDisplayTitle = true
	opts.DualColumn = false
	opts.AppName = "Theme Demo"
	opts.KBControllers = []model.KeyboardController{&ThemeController{}}

	app := model.NewApp(opts)
	app.With(
		model.WithThemeList(darkTheme, lightTheme, vscodeTheme),
		model.WithThemeSwitchKey("T"),
		model.WithMainMenu(NewMainMenu(), &model.MenuItem{Title: "Theme Demo"}),
	)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
