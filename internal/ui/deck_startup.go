package ui

import (
	"math"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/motion"
	"github.com/anhoder/foxful-cli/style"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/go-musicfox/go-musicfox/internal/ui/deck"
)

// StartupDeck names the deck startup animation in the configuration.
const StartupDeck = "deck"

// The deck startup's timeline, for a two second start; shorter starts run
// it faster.
const (
	startupFly       = 1100 * time.Millisecond // the cassette comes in from the depth
	startupWordmark  = 500 * time.Millisecond  // the first letter starts
	startupLetterGap = 70 * time.Millisecond
	startupLetter    = 300 * time.Millisecond
	startupLine      = 900 * time.Millisecond // the loading line fades in
	startupLineFade  = 300 * time.Millisecond
	startupNominal   = 2 * time.Second
)

const startupWord = "TAPEFOX"

// DeckStartup draws the startup page: the cassette turns in from the depth
// of the screen and settles, the wordmark comes in letter by letter under
// it, then a hairline fills as loading goes on.
type DeckStartup struct {
	renderer deck.Renderer
	scr      npScreen
}

func NewDeckStartup() *DeckStartup { return &DeckStartup{} }

// View is the startup page's model.StartupOptions.View.
func (d *DeckStartup) View(a *model.App, w, h int, elapsed, total time.Duration) string {
	if a.MotionLevel() == motion.Off {
		elapsed = max(total, startupNominal) // the last frame, still
	}
	scale := 1.0
	if total > 0 && total < startupNominal {
		scale = float64(total) / float64(startupNominal)
	}
	at := func(d time.Duration) time.Duration { return time.Duration(float64(d) * scale) }
	ease := func(start, dur time.Duration) float64 {
		return motion.OutCubic(min(max(float64(elapsed-at(start))/float64(at(dur)), 0), 1))
	}
	if elapsed < at(startupFly) {
		a.Animate(50 * time.Millisecond)
	} else if elapsed < total {
		a.AnimateSlow(100 * time.Millisecond)
	}

	pal := npPaletteFromTheme(style.CurrentStyleSet().Theme())
	scr := &d.scr
	scr.reset(w, h)
	scr.pal = pal

	aspect := deckAspect(a)
	if aspect <= 0 {
		aspect = 2
	}
	ch := min(max(h*42/100, 6), 16)
	cw := min(int(float64(ch)*aspect*deck.FitAspect+0.5), w-4)
	top := max((h-(ch+6))*2/5, 0)
	box := deck.Rect{X: float64((w - cw) / 2), Y: float64(top), W: float64(cw), H: float64(ch)}

	// The cassette turns in from the depth, side B first, and settles.
	e := ease(0, startupFly)
	rest := deck.RestPose()
	pose := deck.Pose{
		Yaw:   rest.Yaw - (1-e)*200*math.Pi/180,
		Pitch: rest.Pitch + (1-e)*25*math.Pi/180,
		Roll:  -(1 - e) * 12 * math.Pi / 180,
		Y:     -(1 - e) * 6,
		Scale: 0.3 + 0.7*e,
	}
	spin := elapsed.Seconds() * 2.4
	f := d.renderer.Render(deck.View{W: w, H: h, Tier: deckTier(a), Aspect: aspect, Fit: box}, deckPaletteOf(pal), pose,
		deck.State{Radii: deck.PackRadii(0.25, false), Angles: [2]float64{spin, spin + 1.3}, Lift: deck.RestLift + (1-e)*6})
	scr.drawDeck(f, 0, 0)

	// The wordmark, a letter at a time.
	word := npSpaced(startupWord)
	wx, wy := (w-npWidth(word))/2, top+ch+2
	for i, r := range startupWord {
		k := ease(startupWordmark+time.Duration(i)*startupLetterGap, startupLetter)
		scr.text(wx+2*i, wy, string(r), uv.Style{Fg: npLerp(pal.Bg, pal.Text, k), Attrs: uv.AttrBold}, 1)
	}

	// The loading line fills under it.
	if k := ease(startupLine, startupLineFade); k > 0 && wy+2 < h {
		lw := npWidth(word)
		done := lw
		if total > 0 {
			done = int(math.Round(float64(lw) * min(float64(elapsed)/float64(total), 1)))
		}
		scr.repeat(wx, wy+2, done, "━", uv.Style{Fg: npLerp(pal.Bg, pal.Accent, k)})
		scr.repeat(wx+done, wy+2, lw-done, "─", uv.Style{Fg: npLerp(pal.Bg, pal.Rule, k)})
	}
	scr.fillBackground()
	return scr.render()
}
