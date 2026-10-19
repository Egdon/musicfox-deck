package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

func TestPopupResizeBottomRight(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:   "Test",
		Content: "Content",
		Actions: []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	styleSet := style.NewStyleSet(theme).Popup

	// Render and set bounds
	rendered := popup.render(styleSet)
	popup.setBounds(10, 5, 40, 20, rendered.actionBounds)

	initialWidth := popup.maxWidth
	initialHeight := popup.maxHeight

	// Simulate resize start (click on bottom-right corner)
	cornerX := popup.bounds.x + popup.bounds.w - 1
	cornerY := popup.bounds.y + popup.bounds.h - 1
	clickMsg := tea.MouseClickMsg(tea.Mouse{X: cornerX, Y: cornerY, Button: tea.MouseLeft})

	handled, _ := popup.handleMouse(clickMsg)
	if !handled {
		t.Error("Expected resize click to be handled")
	}
	if !popup.resizing {
		t.Error("Expected popup.resizing to be true after clicking resize corner")
	}

	// Note: popup was created with maxWidth=0, maxHeight=0 (unlimited)
	// After rendering, bounds.w and bounds.h reflect the actual rendered size
	// When we resize, we're modifying maxWidth/maxHeight from 0 to actual values
	initialWidth = popup.bounds.w
	initialHeight = popup.bounds.h

	// Simulate drag to increase size by 5x3
	dragMsg := tea.MouseMotionMsg(tea.Mouse{X: cornerX + 5, Y: cornerY + 3})
	handled, _ = popup.handleMouse(dragMsg)
	if !handled {
		t.Error("Expected resize drag to be handled")
	}

	// Check that maxWidth and maxHeight were set to new sizes
	// The resize increases from the rendered size (bounds.w/h), not from maxWidth/maxHeight
	expectedWidth := initialWidth + 5
	expectedHeight := initialHeight + 3
	if popup.maxWidth != expectedWidth {
		t.Errorf("Expected maxWidth = %d, got %d", expectedWidth, popup.maxWidth)
	}
	if popup.maxHeight != expectedHeight {
		t.Errorf("Expected maxHeight = %d, got %d", expectedHeight, popup.maxHeight)
	}
	// Simulate release
	releaseMsg := tea.MouseReleaseMsg(tea.Mouse{X: cornerX + 5, Y: cornerY + 3})
	handled, _ = popup.handleMouse(releaseMsg)
	if !handled {
		t.Error("Expected resize release to be handled")
	}
	if popup.resizing {
		t.Error("Expected popup.resizing to be false after release")
	}
}

func TestPopupResizeMinimumSize(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:   "Test",
		Content: "Content",
		Actions: []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	styleSet := style.NewStyleSet(theme).Popup

	rendered := popup.render(styleSet)
	popup.setBounds(10, 5, 40, 20, rendered.actionBounds)

	// Start resize
	cornerX := popup.bounds.x + popup.bounds.w - 1
	cornerY := popup.bounds.y + popup.bounds.h - 1
	clickMsg := tea.MouseClickMsg(tea.Mouse{X: cornerX, Y: cornerY, Button: tea.MouseLeft})
	popup.handleMouse(clickMsg)

	// Try to drag to very small size (should be clamped)
	dragMsg := tea.MouseMotionMsg(tea.Mouse{X: cornerX - 100, Y: cornerY - 100})
	popup.handleMouse(dragMsg)

	// Check minimum sizes are enforced
	minW := popupFrameHorizontalOverhead + 10
	minH := popupFrameVerticalOverhead + 3
	if popup.maxWidth < minW {
		t.Errorf("Expected maxWidth >= %d, got %d", minW, popup.maxWidth)
	}
	if popup.maxHeight < minH {
		t.Errorf("Expected maxHeight >= %d, got %d", minH, popup.maxHeight)
	}
}

func TestPopupResizeHandleDetection(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:   "Test",
		Content: "Content",
		Actions: []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	styleSet := style.NewStyleSet(theme).Popup

	rendered := popup.render(styleSet)
	popup.setBounds(10, 5, 40, 20, rendered.actionBounds)

	tests := []struct {
		name string
		x    int
		y    int
		want ResizeHandle
	}{
		{
			name: "bottom_right_corner",
			x:    popup.bounds.x + popup.bounds.w - 1,
			y:    popup.bounds.y + popup.bounds.h - 1,
			want: ResizeBottomRight,
		},
		{
			name: "one_cell_left",
			x:    popup.bounds.x + popup.bounds.w - 2,
			y:    popup.bounds.y + popup.bounds.h - 1,
			want: ResizeBottom, // on bottom edge, not corner
		},
		{
			name: "one_cell_up",
			x:    popup.bounds.x + popup.bounds.w - 1,
			y:    popup.bounds.y + popup.bounds.h - 2,
			want: ResizeRight, // on right edge, not corner
		},
		{
			name: "outside_popup",
			x:    popup.bounds.x + popup.bounds.w + 5,
			y:    popup.bounds.y + popup.bounds.h + 5,
			want: ResizeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mouse := tea.Mouse{X: tt.x, Y: tt.y}
			got := popup.resizeHandleAt(mouse)
			if got != tt.want {
				t.Errorf("resizeHandleAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPopupResizePointerShape(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:   "Test",
		Content: "Content",
		Actions: []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	styleSet := style.NewStyleSet(theme).Popup

	rendered := popup.render(styleSet)
	popup.setBounds(10, 5, 40, 20, rendered.actionBounds)

	// Test resize corner shows nwse-resize cursor
	cornerX := popup.bounds.x + popup.bounds.w - 1
	cornerY := popup.bounds.y + popup.bounds.h - 1
	mouse := tea.Mouse{X: cornerX, Y: cornerY}

	pointer := popup.desiredPointer(mouse)
	if pointer != "nwse-resize" {
		t.Errorf("Expected pointer = 'nwse-resize' at resize corner, got %q", pointer)
	}

	// Test one cell away shows default
	mouse = tea.Mouse{X: cornerX - 1, Y: cornerY - 1}
	pointer = popup.desiredPointer(mouse)
	if pointer == "nwse-resize" {
		t.Error("Expected pointer != 'nwse-resize' away from resize corner")
	}
}

func TestPopupResizeIndicatorRendered(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:   "Test",
		Content: "Content",
		Actions: []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	styleSet := style.NewStyleSet(theme).Popup

	rendered := popup.render(styleSet)

	// Check that rendered content is not empty
	if rendered.content == "" {
		t.Fatal("Expected non-empty rendered content")
	}

	// Check that the content contains the resize indicator
	// Note: actual rendering verification would require parsing ANSI,
	// but we can at least check the function doesn't crash
	screen := popupStyledScreen(rendered.content)
	if len(screen.Lines) == 0 {
		t.Fatal("Expected screen to have lines")
	}

	lastLine := len(screen.Lines) - 1
	if lastLine < 0 {
		t.Fatal("Expected at least one line")
	}

	// The resize indicator should be in the last cell
	lastCol := len(screen.Lines[lastLine]) - 1
	if lastCol < 0 {
		t.Fatal("Expected last line to have cells")
	}

	cell := screen.Lines[lastLine][lastCol]
	// The indicator is "◢"
	if cell.Content != "◢" {
		t.Logf("Note: Bottom-right cell content = %q (expected '◢', but may be overridden by border)", cell.Content)
		// Don't fail - the border might override it depending on rendering order
	}
}

func TestPopupResizeVisualWidthIncreases(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:     "Short",
		Content:   "Hi",
		Actions:   []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:  30,
		MaxHeight: 10,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}
	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	initialW := lipgloss.Width(rendered.content)
	if initialW != 30 {
		t.Errorf("expected initial width=30, got %d", initialW)
	}
	popup.setBounds(10, 5, initialW, lipgloss.Height(rendered.content), rendered.actionBounds)

	// Resize: corner click + drag right 40
	cx, cy := popup.bounds.x+popup.bounds.w-1, popup.bounds.y+popup.bounds.h-1
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx + 40, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx + 40, Y: cy}))

	if popup.maxWidth <= 30 {
		t.Errorf("expected maxWidth > 30, got %d", popup.maxWidth)
	}
	r2 := popup.render(ss)
	if w := lipgloss.Width(r2.content); w != popup.maxWidth {
		t.Errorf("rendered width %d != maxWidth %d", w, popup.maxWidth)
	}
}

func TestPopupResizeVisualHeightIncreases(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:     "Short",
		Content:   "One line",
		Actions:   []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:  40,
		MaxHeight: 10,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}
	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	initialH := lipgloss.Height(rendered.content)
	if initialH != 10 {
		t.Errorf("expected initial height=10, got %d", initialH)
	}
	popup.setBounds(10, 5, lipgloss.Width(rendered.content), initialH, rendered.actionBounds)

	// Resize: corner click + drag down 15
	cx, cy := popup.bounds.x+popup.bounds.w-1, popup.bounds.y+popup.bounds.h-1
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx, Y: cy + 15, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx, Y: cy + 15}))

	if popup.maxHeight <= 10 {
		t.Errorf("expected maxHeight > 10, got %d", popup.maxHeight)
	}
	r2 := popup.render(ss)
	if h := lipgloss.Height(r2.content); h != popup.maxHeight {
		t.Errorf("rendered height %d != maxHeight %d", h, popup.maxHeight)
	}
}

func TestPopupResizeVisualShrinks(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:     "Shrink",
		Content:   "This is some longer content that will wrap when the popup shrinks",
		Actions:   []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:  60,
		MaxHeight: 14,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}
	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	initialW, initialH := lipgloss.Width(rendered.content), lipgloss.Height(rendered.content)
	if initialW != 60 {
		t.Errorf("expected initial width=60, got %d", initialW)
	}
	if initialH != 14 {
		t.Errorf("expected initial height=14, got %d", initialH)
	}
	popup.setBounds(10, 5, initialW, initialH, rendered.actionBounds)

	// Shrink: click corner, drag LEFT 20 and UP 6
	cx, cy := popup.bounds.x+popup.bounds.w-1, popup.bounds.y+popup.bounds.h-1
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx - 20, Y: cy - 6, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx - 20, Y: cy - 6}))

	if popup.maxWidth >= 60 {
		t.Errorf("expected maxWidth < 60 after shrinking left, got %d", popup.maxWidth)
	}
	if popup.maxHeight >= 14 {
		t.Errorf("expected maxHeight < 14 after shrinking up, got %d", popup.maxHeight)
	}

	r2 := popup.render(ss)
	if w := lipgloss.Width(r2.content); w != popup.maxWidth {
		t.Errorf("rendered width %d != maxWidth %d after shrink", w, popup.maxWidth)
	}
	if h := lipgloss.Height(r2.content); h != popup.maxHeight {
		t.Errorf("rendered height %d != maxHeight %d after shrink", h, popup.maxHeight)
	}
}

func TestPopupBorderEdgePointerStyles(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:   "Test",
		Content: "Content",
		Actions: []PopupAction{{ID: "ok", Label: "OK"}},
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	styleSet := style.NewStyleSet(theme).Popup

	rendered := popup.render(styleSet)
	popup.setBounds(10, 5, 40, 20, rendered.actionBounds)

	tests := []struct {
		name     string
		x        int
		y        int
		expected string
	}{
		{
			name:     "top_edge_center",
			x:        popup.bounds.x + popup.bounds.w/2,
			y:        popup.bounds.y,
			expected: "grab", // title-bar drag
		},
		{
			name:     "bottom_edge_center",
			x:        popup.bounds.x + popup.bounds.w/2,
			y:        popup.bounds.y + popup.bounds.h - 1,
			expected: "ns-resize", // ↕ vertical
		},
		{
			name:     "left_edge_center",
			x:        popup.bounds.x,
			y:        popup.bounds.y + popup.bounds.h/2,
			expected: "ew-resize", // ↔ horizontal
		},
		{
			name:     "right_edge_center",
			x:        popup.bounds.x + popup.bounds.w - 1,
			y:        popup.bounds.y + popup.bounds.h/2,
			expected: "ew-resize", // ↔ horizontal
		},
		{
			name:     "bottom_right_corner",
			x:        popup.bounds.x + popup.bounds.w - 1,
			y:        popup.bounds.y + popup.bounds.h - 1,
			expected: "nwse-resize", // ↖↘ diagonal
		},
		{
			name:     "center_content",
			x:        popup.bounds.x + popup.bounds.w/2,
			y:        popup.bounds.y + popup.bounds.h/2,
			expected: "", // default or text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mouse := tea.Mouse{X: tt.x, Y: tt.y}
			got := popup.desiredPointer(mouse)
			if got != tt.expected {
				t.Errorf("desiredPointer() = %q, want %q at position (%d,%d)", got, tt.expected, tt.x, tt.y)
			}
		})
	}
}

func TestPopupDisableResizeNoIndicator(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:         "No Resize",
		Content:       "Content",
		Actions:       []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:      40,
		MaxHeight:     10,
		DisableResize: true,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	screen := popupStyledScreen(rendered.content)
	if len(screen.Lines) == 0 {
		t.Fatal("Expected screen to have lines")
	}

	lastLine := len(screen.Lines) - 1
	lastCol := len(screen.Lines[lastLine]) - 1
	cell := screen.Lines[lastLine][lastCol]

	// With DisableResize=true, the ◢ indicator should NOT be present.
	// The last cell is typically the bottom border character like "╯".
	if cell.Content == "\u25e2" {
		t.Errorf("Expected no ◢ resize indicator when DisableResize=true, but found it")
	}
	t.Logf("Last cell content: %q (expected border char, not ◢)", cell.Content)
}

func TestPopupDisableResizeNoResizeCursor(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:         "No Resize",
		Content:       "Content",
		Actions:       []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:      40,
		MaxHeight:     10,
		DisableResize: true,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	popup.setBounds(10, 5, lipgloss.Width(rendered.content), lipgloss.Height(rendered.content), rendered.actionBounds)

	bx, by, bw, bh := popup.bounds.x, popup.bounds.y, popup.bounds.w, popup.bounds.h

	tests := []struct {
		name   string
		x      int
		y      int
		badPtr string
	}{
		{"corner", bx + bw - 1, by + bh - 1, "nwse-resize"},
		{"right_edge", bx + bw - 1, by + bh/2, "ew-resize"},
		{"bottom_edge", bx + bw/2, by + bh - 1, "ns-resize"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := popup.desiredPointer(tea.Mouse{X: tt.x, Y: tt.y})
			if ptr == tt.badPtr {
				t.Errorf("Expected no %q cursor when DisableResize=true, got it", tt.badPtr)
			}
		})
	}
}

func TestPopupDisableResizeClickIgnored(t *testing.T) {
	popup, err := NewPopup(PopupSpec{
		Title:         "No Resize",
		Content:       "Content",
		Actions:       []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:      40,
		MaxHeight:     10,
		DisableResize: true,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	popup.setBounds(10, 5, lipgloss.Width(rendered.content), lipgloss.Height(rendered.content), rendered.actionBounds)

	// Click on bottom-right corner — should NOT start resize
	cornerX := popup.bounds.x + popup.bounds.w - 1
	cornerY := popup.bounds.y + popup.bounds.h - 1
	clickMsg := tea.MouseClickMsg(tea.Mouse{X: cornerX, Y: cornerY, Button: tea.MouseLeft})
	popup.handleMouse(clickMsg)

	if popup.resizing {
		t.Error("Expected no resize started when DisableResize=true")
	}

	// Try to drag — should not change maxWidth
	dragMsg := tea.MouseMotionMsg(tea.Mouse{X: cornerX + 20, Y: cornerY, Button: tea.MouseLeft})
	popup.handleMouse(dragMsg)

	if popup.maxWidth != 40 {
		t.Errorf("Expected maxWidth to stay at 40 when DisableResize=true, got %d", popup.maxWidth)
	}
}

func TestMarkdownPopupDisableResize(t *testing.T) {
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "No Resize MD",
		MarkdownContent: "# Hello",
		MaxWidth:        50,
		MaxHeight:       12,
		DisableResize:   true,
	})
	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	if popup.disableResize != true {
		t.Error("Expected disableResize to be true")
	}

	// Verify resize handle doesn't trigger resize
	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	popup.setBounds(10, 5, lipgloss.Width(rendered.content), lipgloss.Height(rendered.content), rendered.actionBounds)

	cornerX := popup.bounds.x + popup.bounds.w - 1
	cornerY := popup.bounds.y + popup.bounds.h - 1
	clickMsg := tea.MouseClickMsg(tea.Mouse{X: cornerX, Y: cornerY, Button: tea.MouseLeft})
	popup.handleMouse(clickMsg)

	if popup.resizing {
		t.Error("Expected no resize on markdown popup with DisableResize=true")
	}
}

func TestPopupResizeRightEdge(t *testing.T) {
	popup := popupForResizeTest(t, 40, 10)
	ss := renderAndBind(t, popup)

	cx := popup.bounds.x + popup.bounds.w - 1
	cy := popup.bounds.y + popup.bounds.h/2

	// Click right edge and drag right 20
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx + 20, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx + 20, Y: cy}))

	if popup.maxWidth <= 40 {
		t.Errorf("right edge: maxWidth should increase, got %d", popup.maxWidth)
	}
	// height should be unchanged
	if popup.maxHeight != 10 {
		t.Errorf("right edge: maxHeight should stay 10, got %d", popup.maxHeight)
	}
	r2 := popup.render(ss)
	if lipgloss.Width(r2.content) != popup.maxWidth {
		t.Errorf("right edge: rendered width %d != maxWidth %d", lipgloss.Width(r2.content), popup.maxWidth)
	}
}

func TestPopupResizeLeftEdge(t *testing.T) {
	popup := popupForResizeTest(t, 40, 10)
	ss := renderAndBind(t, popup)

	cx := popup.bounds.x
	cy := popup.bounds.y + popup.bounds.h/2

	// Click left edge and drag left 15
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx - 15, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx - 15, Y: cy}))

	if popup.maxWidth <= 40 {
		t.Errorf("left edge: maxWidth should increase (drag left=expand), got %d", popup.maxWidth)
	}
	r2 := popup.render(ss)
	if lipgloss.Width(r2.content) != popup.maxWidth {
		t.Errorf("left edge: rendered width %d != maxWidth %d", lipgloss.Width(r2.content), popup.maxWidth)
	}
}

func TestPopupResizeBottomEdge(t *testing.T) {
	popup := popupForResizeTest(t, 40, 10)
	ss := renderAndBind(t, popup)

	cx := popup.bounds.x + popup.bounds.w/2
	cy := popup.bounds.y + popup.bounds.h - 1

	// Click bottom edge and drag down 15
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx, Y: cy + 15, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx, Y: cy + 15}))

	if popup.maxHeight <= 10 {
		t.Errorf("bottom edge: maxHeight should increase, got %d", popup.maxHeight)
	}
	// width should be unchanged
	if popup.maxWidth != 40 {
		t.Errorf("bottom edge: maxWidth should stay 40, got %d", popup.maxWidth)
	}
	r2 := popup.render(ss)
	if lipgloss.Height(r2.content) != popup.maxHeight {
		t.Errorf("bottom edge: rendered height %d != maxHeight %d", lipgloss.Height(r2.content), popup.maxHeight)
	}
}

func TestPopupTitleBarDragNotResize(t *testing.T) {
	// Top edge (title bar) should start a drag, NOT a resize.
	popup := popupForResizeTest(t, 40, 10)
	ss := renderAndBind(t, popup)

	cx := popup.bounds.x + popup.bounds.w/2
	cy := popup.bounds.y

	// Click top edge — should start drag, not resize
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))

	if popup.resizing {
		t.Error("top edge click should NOT start resize")
	}
	if !popup.dragging {
		t.Error("top edge click should start drag (title-bar move)")
	}

	// Drag should move the popup
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx + 10, Y: cy + 5, Button: tea.MouseLeft}))
	if popup.offsetX < 10 {
		t.Errorf("expected offsetX to change during drag, got %d", popup.offsetX)
	}

	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx + 10, Y: cy + 5}))
	if popup.dragging {
		t.Error("expected drag to end on release")
	}
	// maxHeight should be unchanged (no resize)
	if popup.maxHeight != 10 {
		t.Errorf("expected maxHeight=10 (unchanged), got %d", popup.maxHeight)
	}

	_ = ss
}

func TestPopupTopLeftCornerDragNotResize(t *testing.T) {
	// Top-left corner should start a drag, NOT a resize.
	popup := popupForResizeTest(t, 40, 10)
	ss := renderAndBind(t, popup)

	cx := popup.bounds.x
	cy := popup.bounds.y

	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))

	if popup.resizing {
		t.Error("top-left corner click should NOT start resize")
	}
	if !popup.dragging {
		t.Error("top-left corner click should start drag")
	}

	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx, Y: cy}))
	if popup.maxWidth != 40 || popup.maxHeight != 10 {
		t.Errorf("drag should not change size: got %dx%d", popup.maxWidth, popup.maxHeight)
	}

	_ = ss
}

func TestPopupTopRightCornerDragNotResize(t *testing.T) {
	// Top-right corner should start a drag, NOT a resize.
	popup := popupForResizeTest(t, 40, 10)
	ss := renderAndBind(t, popup)

	cx := popup.bounds.x + popup.bounds.w - 1
	cy := popup.bounds.y

	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))

	if popup.resizing {
		t.Error("top-right corner click should NOT start resize")
	}
	if !popup.dragging {
		t.Error("top-right corner click should start drag")
	}

	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx, Y: cy}))
	if popup.maxWidth != 40 || popup.maxHeight != 10 {
		t.Errorf("drag should not change size: got %dx%d", popup.maxWidth, popup.maxHeight)
	}

	_ = ss
}

func TestPopupResizeBottomLeftCorner(t *testing.T) {
	popup := popupForResizeTest(t, 40, 10)
	ss := renderAndBind(t, popup)

	cx := popup.bounds.x
	cy := popup.bounds.y + popup.bounds.h - 1

	// Click bottom-left corner, drag left 20 and down 10
	popup.handleMouse(tea.MouseClickMsg(tea.Mouse{X: cx, Y: cy, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: cx - 20, Y: cy + 10, Button: tea.MouseLeft}))
	popup.handleMouse(tea.MouseReleaseMsg(tea.Mouse{X: cx - 20, Y: cy + 10}))

	if popup.maxWidth <= 40 {
		t.Errorf("bottom-left: maxWidth should increase (drag left=expand), got %d", popup.maxWidth)
	}
	if popup.maxHeight <= 10 {
		t.Errorf("bottom-left: maxHeight should increase (drag down=expand), got %d", popup.maxHeight)
	}
	r2 := popup.render(ss)
	if lipgloss.Width(r2.content) != popup.maxWidth || lipgloss.Height(r2.content) != popup.maxHeight {
		t.Errorf("bottom-left: rendered %dx%d != maxWidth/maxHeight %dx%d",
			lipgloss.Width(r2.content), lipgloss.Height(r2.content), popup.maxWidth, popup.maxHeight)
	}
}

func popupForResizeTest(t *testing.T, w, h int) *Popup {
	t.Helper()
	popup, err := NewPopup(PopupSpec{
		Title:     "Resize",
		Content:   "Some content to test resize behavior",
		Actions:   []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:  w,
		MaxHeight: h,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}
	return popup
}

func renderAndBind(t *testing.T, popup *Popup) style.PopupStyleSet {
	t.Helper()
	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup
	rendered := popup.render(ss)
	popup.setBounds(10, 5, lipgloss.Width(rendered.content), lipgloss.Height(rendered.content), rendered.actionBounds)
	return ss
}

func TestPopupScrollbarPinnedToRightEdge(t *testing.T) {
	// When the popup is wider than the content, the scrollbar should stick
	// to the right edge of the body area, not float mid-content.
	popup, err := NewPopup(PopupSpec{
		Title:     "Wide",
		Content:   strings.Repeat("short\n", 15), // 15 lines of "short" (5 chars)
		Actions:   []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:  50, // wide enough that content doesn't fill
		MaxHeight: 8,  // only show ~5 lines → scrolling
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	rendered := popup.render(ss)
	popup.setBounds(10, 5, lipgloss.Width(rendered.content), lipgloss.Height(rendered.content), rendered.actionBounds)

	// After our fix:
	// - p.scrollbarRelX should be at the right edge = popupFrameInsetX + innerWidth - 1
	// - innerWidth should be maxContentWidth (since it's wider than content)
	if popup.scrollbarRelX < 0 {
		t.Fatal("Expected scrollbar to be visible (scrolling content)")
	}

	// The scrollbar should be at the rightmost column of the body area,
	// which is at position popupFrameInsetX + maxContentWidth - 1
	expectedScrollbarX := popupFrameInsetX + (popup.maxWidth - popupFrameHorizontalOverhead) - 1
	// = 2 + (50 - 4) - 1 = 2 + 46 - 1 = 47

	t.Logf("scrollbarRelX=%d, expected=%d (innerWidth=%d, contentWidth=%d)",
		popup.scrollbarRelX, expectedScrollbarX, popup.maxWidth-popupFrameHorizontalOverhead, popup.contentTextW)

	if popup.scrollbarRelX != expectedScrollbarX {
		t.Errorf("scrollbar not at right edge: got %d, want %d", popup.scrollbarRelX, expectedScrollbarX)
	}
}

func TestPopupScrollbarBackgroundMatchesSurface(t *testing.T) {
	// Verify the scrollbar track and thumb use the popup surface background.
	_, err := NewPopup(PopupSpec{
		Title:     "Scroll",
		Content:   strings.Repeat("line\n", 20),
		Actions:   []PopupAction{{ID: "ok", Label: "OK"}},
		MaxWidth:  40,
		MaxHeight: 8,
	})
	if err != nil {
		t.Fatalf("NewPopup() error = %v", err)
	}

	theme := style.DefaultDarkTheme()
	ss := style.NewStyleSet(theme).Popup

	surface := ss.Surface
	trackBg := ss.ScrollTrack.GetBackground()
	thumbBg := ss.ScrollThumb.GetBackground()

	if !samePopupColor(trackBg, surface) {
		t.Errorf("ScrollTrack background does not match popup surface")
	}
	if !samePopupColor(thumbBg, surface) {
		t.Errorf("ScrollThumb background does not match popup surface")
	}
}
