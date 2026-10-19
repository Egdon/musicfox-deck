package model

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

func TestRenderAppBackgroundFillsMissingCellsWithoutOverwritingExplicitBackground(t *testing.T) {
	previousStyles := style.CurrentStyleSet()
	t.Cleanup(func() { style.SetStyleSet(previousStyles) })

	appBackground := lipgloss.Color("#FAF0C8")
	theme := style.DefaultLightTheme()
	theme.AppBackground = style.Highlight{Bg: appBackground}
	style.SetStyleSet(style.NewStyleSet(theme))

	explicitBackground := lipgloss.Color("#000000")
	content := lipgloss.NewStyle().Background(explicitBackground).Render("X") + " "
	screen := popupStyledScreen(RenderAppBackground(content, 2))

	assertPageBackgroundColor(t, screen.CellAt(0, 0).Style.Bg, explicitBackground)
	assertPageBackgroundColor(t, screen.CellAt(1, 0).Style.Bg, appBackground)
}

func assertPageBackgroundColor(t *testing.T, got, want color.Color) {
	t.Helper()
	if got == nil {
		t.Fatalf("background = nil, want %v", want)
	}
	gotR, gotG, gotB, gotA := got.RGBA()
	wantR, wantG, wantB, wantA := want.RGBA()
	if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
		t.Fatalf("background = %v, want %v", got, want)
	}
}
