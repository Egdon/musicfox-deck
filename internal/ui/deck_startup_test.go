package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/model"
	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

func TestDeckStartupFillsTheScreen(t *testing.T) {
	prev := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = prev })
	app := model.NewApp(model.DefaultOptions())
	d := NewDeckStartup()
	for _, size := range [][2]int{{110, 32}, {80, 24}, {40, 12}, {20, 6}} {
		w, h := size[0], size[1]
		for _, at := range []time.Duration{0, 300 * time.Millisecond, 900 * time.Millisecond, 2 * time.Second, 5 * time.Second} {
			lines := strings.Split(d.View(app, w, h, at, 2*time.Second), "\n")
			if len(lines) != h {
				t.Fatalf("%dx%d at %v: %d lines", w, h, at, len(lines))
			}
			for i, l := range lines {
				if got := ansix.Width(l); got != w {
					t.Fatalf("%dx%d at %v: line %d is %d cells", w, h, at, i, got)
				}
			}
		}
	}
	if end := ansi.Strip(d.View(app, 80, 24, 3*time.Second, 2*time.Second)); !strings.Contains(end, npSpaced(startupWord)) {
		t.Fatal("the wordmark shows once the cassette has settled")
	}
	if start := ansi.Strip(d.View(app, 80, 24, 0, 2*time.Second)); strings.Contains(start, "━") {
		t.Fatal("the loading line waits for the cassette")
	}
}
