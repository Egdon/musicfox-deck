package model

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

func TestAppThemePairControlsSystemAppearance(t *testing.T) {
	previousStyles := style.CurrentStyleSet()
	previousDark := style.HasDarkBackground()
	t.Cleanup(func() {
		style.SetStyleSet(previousStyles)
		style.SetDarkBackground(previousDark)
	})

	dark := style.DefaultDarkTheme()
	dark.Primary = lipgloss.Color("#112233")
	light := style.DefaultLightTheme()
	light.Primary = lipgloss.Color("#AABBCC")

	app := NewApp(DefaultOptions())
	app.SetThemePair(dark, light)

	app.onBackgroundChanged(false)
	assertThemePairColor(t, style.CurrentStyleSet().SelectedItem.GetForeground(), light.Primary, "light primary")
	app.onBackgroundChanged(true)
	assertThemePairColor(t, style.CurrentStyleSet().SelectedItem.GetForeground(), dark.Primary, "dark primary")
}

func assertThemePairColor(t *testing.T, got, want color.Color, label string) {
	t.Helper()
	gotR, gotG, gotB, gotA := got.RGBA()
	wantR, wantG, wantB, wantA := want.RGBA()
	if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
		t.Fatalf("%s = %#v, want %#v", label, got, want)
	}
}
