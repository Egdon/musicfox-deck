package model

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

type testMenu struct {
	DefaultMenu
	items []MenuItem
}

func (m *testMenu) GetMenuKey() string {
	return "test"
}

func (m *testMenu) MenuViews() []MenuItem {
	return m.items
}

func (m *testMenu) ContextMenuItems(_ *App, _ int) []ContextMenuItem {
	return []ContextMenuItem{
		{ID: "action1", Label: "Action 1"},
		{ID: "action2", Label: "Action 2"},
	}
}

func (m *testMenu) ContextMenuAction(_ *App, _ int, item ContextMenuItem) (Page, tea.Cmd) {
	// Return nil to indicate no navigation
	return nil, nil
}

func TestContextMenuComputePositionFlipsRight(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
		{ID: "b", Label: "Action B"},
	}
	cm := NewContextMenu(menu, 0, items, 75, 10) // near right edge

	// Simulate a 20-wide menu at x=75 in an 80-wide terminal
	x, y := cm.computePosition(80, 24, 20, 5)

	// Should flip left: x = 75 - 20 = 55
	if x > 75 {
		t.Errorf("expected menu to flip left, got x=%d", x)
	}
	if y != 11 {
		t.Errorf("expected y=11, got y=%d", y)
	}
}

func TestContextMenuPositionLeavesClickedRowVisible(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{{ID: "a", Label: "Action A"}}
	cm := NewContextMenu(menu, 0, items, 10, 10)

	_, y := cm.computePosition(80, 24, 20, 5)
	if y != 11 {
		t.Errorf("menu y=%d, want 11 (one row below clicked row)", y)
	}
	if y <= cm.mouseY {
		t.Errorf("menu y=%d overlaps clicked row y=%d", y, cm.mouseY)
	}
}

func TestContextMenuComputePositionClampsToBottom(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
		{ID: "b", Label: "Action B"},
	}
	cm := NewContextMenu(menu, 0, items, 10, 22) // near bottom edge

	// A 5-tall menu in a 24-row terminal should align with the bottom edge.
	x, y := cm.computePosition(80, 24, 20, 5)
	if x != 10 {
		t.Errorf("expected x=10, got x=%d", x)
	}
	if y != 19 {
		t.Errorf("expected menu to clamp down to y=19, got y=%d", y)
	}
}

func TestContextMenuComputePositionClampsToScreen(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
	}
	cm := NewContextMenu(menu, 0, items, 100, 100) // far outside

	x, y := cm.computePosition(80, 24, 20, 5)

	// Should clamp to screen bounds
	if x < 0 || x+20 > 80 {
		t.Errorf("x=%d out of bounds for 80-wide screen with 20-wide menu", x)
	}
	if y < 0 || y+5 > 24 {
		t.Errorf("y=%d out of bounds for 24-tall screen with 5-tall menu", y)
	}
}

func TestContextMenuKeyboardNavigation(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
		{Separator: true},
		{ID: "b", Label: "Action B", Disabled: true},
		{ID: "c", Label: "Action C"},
	}
	cm := NewContextMenu(menu, 0, items, 10, 10)

	// Opening a context menu must not preselect an action.
	if cm.focused != -1 {
		t.Errorf("initial focused=%d, want -1", cm.focused)
	}

	// Down selects the first selectable item.
	cm.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if cm.focused != 0 {
		t.Errorf("after down, focused=%d, want 0", cm.focused)
	}

	// Down skips the separator and disabled item.
	cm.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if cm.focused != 3 {
		t.Errorf("after second down, focused=%d, want 3 (skipping separator and disabled)", cm.focused)
	}

	// Up returns to the first selectable item.
	cm.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if cm.focused != 0 {
		t.Errorf("after up, focused=%d, want 0", cm.focused)
	}
}

func TestContextMenuEnterSelectsItem(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
		{ID: "b", Label: "Action B"},
	}
	cm := NewContextMenu(menu, 0, items, 10, 10)

	// Focus on index 1
	cm.focused = 1

	// Press enter
	cm.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if !cm.dismissed() {
		t.Error("expected menu to be dismissed after enter")
	}
	if cm.selected == nil || cm.selected.ID != "b" {
		t.Errorf("expected selected item ID='b', got %v", cm.selected)
	}
}

func TestContextMenuEscapeDismisses(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
	}
	cm := NewContextMenu(menu, 0, items, 10, 10)

	cm.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))

	if !cm.dismissed() {
		t.Error("expected menu to be dismissed after escape")
	}
	if !cm.isCanceled {
		t.Error("expected menu to be canceled after escape")
	}
	if cm.selected != nil {
		t.Error("expected no selected item after escape")
	}
}

func TestContextMenuItemAtHitTest(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
		{ID: "b", Label: "Action B"},
	}
	cm := NewContextMenu(menu, 0, items, 10, 10)

	// Render to get bounds
	rendered := cm.renderModal(style.NewStyleSet(style.DefaultDarkTheme()), 80, 24)
	cm.setModalBounds(10, 10, 20, 4, rendered.itemBounds)

	// itemBounds[0] should be at (11, 11) with width=innerWidth (depends on label)
	// Test that clicking inside first item's bounds returns index 0
	idx := cm.itemAt(cm.itemBounds[0].x, cm.itemBounds[0].y)
	if idx != 0 {
		t.Errorf("itemAt(%d, %d) = %d, want 0", cm.itemBounds[0].x, cm.itemBounds[0].y, idx)
	}

	// Clicking outside any bounds should return -1
	idx = cm.itemAt(0, 0)
	if idx != -1 {
		t.Errorf("itemAt(0, 0) = %d, want -1", idx)
	}
}

func TestContextMenuMouseClickActivatesItem(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
		{ID: "b", Label: "Action B"},
	}
	cm := NewContextMenu(menu, 0, items, 10, 10)

	// Render and set bounds
	rendered := cm.renderModal(style.NewStyleSet(style.DefaultDarkTheme()), 80, 24)
	cm.setModalBounds(10, 10, 20, 4, rendered.itemBounds)

	// Simulate left-click on first item
	cm.hovered = 0 // Simulate hover state
	msg := tea.MouseClickMsg(tea.Mouse{
		X:      cm.itemBounds[0].x,
		Y:      cm.itemBounds[0].y,
		Button: tea.MouseLeft,
	})
	handled, _ := cm.handleMouse(msg)

	if !handled {
		t.Error("expected mouse click to be handled")
	}
	if !cm.dismissed() {
		t.Error("expected menu to be dismissed after click")
	}
	if cm.selected == nil || cm.selected.ID != "a" {
		t.Errorf("expected selected item ID='a', got %v", cm.selected)
	}
}

func TestContextMenuCompleteInvokesAction(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{
		{ID: "a", Label: "Action A"},
	}
	cm := NewContextMenu(menu, 0, items, 10, 10)
	cm.selected = &items[0]

	app := NewApp(DefaultOptions())
	page, cmd := cm.complete(app)

	// testMenu.ContextMenuAction returns (nil, nil)
	if page != nil {
		t.Errorf("expected page=nil, got %v", page)
	}
	if cmd != nil {
		t.Errorf("expected cmd=nil, got %v", cmd)
	}
}

func TestContextMenuEmptyItemsRendersMinWidth(t *testing.T) {
	menu := &testMenu{items: []MenuItem{{Title: "Test"}}}
	items := []ContextMenuItem{}
	cm := NewContextMenu(menu, 0, items, 10, 10)

	rendered := cm.renderModal(style.NewStyleSet(style.DefaultDarkTheme()), 80, 24)

	if rendered.content != "" {
		t.Error("expected empty content for empty items")
	}
}

func TestContextMenuHoverStyleUsesThemedBackgroundWithoutUnderline(t *testing.T) {
	hoverBackground := lipgloss.Color("#123456")
	theme := style.DefaultDarkTheme()
	theme.Popup.ContextMenuItemHover = style.Highlight{Bg: hoverBackground}
	styles := style.NewStyleSet(theme)

	cm := NewContextMenu(&testMenu{}, 0, []ContextMenuItem{{ID: "a", Label: "Action"}}, 0, 0)
	cm.hovered = 0
	itemStyle := cm.itemStyle(styles, 0)

	if got := itemStyle.GetBackground(); got != hoverBackground {
		t.Errorf("hover background = %v, want %v", got, hoverBackground)
	}
	if itemStyle.GetUnderline() {
		t.Error("context menu hover style must not underline text by default")
	}
}

func TestContextMenuChromeStylesUseThemedColors(t *testing.T) {
	borderColor := lipgloss.Color("#234567")
	separatorColor := lipgloss.Color("#345678")
	theme := style.DefaultDarkTheme()
	theme.Popup.ContextMenuBorder = borderColor
	theme.Popup.ContextMenuSeparator = separatorColor
	styles := style.NewStyleSet(theme)

	if got := styles.Popup.ContextMenuFrame.GetBorderTopForeground(); got != borderColor {
		t.Errorf("border color = %v, want %v", got, borderColor)
	}
	if got := styles.Popup.ContextMenuSeparator.GetForeground(); got != separatorColor {
		t.Errorf("separator color = %v, want %v", got, separatorColor)
	}
}

func TestContextMenuDefaultHoverStyleHasBackground(t *testing.T) {
	styles := style.NewStyleSet(style.DefaultDarkTheme())
	if styles.Popup.ContextMenuItemHover.GetBackground() == styles.Popup.Surface {
		t.Error("default context menu hover background must differ from the popup surface")
	}
	if styles.Popup.ContextMenuItemHover.GetUnderline() {
		t.Error("default context menu hover style must not underline text")
	}
}

func TestContextMenuDefaultChromeColors(t *testing.T) {
	tests := []struct {
		name  string
		theme style.Theme
		want  string
	}{
		{name: "dark", theme: style.DefaultDarkTheme(), want: "#5C5C5C"},
		{name: "light", theme: style.DefaultLightTheme(), want: "#A8A8A8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := style.NewStyleSet(tt.theme)
			want := lipgloss.Color(tt.want)
			if got := styles.Popup.ContextMenuFrame.GetBorderTopForeground(); got != want {
				t.Errorf("default border color = %v, want %v", got, want)
			}
			if got := styles.Popup.ContextMenuSeparator.GetForeground(); got != want {
				t.Errorf("default separator color = %v, want %v", got, want)
			}
		})
	}
}

func TestContextMenuItemStyleLeavesForegroundUnset(t *testing.T) {
	styles := style.NewStyleSet(style.DefaultDarkTheme())
	if got := styles.Popup.ContextMenuItem.GetForeground(); got != (lipgloss.NoColor{}) {
		t.Errorf("default context menu item color = %v, want no explicit foreground", got)
	}
}

func TestContextMenuItemStylesUsePopupTheme(t *testing.T) {
	normalColor := lipgloss.Color("#111111")
	focusedColor := lipgloss.Color("#222222")
	disabledColor := lipgloss.Color("#333333")
	theme := style.DefaultDarkTheme()
	theme.Popup.ContextMenuItem = style.Highlight{Fg: normalColor}
	theme.Popup.ContextMenuItemFocused = style.Highlight{Fg: focusedColor}
	theme.Popup.ContextMenuItemDisabled = style.Highlight{Fg: disabledColor}
	styles := style.NewStyleSet(theme)

	cm := NewContextMenu(&testMenu{}, 0, []ContextMenuItem{
		{ID: "normal", Label: "Normal"},
		{ID: "focused", Label: "Focused"},
		{ID: "disabled", Label: "Disabled", Disabled: true},
	}, 0, 0)
	cm.focused = 1

	if got := cm.itemStyle(styles, 0).GetForeground(); got != normalColor {
		t.Errorf("normal item color = %v, want %v", got, normalColor)
	}
	if got := cm.itemStyle(styles, 1).GetForeground(); got != focusedColor {
		t.Errorf("focused item color = %v, want %v", got, focusedColor)
	}
	if got := cm.itemStyle(styles, 2).GetForeground(); got != disabledColor {
		t.Errorf("disabled item color = %v, want %v", got, disabledColor)
	}
}

func TestContextMenuDefaultsToUnlimitedDimensions(t *testing.T) {
	items := []ContextMenuItem{
		{ID: "a", Label: "A context menu label that stays complete"},
		{ID: "b", Label: "Second item"},
	}
	cm := NewContextMenu(&testMenu{}, 0, items, 0, 0)
	rendered := cm.renderModal(style.NewStyleSet(style.DefaultDarkTheme()), 80, 24)
	plain := ansi.Strip(rendered.content)

	if cm.maxWidth != 0 || cm.maxHeight != 0 {
		t.Fatalf("default limits = (%d, %d), want unlimited (0, 0)", cm.maxWidth, cm.maxHeight)
	}
	if !strings.Contains(plain, items[0].Label) || strings.Contains(plain, "…") {
		t.Fatalf("default rendering unexpectedly truncated label: %q", plain)
	}
	if got := lipgloss.Height(rendered.content); got != len(items)+contextMenuFrameOverhead {
		t.Fatalf("default height = %d, want %d", got, len(items)+contextMenuFrameOverhead)
	}
}

func TestContextMenuMaxWidthTruncatesLabelWithEllipsis(t *testing.T) {
	cm := newContextMenu(
		&testMenu{},
		0,
		[]ContextMenuItem{{ID: "a", Label: "1234567890ABC"}},
		0,
		0,
		ContextMenuOptions{MaxWidth: 12},
	)
	rendered := cm.renderModal(style.NewStyleSet(style.DefaultDarkTheme()), 80, 24)
	plain := ansi.Strip(rendered.content)

	if got := lipgloss.Width(rendered.content); got != 12 {
		t.Fatalf("rendered width = %d, want configured maximum 12", got)
	}
	if !strings.Contains(plain, "1234567…") {
		t.Fatalf("truncated label does not contain expected ellipsis: %q", plain)
	}
	if strings.Contains(plain, "1234567890ABC") {
		t.Fatalf("rendered menu still contains the complete overlong label: %q", plain)
	}
}

func TestContextMenuMaxHeightRendersScrollbarAndMouseWheelScrolls(t *testing.T) {
	items := []ContextMenuItem{
		{ID: "one", Label: "One"},
		{ID: "two", Label: "Two"},
		{ID: "three", Label: "Three"},
		{ID: "four", Label: "Four"},
		{ID: "five", Label: "Five"},
	}
	cm := newContextMenu(&testMenu{}, 0, items, 0, 0, ContextMenuOptions{MaxHeight: 5})
	styles := style.NewStyleSet(style.DefaultDarkTheme())
	rendered := cm.renderModal(styles, 80, 24)
	plain := ansi.Strip(rendered.content)

	if got := lipgloss.Height(rendered.content); got != 5 {
		t.Fatalf("rendered height = %d, want configured maximum 5", got)
	}
	if !strings.Contains(plain, "│") || !strings.Contains(plain, "█") {
		t.Fatalf("overflowing context menu has no visible scrollbar: %q", plain)
	}
	if !strings.Contains(plain, "One") || !strings.Contains(plain, "Three") || strings.Contains(plain, "Four") {
		t.Fatalf("initial viewport is incorrect: %q", plain)
	}
	if rendered.itemBounds[3].h != 0 {
		t.Fatalf("off-screen item has clickable bounds: %+v", rendered.itemBounds[3])
	}

	w := lipgloss.Width(rendered.content)
	h := lipgloss.Height(rendered.content)
	cm.setModalBounds(0, 0, w, h, rendered.itemBounds)
	handled, _ := cm.handleMouse(tea.MouseWheelMsg(tea.Mouse{
		X:      1,
		Y:      1,
		Button: tea.MouseWheelDown,
	}))
	if !handled || cm.scrollOffset != 1 {
		t.Fatalf("wheel down handled=%v scrollOffset=%d, want true and 1", handled, cm.scrollOffset)
	}

	rendered = cm.renderModal(styles, 80, 24)
	plain = ansi.Strip(rendered.content)
	if strings.Contains(plain, "One") || !strings.Contains(plain, "Four") {
		t.Fatalf("viewport did not advance after wheel down: %q", plain)
	}
}

func TestMainContextMenuUsesConfiguredLimits(t *testing.T) {
	app, main := newMainForMenuMouseTest(t)
	app.With(WithContextMenuOptions(ContextMenuOptions{MaxWidth: 18, MaxHeight: 4}))
	main.mouseClickHandle(tea.Mouse{X: 0, Y: main.menuListStartRow, Button: tea.MouseRight}, app)

	if !app.HasPopup() {
		t.Fatal("right click did not open a context menu")
	}
	cm, ok := app.modalStack[len(app.modalStack)-1].(*ContextMenu)
	if !ok {
		t.Fatalf("top modal = %T, want *ContextMenu", app.modalStack[len(app.modalStack)-1])
	}
	if cm.maxWidth != 18 || cm.maxHeight != 4 {
		t.Fatalf("context menu limits = (%d, %d), want (18, 4)", cm.maxWidth, cm.maxHeight)
	}
}

func TestTransparentThemeContextMenuHoverUsesPrimary(t *testing.T) {
	primary := lipgloss.Color("#E040FB")
	theme := style.DefaultDarkTheme()
	theme.Primary = primary
	theme.AppBackground = style.Highlight{Bg: lipgloss.NoColor{}}
	theme.Popup.Surface = lipgloss.NoColor{}
	styles := style.NewStyleSet(theme)

	cm := NewContextMenu(&testMenu{}, 0, []ContextMenuItem{{ID: "action", Label: "Action"}}, 0, 0)
	cm.hovered = 0
	if got := cm.itemStyle(styles, 0).GetForeground(); got != primary {
		t.Fatalf("transparent hover foreground = %v, want primary %v", got, primary)
	}
}
