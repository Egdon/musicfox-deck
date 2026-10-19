package model

import (
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	uv "github.com/charmbracelet/ultraviolet"
)

func TestNormalizePopupSurfaceOverridesContentBackground(t *testing.T) {
	surface := color.NRGBA{R: 28, G: 31, B: 38, A: 255}
	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#FF0000")).
		Render("A") + "\x1b[7mB\x1b[0m"

	normalized := normalizePopupSurface(content, surface)
	screen := popupStyledScreen(normalized)
	for _, line := range screen.Lines {
		for _, cell := range line {
			if cell.IsZero() {
				continue
			}
			if !samePopupColor(cell.Style.Bg, surface) {
				t.Fatalf("cell %q background = %#v, want popup surface %#v", cell.Content, cell.Style.Bg, surface)
			}
			if cell.Style.Attrs&uv.AttrReverse != 0 {
				t.Fatalf("cell %q retained reverse video", cell.Content)
			}
		}
	}
}

func TestPopupRenderUsesSurfaceForScrolledContent(t *testing.T) {
	surface := color.NRGBA{R: 24, G: 28, B: 36, A: 255}
	theme := style.DefaultDarkTheme()
	theme.Popup.Surface = surface
	styles := style.NewStyleSet(theme).Popup
	content := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Render("one") + "\n" +
		lipgloss.NewStyle().Background(lipgloss.Color("#00FF00")).Render("two") + "\n" +
		lipgloss.NewStyle().Background(lipgloss.Color("#0000FF")).Render("one") + "\n" +
		lipgloss.NewStyle().Background(lipgloss.Color("#FFFF00")).Render("one")
	if !samePopupColor(styles.ScrollTrack.GetBackground(), surface) || !samePopupColor(styles.ScrollThumb.GetBackground(), surface) {
		t.Fatal("scrollbar styles must use the popup surface background")
	}

	popup, err := NewPopup(PopupSpec{
		Title:     "Surface",
		Content:   content,
		MaxHeight: 6,
		Actions:   []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	rendered := popup.render(styles)
	if !popup.isContentScrollable() {
		t.Fatal("expected popup body to scroll")
	}
	screen := popupStyledScreen(rendered.content)
	for _, line := range screen.Lines {
		for _, cell := range line {
			if cell.Content != "o" && cell.Content != "n" && cell.Content != "e" {
				continue
			}
			if !samePopupColor(cell.Style.Bg, surface) {
				t.Fatalf("visible content cell %q background = %#v, want popup surface %#v", cell.Content, cell.Style.Bg, surface)
			}
		}
	}
}

func TestPopupMaxWidthIncludesFrame(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Content:  "this content must be constrained by the popup width",
		MaxWidth: 12,
		Actions:  []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	rendered := popup.render(style.NewStyleSet(style.DefaultDarkTheme()).Popup)
	if got, want := lipgloss.Width(rendered.content), 12; got > want {
		t.Fatalf("popup width = %d, want at most %d including frame", got, want)
	}
}

func TestPopupFrameFillsContentRowAfterStyledContent(t *testing.T) {
	surface := color.NRGBA{R: 30, G: 34, B: 42, A: 255}
	theme := style.DefaultDarkTheme()
	theme.Popup.Surface = surface
	popup, err := NewPopup(PopupSpec{Content: lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Render("X")})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	rendered := popup.render(style.NewStyleSet(theme).Popup)
	screen := popupStyledScreen(rendered.content)
	for y := 1; y < len(screen.Lines)-1; y++ {
		for x := 1; x < len(screen.Lines[y])-1; x++ {
			cell := screen.CellAt(x, y)
			if !samePopupColor(cell.Style.Bg, surface) {
				t.Fatalf("interior cell (%d, %d) %q background = %#v, want popup surface %#v", x, y, cell.Content, cell.Style.Bg, surface)
			}
		}
	}
}

func TestPopupFrameFillsBorderCellsWithSurface(t *testing.T) {
	surface := color.NRGBA{R: 30, G: 34, B: 42, A: 255}
	theme := style.DefaultDarkTheme()
	theme.Popup.Surface = surface
	popup, err := NewPopup(PopupSpec{Content: "content"})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	screen := popupStyledScreen(popup.render(style.NewStyleSet(theme).Popup).content)
	for y, line := range screen.Lines {
		for x := range line {
			if x != 0 && x != len(line)-1 && y != 0 && y != len(screen.Lines)-1 {
				continue
			}
			cell := screen.CellAt(x, y)
			if !samePopupColor(cell.Style.Bg, surface) {
				t.Fatalf("border cell (%d, %d) %q background = %#v, want popup surface %#v", x, y, cell.Content, cell.Style.Bg, surface)
			}
		}
	}
}

func TestPopupInteriorHasNoUnstyledGaps(t *testing.T) {
	surface := color.NRGBA{R: 30, G: 34, B: 42, A: 255}
	theme := style.DefaultDarkTheme()
	theme.Popup.Surface = surface
	popup, err := NewPopup(PopupSpec{
		Title:   "Anchor: Bottom-Center",
		Content: "This popup is anchored at Bottom-Center.\nOffset: (0, 0)\n\nPress ESC to dismiss.",
		Actions: []PopupAction{
			{ID: "confirm", Label: "Confirm"},
			{ID: "cancel", Label: "Cancel", IsCancel: true},
		},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	screen := popupStyledScreen(popup.render(style.NewStyleSet(theme).Popup).content)
	for y := range screen.Lines {
		for x := range screen.Lines[y] {
			cell := screen.CellAt(x, y)
			if cell == nil || cell.IsZero() {
				continue
			}
			if cell.Style.Bg == nil {
				t.Fatalf("interior cell (%d, %d) %q has no background; popup surface must fill every cell", x, y, cell.Content)
			}
		}
	}
}

func TestAppCompositePopupsPreservesPopupSurface(t *testing.T) {
	surface := color.NRGBA{R: 30, G: 34, B: 42, A: 255}
	theme := style.DefaultDarkTheme()
	theme.Popup.Surface = surface
	previousStyles := style.CurrentStyleSet()
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(previousStyles) })

	popup, err := NewPopup(PopupSpec{Content: lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Render("X")})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}
	app := &App{windowWidth: 80, windowHeight: 24}
	app.ShowPopup(popup)
	base := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Width(80).Height(24).Render("")
	screen := popupStyledScreen(app.compositeModals(base))

	contentFound := false
	for _, line := range screen.Lines {
		for _, cell := range line {
			if cell.Content != "X" {
				continue
			}
			contentFound = true
			if !samePopupColor(cell.Style.Bg, surface) {
				t.Fatalf("composited content background = %#v, want popup surface %#v", cell.Style.Bg, surface)
			}
		}
	}
	if !contentFound {
		t.Fatal("composite output omitted popup content")
	}
	border := screen.CellAt(popup.bounds.x, popup.bounds.y)
	if !samePopupColor(border.Style.Bg, surface) {
		t.Fatalf("composited border background = %#v, want popup surface %#v", border.Style.Bg, surface)
	}
}
func TestPopupActionBoundsMatchRenderedActions(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Content: "Choose an action",
		Actions: []PopupAction{
			{ID: "confirm", Label: "Confirm"},
			{ID: "cancel", Label: "Cancel", IsCancel: true},
		},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	rendered := popup.render(style.NewStyleSet(style.DefaultDarkTheme()).Popup)
	if got, want := len(rendered.actionBounds), 2; got != want {
		t.Fatalf("action bounds = %d, want %d", got, want)
	}
	popup.setBounds(7, 11, lipgloss.Width(rendered.content), lipgloss.Height(rendered.content), rendered.actionBounds)
	for index, bounds := range popup.actionBounds {
		if got := popup.actionAt(bounds.x, bounds.y); got != index {
			t.Fatalf("actionAt(%d, %d) = %d, want %d", bounds.x, bounds.y, got, index)
		}
	}
}

func TestNewPopupRejectsInvalidActions(t *testing.T) {
	_, err := NewPopup(PopupSpec{Actions: []PopupAction{{ID: "same", Label: "One"}, {ID: "same", Label: "Two"}}})
	if err == nil {
		t.Fatal("NewPopup() accepted duplicate action IDs")
	}

	_, err = NewPopup(PopupSpec{Actions: []PopupAction{{ID: "escape", Label: "Esc", IsCancel: true}, {ID: "outside", Label: "Outside", IsCancel: true}}})
	if err == nil {
		t.Fatal("NewPopup() accepted multiple cancel actions")
	}
}

func TestPopupEscapeSelectsCancelAction(t *testing.T) {
	popup, err := NewPopup(PopupSpec{Actions: []PopupAction{{ID: "cancel", Label: "Cancel", IsCancel: true}}})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	popup.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	result := popup.consumeResult()
	if result == nil {
		t.Fatal("escape did not dismiss popup")
	}
	if result.ActionID != "cancel" || result.Cause != PopupDismissEscape {
		t.Fatalf("escape result = %+v, want cancel action with escape cause", *result)
	}
}

// newScrollablePopup builds a popup whose content overflows the visible area,
// renders it with the default dark theme, and anchors it at screen (0,0) so
// popup-relative coordinates equal screen coordinates.
func newScrollablePopup(t *testing.T, actions ...PopupAction) (*Popup, style.PopupStyleSet) {
	t.Helper()
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line" + string(rune('0'+i))
	}
	styles := style.NewStyleSet(style.DefaultDarkTheme()).Popup
	popup, err := NewPopup(PopupSpec{
		Content:   strings.Join(lines, "\n"),
		MaxHeight: 9,
		Actions:   actions,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}
	rendered := popup.render(styles)
	if !popup.isContentScrollable() {
		t.Fatal("expected popup content to be scrollable")
	}
	w := lipgloss.Width(rendered.content)
	h := lipgloss.Height(rendered.content)
	popup.setBounds(0, 0, w, h, rendered.actionBounds)
	return popup, styles
}

func TestPopupScrollbarThumbDrag(t *testing.T) {
	popup, styles := newScrollablePopup(t)
	if popup.thumbRelY != popupFrameInsetY {
		t.Fatalf("thumbRelY = %d, want %d at offset 0", popup.thumbRelY, popupFrameInsetY)
	}

	// Grab the thumb.
	handled, _ := popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: popup.scrollbarRelX, Y: popup.thumbRelY, Button: tea.MouseLeft}))
	if !handled || !popup.scrollDragging {
		t.Fatalf("clicking thumb did not start scroll drag (handled=%v dragging=%v)", handled, popup.scrollDragging)
	}

	// Drag to the bottom of the scroll track.
	bottom := popup.bodyRelY + popup.visibleRows - 1
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: popup.scrollbarRelX, Y: bottom, Button: tea.MouseLeft}))
	if popup.scrollOffset != popup.maxScrollOffset() {
		t.Fatalf("scrollOffset after drag = %d, want max %d", popup.scrollOffset, popup.maxScrollOffset())
	}

	// Release ends the drag.
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: popup.scrollbarRelX, Y: bottom}))
	if popup.scrollDragging {
		t.Fatal("scroll drag did not end on release")
	}
	_ = styles
}

func TestPopupScrollbarTrackJump(t *testing.T) {
	popup, _ := newScrollablePopup(t)
	bottom := popup.bodyRelY + popup.visibleRows - 1
	// Click the track below the thumb (not on the thumb itself).
	handled, _ := popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: popup.scrollbarRelX, Y: bottom, Button: tea.MouseLeft}))
	if !handled {
		t.Fatal("track click not handled")
	}
	if popup.scrollDragging {
		t.Fatal("track click should not start a thumb drag")
	}
	if popup.scrollOffset != popup.maxScrollOffset() {
		t.Fatalf("track click scrollOffset = %d, want max %d", popup.scrollOffset, popup.maxScrollOffset())
	}
}

func TestPopupPointerShapes(t *testing.T) {
	popup, _ := newScrollablePopup(t, PopupAction{ID: "ok", Label: "OK"})

	// Over the scrollbar thumb -> pointer.
	if got := popup.desiredPointer(tea.Mouse{X: popup.scrollbarRelX, Y: popup.thumbRelY}); got != "pointer" {
		t.Fatalf("pointer over thumb = %q, want pointer", got)
	}
	// Over content text -> text cursor.
	if got := popup.desiredPointer(tea.Mouse{X: popup.bodyRelX, Y: popup.bodyRelY}); got != "text" {
		t.Fatalf("pointer over content = %q, want text", got)
	}
	// Over an action button -> pointer.
	if len(popup.actionBounds) == 0 {
		t.Fatal("expected action bounds")
	}
	ab := popup.actionBounds[0]
	if got := popup.desiredPointer(tea.Mouse{X: ab.x, Y: ab.y}); got != "pointer" {
		t.Fatalf("pointer over action = %q, want pointer", got)
	}
	// Outside the popup -> default (empty).
	if got := popup.desiredPointer(tea.Mouse{X: popup.bounds.x + popup.bounds.w + 5, Y: popup.bounds.y}); got != "" {
		t.Fatalf("pointer outside popup = %q, want empty", got)
	}
}

func TestPopupTextSelectionCopiesToClipboard(t *testing.T) {
	popup, _ := newScrollablePopup(t)

	// Press at the first content column of the first visible line.
	handled, _ := popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: popup.bodyRelX, Y: popup.bodyRelY, Button: tea.MouseLeft}))
	if !handled || !popup.selecting {
		t.Fatalf("content press did not start selection (handled=%v selecting=%v)", handled, popup.selecting)
	}

	// Drag across the full width of the first line (past its last column).
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: popup.bodyRelX + popup.contentTextW, Y: popup.bodyRelY, Button: tea.MouseLeft}))

	// Release finalizes: expect a clipboard command and the extracted text.
	_, cmd := popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: popup.bodyRelX + popup.contentTextW, Y: popup.bodyRelY}))
	if cmd == nil {
		t.Fatal("release did not produce a clipboard command")
	}
	if popup.selecting {
		t.Fatal("selection should end on release")
	}
	if got := popup.selectionText(); got != "line0" {
		t.Fatalf("selectionText() = %q, want %q", got, "line0")
	}
}

func TestPopupSelectionSpansMultipleLines(t *testing.T) {
	popup, _ := newScrollablePopup(t)

	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: popup.bodyRelX, Y: popup.bodyRelY, Button: tea.MouseLeft}))
	// Drag to the end of the third visible line.
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: popup.bodyRelX + popup.contentTextW, Y: popup.bodyRelY + 2, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: popup.bodyRelX + popup.contentTextW, Y: popup.bodyRelY + 2}))

	if got, want := popup.selectionText(), "line0\nline1\nline2"; got != want {
		t.Fatalf("selectionText() = %q, want %q", got, want)
	}
}

func TestPopupEscapeClearsSelectionBeforeDismiss(t *testing.T) {
	popup, _ := newScrollablePopup(t)
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: popup.bodyRelX, Y: popup.bodyRelY, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: popup.bodyRelX + popup.contentTextW, Y: popup.bodyRelY, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: popup.bodyRelX + popup.contentTextW, Y: popup.bodyRelY}))
	if !popup.hasSelection {
		t.Fatal("expected an active selection")
	}

	// First Esc clears the selection without dismissing.
	popup.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if popup.hasSelection {
		t.Fatal("first Esc should clear the selection")
	}
	if popup.dismissed() {
		t.Fatal("first Esc should not dismiss the popup")
	}

	// Second Esc dismisses.
	popup.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if !popup.dismissed() {
		t.Fatal("second Esc should dismiss the popup")
	}
}

func TestPopupCustomCloseKeyDismisses(t *testing.T) {
	popup, err := NewPopup(PopupSpec{CloseKeys: []string{"q"}})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	// Esc is not a configured close key, so it must not dismiss.
	popup.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if popup.dismissed() {
		t.Fatal("esc dismissed popup that only closes on q")
	}

	popup.update(tea.KeyPressMsg(tea.Key{Code: 'q'}))
	result := popup.consumeResult()
	if result == nil {
		t.Fatal("q did not dismiss popup")
	}
	if result.Cause != PopupDismissKey || result.Key != "q" {
		t.Fatalf("q result = %+v, want key cause with key %q", *result, "q")
	}
}

func TestPopupEmptyCloseKeysDisablesKeyDismiss(t *testing.T) {
	popup, err := NewPopup(PopupSpec{CloseKeys: []string{}})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	popup.update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if popup.dismissed() {
		t.Fatal("esc dismissed popup with key dismissal disabled")
	}
}

func TestNewPopupRejectsEmptyCloseKey(t *testing.T) {
	if _, err := NewPopup(PopupSpec{CloseKeys: []string{""}}); err == nil {
		t.Fatal("NewPopup() accepted empty close key")
	}
}

func TestPopupDisableOutsideClickKeepsOpen(t *testing.T) {
	popup, err := NewPopup(PopupSpec{DisableOutsideClick: true})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	if popup.dismissOutside() {
		t.Fatal("dismissOutside() dismissed a popup with DisableOutsideClick")
	}
	if popup.dismissed() {
		t.Fatal("popup dismissed despite DisableOutsideClick")
	}
}

func TestPopupOutsideClickDismissesByDefault(t *testing.T) {
	popup, err := NewPopup(PopupSpec{})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	if !popup.dismissOutside() {
		t.Fatal("dismissOutside() did not dismiss default popup")
	}
	result := popup.consumeResult()
	if result == nil || result.Cause != PopupDismissOutsideClick {
		t.Fatalf("outside-click result = %+v, want outside-click cause", result)
	}
}

func samePopupColor(got, want color.Color) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	gr, gg, gb, ga := got.RGBA()
	wr, wg, wb, wa := want.RGBA()
	return gr == wr && gg == wg && gb == wb && ga == wa
}
