package ui

import (
	"slices"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/ui/deck"
)

// The pacer turns the reels on every frame but takes a floating cassette's
// pose only on every other one, and at once while it moves briskly.
func TestDeckPacerHoldsTheFloatingPose(t *testing.T) {
	var c deckPacer
	var r deck.Renderer
	v := deck.View{W: 40, H: 14, Tier: deck.Octant}
	pal := deckPaletteOf(npWarmNight)
	pose := deck.RestPose()
	st := deck.State{Radii: deck.ReelRadii(0.3), Lift: deck.RestLift}
	full := func(pose deck.Pose, st deck.State) []deck.Cell {
		var r deck.Renderer
		return r.Render(v, pal, pose, st).Cells
	}
	t0 := time.Unix(1_800_000_000, 0)
	c.frame(t0, &r, v, pal, pose, st, false)

	moved, turned := pose, st
	moved.Yaw += 0.02
	turned.Angles = [2]float64{1, 2}
	turned.Radii = deck.ReelRadii(0.31)
	if got := c.frame(t0.Add(34*time.Millisecond), &r, v, pal, moved, turned, false); !slices.Equal(got.Cells, full(pose, deck.State{Radii: st.Radii, Angles: turned.Angles, Lift: st.Lift})) {
		t.Fatal("one frame on, the reels should turn on the cassette as it was")
	}
	if got := c.frame(t0.Add(68*time.Millisecond), &r, v, pal, moved, turned, false); !slices.Equal(got.Cells, full(moved, turned)) {
		t.Fatal("two frames on, the cassette should take its new pose")
	}

	moved.Yaw += 0.02
	if got := c.frame(t0.Add(102*time.Millisecond), &r, v, pal, moved, turned, true); !slices.Equal(got.Cells, full(moved, turned)) {
		t.Fatal("a brisk cassette should take its new pose on every frame")
	}
}
