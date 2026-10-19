package model

import (
	"testing"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderAppBackgroundExcludesCoverRect(t *testing.T) {
	view := renderAppBackground(
		"abcd\nefgh",
		4,
		lipgloss.NewStyle().Background(lipgloss.Color("#1E1E2E")),
		backgroundRect{x: 1, y: 0, w: 2, h: 1},
	)

	screen := uv.NewScreenBuffer(4, 2)
	screen.Method = ansi.GraphemeWidth
	uv.NewStyledString(view).Draw(screen, screen.Bounds())

	for _, x := range []int{0, 3} {
		if cell := screen.CellAt(x, 0); cell == nil || cell.Style.Bg == nil {
			t.Fatalf("outside cover cell (%d, 0) has no app background", x)
		}
	}
	for _, x := range []int{1, 2} {
		if cell := screen.CellAt(x, 0); cell == nil || cell.Style.Bg != nil {
			t.Fatalf("cover cell (%d, 0) has background %v, want transparent", x, cell.Style.Bg)
		}
	}
	for x := range 4 {
		if cell := screen.CellAt(x, 1); cell == nil || cell.Style.Bg == nil {
			t.Fatalf("outside cover row cell (%d, 1) has no app background", x)
		}
	}
}
