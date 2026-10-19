// Layout application example — demonstrates the layout package inside a
// foxful-cli framework application. Uses model.App, model.Page, model.Menu,
// and model.KeyboardController (no raw bubbletea).
//
// Keys:
//
//	o — popup: 2×3 grid built with JoinHorizontal + JoinVertical
//	p — navigate to the full-screen LayoutPage
//	c — popup: PlaceCenter centering demo
//	j — popup: JoinHorizontal + JoinVertical cards demo
//	esc / b — return to main menu (from LayoutPage)
//	q / ctrl+c — quit
package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

// ── tick message for re-rendering ──

type layoutTickMsg struct{}

// ── MainMenu ────────────────────────────────────────────────────────

var mainMenu = NewMainMenu()

func mustPopup(spec model.PopupSpec) *model.Popup {
	popup, err := model.NewPopup(spec)
	if err != nil {
		panic(err)
	}
	return popup
}

type MainMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewMainMenu() *MainMenu {
	m := &MainMenu{}
	m.menus = []model.MenuItem{
		{Title: "Demo 1: Join Grid", Subtitle: "Press 'o' — 2×3 grid via JoinHorizontal + JoinVertical"},
		{Title: "Demo 2: Layout Page", Subtitle: "Press 'p' — PlaceCenter + info cards"},
		{Title: "Demo 3: PlaceCenter", Subtitle: "Press 'c' — centered content in a bounding box"},
		{Title: "Demo 4: Join Cards", Subtitle: "Press 'j' — JoinHorizontal + JoinVertical composition"},
	}
	return m
}

func (m *MainMenu) IsSearchable() bool                     { return true }
func (m *MainMenu) GetMenuKey() string                     { return "layout_main_menu" }
func (m *MainMenu) MenuViews() []model.MenuItem            { return m.menus }
func (m *MainMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

// ── LayoutController ────────────────────────────────────────────────
// Implements model.KeyboardController to intercept key presses on the
// main menu page.

type LayoutController struct{}

func (c *LayoutController) KeyMsgHandle(msg tea.KeyMsg, a *model.App) (bool, model.Page, tea.Cmd) {
	switch msg.String() {
	case "o", "O":
		showGridPopup(a)
		return true, nil, nil
	case "p", "P":
		return true, NewLayoutPage(), nil
	case "c", "C":
		showPlaceCenterPopup(a)
		return true, nil, nil
	case "j", "J":
		showJoinPopup(a)
		return true, nil, nil
	}
	return false, nil, nil
}

// showGridPopup builds a 2×3 labeled grid inside a rounded-border box
// and displays it as a custom popup.
func showGridPopup(a *model.App) {
	styles := style.CurrentStyleSet()

	// Build a cell style with an accent-colored rounded border.
	cellStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.BrightCyan).
		Width(15).Height(3).
		Align(lipgloss.Center, lipgloss.Center).
		Padding(0, 1)

	// Six cells for a 2-row × 3-column grid.
	cells := []string{
		cellStyle.Render("Songs"),
		cellStyle.Render("Albums"),
		cellStyle.Render("Artists"),
		cellStyle.Render("Playlists"),
		cellStyle.Render("Genres"),
		cellStyle.Render("Podcasts"),
	}

	// Join 3 cells per row horizontally, then stack 2 rows vertically.
	row1 := layout.JoinHorizontal(layout.Top, cells[0], cells[1], cells[2])
	row2 := layout.JoinHorizontal(layout.Top, cells[3], cells[4], cells[5])
	grid := layout.JoinVertical(layout.Left, row1, row2)

	// Wrap the grid in a rounded-border box.
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.BrightCyan).
		Padding(1, 2).
		Render(grid)

	popup := mustPopup(model.PopupSpec{
		Title:   "Layout Demo",
		Content: box,
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK"},
			{ID: "cancel", Label: "Cancel", IsCancel: true},
		},
		OnResult: func(model.PopupResult) {
			_ = styles.Muted.Render("popup dismissed")
		},
	})
	a.ShowPopup(popup)
}

// showPlaceCenterPopup demonstrates layout.PlaceCenter — centering content
// in a fixed-size bounding box.
func showPlaceCenterPopup(a *model.App) {
	styles := style.CurrentStyleSet()

	// Build centered logo content
	logo := lipgloss.NewStyle().
		Foreground(lipgloss.BrightGreen).
		Bold(true).
		Render("┌─────────────┐\n│ foxful-cli   │\n│ PlaceCenter  │\n└─────────────┘")

	// Center in a 30×8 box
	centered := layout.PlaceCenter(logo, 30, 8)

	// Wrap in a visible border to show the bounding box dimensions
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.BrightMagenta).
		Padding(1, 2).
		Render(centered)

	popup := mustPopup(model.PopupSpec{
		Title:   "PlaceCenter Demo",
		Content: box,
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK"},
		},
		OnResult: func(model.PopupResult) {
			_ = styles.Muted.Render("")
		},
	})
	a.ShowPopup(popup)
}

// showJoinPopup demonstrates layout.JoinHorizontal + layout.JoinVertical
// by composing three info cards into a vertical stack.
func showJoinPopup(a *model.App) {
	styles := style.CurrentStyleSet()

	card := func(title, desc string) string {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.BrightBlue).
			Padding(1, 2).
			Width(40).
			Render(
				styles.MenuTitle.Render(title) + "\n" +
					styles.Subtitle.Render(desc),
			)
	}

	c1 := card("JoinHorizontal", "Side-by-side layout\nfor columns and toolbars")
	c2 := card("JoinVertical", "Stacking layout\nfor sections and lists")
	c3 := card("Place / PlaceCenter", "Centering in a bounding box\nfor splash screens and dialogs")

	stack := layout.JoinVertical(layout.Left, c1, c2, c3)

	popup := mustPopup(model.PopupSpec{
		Title:   "Join Demo",
		Content: stack,
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK"},
		},
		OnResult: func(model.PopupResult) {
			_ = styles.Muted.Render("")
		},
	})
	a.ShowPopup(popup)
}

// ── LayoutPage ──────────────────────────────────────────────────────
// A full-screen page that demonstrates layout.PlaceCenter,
// layout.JoinVertical, and layout.JoinHorizontal.

type LayoutPage struct {
	app *model.App
}

func NewLayoutPage() *LayoutPage {
	return &LayoutPage{}
}

func (p *LayoutPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return false
}

func (p *LayoutPage) Type() model.PageType {
	return "layout"
}

func (p *LayoutPage) Msg() tea.Msg {
	return layoutTickMsg{}
}

func (p *LayoutPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	p.app = a

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b", "B":
			return a.Main(), nil
		}
	case layoutTickMsg:
		return p, a.RerenderCmd(true)
	}

	return p, nil
}

func (p *LayoutPage) View(a *model.App) string {
	w, h := a.WindowWidth(), a.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}

	styles := style.CurrentStyleSet()

	// ── 1. Logo area — centered with PlaceCenter ──

	logo := lipgloss.NewStyle().
		Bold(true).
		Padding(1, 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.BrightGreen).
		Align(lipgloss.Center).
		Render("foxful-cli\nLayout Demo")

	// Center the logo in the upper third of the window.
	logoArea := layout.PlaceCenter(logo, w, h/3)

	// ── 2. Info cards — two boxes side-by-side with JoinHorizontal ──

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.BrightBlue).
		Padding(1, 2).
		Width(30).Height(7).
		Align(lipgloss.Left, lipgloss.Top)

	leftCard := cardStyle.Render(
		styles.MenuTitle.Render("Layout Primitives") + "\n\n" +
			"  JoinVertical\n" +
			"  JoinHorizontal\n" +
			"  PlaceCenter",
	)

	rightCard := cardStyle.Render(
		styles.MenuTitle.Render("Key Bindings") + "\n\n" +
			"  o — Show popup\n" +
			"  p — This page\n" +
			"  esc / b — Go back",
	)

	infoRow := layout.JoinHorizontal(layout.Top, leftCard, "  ", rightCard)
	infoCards := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(infoRow)

	// ── 3. Footer hint bar — built with JoinHorizontal ──

	footer := layout.JoinHorizontal(layout.Top,
		styles.Muted.Render(" esc/b: Back to menu "),
		styles.Muted.Render(" | "),
		styles.Info.Render(fmt.Sprintf(" Window: %d×%d ", w, h)),
		styles.Muted.Render(" | "),
		styles.Subtitle.Render(fmt.Sprintf(" %s ", time.Now().Format("15:04:05"))),
	)

	footerBar := lipgloss.NewStyle().
		Width(w).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.BrightBlack).
		Padding(0, 1).
		Render(footer)

	// ── Compose the whole page vertically ──

	body := layout.JoinVertical(layout.Left,
		logoArea,
		infoCards,
		"",
		footerBar,
	)

	// Fill remaining height to prevent display artifacts.
	if lipgloss.Height(body) < h-1 {
		body = lipgloss.NewStyle().Height(h - 1).Render(body)
	}

	return body
}

// ── main ────────────────────────────────────────────────────────────

func main() {
	ops := model.DefaultOptions()
	ops.EnableStartup = false      // skip the 2-second startup animation
	ops.WhetherDisplayTitle = true // show the app name title bar
	ops.AppName = "Layout Demo"
	ops.KBControllers = []model.KeyboardController{&LayoutController{}}
	ops.Ticker = model.DefaultTicker(500 * time.Millisecond) // keeps the footer clock ticking

	app := model.NewApp(ops)
	app.With(model.WithMainMenu(mainMenu, nil))

	fmt.Println(app.Run())
}
