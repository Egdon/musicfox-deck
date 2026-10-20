package ui

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

func TestWarmNightPaletteMatchesTheme(t *testing.T) {
	configs.LoadThemeRegistry(t.TempDir())
	reg := configs.CurrentThemeRegistry()
	reg.SelectTheme("Warm Night", true)
	ss := reg.CurrentStyleSet(true)
	if ss == nil || reg.CurrentName(true) != "Warm Night" {
		t.Fatal("Warm Night theme missing")
	}
	if got := npPaletteFromTheme(ss.Theme()); got != npWarmNight {
		t.Fatalf("palette from theme = %+v\nwant %+v", got, npWarmNight)
	}
	if npGhostOf(npWarmNight.Bg, npWarmNight.Rule) != npWarmNight.Ghost {
		t.Fatal("fallback ghost is not derived from Bg and Rule")
	}
}
