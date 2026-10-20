package player

import (
	"testing"
	"time"

	"github.com/gopxl/beep/effects"
)

// The interface reads these on every frame; they must not wait while a song
// loads, which holds the player's lock across the network.
func TestBeepReadsDoNotWaitForALoadingSong(t *testing.T) {
	p := &beepPlayer{volume: &effects.Volume{Base: 2, Volume: -1.25}}
	p.storeLevel()
	music := URLMusic{URL: "https://example.com/a.mp3"}
	p.current.Store(&music)
	p.passed.Store(int64(42 * time.Second))

	p.l.Lock() // a song is loading
	defer p.l.Unlock()
	done := make(chan struct{})
	var passed time.Duration
	var volume int
	var cur URLMusic
	go func() {
		passed, volume, cur = p.PassedTime(), p.Volume(), p.CurMusic()
		_ = p.PlayedTime()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("reading the player waited for the loading song")
	}
	if passed != 42*time.Second || volume != 75 || cur.URL != music.URL {
		t.Fatalf("passed %v, volume %d, music %q", passed, volume, cur.URL)
	}
}
