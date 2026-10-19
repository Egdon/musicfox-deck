package model

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/colorprofile"
)

func TestStartupRerendersDoNotAdvanceLoadingDuration(t *testing.T) {
	options := DefaultOptions()
	next := &Main{}
	startup := NewStartup(&options.StartupOptions, next)
	app := NewApp(options)
	app.page = startup

	started := time.Now()
	rerenders := int(options.LoadingDuration/options.TickDuration) + 1
	for i := range rerenders {
		msg := app.RerenderCmd(false)()
		page, _ := startup.Update(msg, app)
		if page == next {
			t.Fatalf("startup completed after %d immediate rerenders in %s, before loading duration %s", i+1, time.Since(started), options.LoadingDuration)
		}
	}
}

func TestStartupCompletesAfterLoadingDuration(t *testing.T) {
	options := DefaultOptions()
	next := &Main{}
	startup := NewStartup(&options.StartupOptions, next)
	app := NewApp(options)
	app.page = startup
	startup.startedAt = time.Now().Add(-options.LoadingDuration)

	page, _ := startup.Update(tickStartupMsg{}, app)
	if page != next {
		t.Fatalf("startup page = %T, want next page after loading duration", page)
	}
	if startup.loadedDuration != options.LoadingDuration {
		t.Errorf("loaded duration = %s, want %s", startup.loadedDuration, options.LoadingDuration)
	}
}

func TestStartupViewUsesAppBackground(t *testing.T) {
	previousStyles := style.CurrentStyleSet()
	previousProfile := util.TermProfile
	t.Cleanup(func() {
		style.SetStyleSet(previousStyles)
		util.TermProfile = previousProfile
	})

	appBackground := lipgloss.Color("#FAF0C8")
	theme := style.DefaultLightTheme()
	theme.AppBackground = style.Highlight{Bg: appBackground}
	style.SetStyleSet(style.NewStyleSet(theme))
	util.TermProfile = colorprofile.TrueColor

	for _, animation := range []StartupAnimation{StartupAnimationSequence, StartupAnimationMatrixRain} {
		t.Run(string(animation), func(t *testing.T) {
			options := DefaultOptions()
			app := NewApp(options)
			app.windowWidth = 80
			app.windowHeight = 24
			startup := NewStartup(&StartupOptions{
				Welcome:         "musicfox",
				Animation:       animation,
				LoadingDuration: time.Second,
			}, &Main{})
			startup.startedAt = time.Now()

			screen := renderMainToScreen(startup.View(app), app.windowWidth, app.windowHeight)
			wantR, wantG, wantB, wantA := appBackground.RGBA()
			for y := range screen.Lines {
				for x := range screen.Lines[y] {
					cell := screen.CellAt(x, y)
					if cell == nil || cell.Width == 0 {
						continue
					}
					if cell.Style.Bg == nil {
						t.Fatalf("cell (%d,%d) content=%q has no app background", x, y, cell.Content)
					}
					gotR, gotG, gotB, gotA := cell.Style.Bg.RGBA()
					if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
						t.Fatalf("cell (%d,%d) background = %v, want app background %v", x, y, cell.Style.Bg, appBackground)
					}
				}
			}
		})
	}
}

func TestStartupViewLeavesTransparentAppBackgroundUnpainted(t *testing.T) {
	previousStyles := style.CurrentStyleSet()
	t.Cleanup(func() { style.SetStyleSet(previousStyles) })

	transparentTheme := style.DefaultLightTheme()
	transparentTheme.AppBackground = style.Highlight{Bg: lipgloss.NoColor{}}
	for _, tt := range []struct {
		name  string
		theme style.Theme
	}{
		{name: "default", theme: style.DefaultLightTheme()},
		{name: "transparent", theme: transparentTheme},
	} {
		t.Run(tt.name, func(t *testing.T) {
			style.SetStyleSet(style.NewStyleSet(tt.theme))
			options := DefaultOptions()
			app := NewApp(options)
			app.windowWidth = 80
			app.windowHeight = 24
			startup := NewStartup(&StartupOptions{Welcome: "musicfox", Animation: StartupAnimationSequence}, &Main{})

			screen := renderMainToScreen(startup.View(app), app.windowWidth, app.windowHeight)
			cell := screen.CellAt(0, 0)
			if cell == nil {
				t.Fatal("top-left cell is missing")
			}
			if cell.Style.Bg != nil {
				t.Fatalf("top-left background = %#v, want unset", cell.Style.Bg)
			}
		})
	}
}
