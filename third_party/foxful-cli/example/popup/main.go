// Popup example — demonstrates modal dialog usage in foxful-cli.
//
// Press Enter or double-click a menu item to trigger the corresponding popup.
// ESC dismisses the topmost popup.
//
// Mouse support is enabled. When a popup is active, mouse events are
// intercepted to prevent accidental clicks on the background menu.
package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

var mainMenu = NewMainMenu()

func mustPopup(spec model.PopupSpec) *model.Popup {
	popup, err := model.NewPopup(spec)
	if err != nil {
		panic(err)
	}
	return popup
}

func showAnchoredPopup(a *model.App, anchor model.PopupAnchor, offsetX, offsetY int) {
	var anchorName string
	switch anchor {
	case model.AnchorTopLeft:
		anchorName = "Top-Left"
	case model.AnchorTopCenter:
		anchorName = "Top-Center"
	case model.AnchorTopRight:
		anchorName = "Top-Right"
	case model.AnchorBottomLeft:
		anchorName = "Bottom-Left"
	case model.AnchorBottomCenter:
		anchorName = "Bottom-Center"
	case model.AnchorBottomRight:
		anchorName = "Bottom-Right"
	case model.AnchorCustom:
		anchorName = "Custom"
	default:
		anchorName = "Center"
	}

	msg := fmt.Sprintf("This popup is anchored at %s.\nOffset: (%d, %d)\n\nPress ESC to dismiss.", anchorName, offsetX, offsetY)
	popup := mustPopup(model.PopupSpec{
		Title:   fmt.Sprintf("Anchor: %s", anchorName),
		Content: msg,
		Actions: []model.PopupAction{
			{ID: "confirm", Label: "Confirm"},
			{ID: "cancel", Label: "Cancel", IsCancel: true},
		},
		Anchor:  anchor,
		OffsetX: offsetX,
		OffsetY: offsetY,
		OnResult: func(model.PopupResult) {
			_ = style.Dim("dismissed")
		},
	})
	a.ShowPopup(popup)
}

func stackTwoPopups(a *model.App) {
	// Track whether bottom popup has been dismissed
	// First popup: bottom layer
	bottom := mustPopup(model.PopupSpec{
		Title:   "Bottom Popup",
		Content: "Bottom popup — press ESC to dismiss me first",
		Actions: []model.PopupAction{
			{ID: "confirm", Label: "Confirm"},
			{ID: "cancel", Label: "Cancel", IsCancel: true},
		},
		Anchor: model.AnchorBottomCenter,
		OnResult: func(model.PopupResult) {
			_ = style.Dim("bottom dismissed")
		},
	})
	a.ShowPopup(bottom)

	top := mustPopup(model.PopupSpec{
		Title:   "Top Popup",
		Content: "Top popup — I block the one below!\n\nPress ESC to dismiss me first, then dismiss the bottom one.",
		Actions: []model.PopupAction{
			{ID: "confirm", Label: "Confirm"},
			{ID: "cancel", Label: "Cancel", IsCancel: true},
		},
		OnResult: func(model.PopupResult) {
			_ = style.Dim("top dismissed")
		},
	})
	a.ShowPopup(top)
}

func showConfirmPopup(a *model.App) {
	body := "Press ENTER to confirm, ESC to cancel.\n\nThis is a simple popup test."
	popup := mustPopup(model.PopupSpec{
		Title:   "Test Popup",
		Content: body,
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK"},
			{ID: "cancel", Label: "Cancel", IsCancel: true},
		},
		OnResult: func(model.PopupResult) {
			_ = style.Dim("dismissed")
		},
	})
	a.ShowPopup(popup)
}

type MainMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewMainMenu() *MainMenu {
	m := &MainMenu{}
	m.menus = []model.MenuItem{
		{Title: "Toggle centered popup"},
		{Title: "Top-left popup"},
		{Title: "Bottom-center popup"},
		{Title: "Top-right popup"},
		{Title: "Custom position popup"},
		{Title: "Stack two popups"},
		{Title: "Dismiss top popup"},
	}
	return m
}

func (m *MainMenu) IsSearchable() bool                     { return true }
func (m *MainMenu) GetMenuKey() string                     { return "main" }
func (m *MainMenu) MenuViews() []model.MenuItem            { return m.menus }
func (m *MainMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

// Action triggers the corresponding popup when a menu item is activated
// via Enter or double-click, instead of requiring a keyboard controller.
func (m *MainMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	switch index {
	case 0:
		if app.HasPopup() {
			app.DismissPopup()
		} else {
			showConfirmPopup(app)
		}
	case 1:
		showAnchoredPopup(app, model.AnchorTopLeft, 0, 0)
	case 2:
		showAnchoredPopup(app, model.AnchorBottomCenter, 0, 0)
	case 3:
		showAnchoredPopup(app, model.AnchorTopRight, 0, 0)
	case 4:
		showAnchoredPopup(app, model.AnchorCustom, 15, 8)
	case 5:
		stackTwoPopups(app)
	case 6:
		app.DismissPopup()
	}
	return nil, app.RerenderCmd(true)
}

func main() {
	ops := model.DefaultOptions()
	ops.AppName = "Popup Test"
	ops.MainMenu = mainMenu
	ops.Ticker = model.DefaultTicker(300 * time.Millisecond)

	app := model.NewApp(ops)

	// Show popup automatically after 1 second.
	// Rerender(true) is called right after ShowPopup to force an immediate
	// view cycle so popup bounds are set for mouse hit-testing.
	// go func() {
	// 	time.Sleep(1 * time.Second)
	// 	showConfirmPopup(app)
	// 	app.Rerender(true)
	// }()

	fmt.Println(app.Run())
}
