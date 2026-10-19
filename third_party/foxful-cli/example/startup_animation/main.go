// startup_animation demonstrates every built-in startup animation.
//
// Run from the repository root:
//
//	go run ./example/startup_animation -animation sequence
//	go run ./example/startup_animation -animation matrix-rain -duration 4s
//	go run ./example/startup_animation -animation particle-burst -reduced-motion
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
)

var animations = map[string]model.StartupAnimation{
	"fade-in":        model.StartupAnimationFadeIn,
	"rainbow-wave":   model.StartupAnimationRainbowWave,
	"spinner":        model.StartupAnimationSpinner,
	"slide-in":       model.StartupAnimationSlideIn,
	"glitch":         model.StartupAnimationGlitch,
	"sequence":       model.StartupAnimationSequence,
	"matrix-rain":    model.StartupAnimationMatrixRain,
	"particle-burst": model.StartupAnimationParticleBurst,
}

type demoMenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func newDemoMenu() *demoMenu {
	return &demoMenu{items: []model.MenuItem{
		{Title: "Animation complete", Subtitle: "The startup page handed control back to the application."},
		{Title: "Run another mode", Subtitle: "Quit and choose a different -animation value."},
		{Title: "Quit", Subtitle: "Press q or Ctrl+C."},
	}}
}

func (m *demoMenu) GetMenuKey() string { return "startup_animation_demo" }
func (m *demoMenu) MenuViews() []model.MenuItem {
	return m.items
}
func (m *demoMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

func main() {
	animationName := flag.String("animation", "sequence", "startup animation; use -list to show available values")
	duration := flag.Duration("duration", 3*time.Second, "startup animation duration")
	reducedMotion := flag.Bool("reduced-motion", false, "render the static accessible fallback")
	list := flag.Bool("list", false, "list animation values and exit")
	flag.Parse()

	if *list {
		printAnimations()
		return
	}

	animation, ok := animations[strings.ToLower(*animationName)]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown animation %q\n\n", *animationName)
		printAnimations()
		os.Exit(2)
	}
	if *duration < 0 {
		fmt.Fprintln(os.Stderr, "-duration must not be negative")
		os.Exit(2)
	}

	opts := model.DefaultOptions()
	opts.AppName = "Startup Animation Gallery"
	opts.Welcome = "FOXFUL"
	opts.Animation = animation
	opts.LoadingDuration = *duration
	opts.ReducedMotion = *reducedMotion
	opts.DualColumn = false
	opts.StatusBar = &model.DefaultStatusBar{}

	app := model.NewApp(opts)
	app.With(model.WithMainMenu(newDemoMenu(), &model.MenuItem{
		Title: "Startup Animation Gallery",
	}))

	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "startup animation example:", err)
		os.Exit(1)
	}
}

func printAnimations() {
	names := make([]string, 0, len(animations))
	for name := range animations {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println("Available startup animations:")
	for _, name := range names {
		fmt.Println("  " + name)
	}
}
